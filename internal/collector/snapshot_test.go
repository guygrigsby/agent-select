package collector

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/guygrigsby/agent-routing/routing"
)

func ms(t time.Time) time.Time { return t.Truncate(time.Millisecond) }

func seedSnapshotBase(t *testing.T, p *Ports, sourceID string, redistribution string) (obsID string) {
	t.Helper()
	now := time.Now().UTC()
	mustExec(t, p.Pool,
		`INSERT INTO sources (id,url,source_kind,parser_version,freshness_ceiling_seconds,enabled)
		 VALUES ($1,$2,'vercel_catalog','p1',3600,true)`, sourceID, "https://"+sourceID)
	mustExec(t, p.Pool, `
		INSERT INTO source_policies (source_id,config_hash,research,extraction,redistribution,
		  terms_url,license_text,checked_at,effective_at,expires_at)
		VALUES ($1,$2,'allowed','allowed',$3,'https://t','license',$4,$5,$6)`,
		sourceID, hash32([]byte("policy-"+sourceID)), redistribution,
		now.Add(-time.Hour), now.Add(-2*time.Hour), now.Add(48*time.Hour))
	body := []byte(`{"models":[{"id":"u","object":"model"}],"src":"` + sourceID + `"}`)
	mustExec(t, p.Pool, `INSERT INTO content_blobs (hash,bytes) VALUES ($1,$2)`,
		hash32(body), body)
	obsID = "obs-" + sourceID
	mustExec(t, p.Pool, `
		INSERT INTO observations (id,source_id,resource_url,retrieved_at,http_status,outcome,
		  body_hash,headers_bytes,collector_version)
		VALUES ($1,$2,$3,$4,200,'success',$5,'{}','cv1')`,
		obsID, sourceID, "https://"+sourceID+"/models", now, hash32(body))
	return obsID
}

// seedSnapshotOffer seeds one fully-projectable offer under obs with a
// capability, limit, price, bound revision and one benchmark signal.
func seedSnapshotOffer(t *testing.T, p *Ports, obsID, sourceID string) {
	t.Helper()
	now := time.Now().UTC()
	mustExec(t, p.Pool, `
		INSERT INTO offers (id,provider,upstream_id,catalog_observation_id,advertised_kind)
		VALUES ('of1','prov','u',$1,'language')`, obsID)
	mustExec(t, p.Pool, `
		INSERT INTO offer_listing_facts (offer_id,observation_id,action,observed_at)
		VALUES ('of1',$1,'listed',$2)`, obsID, now)
	// Evidence assertion: capability text_input supported.
	mustExec(t, p.Pool, `
		INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,
		  verbatim,parser_version,evidence_class,payload_fingerprint)
		VALUES ('as-cap',$1,$2,'u','loc-1','{}'::bytea,'p1','provider_reported',$3)`,
		obsID, sourceID, hash32([]byte("cap")))
	mustExec(t, p.Pool, `
		INSERT INTO capability_assertions (assertion_id,capability,support,operation,direction)
		VALUES ('as-cap','text_input','supported','text_assistance','input')`)
	// Limit: context_tokens maximum 128000.
	mustExec(t, p.Pool, `
		INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,
		  verbatim,parser_version,evidence_class,payload_fingerprint)
		VALUES ('as-lim',$1,$2,'u','loc-2','{}'::bytea,'p1','provider_reported',$3)`,
		obsID, sourceID, hash32([]byte("lim")))
	mustExec(t, p.Pool, `
		INSERT INTO limit_assertions (assertion_id,limit_name,value,unit,bound,scope,operation,role)
		VALUES ('as-lim','context_tokens',128000,'token','maximum','combined','text_assistance','request')`)
	// Price: USD per million tokens, exact 3.50 for input_tokens.
	mustExec(t, p.Pool, `
		INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,
		  verbatim,parser_version,evidence_class,payload_fingerprint)
		VALUES ('as-pr',$1,$2,'u','loc-3','{}'::bytea,'p1','provider_reported',$3)`,
		obsID, sourceID, hash32([]byte("pr")))
	mustExec(t, p.Pool, `
		INSERT INTO price_assertions (assertion_id,offer_id,currency,amount,unit,category,
		  quote_knowledge,scope_hash)
		VALUES ('as-pr','of1','USD',3.50,'million_tokens','input_tokens','exact',$1)`,
		hash32([]byte("price-scope")))
	// Revision and current binding.
	mustExec(t, p.Pool, `
		INSERT INTO model_revisions (id,publisher,revision_name,identity_observation_id,identity_locator)
		VALUES ('rev1','prov','u-2026-04-01',$1,'rev-loc')`, obsID)
	mustExec(t, p.Pool, `
		INSERT INTO identity_authority_policies (version,content_hash,recorded_at)
		VALUES ('iap1',$1,$2)`, hash32([]byte("iap")), now)
	mustExec(t, p.Pool, `
		INSERT INTO identity_freshness_policies (version,content_hash,recorded_at)
		VALUES ('ifp1',$1,$2)`, hash32([]byte("ifp")), now)
	mustExec(t, p.Pool, `
		INSERT INTO identity_bindings (id,offer_id,revision_id,evidence_assertion_id,
		  authority_policy_version,freshness_policy_version,effective_at,recorded_at)
		VALUES ('bnd1','of1','rev1','as-cap','iap1','ifp1',$1,$1)`, now.Add(-2*time.Hour))
	// Benchmark measurement bound to rev1.
	mustExec(t, p.Pool, `
		INSERT INTO benchmark_runs (id,observation_id,locator,identity_kind,identity_value,content_hash)
		VALUES ('run1',$1,'bench-loc','publisher_run','run-2026-04',$2)`,
		obsID, hash32([]byte("run")))
	mustExec(t, p.Pool, `
		INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,
		  verbatim,parser_version,evidence_class,payload_fingerprint)
		VALUES ('as-bm',$1,$2,'u','loc-4','{}'::bytea,'p1','third_party_reported',$3)`,
		obsID, sourceID, hash32([]byte("bm")))
	mustExec(t, p.Pool, `
		INSERT INTO benchmark_measurements (assertion_id,run_id,benchmark_label,metric_label,
		  raw_value,unit,direction)
		VALUES ('as-bm','run1','swe_verified','pass_rate',86.0,'%','higher')`)
	mustExec(t, p.Pool, `
		INSERT INTO measurement_revision_facts (assertion_id,revision_id,identity_evidence_id)
		VALUES ('as-bm','rev1','as-cap')`)
}

func TestCompileSnapshotLifecycle(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := ms(time.Now().UTC())
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	mustExec(t, p.Pool, `
		INSERT INTO coverage_findings (observation_id,subject_upstream_id,channel,outcome,detail)
		VALUES ($1,'u','vercel_catalog','researched','resolved')`, obs)

	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), now); err != nil {
		t.Fatal(err)
	}
	policyHex := hexHash(hash32([]byte("policy-s1")))
	res, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs:                  []InputRef{{ObservationID: obs, SourceID: "s1", PolicyVersion: policyHex}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.ArtifactHash) != 32 || len(res.CoverageDigest) != 32 {
		t.Fatalf("compile result %+v", res)
	}
	if res.ArtifactLocator != artifactLocator(res.ArtifactHash) {
		t.Fatalf("locator %s", res.ArtifactLocator)
	}

	// The staged artifact decodes through the routing contract.
	var artifactBytes []byte
	if err := p.Pool.QueryRow(ctx,
		`SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
		res.ArtifactHash).Scan(&artifactBytes); err != nil {
		t.Fatal(err)
	}
	snap, err := routing.DecodeSnapshot(artifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	data := snap.Data()
	if len(data.Offers) != 1 {
		t.Fatalf("offers %+v", data.Offers)
	}
	o := data.Offers[0]
	if o.ID != "prov/u" || o.Revision != "u-2026-04-01" {
		t.Fatalf("offer identity %+v", o)
	}
	if len(o.Capabilities) != 1 || o.Capabilities[0] != "text_input" {
		t.Fatalf("capabilities %+v", o.Capabilities)
	}
	if len(o.Limits) != 1 || o.Limits[0].Resource != "context_tokens" ||
		o.Limits[0].Maximum != 128000 {
		t.Fatalf("limits %+v", o.Limits)
	}
	if len(o.Prices) != 1 || o.Prices[0].Component != "input_tokens" ||
		o.Prices[0].USDPerMillion != "3.5" {
		t.Fatalf("prices %+v", o.Prices)
	}
	if len(o.Signals) != 1 || o.Signals[0].Task != "code" ||
		o.Signals[0].Value != 8600 || o.Signals[0].Weight != 100 ||
		o.Signals[0].Basis != "published" || o.Signals[0].Revision != "u-2026-04-01" {
		t.Fatalf("signals %+v", o.Signals)
	}
	if data.Distribution != "local" || data.Policy != "educated-v1" || data.Schema != 1 {
		t.Fatalf("snapshot header %+v", data)
	}

	// Identical replay returns the identical hash without a second event.
	replay, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs:                  []InputRef{{ObservationID: obs, SourceID: "s1", PolicyVersion: policyHex}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(replay.ArtifactHash) != string(res.ArtifactHash) {
		t.Fatalf("replay hash %x != %x", replay.ArtifactHash, res.ArtifactHash)
	}
	if n := countEventsName(t, p, hexHash(res.ArtifactHash), "SnapshotCompiled"); n != 1 {
		t.Fatalf("SnapshotCompiled events = %d, want 1", n)
	}
}

func TestCompileSnapshotGuardArms(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := ms(time.Now().UTC())
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), now); err != nil {
		t.Fatal(err)
	}
	base := CompileSnapshotParams{
		Inputs:                  []InputRef{{ObservationID: obs, SourceID: "s1", PolicyVersion: hexHash(hash32([]byte("policy-s1")))}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	}
	if _, err := p.CompileSnapshot(ctx, base); err != nil {
		t.Fatal(err)
	}
	bad := base
	bad.Distribution = "sideways"
	if _, err := p.CompileSnapshot(ctx, bad); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("distribution token accepted")
	}
	bad = base
	bad.ProjectionPolicyVersion = "pp-ghost"
	if _, err := p.CompileSnapshot(ctx, bad); codeOf(t, err) != CodeNotFound {
		t.Fatal("unknown projection policy accepted")
	}
	bad = base
	bad.KnowledgeCutoff = base.CreatedAt.Add(time.Minute)
	if _, err := p.CompileSnapshot(ctx, bad); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("cutoff after created accepted")
	}
	bad = base
	bad.CreatedAt = base.CreatedAt.Add(time.Microsecond)
	if _, err := p.CompileSnapshot(ctx, bad); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("sub-millisecond instant accepted")
	}
	bad = base
	bad.Inputs = []InputRef{{ObservationID: "obs-ghost", SourceID: "s1", PolicyVersion: "x"}}
	if _, err := p.CompileSnapshot(ctx, bad); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("unknown observation accepted")
	}
	bad = base
	bad.Inputs = append(bad.Inputs, bad.Inputs[0])
	if _, err := p.CompileSnapshot(ctx, bad); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("duplicate input observation accepted")
	}
}

func TestCompileSnapshotRightsExclusion(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := ms(time.Now().UTC())
	obsDenied := seedSnapshotBase(t, p, "s-denied", "denied") // redistribution denied
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), now); err != nil {
		t.Fatal(err)
	}
	res, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs: []InputRef{{
			ObservationID: obsDenied,
			SourceID:      "s-denied",
			PolicyVersion: hexHash(hash32([]byte("policy-s-denied"))),
		}},
		Distribution:            "distributable",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	var artifactBytes []byte
	if err := p.Pool.QueryRow(ctx,
		`SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
		res.ArtifactHash).Scan(&artifactBytes); err != nil {
		t.Fatal(err)
	}
	snap, err := routing.DecodeSnapshot(artifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Data().Inputs) != 0 || len(snap.Data().Offers) != 0 {
		t.Fatalf("denied source contributed: %+v", snap.Data())
	}

	// Local distribution tolerates the denied redistribution.
	resLocal, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs: []InputRef{{
			ObservationID: obsDenied,
			SourceID:      "s-denied",
			PolicyVersion: hexHash(hash32([]byte("policy-s-denied"))),
		}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(resLocal.ArtifactHash) == string(res.ArtifactHash) {
		t.Fatal("distribution did not change the artifact")
	}

	// A same-time conflicting governing set excludes the source entirely.
	var firstEffective time.Time
	if err := p.Pool.QueryRow(ctx, `
		SELECT effective_at FROM source_policies
		WHERE source_id = 's-denied' AND config_hash = $1`,
		hash32([]byte("policy-s-denied"))).Scan(&firstEffective); err != nil {
		t.Fatal(err)
	}
	mustExec(t, p.Pool, `
		INSERT INTO source_policies (source_id,config_hash,research,extraction,redistribution,
		  terms_url,license_text,checked_at,effective_at,expires_at)
		VALUES ('s-denied',$1,'allowed','allowed','allowed','https://t','license',$2,$3,$4)`,
		hash32([]byte("policy-conflict")), now.Add(-time.Hour),
		firstEffective, now.Add(48*time.Hour))
	resConflict, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs: []InputRef{{
			ObservationID: obsDenied,
			SourceID:      "s-denied",
			PolicyVersion: hexHash(hash32([]byte("policy-s-denied"))),
		}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	var bytes2 []byte
	if err := p.Pool.QueryRow(ctx,
		`SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
		resConflict.ArtifactHash).Scan(&bytes2); err != nil {
		t.Fatal(err)
	}
	if s2, err := routing.DecodeSnapshot(bytes2); err != nil || len(s2.Data().Inputs) != 0 {
		t.Fatalf("conflicting policy set contributed: %v %+v", err, s2.Data())
	}
}

func TestValidateAndPublishSnapshot(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := ms(time.Now().UTC())
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), now); err != nil {
		t.Fatal(err)
	}
	res, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs: []InputRef{{ObservationID: obs, SourceID: "s1",
			PolicyVersion: hexHash(hash32([]byte("policy-s1")))}},
		Distribution:            "distributable",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.RegisterPublicationChannel(ctx, "lane-local", "local", now); err != nil {
		t.Fatal(err)
	}
	if _, err := p.RegisterPublicationChannel(ctx, "lane-dist", "distributable", now); err != nil {
		t.Fatal(err)
	}

	// Publication requires an accepted validation.
	if _, err := p.PublishSnapshot(ctx, res.ArtifactHash, "lane-dist"); codeOf(t, err) != CodeConflict {
		t.Fatalf("unvalidated publish accepted: %v", err)
	}
	v, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash: res.ArtifactHash, ValidatorVersion: "vv1"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Outcome != "accepted" || len(v.Findings) != 0 {
		t.Fatalf("validation %+v", v)
	}
	// Replay is idempotent.
	v2, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash: res.ArtifactHash, ValidatorVersion: "vv1"})
	if err != nil || v2.Outcome != "accepted" {
		t.Fatalf("validation replay %+v %v", v2, err)
	}
	if n := countEventsName(t, p, hexHash(res.ArtifactHash), "SnapshotValidated"); n != 1 {
		t.Fatalf("SnapshotValidated events = %d", n)
	}

	pub, err := p.PublishSnapshot(ctx, res.ArtifactHash, "lane-dist")
	if err != nil {
		t.Fatal(err)
	}
	if pub.ArtifactLocator != artifactLocator(res.ArtifactHash) {
		t.Fatalf("publication locator %s", pub.ArtifactLocator)
	}
	// Replay returns the same publication without a second event.
	pub2, err := p.PublishSnapshot(ctx, res.ArtifactHash, "lane-dist")
	if err != nil || pub2.PublicationID != pub.PublicationID {
		t.Fatalf("publication replay %+v %v", pub2, err)
	}
	if n := countEventsName(t, p, hexHash(res.ArtifactHash), "SnapshotPublished"); n != 1 {
		t.Fatalf("SnapshotPublished events = %d", n)
	}

	// A revocation committed before publication fails the gate.
	obsRev := seedSnapshotBase(t, p, "s2", "allowed")
	resRev, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs: []InputRef{{ObservationID: obsRev, SourceID: "s2",
			PolicyVersion: hexHash(hash32([]byte("policy-s2")))}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash: resRev.ArtifactHash, ValidatorVersion: "vv1"}); err != nil {
		t.Fatal(err)
	}
	mustExec(t, p.Pool, `
		INSERT INTO source_policies (source_id,config_hash,research,extraction,redistribution,
		  terms_url,license_text,checked_at,effective_at,expires_at)
		VALUES ('s2',$1,'denied','denied','denied','https://t','license',$2,$3,$4)`,
		hash32([]byte("policy-s2-revoked")), now.Add(-time.Minute),
		now.Add(-time.Minute), now.Add(48*time.Hour))
	if _, err := p.PublishSnapshot(ctx, resRev.ArtifactHash, "lane-local"); codeOf(t, err) != CodeForbidden {
		t.Fatalf("revoked rights published: %v", err)
	}

	// Unknown hash and channel arms.
	if _, err := p.PublishSnapshot(ctx, hash32([]byte("ghost")), "lane-dist"); codeOf(t, err) != CodeNotFound {
		t.Fatal("unknown artifact published")
	}
	if _, err := p.PublishSnapshot(ctx, res.ArtifactHash, "lane-ghost"); codeOf(t, err) != CodeNotFound {
		t.Fatal("unknown channel published")
	}
}

func TestValidateSnapshotRejectsCorruptBytes(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := ms(time.Now().UTC())
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), now); err != nil {
		t.Fatal(err)
	}
	res, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs: []InputRef{{ObservationID: obs, SourceID: "s1",
			PolicyVersion: hexHash(hash32([]byte("policy-s1")))}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, p.Pool, `UPDATE snapshots SET artifact_bytes = $2 WHERE artifact_hash = $1`,
		res.ArtifactHash, []byte("not a snapshot"))
	v, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash: res.ArtifactHash, ValidatorVersion: "vv1"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Outcome != "rejected" {
		t.Fatalf("corrupt artifact validated: %+v", v)
	}
	// The rejection is recorded and replayed.
	var outcome string
	var findings []byte
	if err := p.Pool.QueryRow(ctx, `
		SELECT outcome, findings_bytes FROM snapshot_validations
		WHERE artifact_hash = $1 AND validator_version = 'vv1'`,
		res.ArtifactHash).Scan(&outcome, &findings); err != nil {
		t.Fatal(err)
	}
	if outcome != "rejected" {
		t.Fatalf("stored outcome %s", outcome)
	}
	var arr []ValidationFinding
	if err := json.Unmarshal(findings, &arr); err != nil || len(arr) == 0 {
		t.Fatalf("findings %s %v", findings, err)
	}
	for _, f := range arr {
		if f.Code == "" || f.Detail == "" {
			t.Fatalf("finding not closed-schema: %+v", f)
		}
	}
}

// seedExtraAssertion inserts a bare assertion row for projection fixtures.
func seedExtraAssertion(t *testing.T, p *Ports, obsID, sourceID, asID, class string) {
	t.Helper()
	mustExec(t, p.Pool, `
		INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,
		  verbatim,parser_version,evidence_class,payload_fingerprint)
		VALUES ($1,$2,$3,'u',$4,'{}'::bytea,'p1',$5,$6)`,
		asID, obsID, sourceID, "loc-"+asID, class, hash32([]byte(asID)))
}

// compileLocal compiles the observation as a local snapshot under the
// already-registered pp-1, returning the decoded artifact.
func compileLocal(t *testing.T, p *Ports, obs, src string) (CompileSnapshotResult, *routing.Snapshot) {
	t.Helper()
	ctx := context.Background()
	now := ms(time.Now().UTC())
	res, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs:                  []InputRef{{ObservationID: obs, SourceID: src, PolicyVersion: hexHash(hash32([]byte("policy-" + src)))}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	var bytes_ []byte
	if err := p.Pool.QueryRow(ctx,
		`SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
		res.ArtifactHash).Scan(&bytes_); err != nil {
		t.Fatal(err)
	}
	snap, err := routing.DecodeSnapshot(bytes_)
	if err != nil {
		t.Fatal(err)
	}
	return res, snap
}

// TestSnapshotPriceLexemeIntegrity pins the exact-decimal lexeme rules:
// integral lexemes keep their digits, zero stays expressible, fractional
// zeros trim and disagreeing exacts omit the component.
func TestSnapshotPriceLexemeIntegrity(t *testing.T) {
	p := mustPort(t)
	obs := seedSnapshotBase(t, p, "plx", "allowed")
	seedSnapshotOffer(t, p, obs, "plx")
	if _, err := p.RegisterProjectionPolicy(context.Background(), "pp-1", validProjectionPolicy(t), ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	// Replace the seeded 3.50 exact with the corruption-sensitive set.
	mustExec(t, p.Pool, `DELETE FROM price_assertions WHERE assertion_id = 'as-pr'`)
	seedExtraAssertion(t, p, obs, "plx", "as-p100", "provider_reported")
	mustExec(t, p.Pool, `
		INSERT INTO price_assertions (assertion_id,offer_id,currency,amount,unit,category,
		  quote_knowledge,scope_hash)
		VALUES ('as-p100','of1','USD',100,'million_tokens','input_tokens','exact',$1)`,
		hash32([]byte("p100")))
	seedExtraAssertion(t, p, obs, "plx", "as-p0", "provider_reported")
	mustExec(t, p.Pool, `
		INSERT INTO price_assertions (assertion_id,offer_id,currency,amount,unit,category,
		  quote_knowledge,scope_hash)
		VALUES ('as-p0','of1','USD',0,'million_tokens','output_tokens','exact',$1)`,
		hash32([]byte("p0")))
	seedExtraAssertion(t, p, obs, "plx", "as-pdis1", "provider_reported")
	mustExec(t, p.Pool, `
		INSERT INTO price_assertions (assertion_id,offer_id,currency,amount,unit,category,
		  quote_knowledge,scope_hash)
		VALUES ('as-pdis1','of1','USD',100,'million_tokens','cached_input_tokens','exact',$1)`,
		hash32([]byte("pdis1")))
	seedExtraAssertion(t, p, obs, "plx", "as-pdis2", "provider_reported")
	mustExec(t, p.Pool, `
		INSERT INTO price_assertions (assertion_id,offer_id,currency,amount,unit,category,
		  quote_knowledge,scope_hash)
		VALUES ('as-pdis2','of1','USD',1.0,'million_tokens','cached_input_tokens','exact',$1)`,
		hash32([]byte("pdis2")))
	_, snap := compileLocal(t, p, obs, "plx")
	prices := snap.Data().Offers[0].Prices
	byComponent := map[string]string{}
	for _, pr := range prices {
		byComponent[pr.Component] = pr.USDPerMillion
	}
	if byComponent["input_tokens"] != "100" {
		t.Fatalf("integral lexeme corrupted: %q", byComponent["input_tokens"])
	}
	if byComponent["output_tokens"] != "0" {
		t.Fatalf("zero price must stay expressible: %q", byComponent["output_tokens"])
	}
	if _, ok := byComponent["cached_input_tokens"]; ok {
		t.Fatalf("disagreeing exacts must omit the component: %+v", prices)
	}
}

// TestSnapshotExactAboveMaximumOmits pins the both-ends exact fold: an exact
// bound above every maximum is an empty intersection, omitted.
func TestSnapshotExactAboveMaximumOmits(t *testing.T) {
	p := mustPort(t)
	obs := seedSnapshotBase(t, p, "exa", "allowed")
	seedSnapshotOffer(t, p, obs, "exa")
	if _, err := p.RegisterProjectionPolicy(context.Background(), "pp-1", validProjectionPolicy(t), ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	seedExtraAssertion(t, p, obs, "exa", "as-exact", "provider_reported")
	mustExec(t, p.Pool, `
		INSERT INTO limit_assertions (assertion_id,limit_name,value,unit,bound,scope,operation,role)
		VALUES ('as-exact','context_tokens',200000,'token','exact','combined','text_assistance','request')`)
	_, snap := compileLocal(t, p, obs, "exa")
	for _, l := range snap.Data().Offers[0].Limits {
		if l.Resource == "context_tokens" {
			t.Fatalf("exact 200000 above maximum 128000 must omit, got %+v", l)
		}
	}
}

// TestSnapshotSignalBoundaryAndNaN pins post-round skipping: values rounding
// outside 0..10000 are skipped without failing the compile, and NaN never
// reaches the artifact.
func TestSnapshotSignalBoundaryAndNaN(t *testing.T) {
	p := mustPort(t)
	obs := seedSnapshotBase(t, p, "sgb", "allowed")
	seedSnapshotOffer(t, p, obs, "sgb")
	if _, err := p.RegisterProjectionPolicy(context.Background(), "pp-1", validProjectionPolicy(t), ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	// Replace the seeded 86.0% measurement with a boundary and a NaN row.
	mustExec(t, p.Pool, `DELETE FROM measurement_revision_facts WHERE assertion_id = 'as-bm'`)
	mustExec(t, p.Pool, `DELETE FROM benchmark_measurements WHERE assertion_id = 'as-bm'`)
	mustExec(t, p.Pool, `DELETE FROM assertions WHERE id = 'as-bm'`)
	for _, tc := range []struct {
		id, raw string
	}{
		{"as-b1", "100.005"}, // rounds to 10001: skipped
		{"as-b2", "100.000"}, // rounds to 10000: carried
		{"as-b3", `NaN`},
	} {
		seedExtraAssertion(t, p, obs, "sgb", tc.id, "third_party_reported")
		mustExec(t, p.Pool, `
			INSERT INTO benchmark_measurements (assertion_id,run_id,benchmark_label,metric_label,
			  raw_value,unit,direction)
			VALUES ($1,'run1','swe_verified','pass_rate',$2,'%','higher')`, tc.id, tc.raw)
		mustExec(t, p.Pool, `
			INSERT INTO measurement_revision_facts (assertion_id,revision_id,identity_evidence_id)
			VALUES ($1,'rev1','as-cap')`, tc.id)
	}
	_, snap := compileLocal(t, p, obs, "sgb")
	signals := snap.Data().Offers[0].Signals
	if len(signals) != 1 || signals[0].ID != "as-b2" || signals[0].Value != 10000 {
		t.Fatalf("boundary signals: %+v", signals)
	}
}

// TestSnapshotBasisByEvidenceClass pins basis_rules keying by the
// measurement's evidence class: a provider_reported measurement projects
// basis publisher_claim.
func TestSnapshotBasisByEvidenceClass(t *testing.T) {
	p := mustPort(t)
	obs := seedSnapshotBase(t, p, "bas", "allowed")
	seedSnapshotOffer(t, p, obs, "bas")
	if _, err := p.RegisterProjectionPolicy(context.Background(), "pp-1", validProjectionPolicy(t), ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	mustExec(t, p.Pool, `UPDATE assertions SET evidence_class = 'provider_reported' WHERE id = 'as-bm'`)
	_, snap := compileLocal(t, p, obs, "bas")
	signals := snap.Data().Offers[0].Signals
	if len(signals) != 1 || signals[0].Basis != "publisher_claim" {
		t.Fatalf("basis must follow evidence class: %+v", signals)
	}
}

// TestSnapshotFindingsClosedSchema pins findings_bytes as the canonical
// encoding of the closed finding objects.
func TestSnapshotFindingsClosedSchema(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	obs := seedSnapshotBase(t, p, "fin", "allowed")
	seedSnapshotOffer(t, p, obs, "fin")
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	res, _ := compileLocal(t, p, obs, "fin")
	// Corrupt the staged bytes in place (test DB only) to force a rejection.
	mustExec(t, p.Pool, `UPDATE snapshots SET artifact_bytes = $1 WHERE artifact_hash = $2`,
		[]byte(`{"schema":1}`), res.ArtifactHash)
	v, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash: res.ArtifactHash, ValidatorVersion: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Outcome != "rejected" || len(v.Findings) == 0 {
		t.Fatalf("validation %+v", v)
	}
	codes := map[string]bool{}
	for _, f := range v.Findings {
		if f.OfferID != "" || f.Detail == "" {
			t.Fatalf("finding shape %+v", f)
		}
		codes[f.Code] = true
	}
	if !codes["artifact_hash_mismatch"] {
		t.Fatalf("codes %v", codes)
	}
	var stored []byte
	if err := p.Pool.QueryRow(ctx,
		`SELECT findings_bytes FROM snapshot_validations WHERE artifact_hash = $1`,
		res.ArtifactHash).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	var decoded []ValidationFinding
	if err := strictDecode(stored, &decoded); err != nil {
		t.Fatalf("findings bytes are not the closed schema: %v", err)
	}
	if len(decoded) != len(v.Findings) {
		t.Fatalf("stored findings %+v", decoded)
	}
}

// TestSnapshotStalePolicyVersionExcludesSource pins B2: a source whose every
// input fails the pinned-version check gets no rights-audit row, and the
// artifact validates.
func TestSnapshotStalePolicyVersionExcludesSource(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	obs := seedSnapshotBase(t, p, "stl", "allowed")
	seedSnapshotOffer(t, p, obs, "stl")
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	now := ms(time.Now().UTC())
	res, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs:                  []InputRef{{ObservationID: obs, SourceID: "stl", PolicyVersion: hexHash(hash32([]byte("wrong-policy")))}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	var audits, inputs int
	if err := p.Pool.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM snapshot_rights_audits WHERE artifact_hash = $1),
		       (SELECT count(*) FROM snapshot_inputs WHERE artifact_hash = $1)`,
		res.ArtifactHash).Scan(&audits, &inputs); err != nil {
		t.Fatal(err)
	}
	if audits != 0 || inputs != 0 {
		t.Fatalf("stale input must leave no audit/input rows: %d/%d", audits, inputs)
	}
	v, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash: res.ArtifactHash, ValidatorVersion: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Outcome != "accepted" {
		t.Fatalf("empty-input artifact must validate: %+v", v)
	}
}

// TestSnapshotIdenticalBytesCollisionConflict pins N1: a different parameter
// set producing identical artifact bytes is conflict, never transient.
func TestSnapshotIdenticalBytesCollisionConflict(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	obs := seedSnapshotBase(t, p, "col", "allowed")
	seedSnapshotOffer(t, p, obs, "col")
	now := ms(time.Now().UTC())
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), now); err != nil {
		t.Fatal(err)
	}
	pp2 := canonical(t, map[string]any{
		"version":                    "pp-2",
		"weights":                    map[string]any{"code": int64(100), "general": int64(50)},
		"task_map":                   map[string]any{"swe_verified": "code"},
		"basis_rules":                permittedBasis(),
		"zero_image_charge_evidence": true, // off-artifact difference only
	})
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-2", pp2, now); err != nil {
		t.Fatal(err)
	}
	policyHex := hexHash(hash32([]byte("policy-col")))
	params := CompileSnapshotParams{
		Inputs:                  []InputRef{{ObservationID: obs, SourceID: "col", PolicyVersion: policyHex}},
		Distribution:            "local",
		ProjectionPolicyVersion: "pp-1",
		CreatedAt:               now,
		KnowledgeCutoff:         now.Add(-time.Minute),
	}
	if _, err := p.CompileSnapshot(ctx, params); err != nil {
		t.Fatal(err)
	}
	params.ProjectionPolicyVersion = "pp-2"
	_, err := p.CompileSnapshot(ctx, params)
	if err == nil || codeOfErr(err) != CodeConflict {
		t.Fatalf("identical-bytes different-params compile must be conflict, got %v", err)
	}
	// Same parameters again: replay, no second row.
	params.ProjectionPolicyVersion = "pp-1"
	if _, err := p.CompileSnapshot(ctx, params); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := p.Pool.QueryRow(ctx, `SELECT count(*) FROM snapshots`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("replay must not add rows: %d", n)
	}
}
