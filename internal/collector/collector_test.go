package collector

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

// startPostgres launches one real Postgres container per test needing it.
func startPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_PASSWORD": "test", "POSTGRES_DB": "collector"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("postgres container unavailable: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })
	host, err := c.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}
	dsn := fmt.Sprintf("postgres://postgres:test@%s:%s/collector", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

func hash32(parts ...[]byte) []byte {
	var all []byte
	for _, p := range parts {
		all = append(all, p...)
	}
	h := sha256.Sum256(all)
	return h[:]
}

func rid(t *testing.T) string {
	t.Helper()
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", b[:])
}

func mustExec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %s: %v", sql, err)
	}
}

func wantFail(t *testing.T, name string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected rejection, got success", name)
	}
	t.Logf("%s rejected: %v", name, err)
}

// insertSourceWithObservation creates one source and one success observation
// and returns the observation id.
func insertSourceWithObservation(t *testing.T, pool *pgxpool.Pool, sourceID, kind string) string {
	t.Helper()
	body := []byte("body")
	mustExec(t, pool,
		`INSERT INTO sources (id,url,source_kind,parser_version,freshness_ceiling_seconds,enabled)
		 VALUES ($1,$2,$3,'p1',3600,true)`, sourceID, "https://"+sourceID, kind)
	mustExec(t, pool, `INSERT INTO content_blobs (hash,bytes) VALUES ($1,$2)`,
		hash32(body), body)
	obsID := "obs-" + rid(t)
	mustExec(t, pool,
		`INSERT INTO observations (id,source_id,resource_url,retrieved_at,http_status,outcome,body_hash,headers_bytes,collector_version)
		 VALUES ($1,$2,$3,$4,200,'success',$5,$6,'c1')`,
		obsID, sourceID, "https://"+sourceID+"/res", time.Now().UTC(), hash32(body), []byte("{}"))
	return obsID
}

func TestMigrationsApplyAndAreIdempotent(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	names, err := MigrationFiles()
	if err != nil {
		t.Fatal(err)
	}
	applied, err := AppliedMigrations(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != len(names) {
		t.Fatalf("applied %d migrations, want %d", len(applied), len(names))
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	// The research-only JSON/JSONB ban also holds on the authority schema.
	var n int
	err = pool.QueryRow(ctx, `
		SELECT count(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND udt_name IN ('json','jsonb')`).Scan(&n)
	if err != nil || n != 0 {
		t.Fatalf("json/jsonb columns present: n=%d err=%v", n, err)
	}
}

func TestObservationStatusCheckRejectsDishonestPairs(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	body := []byte("b")
	mustExec(t, pool,
		`INSERT INTO sources (id,url,source_kind,parser_version,freshness_ceiling_seconds,enabled)
		 VALUES ('s1','https://s1','provider_api','p1',3600,true)`)
	mustExec(t, pool, `INSERT INTO content_blobs (hash,bytes) VALUES ($1,$2)`, hash32(body), body)
	base := `INSERT INTO observations (id,source_id,resource_url,retrieved_at,http_status,outcome,body_hash,headers_bytes,collector_version)
		 VALUES ($1,'s1','u',$2,%d,'%s',$3,'{}','c1')`
	ins := func(id string, status int, outcome string) error {
		_, err := pool.Exec(ctx, fmt.Sprintf(base, status, outcome), id, time.Now().UTC(), hash32(body))
		return err
	}
	// Remote-style success requires 200: 500 and 403 both fail.
	wantFail(t, "success with 500", ins("o1", 500, "success"))
	wantFail(t, "success with 403", ins("o1", 403, "success"))
	// unchanged requires 200 or 304.
	if err := ins("o2", 200, "unchanged"); err != nil {
		t.Fatalf("unchanged with 200: %v", err)
	}
	if err := ins("o3", 304, "unchanged"); err != nil {
		t.Fatalf("unchanged with 304: %v", err)
	}
	wantFail(t, "unchanged with 500", ins("o4", 500, "unchanged"))
	// Honest absence marker: success with status 0 is admitted by the CHECK.
	// Origin (host_local only) is port-enforced and sweep-verified, not a
	// column CHECK.
	if err := ins("o5", 0, "success"); err != nil {
		t.Fatalf("success with status 0: %v", err)
	}
	// source_blocked outside its status set fails.
	wantFail(t, "source_blocked with 404", ins("o6", 404, "source_blocked"))
}

func TestJobAttemptFencingAndRetryChecks(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	now := time.Now().UTC()
	mustExec(t, pool,
		`INSERT INTO sources (id,url,source_kind,parser_version,freshness_ceiling_seconds,enabled)
		 VALUES ('s1','https://s1','provider_api','p1',3600,true)`)
	mustExec(t, pool,
		`INSERT INTO jobs (id,source_id,resource_url,stage,input_hash,parser_version,created_at,attempts,fencing_token)
		 VALUES ('j1','s1','https://s1/x','fetch',$1,'p1',$2,0,0)`, hash32([]byte("w")), now)
	// Attempt referencing a lease that does not exist is rejected by the
	// composite FK: fencing is structural.
	_, err := pool.Exec(ctx,
		`INSERT INTO job_attempts (job_id,token,observed_at,outcome,detail,retry_not_before)
		 VALUES ('j1',1,$1,'observed','',$1)`, now)
	wantFail(t, "attempt without lease", err)
	mustExec(t, pool,
		`INSERT INTO job_leases (job_id,token,worker_id,leased_at,expires_at)
		 VALUES ('j1',1,'w1',$1,$2)`, now, now.Add(time.Minute))
	// deferred requires retry_not_before strictly in the future of observed_at.
	_, err = pool.Exec(ctx,
		`INSERT INTO job_attempts (job_id,token,observed_at,outcome,detail,retry_not_before)
		 VALUES ('j1',1,$1,'deferred','',$1)`, now)
	wantFail(t, "deferred without future retry", err)
	// Terminal outcomes pin retry_not_before to observed_at.
	_, err = pool.Exec(ctx,
		`INSERT INTO job_attempts (job_id,token,observed_at,outcome,detail,retry_not_before)
		 VALUES ('j1',1,$1,'quarantined','',$2)`, now, now.Add(time.Minute))
	wantFail(t, "quarantined with shifted retry_not_before", err)
	mustExec(t, pool,
		`INSERT INTO job_attempts (job_id,token,observed_at,outcome,detail,retry_not_before)
		 VALUES ('j1',1,$1,'deferred','',$2)`, now, now.Add(time.Minute))
}

func TestPublicationTriggerRejectsClassMismatchAndRetiredChannel(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	now := time.Now().UTC()
	mustExec(t, pool, `INSERT INTO publication_channels (channel,channel_class,created_at)
		 VALUES ('lane-local','local',$1),('lane-dist','distributable',$1)`, now)
	mustExec(t, pool,
		`INSERT INTO projection_policies (version,content_hash,content_bytes,recorded_at)
		 VALUES ('pp1',$1,$2,$3)`, hash32([]byte("pp")), []byte("pp"), now)
	body := []byte("artifact")
	ah := hash32(body)
	mustExec(t, pool,
		`INSERT INTO snapshots (artifact_hash,schema_version,policy,distribution,projection_policy_version,created_at,knowledge_cutoff,artifact_bytes,input_manifest_hash,coverage_digest)
		 VALUES ($1,1,'educated-v1','local','pp1',$2,$2,$3,$4,$5)`,
		ah, now, body, hash32([]byte("m")), hash32([]byte("c")))
	pub := func(id, channel, class string) error {
		_, err := pool.Exec(ctx,
			`INSERT INTO snapshot_publications (id,artifact_hash,channel,channel_class,rights_decision,artifact_locator,published_at)
			 VALUES ($1,$2,$3,$4,'allowed','loc',$5)`, id, ah, channel, class, now)
		return err
	}
	// Class mismatch is rejected structurally.
	wantFail(t, "class mismatch", pub("p1", "lane-local", "distributable"))
	// Matching class on an open lane succeeds.
	if err := pub("p2", "lane-local", "local"); err != nil {
		t.Fatalf("open lane publication: %v", err)
	}
	// Retirement is an append-only child row and blocks later publication.
	mustExec(t, pool,
		`INSERT INTO publication_channel_retirements (channel,retired_at)
		 VALUES ('lane-dist',$1)`, now)
	wantFail(t, "retired channel", pub("p3", "lane-dist", "distributable"))
	// The class still matches; only retirement rejects. Idempotent replay on
	// the same hash+channel is the UNIQUE constraint.
	err := pub("p4", "lane-local", "local")
	if err == nil {
		t.Fatal("duplicate publication on same hash+channel accepted")
	}
}

func TestValidatedOutcomePairingTrigger(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	now := time.Now().UTC()
	obsID := insertSourceWithObservation(t, pool, "s-host", "host_local")
	mk := func(receipt, outcome string) {
		t.Helper()
		mustExec(t, pool,
			`INSERT INTO outcome_facts (receipt_id,observation_id,request_id,offer_id,task,outcome,recorded_at)
			 VALUES ($1,$2,'req-1','prov/up','code',$3,$4)`, receipt, obsID, outcome, now)
	}
	mk("r-unval", "unvalidated_completion")
	mk("r-val", "validated_success")
	// A validation child on an unvalidated parent is rejected by the trigger.
	_, err := pool.Exec(ctx,
		`INSERT INTO validated_outcomes (receipt_id,validator_version,recorded_at)
		 VALUES ('r-unval','v1',$1)`, now)
	wantFail(t, "validated child on unvalidated parent", err)
	// On a validated parent it is admitted; the reverse direction (validated
	// parent without a child) is the sweep's job, port-written together.
	mustExec(t, pool,
		`INSERT INTO validated_outcomes (receipt_id,validator_version,recorded_at)
		 VALUES ('r-val','v1',$1)`, now)
	// Empty validator_version is structurally rejected.
	mk("r-val2", "validated_failure")
	_, err = pool.Exec(ctx,
		`INSERT INTO validated_outcomes (receipt_id,validator_version,recorded_at)
		 VALUES ('r-val2','',$1)`, now)
	wantFail(t, "empty validator_version", err)
}

func TestOutcomeUsageSplitAndKnownValueCheck(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	now := time.Now().UTC()
	obsID := insertSourceWithObservation(t, pool, "s-host2", "host_local")
	mustExec(t, pool,
		`INSERT INTO outcome_facts (receipt_id,observation_id,request_id,offer_id,task,outcome,recorded_at)
		 VALUES ('r1',$1,'req-1','prov/up','general','indeterminate',$2)`, obsID, now)
	// Known values are nonnegative.
	_, err := pool.Exec(ctx,
		`INSERT INTO outcome_usage_known (receipt_id,component,value)
		 VALUES ('r1','input_tokens',-1)`)
	wantFail(t, "negative known usage", err)
	// Unknown rows carry no value at all; a component can appear on exactly
	// one side because each PK is per-table only, so disjointness is the
	// port/sweep invariant. Exhaustiveness is visible to a sweep query.
	mustExec(t, pool,
		`INSERT INTO outcome_usage_unknown (receipt_id,component)
		 VALUES ('r1','input_tokens'),('r1','elapsed_millis')`)
	mustExec(t, pool,
		`INSERT INTO outcome_usage_known (receipt_id,component,value)
		 VALUES ('r1','output_tokens',12)`)
	var covered int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM (
			SELECT component FROM outcome_usage_known WHERE receipt_id='r1'
			UNION ALL
			SELECT component FROM outcome_usage_unknown WHERE receipt_id='r1') u
		WHERE component IN ('input_tokens','output_tokens','input_images','output_images','references','searches','elapsed_millis')`).Scan(&covered); err != nil {
		t.Fatal(err)
	}
	if covered != 3 {
		t.Fatalf("merged usage projection covers %d components, want 3", covered)
	}
	// The component vocabulary is closed: a fabricated component is rejected.
	_, err = pool.Exec(ctx,
		`INSERT INTO outcome_usage_known (receipt_id,component,value)
		 VALUES ('r1','gpu_seconds',1)`)
	wantFail(t, "unknown usage component", err)
}

func TestOutboxEventNameVocabulary(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	seq := 0
	ins := func(name string) error {
		seq++
		_, err := pool.Exec(ctx,
			`INSERT INTO outbox_events (event_id,aggregate_id,aggregate_sequence,event_name,schema_version,occurred_at,payload_hash,payload)
			 VALUES ($1,'agg',$2,$3,1,$4,$5,'{}')`, "e-"+rid(t), seq, name, time.Now().UTC(), hash32([]byte(name)))
		return err
	}
	for _, ok := range []string{"SourcePolicyChanged", "ObservationRecorded", "ReceiptRecorded", "IdentityPolicyAvailable"} {
		if err := ins(ok); err != nil {
			t.Fatalf("valid event %s rejected: %v", ok, err)
		}
	}
	wantFail(t, "fabricated event name", ins("SnapshotMiracled"))
	// Sequences are unique per aggregate: replay conflicts structurally.
	if err := ins("JobCompleted"); err != nil {
		t.Fatalf("first JobCompleted: %v", err)
	}
	_, err := pool.Exec(ctx,
		`INSERT INTO outbox_events (event_id,aggregate_id,aggregate_sequence,event_name,schema_version,occurred_at,payload_hash,payload)
		 VALUES ($1,'agg',1,'JobDeferred',1,$2,$3,'{}')`, "e-"+rid(t), time.Now().UTC(), hash32([]byte("d")))
	wantFail(t, "duplicate aggregate sequence", err)
}

func TestLimitAssertionLegalCombinations(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	obsID := insertSourceWithObservation(t, pool, "s-img", "provider_site")
	mkLimit := func(name, unit, bound, scope, op, role string, value int64) error {
		id := "a-" + rid(t)
		if _, err := pool.Exec(ctx,
			`INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,verbatim,parser_version,evidence_class,payload_fingerprint)
			 VALUES ($1,$2,'s-img','up','loc',$3,'p1','provider_reported',$4)`,
			id, obsID, []byte("v"), hash32([]byte(id))); err != nil {
			t.Fatal(err)
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO limit_assertions (assertion_id,limit_name,value,unit,bound,scope,operation,role)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, name, value, unit, bound, scope, op, role)
		return err
	}
	// Legal: output token maximum on request role.
	if err := mkLimit("output_tokens", "token", "maximum", "output", "text_assistance", "request", 4096); err != nil {
		t.Fatalf("legal limit rejected: %v", err)
	}
	// Illegal: mask_count outside image_editing.
	wantFail(t, "mask_count on text_assistance",
		mkLimit("mask_count", "image", "maximum", "input", "text_assistance", "mask", 1))
	// Illegal: unit/scope/role combination mismatch.
	wantFail(t, "input_tokens with output scope",
		mkLimit("input_tokens", "token", "maximum", "output", "text_assistance", "request", 10))
	// Illegal: image_count with reference role.
	wantFail(t, "image_count with reference role",
		mkLimit("image_count", "image", "maximum", "input", "image_generation", "reference", 2))
}

func TestPriceScopeKnowledgeEncoding(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	obsID := insertSourceWithObservation(t, pool, sID("price"), "provider_api")
	offer := "off-" + rid(t)
	mustExec(t, pool,
		`INSERT INTO offers (id,provider,upstream_id,catalog_observation_id,advertised_kind)
		 VALUES ($1,'prov','up',$2,'language')`, offer, obsID)
	a := "a-" + rid(t)
	mustExec(t, pool,
		`INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,verbatim,parser_version,evidence_class,payload_fingerprint)
		 VALUES ($1,$2,$3,'up','loc',$4,'p1','provider_reported',$5)`,
		a, obsID, sID("price"), []byte("v"), hash32([]byte(a)))
	mustExec(t, pool,
		`INSERT INTO price_assertions (assertion_id,offer_id,currency,amount,unit,category,quote_knowledge,scope_hash)
		 VALUES ($1,$2,'USD',1.5,'million_tokens','input_tokens','exact',$3)`,
		a, offer, hash32([]byte("scope")))
	// Known scope dimension requires bytes; unknown requires empty bytes.
	_, err := pool.Exec(ctx,
		`INSERT INTO price_scope_items (assertion_id,dimension,knowledge,value_bytes)
		 VALUES ($1,'region','known',$2)`, a, []byte{})
	wantFail(t, "known dimension with empty bytes", err)
	_, err = pool.Exec(ctx,
		`INSERT INTO price_scope_items (assertion_id,dimension,knowledge,value_bytes)
		 VALUES ($1,'region','unknown',$2)`, a, []byte("us"))
	wantFail(t, "unknown dimension with bytes", err)
	mustExec(t, pool,
		`INSERT INTO price_scope_items (assertion_id,dimension,knowledge,value_bytes)
		 VALUES ($1,'region','known',$2)`, a, []byte("us"))
}

func sID(prefix string) string { return "s-" + prefix }

func TestSnapshotFamilyConstraintPins(t *testing.T) {
	pool := startPostgres(t)
	ctx := context.Background()
	now := time.Now().UTC()
	mustExec(t, pool,
		`INSERT INTO projection_policies (version,content_hash,content_bytes,recorded_at)
		 VALUES ('pp1',$1,$2,$3)`, hash32([]byte("pp")), []byte("pp"), now)
	body := []byte("artifact")
	// Wrong schema version and wrong policy token are structurally rejected.
	_, err := pool.Exec(ctx,
		`INSERT INTO snapshots (artifact_hash,schema_version,policy,distribution,projection_policy_version,created_at,knowledge_cutoff,artifact_bytes,input_manifest_hash,coverage_digest)
		 VALUES ($1,2,'educated-v1','local','pp1',$2,$2,$3,$4,$5)`,
		hash32(body, []byte{2}), now, body, hash32([]byte("m")), hash32([]byte("c")))
	wantFail(t, "schema_version 2", err)
	_, err = pool.Exec(ctx,
		`INSERT INTO snapshots (artifact_hash,schema_version,policy,distribution,projection_policy_version,created_at,knowledge_cutoff,artifact_bytes,input_manifest_hash,coverage_digest)
		 VALUES ($1,1,'calibrated-v9','local','pp1',$2,$2,$3,$4,$5)`,
		hash32(body, []byte{3}), now, body, hash32([]byte("m")), hash32([]byte("c")))
	wantFail(t, "unpinned policy token", err)
}
