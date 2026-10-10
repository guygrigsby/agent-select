package collector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/guygrigsby/agent-select/routing"
)

func testStore(t *testing.T, keep int) *ArtifactStore {
	t.Helper()
	store, err := NewArtifactStore(t.TempDir(), keep)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func successorCfg() SuccessorConfig {
	return SuccessorConfig{
		Channel:               "local",
		ValidatorVersion:      "vv1",
		BootstrapPolicy:       "pp-1",
		BootstrapDistribution: "local",
	}
}

func seedSuccessorBase(t *testing.T, p *Ports) {
	t.Helper()
	ctx := context.Background()
	now := ms(time.Now().UTC())
	if _, err := p.RegisterProjectionPolicy(ctx, "pp-1", validProjectionPolicy(t), now); err != nil {
		t.Fatal(err)
	}
	if _, err := p.RegisterPublicationChannel(ctx, "local", "local", now); err != nil {
		t.Fatal(err)
	}
}

// seedChangedObservation adds a newer success observation for the source
// whose offer carries a different exact input price, so the successor
// projection genuinely changes.
func seedChangedObservation(t *testing.T, p *Ports, sourceID string, amount float64, suffix string) string {
	t.Helper()
	now := time.Now().UTC()
	body := []byte(`{"models":[{"id":"u","object":"model"}],"src":"` + sourceID + `","v":"` + suffix + `"}`)
	mustExec(t, p.Pool, `INSERT INTO content_blobs (hash,bytes) VALUES ($1,$2)`,
		hash32(body), body)
	obsID := "obs-" + sourceID + "-" + suffix
	mustExec(t, p.Pool, `
		INSERT INTO observations (id,source_id,resource_url,retrieved_at,http_status,outcome,
		  body_hash,headers_bytes,collector_version)
		VALUES ($1,$2,$3,$4,200,'success',$5,'{}','cv1')`,
		obsID, sourceID, "https://"+sourceID+"/models", now.Add(time.Minute), hash32(body))
	mustExec(t, p.Pool, `
		INSERT INTO offer_listing_facts (offer_id,observation_id,action,observed_at)
		VALUES ('of1',$1,'listed',$2)`, obsID, now.Add(time.Minute))
	mustExec(t, p.Pool, `
		INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,
		  verbatim,parser_version,evidence_class,payload_fingerprint)
		VALUES ($1,$2,$3,'u','loc-3','{}'::bytea,'p1','provider_reported',$4)`,
		"as-pr-"+suffix, obsID, sourceID, hash32([]byte("pr-"+suffix)))
	mustExec(t, p.Pool, `
		INSERT INTO price_assertions (assertion_id,offer_id,currency,amount,unit,category,
		  quote_knowledge,scope_hash)
		VALUES ($1,'of1','USD',$2,'million_tokens','input_tokens','exact',$3)`,
		"as-pr-"+suffix, amount, hash32([]byte("price-scope-"+suffix)))
	return obsID
}

func countRows(t *testing.T, p *Ports, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := p.Pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestArtifactStoreStageAdoptPrune(t *testing.T) {
	store := testStore(t, 2)
	a := makeArtifact(t, "a")
	b := makeArtifact(t, "b")
	c := makeArtifact(t, "c")

	if err := store.Stage(aHash(t, "a"), a); err != nil {
		t.Fatal(err)
	}
	if err := store.Stage(aHash(t, "a"), a); err != nil {
		t.Fatal(err) // identical restage is a no-op
	}
	if err := store.Adopt(aHash(t, "a")); err != nil {
		t.Fatal(err)
	}
	cur, err := store.Current()
	if err != nil || cur != aHash(t, "a") {
		t.Fatalf("current = %q err=%v", cur, err)
	}
	if err := store.Adopt(aHash(t, "b")); err == nil {
		t.Fatal("adopting an unstaged artifact must fail")
	}

	// Three generations, keep 2: adopting c prunes a, keeps b and c.
	if err := store.Stage(aHash(t, "b"), b); err != nil {
		t.Fatal(err)
	}
	if err := store.Adopt(aHash(t, "b")); err != nil {
		t.Fatal(err)
	}
	if err := store.Stage(aHash(t, "c"), c); err != nil {
		t.Fatal(err)
	}
	if err := store.Adopt(aHash(t, "c")); err != nil {
		t.Fatal(err)
	}
	if err := store.Prune(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Staged(aHash(t, "a")); err == nil {
		t.Fatal("generation a must be pruned beyond the retention window")
	}
	for _, h := range []string{"b", "c"} {
		if _, err := store.Staged(aHash(t, h)); err != nil {
			t.Fatalf("generation %s must be retained: %v", h, err)
		}
	}
	cur, err = store.Current()
	if err != nil || cur != aHash(t, "c") {
		t.Fatalf("current after adoption = %q err=%v", cur, err)
	}
}

func TestNewArtifactStoreFloor(t *testing.T) {
	if _, err := NewArtifactStore(t.TempDir(), 1); err == nil {
		t.Fatal("keep_generations below the floor of 2 must be rejected")
	}
}

func TestAttemptSuccessorBootstrap(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)

	outcome, err := AttemptSuccessor(ctx, p, store, successorCfg(), ms(time.Now().UTC()))
	if err != nil {
		t.Fatal(err)
	}
	if outcome != SuccessorPublished {
		t.Fatalf("bootstrap outcome = %q", outcome)
	}
	pub, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || pub == nil {
		t.Fatalf("no publication after bootstrap: %v", err)
	}
	cur, err := store.Current()
	if err != nil || cur != hexDigest(pub.ArtifactHash) {
		t.Fatalf("store current %q != published %s (err %v)", cur, hexDigest(pub.ArtifactHash), err)
	}
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_retirements`); n != 0 {
		t.Fatalf("bootstrap must not retire anything, retired %d", n)
	}
}

func TestAttemptSuccessorUnchangedConfirmsFreshness(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	before, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || before == nil {
		t.Fatal(err)
	}

	// No evidence change: a later attempt confirms freshness, publishes
	// nothing and leaves the pointer alone.
	outcome, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC().Add(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if outcome != SuccessorConfirmed {
		t.Fatalf("unchanged outcome = %q", outcome)
	}
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_publications`); n != 1 {
		t.Fatalf("unchanged attempt must not publish, publications = %d", n)
	}
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_freshness_confirmations`); n != 1 {
		t.Fatalf("freshness confirmation missing, rows = %d", n)
	}
	cur, _ := store.Current()
	if cur != hexDigest(before.ArtifactHash) {
		t.Fatalf("pointer moved on an unchanged attempt: %s", cur)
	}
}

func TestAttemptSuccessorPublishesOnChange(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	first, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || first == nil {
		t.Fatal(err)
	}

	seedChangedObservation(t, p, "s1", 4.25, "v2")
	outcome, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC().Add(2*time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if outcome != SuccessorPublished {
		t.Fatalf("changed outcome = %q", outcome)
	}

	second, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || second == nil {
		t.Fatal(err)
	}
	if hexDigest(second.ArtifactHash) == hexDigest(first.ArtifactHash) {
		t.Fatal("changed evidence must produce a new artifact")
	}
	// The predecessor is retired under the successor's publication.
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_retirements
		WHERE artifact_hash = $1 AND successor_hash = $2`,
		first.ArtifactHash, second.ArtifactHash); n != 1 {
		t.Fatalf("predecessor not retired: %d rows", n)
	}
	cur, _ := store.Current()
	if cur != hexDigest(second.ArtifactHash) {
		t.Fatalf("pointer not adopted: %s", cur)
	}
	// Retention: both generations remain on disk.
	for _, h := range [][]byte{first.ArtifactHash, second.ArtifactHash} {
		if _, err := store.Staged(hexDigest(h)); err != nil {
			t.Fatalf("generation %s missing from store: %v", hexDigest(h), err)
		}
	}
	// The prior generation stays decodable and readable after retirement.
	b, err := store.Staged(hexDigest(first.ArtifactHash))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := routing.DecodeSnapshot(b); err != nil {
		t.Fatalf("retired generation no longer decodable: %v", err)
	}
}

func TestAttemptSuccessorRevocationExcludesSource(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	first, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || first == nil {
		t.Fatal(err)
	}

	// A newer denial row becomes the governing policy: the source drops out
	// of the successor input set and its offers vanish from the projection.
	now := time.Now().UTC()
	mustExec(t, p.Pool, `
		INSERT INTO source_policies (source_id,config_hash,research,extraction,redistribution,
		  terms_url,license_text,checked_at,effective_at,expires_at)
		VALUES ('s1',$1,'denied','denied','denied','https://t','license',$2,$3,$4)`,
		hash32([]byte("policy-denied")), now, now.Add(time.Minute), now.Add(48*time.Hour))

	outcome, err := AttemptSuccessor(ctx, p, store, cfg, ms(now.Add(2*time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if outcome != SuccessorPublished {
		t.Fatalf("revocation successor outcome = %q", outcome)
	}
	second, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || second == nil {
		t.Fatal(err)
	}
	var artifact []byte
	if err := p.Pool.QueryRow(ctx,
		`SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
		second.ArtifactHash).Scan(&artifact); err != nil {
		t.Fatal(err)
	}
	snap, err := routing.DecodeSnapshot(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(snap.Data().Offers); got != 0 {
		t.Fatalf("revoked source must leave an empty projection, got %d offers", got)
	}
	// The prior generation file stays readable (local use, no destruction).
	if _, err := store.Staged(hexDigest(first.ArtifactHash)); err != nil {
		t.Fatalf("prior generation destroyed on revocation: %v", err)
	}
}

func TestReconcileStoreRepairsMissingAndCorrupt(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	pub, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || pub == nil {
		t.Fatal(err)
	}
	target := hexDigest(pub.ArtifactHash)

	// Crash window between publication and materialization: file and
	// pointer both missing.
	os.Remove(store.artifactPath(target))
	os.Remove(store.currentPath())
	if err := ReconcileStore(ctx, p, store, "local"); err != nil {
		t.Fatal(err)
	}
	cur, err := store.Current()
	if err != nil || cur != target {
		t.Fatalf("reconcile did not restore pointer: %q err=%v", cur, err)
	}

	// Corrupt artifact bytes: reconcile re-stages from the database.
	if err := os.WriteFile(store.artifactPath(target), []byte("corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ReconcileStore(ctx, p, store, "local"); err != nil {
		t.Fatal(err)
	}
	staged, err := store.Staged(target)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(staged)
	if hex.EncodeToString(sum[:]) != target {
		t.Fatal("reconcile did not re-stage the published artifact")
	}
}

func TestRunSuccessorLoopPublishesOnEvent(t *testing.T) {
	p := mustPort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()
	cfg.PollIdle = 20 * time.Millisecond

	// No manual attempt: the loop must publish purely off the outbox
	// trigger event.
	seedObservationEvent(t, p, obs, "s1")

	done := make(chan error, 1)
	go func() { done <- RunSuccessorLoop(ctx, p, store, cfg) }()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		pub, err := latestPublication(ctx, p.Pool, "local")
		if err != nil {
			t.Fatal(err)
		}
		if pub != nil {
			if cur, err := store.Current(); err == nil && cur == hexDigest(pub.ArtifactHash) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	// Every event the run produced, trigger and non-trigger alike, ends up
	// receipted: poll until a scan returns nothing while the loop lives.
	verify := context.Background()
	drainDeadline := time.Now().Add(10 * time.Second)
	for {
		pub, err := latestPublication(verify, p.Pool, "local")
		if err != nil {
			t.Fatal(err)
		}
		if pub != nil {
			if cur, err := store.Current(); err == nil && cur == hexDigest(pub.ArtifactHash) {
				events, err := p.DeliveryScan(verify, "publisher", 10)
				if err != nil {
					t.Fatal(err)
				}
				if len(events) == 0 {
					break
				}
			}
		}
		if time.Now().After(drainDeadline) {
			t.Fatal("loop never published, adopted and drained")
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	<-done

	pub, err := latestPublication(verify, p.Pool, "local")
	if err != nil || pub == nil {
		t.Fatalf("loop never published: %v", err)
	}
	cur, err := store.Current()
	if err != nil || cur != hexDigest(pub.ArtifactHash) {
		t.Fatalf("loop published without adopting: %q err=%v", cur, err)
	}
}

func TestRetireSnapshotRequiresSuccessorPublication(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	first, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || first == nil {
		t.Fatal(err)
	}
	fake := hash32([]byte("not-a-real-successor"))
	// The successor hash exists as bytes but has no publication naming it.
	if err := p.RetireSnapshot(ctx, first.ArtifactHash, fake, "pub-does-not-exist"); err == nil {
		t.Fatal("retirement must require the publication to name the successor")
	}
	if err := p.RetireSnapshot(ctx, first.ArtifactHash, first.ArtifactHash, first.PublicationID); err == nil {
		t.Fatal("self-retirement must be rejected")
	}
}

// seedObservationEvent inserts one ObservationRecorded outbox event directly,
// the exact shape record_attempt emits for a fetch success.
func seedObservationEvent(t *testing.T, p *Ports, obsID, sourceID string) {
	t.Helper()
	payload := map[string]any{
		"observation_id":    obsID,
		"source_id":         sourceID,
		"resource_url":      "https://" + sourceID + "/models",
		"body_hash":         hexHash(hash32([]byte("event-body"))),
		"outcome":           "success",
		"collector_version": "cv1",
		"output_id":         "out-1",
	}
	tx, err := p.Pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if err := emitEvent(context.Background(), tx, obsID, "ObservationRecorded", 1,
		ms(time.Now().UTC()), payload); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// makeArtifact builds a deterministic distinct artifact-shaped blob keyed by
// tag; the store only requires hash-pinned bytes, not a valid snapshot.
func makeArtifact(t *testing.T, tag string) []byte {
	t.Helper()
	b := []byte("artifact-" + tag)
	return b
}

func aHash(t *testing.T, tag string) string {
	t.Helper()
	return hexDigest(hash32([]byte("artifact-" + tag)))
}

// TestSuccessorRetirementConvergesAfterCrash reproduces review blocker 1: a
// crash between publish_snapshot and retire_snapshot leaves the predecessor
// unretired. Redelivery re-derives from the new publication, sees unchanged
// content and would previously confirm freshness forever. The sweep must
// retire the orphan naming the current publication.
func TestSuccessorRetirementConvergesAfterCrash(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	first, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || first == nil {
		t.Fatal(err)
	}

	// Simulate the crash window: evidence changed and a successor was
	// compiled, validated and published, but the process died before
	// materializing the store and retiring the predecessor.
	seedChangedObservation(t, p, "s1", 9.5, "crash")
	attemptAt := ms(time.Now().UTC().Add(time.Minute))
	inputs, err := eligibleInputs(ctx, p.Pool, attemptAt, "local")
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs:                  inputs,
		Distribution:            "local",
		ProjectionPolicyVersion: first.PolicyVersion,
		CreatedAt:               attemptAt,
		KnowledgeCutoff:         attemptAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	validation, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash:     compiled.ArtifactHash,
		ValidatorVersion: cfg.ValidatorVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	if validation.Outcome != "accepted" {
		t.Fatalf("crash-window validation = %q", validation.Outcome)
	}
	if _, err := p.PublishSnapshot(ctx, compiled.ArtifactHash, "local"); err != nil {
		t.Fatal(err)
	}

	// Redelivery: same evidence, so content is unchanged against the
	// current publication. The attempt must confirm freshness AND converge
	// the store and the orphaned retirement.
	outcome, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC().Add(2*time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if outcome != SuccessorConfirmed {
		t.Fatalf("redelivery outcome = %q", outcome)
	}
	second, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || second == nil {
		t.Fatal(err)
	}
	if hexDigest(second.ArtifactHash) == hexDigest(first.ArtifactHash) {
		t.Fatal("crash-window publication missing")
	}
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_retirements
		WHERE artifact_hash = $1 AND successor_hash = $2`,
		first.ArtifactHash, second.ArtifactHash); n != 1 {
		t.Fatalf("orphaned retirement not converged: %d rows", n)
	}
	// The retirement names the successor's publication.
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_retirements r
		JOIN snapshot_publications s ON s.id = r.publication_id
		WHERE r.artifact_hash = $1 AND s.artifact_hash = $2`,
		first.ArtifactHash, second.ArtifactHash); n != 1 {
		t.Fatalf("retirement paired to the wrong publication: %d rows", n)
	}
	cur, err := store.Current()
	if err != nil || cur != hexDigest(second.ArtifactHash) {
		t.Fatalf("store not converged after crash: %q err=%v", cur, err)
	}
}

// TestSuccessorEventPayloadsCarryTrueHex pins review blocker 2: every payload
// field or locator that claims to be an artifact hash carries the artifact's
// own hex digest, not a hash of the digest.
func TestSuccessorEventPayloadsCarryTrueHex(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	first, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || first == nil {
		t.Fatal(err)
	}
	seedChangedObservation(t, p, "s1", 5.5, "hexcheck")
	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC().Add(time.Minute))); err != nil {
		t.Fatal(err)
	}
	second, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || second == nil {
		t.Fatal(err)
	}

	// Every payload field must be a digest of a real publication's
	// artifact, never a hash of that digest.
	digests := map[string]bool{
		hexDigest(first.ArtifactHash):  true,
		hexDigest(second.ArtifactHash): true,
	}
	rows, err := p.Pool.Query(ctx, `SELECT event_name, payload FROM outbox_events
		WHERE event_name IN ('SnapshotCompiled','SnapshotValidated','SnapshotPublished','SnapshotRetired')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := map[string]int{}
	for rows.Next() {
		var name string
		var payload []byte
		if err := rows.Scan(&name, &payload); err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(payload, &fields); err != nil {
			t.Fatalf("%s payload not JSON: %v", name, err)
		}
		if name == "SnapshotRetired" {
			got, _ := fields["retired_hash"].(string)
			if got != hexDigest(first.ArtifactHash) {
				t.Fatalf("SnapshotRetired retired_hash = %q, want predecessor %q", got, hexDigest(first.ArtifactHash))
			}
			succ, _ := fields["successor_hash"].(string)
			if succ != hexDigest(second.ArtifactHash) {
				t.Fatalf("SnapshotRetired successor_hash = %q, want %q", succ, hexDigest(second.ArtifactHash))
			}
			if _, ok := fields["artifact_hash"]; ok {
				t.Fatal("SnapshotRetired payload must use retired_hash, not artifact_hash")
			}
		} else {
			got, _ := fields["artifact_hash"].(string)
			if !digests[got] {
				t.Fatalf("%s artifact_hash = %q, want a publication digest (%s or %s)",
					name, got, hexDigest(first.ArtifactHash), hexDigest(second.ArtifactHash))
			}
			if name == "SnapshotCompiled" || name == "SnapshotPublished" {
				var manifest, coverage []byte
				if err := p.Pool.QueryRow(ctx,
					`SELECT input_manifest_hash, coverage_digest FROM snapshots WHERE artifact_hash = $1`,
					mustDecodeHex(t, got)).Scan(&manifest, &coverage); err != nil {
					t.Fatal(err)
				}
				mh, _ := fields["input_manifest_hash"].(string)
				if name == "SnapshotCompiled" && mh != hexDigest(manifest) {
					t.Fatalf("SnapshotCompiled input_manifest_hash = %q, want the column digest %q", mh, hexDigest(manifest))
				}
				cd, _ := fields["coverage_digest"].(string)
				if cd != hexDigest(coverage) {
					t.Fatalf("%s coverage_digest = %q, want the column digest %q", name, cd, hexDigest(coverage))
				}
			}
		}
		if name == "SnapshotPublished" {
			loc, _ := fields["artifact_locator"].(string)
			want := "snapshot://" + fields["artifact_hash"].(string)
			if loc != want {
				t.Fatalf("SnapshotPublished locator = %q, want %q", loc, want)
			}
		}
		seen[name]++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SnapshotCompiled", "SnapshotValidated", "SnapshotPublished", "SnapshotRetired"} {
		if seen[name] == 0 {
			t.Fatalf("no %s event found", name)
		}
	}
}

// TestProcessOneBatchDebounce pins the debounce contract: one scan batch
// containing several eligible triggers resolves exactly one successor
// attempt (one publication) and receipts every event in the batch.
func TestProcessOneBatchDebounce(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	for i := range 3 {
		seedObservationEvent(t, p, fmt.Sprintf("obs-multi-%d", i), "s1")
	}
	busy, err := processOneBatch(ctx, p, store, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !busy {
		t.Fatal("batch with events must report busy")
	}
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_publications`); n != 1 {
		t.Fatalf("debounce must produce exactly one publication, got %d", n)
	}
	// The attempt's own outbox events land in later batches; keep
	// processing until the consumer drains. Each pass must be a no-op for
	// publications: debounce holds at one publication for the burst.
	for range 20 {
		busy, err = processOneBatch(ctx, p, store, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRows(t, p, `SELECT count(*) FROM snapshot_publications`); n != 1 {
			t.Fatalf("drain pass republished: %d publications", n)
		}
		if !busy {
			break
		}
	}
	if busy {
		t.Fatal("consumer never drained")
	}
	events, err := p.DeliveryScan(ctx, "publisher", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("scan after drain returned %d events", len(events))
	}
}

// TestProcessOneBatchFatalOnMisconfiguration: a permanent invalid_request
// from the attempt (missing bootstrap policy on an empty channel) stops the
// batch processor loudly instead of hot-retrying forever.
func TestProcessOneBatchFatalOnMisconfiguration(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()
	cfg.BootstrapPolicy = "" // no publication exists to inherit from

	seedObservationEvent(t, p, obs, "s1")
	_, err := processOneBatch(ctx, p, store, cfg)
	if err == nil {
		t.Fatal("misconfigured attempt must be fatal")
	}
	var ce *Error
	if !errors.As(err, &ce) || ce.Code != CodeInvalidRequest {
		t.Fatalf("want invalid_request, got %v", err)
	}
}

// TestCurrentMalformedPointerReadsAbsent: a torn CURRENT pointer reads as
// absent so reconciliation re-adopts instead of bricking the store.
func TestCurrentMalformedPointerReadsAbsent(t *testing.T) {
	store := testStore(t, 2)
	if err := os.WriteFile(store.currentPath(), []byte("not-a-hash"), 0o644); err != nil {
		t.Fatal(err)
	}
	cur, err := store.Current()
	if err != nil {
		t.Fatalf("malformed CURRENT must not error: %v", err)
	}
	if cur != "" {
		t.Fatalf("malformed CURRENT must read as absent, got %q", cur)
	}
}

// TestSuccessorConfigFromDefaults pins the daemon mapping: defaults for
// channel, distribution, scan limit, idle and retention, and the validator
// version gating the loop.
func TestSuccessorConfigFromDefaults(t *testing.T) {
	cfg := CollectorConfig{
		SnapshotValidatorVersion: "vv1",
		SnapshotPolicy:           "pp-1",
	}
	sc, root, keep, ok, err := cfg.SuccessorConfigFrom()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("validator version set: loop must be enabled")
	}
	if sc.Channel != "local" || sc.BootstrapDistribution != "local" {
		t.Fatalf("channel defaults wrong: %+v", sc)
	}
	if sc.ValidatorVersion != "vv1" || sc.BootstrapPolicy != "pp-1" {
		t.Fatalf("versions not mapped: %+v", sc)
	}
	if sc.ScanLimit != 50 || sc.PollIdle != 5*time.Second {
		t.Fatalf("loop defaults wrong: %+v", sc)
	}
	if root == "" || keep != 2 {
		t.Fatalf("store defaults wrong: root=%q keep=%d", root, keep)
	}

	cfg.SnapshotValidatorVersion = ""
	if _, _, _, ok, err := cfg.SuccessorConfigFrom(); err != nil || ok {
		t.Fatalf("empty validator version must disable the loop (ok=%v err=%v)", ok, err)
	}
}

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestRetireSnapshotReplayConflict: a replay of the same (artifact,
// publication) pair naming a different successor is a conflict, not a
// silent second retirement.
func TestRetireSnapshotReplayConflict(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 2)
	cfg := successorCfg()

	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	first, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || first == nil {
		t.Fatal(err)
	}
	seedChangedObservation(t, p, "s1", 6.75, "replay")
	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC().Add(time.Minute))); err != nil {
		t.Fatal(err)
	}
	second, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || second == nil {
		t.Fatal(err)
	}

	// The legitimate retirement exists; replaying the same pair with a
	// different successor must conflict.
	other := hash32([]byte("replay-other-successor"))
	if err := p.RetireSnapshot(ctx, first.ArtifactHash, other, second.PublicationID); err == nil {
		t.Fatal("replay with a different successor must conflict")
	} else {
		var ce *Error
		if !errors.As(err, &ce) || ce.Code != CodeConflict {
			t.Fatalf("want conflict, got %v", err)
		}
	}
	// Idempotent replay of the exact same triple stays a no-op.
	if err := p.RetireSnapshot(ctx, first.ArtifactHash, second.ArtifactHash, second.PublicationID); err != nil {
		t.Fatalf("exact replay must be idempotent: %v", err)
	}
}

// TestAttemptSuccessorBlockedResolvesAndLogs pins the semantic-block
// contract: a local-distribution artifact cannot publish to a
// distributable channel, so the attempt resolves blocked, logs through
// the hook, publishes nothing and leaves the store empty.
func TestAttemptSuccessorBlockedResolvesAndLogs(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	now := ms(time.Now().UTC())
	if _, err := p.RegisterPublicationChannel(ctx, "dist", "distributable", now); err != nil {
		t.Fatal(err)
	}
	store := testStore(t, 2)
	cfg := successorCfg()
	cfg.Channel = "dist"
	var logs []string
	cfg.Log = func(format string, args ...any) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}

	outcome, err := AttemptSuccessor(ctx, p, store, cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != SuccessorBlocked {
		t.Fatalf("channel-class mismatch must resolve blocked, got %q", outcome)
	}
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_publications`); n != 0 {
		t.Fatalf("blocked attempt must not publish, got %d", n)
	}
	cur, err := store.Current()
	if err != nil || cur != "" {
		t.Fatalf("blocked attempt must leave the store empty: %q err=%v", cur, err)
	}
	blocked := false
	for _, l := range logs {
		if strings.Contains(l, "blocked") {
			blocked = true
		}
	}
	if !blocked {
		t.Fatalf("blocked outcome must be logged, logs = %v", logs)
	}
}

// TestSweepRetirementsConvergesMultipleOrphans: successive crash windows
// can orphan more than one predecessor; the sweep retires each naming the
// current publication.
func TestSweepRetirementsConvergesMultipleOrphans(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedSuccessorBase(t, p)
	obs := seedSnapshotBase(t, p, "s1", "allowed")
	seedSnapshotOffer(t, p, obs, "s1")
	store := testStore(t, 4)
	cfg := successorCfg()

	// P1 published and materialized normally.
	if _, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}
	first, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || first == nil {
		t.Fatal(err)
	}

	// Two successive crash windows: publish successors P2 and P3 without
	// materializing or retiring.
	at := ms(time.Now().UTC().Add(time.Minute))
	published := [][]byte{first.ArtifactHash}
	for i, amt := range []float64{7.5, 8.5} {
		seedChangedObservation(t, p, "s1", amt, fmt.Sprintf("multi-%d", i))
		at = ms(at.Add(time.Minute))
		inputs, err := eligibleInputs(ctx, p.Pool, at, "local")
		if err != nil {
			t.Fatal(err)
		}
		compiled, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
			Inputs:                  inputs,
			Distribution:            "local",
			ProjectionPolicyVersion: first.PolicyVersion,
			CreatedAt:               at,
			KnowledgeCutoff:         at,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
			ArtifactHash:     compiled.ArtifactHash,
			ValidatorVersion: cfg.ValidatorVersion,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := p.PublishSnapshot(ctx, compiled.ArtifactHash, "local"); err != nil {
			t.Fatal(err)
		}
		published = append(published, compiled.ArtifactHash)
	}
	third, err := latestPublication(ctx, p.Pool, "local")
	if err != nil || third == nil {
		t.Fatal(err)
	}

	// Redelivery converges: both orphans retire naming P3's publication.
	// The current publication itself never retires.
	outcome, err := AttemptSuccessor(ctx, p, store, cfg, ms(time.Now().UTC().Add(5*time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if outcome != SuccessorConfirmed {
		t.Fatalf("redelivery outcome = %q", outcome)
	}
	wantRetired := len(published) - 1 // every predecessor, not the current
	if n := countRows(t, p, `SELECT count(*) FROM snapshot_retirements r
		JOIN snapshot_publications s ON s.id = r.publication_id
		WHERE r.artifact_hash = ANY($1) AND s.artifact_hash = $2`,
		published, third.ArtifactHash); n != wantRetired {
		t.Fatalf("orphans not all retired under the current publication: %d of %d",
			n, wantRetired)
	}
	cur, err := store.Current()
	if err != nil || cur != hexDigest(third.ArtifactHash) {
		t.Fatalf("store not converged: %q err=%v", cur, err)
	}
}
