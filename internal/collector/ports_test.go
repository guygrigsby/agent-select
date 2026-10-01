package collector

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func mustPort(t *testing.T) *Ports {
	t.Helper()
	return New(startPostgres(t))
}

func registerSource(t *testing.T, p *Ports, id, kind string, enabled bool) {
	t.Helper()
	mustExec(t, p.Pool,
		`INSERT INTO sources (id,url,source_kind,parser_version,freshness_ceiling_seconds,enabled)
		 VALUES ($1,$2,$3,'p1',3600,$4)`, id, "https://"+id, kind, enabled)
}

func canonical(t *testing.T, v any) []byte {
	t.Helper()
	b, err := canonicalize(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func fetchEnvelopeDoc(t *testing.T, obs *AttemptObservation) []byte {
	t.Helper()
	return canonical(t, map[string]any{
		"stage":          "fetch",
		"observation_id": obs.ObservationID,
		"resource_url":   obs.ResourceURL,
		"body_hash":      hexHash(obs.BodyBytes),
		"http_status":    int64(obs.HTTPStatus),
		"headers_hash":   hexHash(obs.HeadersBytes),
	})
}

// fetchEnvelopeWithStatus builds a canonical but different fetch envelope.
func fetchEnvelopeDocStatus(t *testing.T, obs *AttemptObservation, status int) []byte {
	t.Helper()
	return canonical(t, map[string]any{
		"stage":          "fetch",
		"observation_id": obs.ObservationID,
		"resource_url":   obs.ResourceURL,
		"body_hash":      hexHash(obs.BodyBytes),
		"http_status":    int64(status),
		"headers_hash":   hexHash(obs.HeadersBytes),
	})
}

func allUsage(known map[string]int64) []ReceiptUsage {
	var u []ReceiptUsage
	for _, c := range usageComponents {
		if v, ok := known[c]; ok {
			u = append(u, ReceiptUsage{Component: c, Known: true, Value: v})
		} else {
			u = append(u, ReceiptUsage{Component: c})
		}
	}
	return u
}

func codeOf(t *testing.T, err error) Code {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	se, ok := err.(*Error)
	if !ok {
		t.Fatalf("error %v is not a storage error", err)
	}
	return se.Code
}

func TestEnqueueJobIdempotencyAndGates(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	registerSource(t, p, "s1", "provider_api", true)
	h := hash32([]byte("w"))
	r1, err := p.EnqueueJob(ctx, "s1", "https://s1/x", "fetch", h, "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	if r1.Existing {
		t.Fatal("first enqueue reported existing")
	}
	r2, err := p.EnqueueJob(ctx, "s1", "https://s1/x", "fetch", h, "p1", now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if !r2.Existing || r2.JobID != r1.JobID {
		t.Fatalf("replay: existing=%v id=%s want id=%s", r2.Existing, r2.JobID, r1.JobID)
	}
	if _, err := p.EnqueueJob(ctx, "nope", "u", "fetch", h, "p1", now); codeOf(t, err) != CodeNotFound {
		t.Fatal("unknown source accepted")
	}
	registerSource(t, p, "s2", "provider_api", false)
	if _, err := p.EnqueueJob(ctx, "s2", "u", "fetch", h, "p1", now); codeOf(t, err) != CodeSourceBlocked {
		t.Fatal("disabled source accepted")
	}
	if _, err := p.EnqueueJob(ctx, "s1", "u", "transmogrify", h, "p1", now); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("bad stage accepted")
	}
}

func leaseFetchJob(t *testing.T, p *Ports, sourceID string) (jobID string, token int64, obs *AttemptObservation, output []byte) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC()
	registerSource(t, p, sourceID, "provider_api", true)
	r, err := p.EnqueueJob(ctx, sourceID, "https://"+sourceID+"/r", "fetch", hash32([]byte("w"+sourceID)), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := p.LeaseJob(ctx, r.JobID, "w1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	obs = &AttemptObservation{
		ObservationID: "obs-" + r.JobID,
		ResourceURL:   "https://" + sourceID + "/r",
		RetrievedAt:   now,
		HTTPStatus:    200,
		Outcome:       "success",
		BodyBytes:     []byte("body"),
		HeadersBytes:  []byte("h"),
	}
	return r.JobID, lease.Token, obs, fetchEnvelopeDoc(t, obs)
}

func countEvents(t *testing.T, p *Ports) int {
	t.Helper()
	var n int
	if err := p.Pool.QueryRow(context.Background(),
		`SELECT count(*) FROM outbox_events`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRecordAttemptFetchDurableSuccessAndEvents(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID, token, obs, output := leaseFetchJob(t, p, "s1")
	res, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: jobID, Token: token, CollectorVersion: "c1",
		RecordedAt: time.Now().UTC(), Observation: obs, OutputBytes: output,
		Assertions: []AssertionInput{{
			SubjectSourceID: "s1", SubjectUpstreamID: "up", Locator: "l",
			Verbatim: []byte("v"), EvidenceClass: "provider_reported", ParserVersion: "p1",
			Capability: &CapabilityAssertion{
				Capability: "tool_use", Support: "supported",
				Operation: "text_assistance", Direction: "not_applicable"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ObservationID != obs.ObservationID || res.AcceptedAssertions != 1 ||
		res.QuarantinedAssertions != 0 || res.OutputID != "out-"+jobID {
		t.Fatalf("result %+v", res)
	}
	// Two events on the observation aggregate: StageOutputRecorded then
	// ObservationRecorded, sequences 1 and 2.
	var names []string
	var seqs []int64
	rows, err := p.Pool.Query(ctx, `
		SELECT event_name, aggregate_sequence FROM outbox_events
		WHERE aggregate_id = $1 ORDER BY aggregate_sequence`, obs.ObservationID)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var n string
		var s int64
		if err := rows.Scan(&n, &s); err != nil {
			t.Fatal(err)
		}
		names = append(names, n)
		seqs = append(seqs, s)
	}
	rows.Close()
	if len(names) != 2 || names[0] != "StageOutputRecorded" || names[1] != "ObservationRecorded" ||
		seqs[0] != 1 || seqs[1] != 2 {
		t.Fatalf("events %v seqs %v", names, seqs)
	}
	// Idempotent replay: same bytes, independent of the lease token.
	res2, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: jobID, Token: token + 99, CollectorVersion: "c1",
		RecordedAt: time.Now().UTC().Add(time.Hour), Observation: obs, OutputBytes: output,
	})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if res2.ObservationID != res.ObservationID || string(res2.OutputHash) != string(res.OutputHash) {
		t.Fatalf("replay result %+v", res2)
	}
	// Same job, different canonical output bytes: conflict.
	bad := fetchEnvelopeDocStatus(t, obs, 201)
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: jobID, Token: token, CollectorVersion: "c1",
		RecordedAt: time.Now().UTC(), Observation: obs, OutputBytes: bad,
	}); codeOf(t, err) != CodeConflict {
		t.Fatal("different output for same job accepted")
	}
	// Envelope not matching the observation bytes: invalid evidence.
	fresh, err := p.EnqueueJob(ctx, "s1", "https://s1/other", "fetch", hash32([]byte("other")), "p1", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	lease2, err := p.LeaseJob(ctx, fresh.JobID, "w1", time.Minute, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	swapped := *obs
	swapped.ObservationID = "obs-" + fresh.JobID
	mismatch := fetchEnvelopeDoc(t, &swapped)
	mismatch[10] = 'x' // break the body hash
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: fresh.JobID, Token: lease2.Token, CollectorVersion: "c1",
		RecordedAt: time.Now().UTC(), Observation: &swapped, OutputBytes: mismatch,
	}); codeOf(t, err) == CodeConflict {
		t.Fatal("mismatched envelope misclassified")
	}
	// Completion requires the recorded output id; then the job is completed.
	if _, err := p.CompleteJob(ctx, jobID, token, res.OutputID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	state, err := p.JobState(ctx, jobID, time.Now().UTC())
	if err != nil || state != "completed" {
		t.Fatalf("state %s err %v", state, err)
	}
	if _, err := p.CompleteJob(ctx, jobID, token, "out-bogus", time.Now().UTC()); codeOf(t, err) != CodeConflict {
		t.Fatal("bogus output accepted")
	}
}

func TestRecordAttemptQuarantinesInvalidAssertions(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID, token, obs, output := leaseFetchJob(t, p, "s1")
	res, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: jobID, Token: token, CollectorVersion: "c1",
		RecordedAt: time.Now().UTC(), Observation: obs, OutputBytes: output,
		Assertions: []AssertionInput{
			{ // structural failure: no child
				SubjectSourceID: "s1", SubjectUpstreamID: "up", Locator: "l",
				Verbatim: []byte("v"), EvidenceClass: "provider_reported", ParserVersion: "p1",
			},
			{ // CHECK failure: mask limits belong to image_editing only
				SubjectSourceID: "s1", SubjectUpstreamID: "up", Locator: "l2",
				Verbatim: []byte("v"), EvidenceClass: "provider_reported", ParserVersion: "p1",
				Limit: &LimitAssertion{
					Name: "mask_count", Value: 1, Unit: "image", Bound: "maximum",
					Scope: "input", Operation: "text_assistance", Role: "mask"},
			},
			{ // legal
				SubjectSourceID: "s1", SubjectUpstreamID: "up", Locator: "l3",
				Verbatim: []byte("v"), EvidenceClass: "third_party_reported", ParserVersion: "p1",
				Claim: &ClaimAssertion{
					Dimension: "code", Polarity: "strength", ClaimText: "fast"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.AcceptedAssertions != 1 || res.QuarantinedAssertions != 2 {
		t.Fatalf("accepted %d quarantined %d", res.AcceptedAssertions, res.QuarantinedAssertions)
	}
	var n int
	if err := p.Pool.QueryRow(ctx, `SELECT count(*) FROM assertions`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("stored assertions %d err %v", n, err)
	}
	// Quarantine left no partial rows: exactly one child table row.
	if err := p.Pool.QueryRow(ctx,
		`SELECT count(*) FROM limit_assertions`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("limit rows %d err %v", n, err)
	}
	// Parser version mismatch is fatal for the whole call.
	jobID2, token2, obs2, output2 := leaseFetchJob(t, p, "s2")
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: jobID2, Token: token2, CollectorVersion: "c1",
		RecordedAt: time.Now().UTC(), Observation: obs2, OutputBytes: output2,
		Assertions: []AssertionInput{{
			SubjectSourceID: "s2", SubjectUpstreamID: "up", Locator: "l",
			Verbatim: []byte("v"), EvidenceClass: "provider_reported", ParserVersion: "p9",
			Claim: &ClaimAssertion{Dimension: "code", Polarity: "strength", ClaimText: "x"},
		}},
	}); codeOf(t, err) != CodeInvalidEvidence {
		t.Fatal("parser mismatch accepted")
	}
	// Stale lease rejects new work.
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: jobID2, Token: token2 + 5, CollectorVersion: "c1",
		RecordedAt: time.Now().UTC(), Observation: obs2, OutputBytes: output2,
	}); codeOf(t, err) != CodeStaleLease {
		t.Fatal("stale lease accepted")
	}
}

func TestRecordAttemptParseResolveCompile(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()

	// Parse stage over an existing fetch observation.
	fetchJob, fetchTok, fetchObs, fetchOut := leaseFetchJob(t, p, "s1")
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: fetchJob, Token: fetchTok, CollectorVersion: "c1",
		RecordedAt: now, Observation: fetchObs, OutputBytes: fetchOut,
	}); err != nil {
		t.Fatal(err)
	}
	registerSource(t, p, "s1b", "provider_api", true)
	pr, err := p.EnqueueJob(ctx, "s1b", "https://s1b/p", "parse", hash32([]byte("parse")), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := p.LeaseJob(ctx, pr.JobID, "w1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	parseOut := canonical(t, map[string]any{
		"stage": "parse", "observation_id": fetchObs.ObservationID,
		"assertion_ids": []string{assertionID(pr.JobID, 0)},
	})
	res, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: pr.JobID, Token: lease.Token, CollectorVersion: "c1", RecordedAt: now,
		OutputBytes: parseOut,
		Assertions: []AssertionInput{{
			SubjectSourceID: "s1", SubjectUpstreamID: "up", Locator: "l",
			Verbatim: []byte("v"), EvidenceClass: "provider_reported", ParserVersion: "p1",
			Claim: &ClaimAssertion{Dimension: "general", Polarity: "weakness", ClaimText: "slow"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ObservationID != fetchObs.ObservationID {
		t.Fatalf("parse observation %s", res.ObservationID)
	}
	// Parse creates no observation of its own and no ObservationRecorded.
	var n int
	if err := p.Pool.QueryRow(ctx,
		`SELECT count(*) FROM observations`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("observations %d err %v", n, err)
	}
	if err := p.Pool.QueryRow(ctx,
		`SELECT count(*) FROM outbox_events WHERE event_name = 'ObservationRecorded'`).
		Scan(&n); err != nil || n != 1 {
		t.Fatalf("ObservationRecorded events %d err %v", n, err)
	}

	// Resolve stage: findings restricted to researched/identity_ambiguous.
	registerSource(t, p, "s1c", "provider_api", true)
	rr, err := p.EnqueueJob(ctx, "s1c", "https://s1c/r", "resolve", hash32([]byte("resolve")), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	rLease, err := p.LeaseJob(ctx, rr.JobID, "w1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	resolveOut := canonical(t, map[string]any{
		"stage": "resolve", "observation_id": fetchObs.ObservationID,
		"finding_ids": []string{"some-model|vercel_catalog"},
	})
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: rr.JobID, Token: rLease.Token, CollectorVersion: "c1", RecordedAt: now,
		OutputBytes: resolveOut,
		CoverageFindings: []CoverageFinding{{
			SubjectUpstreamID: "some-model", Channel: "vercel_catalog",
			Outcome: "blocked", Detail: "nope"}},
	}); codeOf(t, err) != CodeInvalidEvidence {
		t.Fatal("resolve accepted a blocked finding")
	}
	blocked := countEvents(t, p)
	resolveOK := canonical(t, map[string]any{
		"stage": "resolve", "observation_id": fetchObs.ObservationID,
		"finding_ids": []string{"some-model|vercel_catalog"},
	})
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: rr.JobID, Token: rLease.Token, CollectorVersion: "c1", RecordedAt: now,
		OutputBytes: resolveOK,
		CoverageFindings: []CoverageFinding{{
			SubjectUpstreamID: "some-model", Channel: "vercel_catalog",
			Outcome: "researched", Detail: "ok"}},
	}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if err := p.Pool.QueryRow(ctx,
		`SELECT count(*) FROM coverage_findings`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("findings %d err %v", n, err)
	}

	// Compile stage: no observation, null link.
	registerSource(t, p, "s1d", "provider_api", true)
	compileOut := canonical(t, map[string]any{
		"stage": "compile", "artifact_hash": hexHash([]byte("artifact")),
	})
	cr, err := p.EnqueueJob(ctx, "s1d", "https://s1d/c", "compile", hash32([]byte("compile")), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	cLease, err := p.LeaseJob(ctx, cr.JobID, "w1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	cres, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: cr.JobID, Token: cLease.Token, CollectorVersion: "c1", RecordedAt: now,
		OutputBytes: compileOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cres.ObservationID != "" {
		t.Fatalf("compile observation %s", cres.ObservationID)
	}
	var linkObs *string
	if err := p.Pool.QueryRow(ctx,
		`SELECT observation_id FROM stage_output_links WHERE job_id = $1`, cr.JobID).
		Scan(&linkObs); err != nil {
		t.Fatal(err)
	}
	if linkObs != nil {
		t.Fatal("compile link carries an observation")
	}
	_ = blocked
}

func TestDeferJobAndStateDerivation(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	registerSource(t, p, "s1", "provider_api", true)
	r, err := p.EnqueueJob(ctx, "s1", "https://s1/x", "fetch", hash32([]byte("w")), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := p.LeaseJob(ctx, r.JobID, "w1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	// deferred requires a future retry.
	if err := p.DeferJob(ctx, r.JobID, lease.Token, "deferred", "busy", now, now); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("deferred without future retry accepted")
	}
	retry := now.Add(5 * time.Minute)
	if err := p.DeferJob(ctx, r.JobID, lease.Token, "deferred", "busy", now, retry); err != nil {
		t.Fatal(err)
	}
	state, err := p.JobState(ctx, r.JobID, now.Add(time.Minute))
	if err != nil || state != "deferred" {
		t.Fatalf("state %s err %v", state, err)
	}
	// Before the retry deadline leasing conflicts.
	if _, err := p.LeaseJob(ctx, r.JobID, "w2", time.Minute, now.Add(time.Minute)); codeOf(t, err) != CodeConflict {
		t.Fatal("pre-deadline lease granted")
	}
	// After the deadline the job is queued and leaseable.
	state, err = p.JobState(ctx, r.JobID, retry.Add(time.Second))
	if err != nil || state != "queued" {
		t.Fatalf("state %s err %v", state, err)
	}
	lease2, err := p.LeaseJob(ctx, r.JobID, "w2", time.Minute, retry.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	// Terminal deferral: retry pinned to observed_at, no event.
	before := countEvents(t, p)
	terminalAt := retry.Add(2 * time.Second)
	if err := p.DeferJob(ctx, r.JobID, lease2.Token, "quarantined", "bad shape", terminalAt, terminalAt); err != nil {
		t.Fatal(err)
	}
	if countEvents(t, p) != before {
		t.Fatal("terminal deferral emitted an event")
	}
	state, err = p.JobState(ctx, r.JobID, retry.Add(time.Minute))
	if err != nil || state != "quarantined" {
		t.Fatalf("state %s err %v", state, err)
	}
	// Quarantined jobs conflict on lease; reentry needs a changed tuple.
	if _, err := p.LeaseJob(ctx, r.JobID, "w3", time.Minute, retry.Add(time.Minute)); codeOf(t, err) != CodeConflict {
		t.Fatal("quarantined job leased")
	}
	if _, err := p.ReenterJob(ctx, r.JobID, hash32([]byte("w")), "p1", now); codeOf(t, err) != CodeConflict {
		t.Fatal("unchanged reentry accepted")
	}
	re, err := p.ReenterJob(ctx, r.JobID, hash32([]byte("w2")), "p2", now)
	if err != nil {
		t.Fatal(err)
	}
	if re.Existing || re.JobID == r.JobID {
		t.Fatalf("reentry %+v", re)
	}
}

func TestRecoverJobRequiresExpiredLease(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	registerSource(t, p, "s1", "provider_api", true)
	r, err := p.EnqueueJob(ctx, "s1", "https://s1/x", "fetch", hash32([]byte("w")), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := p.LeaseJob(ctx, r.JobID, "w1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.RecoverJob(ctx, r.JobID, lease.Token, now); codeOf(t, err) != CodeStaleLease {
		t.Fatal("unexpired lease recovered")
	}
	state, err := p.RecoverJob(ctx, r.JobID, lease.Token, now.Add(2*time.Minute))
	if err != nil || state != "queued" {
		t.Fatalf("recover state %s err %v", state, err)
	}
	var attempts int64
	if err := p.Pool.QueryRow(ctx,
		`SELECT attempts FROM jobs WHERE id = $1`, r.JobID).Scan(&attempts); err != nil || attempts != 2 {
		t.Fatalf("attempts %d err %v", attempts, err)
	}
	var n int
	if err := p.Pool.QueryRow(ctx,
		`SELECT count(*) FROM job_attempts WHERE job_id = $1`, r.JobID).Scan(&n); err != nil || n != 0 {
		t.Fatalf("recover wrote attempt rows: %d", n)
	}
	if _, err := p.RecoverJob(ctx, r.JobID, 9999, now.Add(3*time.Minute)); codeOf(t, err) != CodeNotFound {
		t.Fatal("unknown lease recovered")
	}
}

func TestLeaseJobSingleActiveLease(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	registerSource(t, p, "s1", "provider_api", true)
	r, err := p.EnqueueJob(ctx, "s1", "https://s1/x", "fetch", hash32([]byte("w")), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.LeaseJob(ctx, "ghost", "w1", time.Minute, now); codeOf(t, err) != CodeNotFound {
		t.Fatal("unknown job leased")
	}
	l1, err := p.LeaseJob(ctx, r.JobID, "w1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	// Same worker replays the same lease.
	l2, err := p.LeaseJob(ctx, r.JobID, "w1", time.Minute, now.Add(time.Second))
	if err != nil || l2.Token != l1.Token || !l2.ExpiresAt.Equal(l1.ExpiresAt) {
		t.Fatalf("replay %+v err %v", l2, err)
	}
	// Another worker conflicts while live.
	if _, err := p.LeaseJob(ctx, r.JobID, "w2", time.Minute, now.Add(time.Second)); codeOf(t, err) != CodeConflict {
		t.Fatal("second active lease granted")
	}
	// After expiry another worker gets a fresh, higher token.
	l3, err := p.LeaseJob(ctx, r.JobID, "w2", time.Minute, now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if l3.Token <= l1.Token {
		t.Fatalf("token %d not monotonic over %d", l3.Token, l1.Token)
	}
	var fencing int64
	if err := p.Pool.QueryRow(ctx,
		`SELECT fencing_token FROM jobs WHERE id = $1`, r.JobID).Scan(&fencing); err != nil ||
		fencing != l3.Token {
		t.Fatalf("fencing %d err %v", fencing, err)
	}
}

func TestRecordReceiptLifecycle(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	registerSource(t, p, "host", "host_local", true)
	receipt := canonical(t, map[string]any{
		"receipt_id": "rcpt-1", "request_id": "req-1", "offer_id": "prov/up",
		"task": "code", "outcome": "validated_success",
		"usage": map[string]any{"input_tokens": int64(100), "output_tokens": int64(50)},
	})
	res, err := p.RecordReceipt(ctx, RecordReceiptParams{
		ReceiptID: "rcpt-1", RequestID: "req-1", OfferID: "prov/up",
		Task: "code", Outcome: "validated_success", ValidatorVersion: "tests-1",
		RecordedAt: now, Usage: allUsage(map[string]int64{"input_tokens": 100, "output_tokens": 50}),
		ReceiptBytes: receipt, CollectorVersion: "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	// usage_hash over the seven members, known carrying value, unknown bare.
	wantDoc := []map[string]any{
		{"component": "elapsed_millis"},
		{"component": "input_images"},
		{"component": "input_tokens", "value": int64(100)},
		{"component": "output_images"},
		{"component": "output_tokens", "value": int64(50)},
		{"component": "references"},
		{"component": "searches"},
	}
	if string(res.UsageHash) != string(canonicalHash(canonical(t, wantDoc))) {
		t.Fatal("usage hash does not match the canonical seven-member document")
	}
	// host_local observation carries status 0 and the receipt body hash.
	var status int
	var bodyHash []byte
	if err := p.Pool.QueryRow(ctx, `
		SELECT http_status, body_hash FROM observations WHERE id = $1`, res.ObservationID).
		Scan(&status, &bodyHash); err != nil {
		t.Fatal(err)
	}
	if status != 0 || string(bodyHash) != string(canonicalHash(receipt)) {
		t.Fatalf("receipt observation status %d", status)
	}
	// Validated split child exists.
	var one int
	if err := p.Pool.QueryRow(ctx,
		`SELECT 1 FROM validated_outcomes WHERE receipt_id = 'rcpt-1'`).Scan(&one); err != nil {
		t.Fatal(err)
	}
	// Idempotent replay with identical bytes.
	res2, err := p.RecordReceipt(ctx, RecordReceiptParams{
		ReceiptID: "rcpt-1", RequestID: "req-1", OfferID: "prov/up",
		Task: "code", Outcome: "validated_success", ValidatorVersion: "tests-1",
		RecordedAt:   now.Add(time.Second),
		Usage:        allUsage(map[string]int64{"input_tokens": 100, "output_tokens": 50}),
		ReceiptBytes: receipt, CollectorVersion: "c1",
	})
	if err != nil || res2.ObservationID != res.ObservationID {
		t.Fatalf("replay %+v err %v", res2, err)
	}
	// Same ID different bytes conflicts.
	other := canonical(t, map[string]any{"receipt_id": "rcpt-1", "x": int64(1)})
	if _, err := p.RecordReceipt(ctx, RecordReceiptParams{
		ReceiptID: "rcpt-1", RequestID: "req-1", OfferID: "prov/up",
		Task: "code", Outcome: "cancelled", RecordedAt: now,
		Usage: allUsage(nil), ReceiptBytes: other, CollectorVersion: "c1",
	}); codeOf(t, err) != CodeConflict {
		t.Fatal("same id different bytes accepted")
	}
	// Unvalidated outcomes carry no validator_version.
	if _, err := p.RecordReceipt(ctx, RecordReceiptParams{
		ReceiptID: "rcpt-2", RequestID: "req-1", OfferID: "prov/up",
		Task: "code", Outcome: "unvalidated_completion", ValidatorVersion: "v",
		RecordedAt: now, Usage: allUsage(nil), ReceiptBytes: other, CollectorVersion: "c1",
	}); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("unvalidated with validator accepted")
	}
	// Missing usage component rejected.
	short := allUsage(nil)[:6]
	if _, err := p.RecordReceipt(ctx, RecordReceiptParams{
		ReceiptID: "rcpt-3", RequestID: "req-1", OfferID: "prov/up",
		Task: "code", Outcome: "cancelled", RecordedAt: now,
		Usage: short, ReceiptBytes: other, CollectorVersion: "c1",
	}); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("partial usage accepted")
	}
}

func TestPublicationChannelLifecycle(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	lane, err := p.RegisterPublicationChannel(ctx, "lane", "local", now)
	if err != nil {
		t.Fatal(err)
	}
	if lane.RetiredAt != nil {
		t.Fatal("new lane retired")
	}
	if _, err := p.RegisterPublicationChannel(ctx, "lane", "local", now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := p.RegisterPublicationChannel(ctx, "lane", "distributable", now); codeOf(t, err) != CodeConflict {
		t.Fatal("class change on live lane accepted")
	}
	retired, err := p.RetirePublicationChannel(ctx, "lane", now.Add(time.Minute))
	if err != nil || retired.RetiredAt == nil {
		t.Fatalf("retire %+v err %v", retired, err)
	}
	stored := *retired.RetiredAt
	again, err := p.RetirePublicationChannel(ctx, "lane", now.Add(2*time.Minute))
	if err != nil || !again.RetiredAt.Equal(stored) {
		t.Fatalf("re-retire %+v err %v", again, err)
	}
	if _, err := p.RegisterPublicationChannel(ctx, "lane", "local", now); codeOf(t, err) != CodeConflict {
		t.Fatal("re-register of retired lane accepted")
	}
	if _, err := p.RetirePublicationChannel(ctx, "ghost", now); codeOf(t, err) != CodeNotFound {
		t.Fatal("unknown channel retired")
	}
}

func validProjectionPolicy(t *testing.T) []byte {
	t.Helper()
	return canonical(t, map[string]any{
		"version":  "pp-1",
		"weights":  map[string]any{"code": int64(100), "general": int64(50)},
		"task_map": map[string]any{"swe_verified": "code"},
		"basis_rules": map[string]any{
			"provider_reported":    "publisher_claim",
			"third_party_reported": "published",
			"competitor_reported":  "published",
			"host_observed":        "host_observed",
		},
		"zero_image_charge_evidence": false,
	})
}

func projectionPolicy(t *testing.T, version string, weights, basis map[string]any) []byte {
	t.Helper()
	return canonical(t, map[string]any{
		"version":                    version,
		"weights":                    weights,
		"task_map":                   map[string]any{"swe_verified": "code"},
		"basis_rules":                basis,
		"zero_image_charge_evidence": false,
	})
}

func permittedBasis() map[string]any {
	return map[string]any{
		"provider_reported":    "publisher_claim",
		"third_party_reported": "published",
		"competitor_reported":  "published",
		"host_observed":        "host_observed",
	}
}

func TestRegisterProjectionPolicy(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	doc := validProjectionPolicy(t)
	if err := p.RegisterProjectionPolicy(ctx, "pp-1", doc, now); err != nil {
		t.Fatal(err)
	}
	if err := p.RegisterProjectionPolicy(ctx, "pp-1", doc, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	// Same version, different canonical content: conflict.
	different := projectionPolicy(t, "pp-1",
		map[string]any{"code": int64(101)}, permittedBasis())
	if err := p.RegisterProjectionPolicy(ctx, "pp-1", different, now); codeOf(t, err) != CodeConflict {
		t.Fatal("version re-registration with different content accepted")
	}
	// Version mismatch inside content.
	if err := p.RegisterProjectionPolicy(ctx, "pp-2", doc, now); codeOf(t, err) != CodeInvalidEvidence {
		t.Fatal("version mismatch accepted")
	}
	// Illegal basis derivation.
	bad := projectionPolicy(t, "pp-3", map[string]any{"code": int64(1)},
		map[string]any{
			"provider_reported":    "published",
			"third_party_reported": "published",
			"competitor_reported":  "published",
			"host_observed":        "host_observed",
		})
	if err := p.RegisterProjectionPolicy(ctx, "pp-3", bad, now); codeOf(t, err) != CodeInvalidEvidence {
		t.Fatal("illegal basis derivation accepted")
	}
	// Weight out of range.
	badW := projectionPolicy(t, "pp-4", map[string]any{"code": int64(10001)}, permittedBasis())
	if err := p.RegisterProjectionPolicy(ctx, "pp-4", badW, now); codeOf(t, err) != CodeInvalidEvidence {
		t.Fatal("out-of-range weight accepted")
	}
	// Non-canonical bytes rejected.
	if err := p.RegisterProjectionPolicy(ctx, "pp-5", []byte("{not json}"), now); codeOf(t, err) != CodeInvalidEvidence {
		t.Fatal("non-canonical content accepted")
	}
}

func TestOutboxDeliveryScanAndReceipts(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	now := time.Now().UTC()
	registerSource(t, p, "host", "host_local", true)
	for i := range 3 {
		id := fmt.Sprintf("rcpt-%d", i)
		body := canonical(t, map[string]any{"receipt_id": id, "n": i})
		if _, err := p.RecordReceipt(ctx, RecordReceiptParams{
			ReceiptID: id, RequestID: "req", OfferID: "prov/up",
			Task: "code", Outcome: "cancelled", RecordedAt: now.Add(time.Duration(i) * time.Second),
			Usage: allUsage(nil), ReceiptBytes: body, CollectorVersion: "c1",
		}); err != nil {
			t.Fatal(err)
		}
	}
	events, err := p.DeliveryScan(ctx, "consumer-a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("events %d", len(events))
	}
	for i := 1; i < len(events); i++ {
		if events[i].OccurredAt.Before(events[i-1].OccurredAt) {
			t.Fatal("delivery scan not ordered by occurred_at")
		}
	}
	if err := p.RecordConsumerReceipt(ctx, "consumer-a", events[0].EventID, now); err != nil {
		t.Fatal(err)
	}
	if err := p.RecordConsumerReceipt(ctx, "consumer-a", events[0].EventID, now); err != nil {
		t.Fatal(err)
	}
	remaining, err := p.DeliveryScan(ctx, "consumer-a", 10)
	if err != nil || len(remaining) != 2 {
		t.Fatalf("remaining %d err %v", len(remaining), err)
	}
	// A second consumer sees everything: receipts are per consumer.
	other, err := p.DeliveryScan(ctx, "consumer-b", 10)
	if err != nil || len(other) != 3 {
		t.Fatalf("consumer-b %d err %v", len(other), err)
	}
	// The ReceiptRecorded payload is canonical and its hash matches.
	if !isCanonical(events[0].Payload) || string(events[0].PayloadHash) != string(canonicalHash(events[0].Payload)) {
		t.Fatal("payload not canonical or hash mismatch")
	}
	// Bad scan arguments are rejected.
	if _, err := p.DeliveryScan(ctx, "", 10); codeOf(t, err) != CodeInvalidRequest {
		t.Fatal("empty consumer accepted")
	}
}
