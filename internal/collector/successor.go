package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/guygrigsby/agent-select/routing"
)

// SuccessorConfig pins the publisher loop's channel identity and bootstrap
// parameters (delivery-storage V8). Inherited parameters (distribution,
// projection policy) come from the channel's current publication once one
// exists.
type SuccessorConfig struct {
	Channel               string
	ValidatorVersion      string
	BootstrapPolicy       string // projection policy version for the first generation
	BootstrapDistribution string // first slice: local
	ScanLimit             int    // delivery scan batch size
	PollIdle              time.Duration
	// Log receives blocked outcomes and non-fatal attempt failures. Nil
	// discards; the daemon wires it to stderr.
	Log func(format string, args ...any)
}

func (c SuccessorConfig) logf(format string, args ...any) {
	if c.Log != nil {
		c.Log(format, args...)
	}
}

// findingsSummary renders validation findings as one line for logs.
func findingsSummary(findings []ValidationFinding) string {
	parts := make([]string, 0, len(findings))
	for _, f := range findings {
		parts = append(parts, f.Code+": "+f.Detail)
	}
	if len(parts) == 0 {
		return "no findings"
	}
	return strings.Join(parts, "; ")
}

// resolveUnchanged completes the unchanged-content path: append the
// freshness confirmation for the current artifact, converge the local store
// and sweep orphaned retirements (a crash between an earlier publish and
// its retire converges here on redelivery).
func resolveUnchanged(ctx context.Context, p *Ports, store *ArtifactStore, cfg SuccessorConfig, prev *latestPublicationRow) error {
	if err := p.ConfirmFreshness(ctx, prev.ArtifactHash, cfg.Channel); err != nil {
		return err
	}
	return ReconcileStore(ctx, p, store, cfg.Channel)
}

// SuccessorOutcome classifies one resolved successor attempt. A storage or
// transport failure is an error, not an outcome; the caller records no
// receipts and the batch redelivers.
type SuccessorOutcome string

const (
	SuccessorPublished SuccessorOutcome = "published"
	SuccessorConfirmed SuccessorOutcome = "confirmed"
	SuccessorBlocked   SuccessorOutcome = "blocked"
)

// RetireSnapshot implements retire_snapshot: records the predecessor's
// retirement under the successor's publication and emits SnapshotRetired.
// Idempotent by (artifact_hash, publication_id).
func (p *Ports) RetireSnapshot(ctx context.Context, artifactHash, successorHash []byte, publicationID string) error {
	if len(artifactHash) != 32 || len(successorHash) != 32 {
		return errf(CodeInvalidRequest, "artifact and successor hashes must be 32 bytes")
	}
	if bytes.Equal(artifactHash, successorHash) {
		return errf(CodeInvalidRequest, "an artifact cannot retire itself")
	}
	if publicationID == "" {
		return errf(CodeInvalidRequest, "publication id is required")
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return storageErr(err)
	}
	defer tx.Rollback(ctx)
	var storedSuccessor []byte
	err = tx.QueryRow(ctx, `
		SELECT successor_hash FROM snapshot_retirements
		WHERE artifact_hash = $1 AND publication_id = $2`,
		artifactHash, publicationID).Scan(&storedSuccessor)
	if err == nil {
		if !bytes.Equal(storedSuccessor, successorHash) {
			return errf(CodeConflict,
				"retirement of %s under %s already names a different successor",
				hexDigest(artifactHash), publicationID)
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return storageErr(err)
	}
	var successorPublished bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM snapshot_publications
		 WHERE artifact_hash = $1 AND id = $2)`,
		successorHash, publicationID).Scan(&successorPublished); err != nil {
		return storageErr(err)
	}
	if !successorPublished {
		return errf(CodeInvalidRequest, "publication %s does not publish successor %s",
			publicationID, hexDigest(successorHash))
	}
	ct, err := tx.Exec(ctx, `
		INSERT INTO snapshot_retirements (artifact_hash,successor_hash,publication_id,retired_at)
		VALUES ($1,$2,$3,$4)`,
		artifactHash, successorHash, publicationID, time.Now().UTC())
	if err != nil {
		return storageErr(err)
	}
	if ct.RowsAffected() == 0 {
		return nil
	}
	if err := emitEvent(ctx, tx, hexHash(artifactHash), "SnapshotRetired", 1,
		time.Now().UTC(), map[string]any{
			"retired_hash":   hexDigest(artifactHash),
			"successor_hash": hexDigest(successorHash),
			"publication_id": publicationID,
		}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// sweepRetirements retires every older publication on the channel that
// has no retirement row, naming the current publication as successor. A
// crash between publish and retire leaves such an orphan; redelivery
// re-derives from the new current publication and would otherwise confirm
// freshness forever without completing the retirement. Called on every
// resolved attempt and at reconciliation.
func (p *Ports) sweepRetirements(ctx context.Context, channel string) error {
	latest, err := latestPublication(ctx, p.Pool, channel)
	if err != nil || latest == nil {
		return err
	}
	rows, err := p.Pool.Query(ctx, `
		SELECT p.artifact_hash FROM snapshot_publications p
		WHERE p.channel = $1 AND p.artifact_hash <> $2
		  AND NOT EXISTS (SELECT 1 FROM snapshot_retirements r
		                  WHERE r.artifact_hash = p.artifact_hash)
		ORDER BY p.published_at, p.id`, channel, latest.ArtifactHash)
	if err != nil {
		return storageErr(err)
	}
	defer rows.Close()
	var orphans [][]byte
	for rows.Next() {
		var h []byte
		if err := rows.Scan(&h); err != nil {
			return storageErr(err)
		}
		orphans = append(orphans, h)
	}
	if err := rows.Err(); err != nil {
		return storageErr(err)
	}
	for _, h := range orphans {
		if err := p.RetireSnapshot(ctx, h, latest.ArtifactHash, latest.PublicationID); err != nil {
			return err
		}
	}
	return nil
}

// ConfirmFreshness implements confirm_snapshot_freshness: append-only
// infrastructure fact that the channel's artifact was checked and unchanged.
func (p *Ports) ConfirmFreshness(ctx context.Context, artifactHash []byte, channel string) error {
	if len(artifactHash) != 32 {
		return errf(CodeInvalidRequest, "artifact hash must be 32 bytes")
	}
	if channel == "" {
		return errf(CodeInvalidRequest, "channel is required")
	}
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO snapshot_freshness_confirmations (artifact_hash,channel,confirmed_at)
		VALUES ($1,$2,now()) ON CONFLICT DO NOTHING`,
		artifactHash, channel)
	if err != nil {
		return storageErr(err)
	}
	return nil
}

type latestPublicationRow struct {
	PublicationID string
	ArtifactHash  []byte
	Distribution  string
	PolicyVersion string
}

// querier is the read surface AttemptSuccessor's queries need, satisfied
// by *pgxpool.Pool and pgx.Tx alike.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// latestPublication loads the channel's current publication: newest
// published_at wins, ties by publication id (the reconcile rule).
func latestPublication(ctx context.Context, q querier, channel string) (*latestPublicationRow, error) {
	row := latestPublicationRow{}
	err := q.QueryRow(ctx, `
		SELECT p.id, p.artifact_hash, s.distribution, s.projection_policy_version
		FROM snapshot_publications p
		JOIN snapshots s ON s.artifact_hash = p.artifact_hash
		WHERE p.channel = $1
		ORDER BY p.published_at DESC, p.id DESC
		LIMIT 1`, channel).Scan(&row.PublicationID, &row.ArtifactHash, &row.Distribution, &row.PolicyVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, storageErr(err)
	}
	return &row, nil
}

// eligibleInputs selects the successor's input set: per source with a
// governing policy allowing the distribution's rights, the latest success
// observation by (observed_at, id), pinning the governing config hash.
func eligibleInputs(ctx context.Context, pool *pgxpool.Pool, now time.Time, distribution string) ([]InputRef, error) {
	var sourceIDs []string
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT source_id FROM observations WHERE outcome = 'success' ORDER BY source_id`)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var src string
		if err := rows.Scan(&src); err != nil {
			return nil, storageErr(err)
		}
		sourceIDs = append(sourceIDs, src)
	}
	if err := rows.Err(); err != nil {
		return nil, storageErr(err)
	}

	inputs := make([]InputRef, 0, len(sourceIDs))
	for _, src := range sourceIDs {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return nil, storageErr(err)
		}
		gp, err := currentGoverningPolicy(ctx, tx, src, now)
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}
		if gp == nil || gp.Research != "allowed" || gp.Extraction != "allowed" {
			tx.Rollback(ctx)
			continue
		}
		if distribution == "distributable" && gp.Redistribution != "allowed" {
			tx.Rollback(ctx)
			continue
		}
		var obsID string
		err = tx.QueryRow(ctx, `
			SELECT id FROM observations
			WHERE source_id = $1 AND outcome = 'success'
			ORDER BY retrieved_at DESC, id DESC LIMIT 1`, src).Scan(&obsID)
		if errors.Is(err, pgx.ErrNoRows) {
			tx.Rollback(ctx)
			continue
		}
		if err != nil {
			tx.Rollback(ctx)
			return nil, storageErr(err)
		}
		if err := tx.Rollback(ctx); err != nil {
			return nil, storageErr(err)
		}
		inputs = append(inputs, InputRef{
			ObservationID: obsID,
			SourceID:      src,
			PolicyVersion: hexHash(gp.ConfigHash),
		})
	}
	return inputs, nil
}

// contentEqualMasks decodes both artifacts, replaces the attempt-time
// metadata members (created_at, knowledge_cutoff) with the Unix epoch
// sentinel and byte-compares the re-encoded canonical forms. Equal bytes
// mean the projection content is unchanged.
func contentEqual(current, successor []byte) (bool, error) {
	cur, err := routing.DecodeSnapshot(current)
	if err != nil {
		return false, errf(CodeInvalidEvidence, "current artifact fails the decode contract: %v", err)
	}
	suc, err := routing.DecodeSnapshot(successor)
	if err != nil {
		return false, errf(CodeInvalidEvidence, "successor artifact fails the decode contract: %v", err)
	}
	mask := time.Unix(0, 0).UTC()
	cd, sd := cur.Data(), suc.Data()
	cd.CreatedAt, cd.KnowledgeCutoff = mask, mask
	sd.CreatedAt, sd.KnowledgeCutoff = mask, mask
	cm, err1 := routing.NewSnapshot(cd)
	sm, err2 := routing.NewSnapshot(sd)
	if err1 != nil || err2 != nil {
		return false, errf(CodeInvalidEvidence, "masked artifact re-encode failed")
	}
	cb, err := routing.EncodeSnapshot(cm)
	if err != nil {
		return false, err
	}
	sb, err := routing.EncodeSnapshot(sm)
	if err != nil {
		return false, err
	}
	return bytes.Equal(cb, sb), nil
}

// AttemptSuccessor resolves one successor attempt for the channel:
// recompile from pinned observations and current policy, then confirm
// freshness, or validate + publish + materialize + adopt + retire.
// Semantic blocks (validation rejected, publish forbidden) resolve as
// SuccessorBlocked with the prior generation intact.
func AttemptSuccessor(ctx context.Context, p *Ports, store *ArtifactStore, cfg SuccessorConfig, now time.Time) (SuccessorOutcome, error) {
	if cfg.Channel == "" || cfg.ValidatorVersion == "" {
		return "", errf(CodeInvalidRequest, "successor channel and validator version are required")
	}
	attemptAt := now.UTC().Truncate(time.Millisecond)

	distribution := cfg.BootstrapDistribution
	policy := cfg.BootstrapPolicy
	if distribution == "" {
		distribution = "local"
	}
	prev, err := latestPublication(ctx, p.Pool, cfg.Channel)
	if err != nil {
		return "", err
	}
	if prev != nil {
		distribution, policy = prev.Distribution, prev.PolicyVersion
	}
	if policy == "" {
		return "", errf(CodeInvalidRequest, "no projection policy: bootstrap configuration is required before the first publication")
	}

	inputs, err := eligibleInputs(ctx, p.Pool, attemptAt, distribution)
	if err != nil {
		return "", err
	}
	compiled, err := p.CompileSnapshot(ctx, CompileSnapshotParams{
		Inputs:                  inputs,
		Distribution:            distribution,
		ProjectionPolicyVersion: policy,
		CreatedAt:               attemptAt,
		KnowledgeCutoff:         attemptAt,
	})
	if err != nil {
		return "", err
	}

	// Unchanged content: confirm freshness, no publication — but still
	// converge the store and retirements, which repairs a crash between a
	// previous publish and its retire and heals a stale local generation.
	if prev != nil && bytes.Equal(prev.ArtifactHash, compiled.ArtifactHash) {
		if err := resolveUnchanged(ctx, p, store, cfg, prev); err != nil {
			return "", err
		}
		return SuccessorConfirmed, nil
	}
	if prev != nil {
		var currentBytes, successorBytes []byte
		if err := p.Pool.QueryRow(ctx, `SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
			prev.ArtifactHash).Scan(&currentBytes); err != nil {
			return "", storageErr(err)
		}
		if err := p.Pool.QueryRow(ctx, `SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
			compiled.ArtifactHash).Scan(&successorBytes); err != nil {
			return "", storageErr(err)
		}
		same, err := contentEqual(currentBytes, successorBytes)
		if err != nil {
			return "", err
		}
		if same {
			if err := resolveUnchanged(ctx, p, store, cfg, prev); err != nil {
				return "", err
			}
			return SuccessorConfirmed, nil
		}
	}

	validation, err := p.ValidateSnapshot(ctx, ValidateSnapshotParams{
		ArtifactHash:     compiled.ArtifactHash,
		ValidatorVersion: cfg.ValidatorVersion,
	})
	if err != nil {
		return "", err
	}
	if validation.Outcome != "accepted" {
		cfg.logf("successor: validation rejected artifact %s: %s",
			hexDigest(compiled.ArtifactHash), findingsSummary(validation.Findings))
		if err := p.sweepRetirements(ctx, cfg.Channel); err != nil {
			return "", err
		}
		return SuccessorBlocked, nil
	}

	if _, err := p.PublishSnapshot(ctx, compiled.ArtifactHash, cfg.Channel); err != nil {
		var ce *Error
		if errors.As(err, &ce) && (ce.Code == CodeForbidden || ce.Code == CodeConflict || ce.Code == CodeSourceBlocked) {
			cfg.logf("successor: publish blocked artifact %s: %s",
				hexDigest(compiled.ArtifactHash), ce.Detail)
			if err := p.sweepRetirements(ctx, cfg.Channel); err != nil {
				return "", err
			}
			return SuccessorBlocked, nil
		}
		return "", err
	}

	// Materialize, adopt, prune, then converge retirements: the sweep
	// retires the predecessor and any older orphan left by a crash between
	// publish and retire. Publication rows are the durable intent;
	// reconciliation repairs a crash anywhere in these filesystem steps.
	if err := ReconcileStore(ctx, p, store, cfg.Channel); err != nil {
		return "", err
	}
	return SuccessorPublished, nil
}

// ReconcileStore stages and adopts the channel's latest published artifact,
// rebuilding a missing or corrupt file from snapshots.artifact_bytes, then
// converges retirements: the spec's convergence sentence binds every
// reconciliation, not only resolved attempts.
func ReconcileStore(ctx context.Context, p *Ports, store *ArtifactStore, channel string) error {
	pub, err := latestPublication(ctx, p.Pool, channel)
	if err != nil {
		return err
	}
	if pub == nil {
		return nil
	}
	var artifact []byte
	if err := p.Pool.QueryRow(ctx, `SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
		pub.ArtifactHash).Scan(&artifact); err != nil {
		return storageErr(err)
	}
	target := hexDigest(pub.ArtifactHash)
	needStage := true
	if staged, err := store.Staged(target); err == nil {
		if bytes.Equal(staged, artifact) {
			needStage = false
		}
	} else {
		var ce *Error
		if !errors.As(err, &ce) || ce.Code != CodeNotFound {
			return err
		}
	}
	if needStage {
		if err := store.Stage(target, artifact); err != nil {
			return err
		}
	}
	current, err := store.Current()
	if err != nil {
		return err
	}
	if current != target {
		if err := store.Adopt(target); err != nil {
			return err
		}
	}
	if err := store.Prune(); err != nil {
		return err
	}
	return p.sweepRetirements(ctx, channel)
}

// eligibleTrigger reports whether one outbox event should attempt a
// successor: ObservationRecorded with outcome success. The decode mirrors
// the event's full closed payload schema (strict decode rejects unknown
// members, so a partial struct would reject every event).
func eligibleTrigger(e OutboxEvent) bool {
	if e.EventName != "ObservationRecorded" {
		return false
	}
	var payload struct {
		ObservationID    string `json:"observation_id"`
		SourceID         string `json:"source_id"`
		ResourceURL      string `json:"resource_url"`
		BodyHash         string `json:"body_hash"`
		Outcome          string `json:"outcome"`
		CollectorVersion string `json:"collector_version"`
		OutputID         string `json:"output_id"`
	}
	if err := strictDecode(e.Payload, &payload); err != nil {
		return false
	}
	return payload.Outcome == "success"
}

// RunSuccessorLoop is the publisher service identity: reconcile the store,
// then consume the delivery outbox forever. One batch of triggers resolves
// one successor attempt; receipts land only after the attempt resolves, so
// storage failures redeliver.
func RunSuccessorLoop(ctx context.Context, p *Ports, store *ArtifactStore, cfg SuccessorConfig) error {
	if err := ReconcileStore(ctx, p, store, cfg.Channel); err != nil {
		return fmt.Errorf("successor: startup reconciliation: %w", err)
	}
	idle := cfg.PollIdle
	if idle <= 0 {
		idle = 5 * time.Second
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		busy, err := processOneBatch(ctx, p, store, cfg)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if busy {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(idle):
		}
	}
}

// processOneBatch scans and resolves one delivery batch. It reports busy
// false when the consumer is idle. Debounce: one batch containing any
// eligible trigger resolves exactly one successor attempt, then every
// event in the batch is receipted. An attempt failing on a storage error
// leaves receipts unwritten so the batch redelivers; a permanent
// invalid_request (misconfiguration) is fatal so the daemon stops loudly
// instead of hot-retrying forever.
func processOneBatch(ctx context.Context, p *Ports, store *ArtifactStore, cfg SuccessorConfig) (bool, error) {
	limit := cfg.ScanLimit
	if limit <= 0 {
		limit = 50
	}
	consumer := "publisher"
	events, err := p.DeliveryScan(ctx, consumer, limit)
	if err != nil {
		return false, err
	}
	if len(events) == 0 {
		return false, nil
	}
	busy := true
	attempt := false
	for _, e := range events {
		if eligibleTrigger(e) {
			attempt = true
			break
		}
	}
	if attempt {
		_, err := AttemptSuccessor(ctx, p, store, cfg, time.Now().UTC())
		if err != nil {
			var ce *Error
			if errors.As(err, &ce) && (ce.Code == CodeInvalidRequest || ce.Code == CodeInvalidEvidence) {
				return false, err // permanent misconfiguration or stored corruption: stop the loop
			}
			cfg.logf("successor: attempt failed, batch will redeliver: %v", err)
			// Leave receipts unwritten; idle false makes the loop back off
			// before rescanning the redelivered batch.
			return false, nil
		}
	}
	now := time.Now().UTC()
	for _, e := range events {
		if err := p.RecordConsumerReceipt(ctx, consumer, e.EventID, now); err != nil {
			if ctx.Err() != nil {
				return false, nil // shutdown: unreceipted events redeliver
			}
			return false, err
		}
	}
	return busy, nil
}
