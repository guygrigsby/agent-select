// Mutation ports of docs/specs/delivery-storage.md over the migrated Postgres
// authority. Every port writes its rows and outbox events in one transaction.
package collector

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ports is the storage port surface. Construct after Migrate.
type Ports struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Ports { return &Ports{Pool: pool} }

func newID(prefix string) string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err) // crypto/rand failure is not recoverable
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}

var stageTokens = map[string]bool{
	"fetch": true, "parse": true, "resolve": true, "compile": true,
}

// ---------------------------------------------------------------------------
// Collector coordinator
// ---------------------------------------------------------------------------

// EnqueueResult is the enqueue_job response: job ID plus whether an identical
// work tuple already existed.
type EnqueueResult struct {
	JobID    string
	Existing bool
}

// EnqueueJob implements enqueue_job. Idempotent by the complete work tuple.
// The source must exist and be enabled.
func (p *Ports) EnqueueJob(ctx context.Context, sourceID, resourceURL, stage string,
	inputHash []byte, parserVersion string, now time.Time) (EnqueueResult, error) {
	if !stageTokens[stage] {
		return EnqueueResult{}, errf(CodeInvalidRequest, "stage %q is not a stage token", stage)
	}
	if len(inputHash) != 32 {
		return EnqueueResult{}, errf(CodeInvalidRequest, "input hash must be 32 bytes")
	}
	if parserVersion == "" || resourceURL == "" || sourceID == "" {
		return EnqueueResult{}, errf(CodeInvalidRequest, "source, url and parser version are required")
	}
	var enabled bool
	err := p.Pool.QueryRow(ctx,
		`SELECT enabled FROM sources WHERE id = $1`, sourceID).Scan(&enabled)
	if err == pgx.ErrNoRows {
		return EnqueueResult{}, errf(CodeNotFound, "source %s", sourceID)
	}
	if err != nil {
		return EnqueueResult{}, storageErr(err)
	}
	if !enabled {
		return EnqueueResult{}, errf(CodeSourceBlocked, "source %s is disabled", sourceID)
	}
	jobID := newID("job")
	err = p.Pool.QueryRow(ctx, `
		INSERT INTO jobs (id,source_id,resource_url,stage,input_hash,parser_version,created_at,attempts,fencing_token)
		VALUES ($1,$2,$3,$4,$5,$6,$7,0,0)
		ON CONFLICT (source_id,resource_url,stage,input_hash,parser_version) DO NOTHING
		RETURNING id`,
		jobID, sourceID, resourceURL, stage, inputHash, parserVersion, now).Scan(&jobID)
	switch {
	case err == pgx.ErrNoRows:
		// Idempotent replay of the complete work tuple: adopt the stored job.
		if err := p.Pool.QueryRow(ctx, `
			SELECT id FROM jobs
			WHERE source_id = $1 AND resource_url = $2 AND stage = $3
			  AND input_hash = $4 AND parser_version = $5`,
			sourceID, resourceURL, stage, inputHash, parserVersion).Scan(&jobID); err != nil {
			return EnqueueResult{}, storageErr(err)
		}
		return EnqueueResult{JobID: jobID, Existing: true}, nil
	case err != nil:
		return EnqueueResult{}, storageErr(err)
	}
	return EnqueueResult{JobID: jobID}, nil
}

// LeaseResult is the lease_job response.
type LeaseResult struct {
	Token     int64
	ExpiresAt time.Time
}

// LeaseJob implements lease_job with job-row fencing: the token is minted
// under the row lock, retries from the same worker return the same live lease
// and any other concurrent lease is a conflict.
func (p *Ports) LeaseJob(ctx context.Context, jobID, workerID string,
	duration time.Duration, now time.Time) (LeaseResult, error) {
	if workerID == "" {
		return LeaseResult{}, errf(CodeInvalidRequest, "worker id is required")
	}
	if duration <= 0 {
		return LeaseResult{}, errf(CodeInvalidRequest, "lease duration must be positive")
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return LeaseResult{}, storageErr(err)
	}
	defer tx.Rollback(ctx)
	var fencingToken int64
	err = tx.QueryRow(ctx,
		`SELECT fencing_token FROM jobs WHERE id = $1 FOR UPDATE`, jobID).Scan(&fencingToken)
	if err == pgx.ErrNoRows {
		return LeaseResult{}, errf(CodeNotFound, "job %s", jobID)
	}
	if err != nil {
		return LeaseResult{}, storageErr(err)
	}
	if done, err := jobCompleted(ctx, tx, jobID); err != nil {
		return LeaseResult{}, err
	} else if done {
		return LeaseResult{}, errf(CodeConflict, "job %s is completed", jobID)
	}
	latest, err := latestAttempt(ctx, tx, jobID)
	if err != nil {
		return LeaseResult{}, err
	}
	if latest.outcome == "quarantined" || latest.outcome == "failed_terminal" {
		return LeaseResult{}, errf(CodeConflict, "job %s is %s; reentry requires a new work tuple", jobID, latest.outcome)
	}
	var leaseWorker string
	var leaseExpires time.Time
	var leaseToken int64
	err = tx.QueryRow(ctx, `
		SELECT token, worker_id, expires_at FROM job_leases
		WHERE job_id = $1 ORDER BY token DESC LIMIT 1`, jobID).
		Scan(&leaseToken, &leaseWorker, &leaseExpires)
	switch {
	case err == pgx.ErrNoRows:
		// fresh lease below
	case err != nil:
		return LeaseResult{}, storageErr(err)
	case leaseExpires.After(now):
		if leaseWorker != workerID {
			return LeaseResult{}, errf(CodeConflict, "job %s holds an active lease for another worker", jobID)
		}
		return LeaseResult{Token: leaseToken, ExpiresAt: leaseExpires}, nil
	}
	if latest.outcome == "deferred" && latest.retryNotBefore.After(now) {
		return LeaseResult{}, errf(CodeConflict, "job %s is deferred until %s", jobID, latest.retryNotBefore.UTC())
	}
	token := fencingToken + 1
	expires := now.Add(duration)
	if _, err := tx.Exec(ctx, `
		UPDATE jobs SET fencing_token = $2, attempts = attempts + 1 WHERE id = $1`,
		jobID, token); err != nil {
		return LeaseResult{}, storageErr(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO job_leases (job_id,token,worker_id,leased_at,expires_at)
		VALUES ($1,$2,$3,$4,$5)`, jobID, token, workerID, now, expires); err != nil {
		return LeaseResult{}, storageErr(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return LeaseResult{}, storageErr(err)
	}
	return LeaseResult{Token: token, ExpiresAt: expires}, nil
}

// RecoverJob implements recover_job: the coordinator verifies the named lease
// has expired, marks the recovery durably, advances the attempt counter once
// and returns the derived state. Idempotent by expired token: the first
// recovery sets recovered_at; replays derive state without mutation. No
// attempt row and no domain event: nothing was observed.
func (p *Ports) RecoverJob(ctx context.Context, jobID string, expiredToken int64,
	now time.Time) (string, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return "", storageErr(err)
	}
	defer tx.Rollback(ctx)
	var expiresAt time.Time
	err = tx.QueryRow(ctx,
		`SELECT expires_at FROM job_leases WHERE job_id = $1 AND token = $2`,
		jobID, expiredToken).Scan(&expiresAt)
	if err == pgx.ErrNoRows {
		return "", errf(CodeNotFound, "lease %d of job %s", expiredToken, jobID)
	}
	if err != nil {
		return "", storageErr(err)
	}
	if expiresAt.After(now) {
		return "", errf(CodeStaleLease, "lease %d of job %s has not expired", expiredToken, jobID)
	}
	tag, err := tx.Exec(ctx, `
		UPDATE job_leases SET recovered_at = $3
		WHERE job_id = $1 AND token = $2 AND recovered_at IS NULL`,
		jobID, expiredToken, now)
	if err != nil {
		return "", storageErr(err)
	}
	if tag.RowsAffected() == 1 {
		// First recovery under this token: count the take-up once.
		if _, err := tx.Exec(ctx,
			`UPDATE jobs SET attempts = attempts + 1 WHERE id = $1`, jobID); err != nil {
			return "", storageErr(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", storageErr(err)
	}
	return p.JobState(ctx, jobID, now)
}

// ReenterJob implements reenter_job: a quarantined or terminal job re-enters
// exclusively through a changed input hash or parser version, which is a new
// work tuple and therefore a new job row.
func (p *Ports) ReenterJob(ctx context.Context, jobID string, newInputHash []byte,
	parserVersion string, now time.Time) (EnqueueResult, error) {
	var sourceID, resourceURL, stage string
	var oldHash []byte
	var oldParser string
	err := p.Pool.QueryRow(ctx, `
		SELECT source_id,resource_url,stage,input_hash,parser_version FROM jobs WHERE id = $1`,
		jobID).Scan(&sourceID, &resourceURL, &stage, &oldHash, &oldParser)
	if err == pgx.ErrNoRows {
		return EnqueueResult{}, errf(CodeNotFound, "job %s", jobID)
	}
	if err != nil {
		return EnqueueResult{}, storageErr(err)
	}
	if string(newInputHash) == string(oldHash) && parserVersion == oldParser {
		return EnqueueResult{}, errf(CodeConflict, "reentry requires a changed input hash or parser version")
	}
	return p.EnqueueJob(ctx, sourceID, resourceURL, stage, newInputHash, parserVersion, now)
}

// JobState derives the job state per delivery-storage.md, evaluated on the
// locked row.
func (p *Ports) JobState(ctx context.Context, jobID string, now time.Time) (string, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return "", storageErr(err)
	}
	defer tx.Rollback(ctx)
	var dummy int
	err = tx.QueryRow(ctx, `SELECT 1 FROM jobs WHERE id = $1 FOR UPDATE`, jobID).Scan(&dummy)
	if err == pgx.ErrNoRows {
		return "", errf(CodeNotFound, "job %s", jobID)
	}
	if err != nil {
		return "", storageErr(err)
	}
	state, err := deriveJobState(ctx, tx, jobID, now)
	if err != nil {
		return "", err
	}
	return state, tx.Commit(ctx)
}

func deriveJobState(ctx context.Context, tx pgx.Tx, jobID string, now time.Time) (string, error) {
	if done, err := jobCompleted(ctx, tx, jobID); err != nil {
		return "", err
	} else if done {
		return "completed", nil
	}
	latest, err := latestAttempt(ctx, tx, jobID)
	if err != nil {
		return "", err
	}
	if latest.outcome == "quarantined" || latest.outcome == "failed_terminal" {
		return "quarantined", nil
	}
	if latest.outcome == "deferred" && latest.retryNotBefore.After(now) {
		return "deferred", nil
	}
	var expiresAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT expires_at FROM job_leases WHERE job_id = $1 ORDER BY token DESC LIMIT 1`,
		jobID).Scan(&expiresAt)
	if err == pgx.ErrNoRows {
		return "queued", nil
	}
	if err != nil {
		return "", storageErr(err)
	}
	if expiresAt.Before(now) || expiresAt.Equal(now) {
		return "queued", nil
	}
	return "leased", nil
}

type attemptFacts struct {
	outcome        string
	retryNotBefore time.Time
}

func latestAttempt(ctx context.Context, tx pgx.Tx, jobID string) (attemptFacts, error) {
	var f attemptFacts
	err := tx.QueryRow(ctx, `
		SELECT outcome, retry_not_before FROM job_attempts
		WHERE job_id = $1 ORDER BY observed_at DESC, token DESC LIMIT 1`, jobID).
		Scan(&f.outcome, &f.retryNotBefore)
	if err == pgx.ErrNoRows {
		return attemptFacts{}, nil
	}
	if err != nil {
		return attemptFacts{}, storageErr(err)
	}
	return f, nil
}

func jobCompleted(ctx context.Context, tx pgx.Tx, jobID string) (bool, error) {
	var one int
	err := tx.QueryRow(ctx,
		`SELECT 1 FROM job_completions WHERE job_id = $1`, jobID).Scan(&one)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, storageErr(err)
	}
	return true, nil
}

// ---------------------------------------------------------------------------
// Collector worker
// ---------------------------------------------------------------------------

// AttemptObservation is the fetch-stage observation input. The worker names
// the observation ID because the stage envelope references it.
type AttemptObservation struct {
	ObservationID      string
	ResourceURL        string
	RetrievedAt        time.Time
	HTTPStatus         int
	Outcome            string
	BodyBytes          []byte
	HeadersBytes       []byte
	PriorObservationID string // required when Outcome is unchanged
}

// CapabilityAssertion is the capability child payload.
type CapabilityAssertion struct {
	Capability, Support, Operation, Direction string
}

// ClaimAssertion is the claim child payload.
type ClaimAssertion struct {
	Dimension, Polarity, ClaimText string
}

// LimitAssertion is the limit child payload.
type LimitAssertion struct {
	Name      string
	Value     int64
	Unit      string
	Bound     string
	Scope     string
	Operation string
	Role      string
}

// ImageFormatAssertion is the accepted-format child payload.
type ImageFormatAssertion struct {
	Operation, Role, Format string
}

// ImageSemanticAssertion is the mask-semantics child payload. Operation
// image_editing and role mask are fixed by the schema.
type ImageSemanticAssertion struct {
	Semantic string
}

// ImageRequirementAssertion is the requirement child payload.
type ImageRequirementAssertion struct {
	Operation, Role, Requirement string
}

// AssertionInput is one typed assertion with exactly one child set.
type AssertionInput struct {
	SubjectSourceID   string
	SubjectUpstreamID string
	Locator           string
	Verbatim          []byte
	EvidenceClass     string
	ParserVersion     string

	Capability       *CapabilityAssertion
	Claim            *ClaimAssertion
	Limit            *LimitAssertion
	ImageFormat      *ImageFormatAssertion
	ImageSemantic    *ImageSemanticAssertion
	ImageRequirement *ImageRequirementAssertion
}

// CoverageFinding is one coverage_findings row.
type CoverageFinding struct {
	SubjectUpstreamID string
	Channel           string
	Outcome           string
	Detail            string
}

// RecordAttemptParams is the record_attempt input.
type RecordAttemptParams struct {
	JobID            string
	Token            int64
	CollectorVersion string
	RecordedAt       time.Time
	Observation      *AttemptObservation // required for fetch, absent otherwise
	Assertions       []AssertionInput
	CoverageFindings []CoverageFinding
	OutputBytes      []byte
}

// RecordAttemptResult is the record_attempt response.
type RecordAttemptResult struct {
	ObservationID         string
	AcceptedAssertions    int
	QuarantinedAssertions int
	OutputID              string
	OutputHash            []byte
}

// fetchEnvelope mirrors the closed fetch stage schema.
type fetchEnvelope struct {
	Stage         string `json:"stage"`
	ObservationID string `json:"observation_id"`
	ResourceURL   string `json:"resource_url"`
	BodyHash      string `json:"body_hash"`
	HTTPStatus    int64  `json:"http_status"`
	HeadersHash   string `json:"headers_hash"`
}

// compileEnvelope mirrors the closed compile stage schema.
type compileEnvelope struct {
	Stage        string `json:"stage"`
	ArtifactHash string `json:"artifact_hash"`
}

// RecordAttempt implements record_attempt: the durable-success path. Every
// invocation produces one typed output envelope; failed fetches go through
// DeferJob. Idempotent by work tuple plus output hash, independent of the
// retried lease token.
func (p *Ports) RecordAttempt(ctx context.Context, params RecordAttemptParams) (RecordAttemptResult, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return RecordAttemptResult{}, storageErr(err)
	}
	defer tx.Rollback(ctx)

	var stage, sourceID, resourceURL, parserVersion string
	var inputHash []byte
	err = tx.QueryRow(ctx, `
		SELECT stage, source_id, resource_url, input_hash, parser_version
		FROM jobs WHERE id = $1 FOR UPDATE`, params.JobID).
		Scan(&stage, &sourceID, &resourceURL, &inputHash, &parserVersion)
	if err == pgx.ErrNoRows {
		return RecordAttemptResult{}, errf(CodeNotFound, "job %s", params.JobID)
	}
	if err != nil {
		return RecordAttemptResult{}, storageErr(err)
	}

	if len(params.OutputBytes) == 0 || !isCanonical(params.OutputBytes) {
		return RecordAttemptResult{}, errf(CodeInvalidEvidence, "stage output bytes must be canonical-v1")
	}
	outputHash := canonicalHash(params.OutputBytes)
	outputID := "out-" + params.JobID

	// Idempotent replay: an identical output hash for this job adopts the
	// stored row regardless of lease state; a different hash conflicts.
	var storedHash []byte
	err = tx.QueryRow(ctx,
		`SELECT output_hash FROM stage_outputs WHERE job_id = $1`, params.JobID).Scan(&storedHash)
	switch {
	case err == nil:
		if string(storedHash) != string(outputHash) {
			return RecordAttemptResult{}, errf(CodeConflict, "job %s already has a different output", params.JobID)
		}
		var obsID *string
		if err := tx.QueryRow(ctx,
			`SELECT observation_id FROM stage_output_links WHERE job_id = $1`,
			params.JobID).Scan(&obsID); err != nil {
			return RecordAttemptResult{}, storageErr(err)
		}
		result := RecordAttemptResult{OutputID: outputID, OutputHash: outputHash}
		if obsID != nil {
			result.ObservationID = *obsID
		}
		return result, nil
	case err != pgx.ErrNoRows:
		return RecordAttemptResult{}, storageErr(err)
	}

	// An active lease for this token is required for new work.
	if err := requireActiveLease(ctx, tx, params.JobID, params.Token, params.RecordedAt); err != nil {
		return RecordAttemptResult{}, err
	}

	observationID := ""
	switch stage {
	case "fetch":
		if params.Observation == nil {
			return RecordAttemptResult{}, errf(CodeInvalidRequest, "fetch requires an observation")
		}
		obs := params.Observation
		var env fetchEnvelope
		if err := strictDecode(params.OutputBytes, &env); err != nil {
			return RecordAttemptResult{}, err
		}
		if env.Stage != "fetch" {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence, "envelope stage %q does not match job stage", env.Stage)
		}
		if env.ObservationID != obs.ObservationID || env.ResourceURL != obs.ResourceURL ||
			env.HTTPStatus != int64(obs.HTTPStatus) {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence, "fetch envelope does not match the observation")
		}
		if env.BodyHash != hexHash(obs.BodyBytes) || env.HeadersHash != hexHash(obs.HeadersBytes) {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence, "fetch envelope hashes do not match the observation bytes")
		}
		if obs.ObservationID == "" || obs.Outcome == "" || params.CollectorVersion == "" {
			return RecordAttemptResult{}, errf(CodeInvalidRequest, "observation id, outcome and collector version are required")
		}
		// Remote fetch success requires a real 200 exchange. Status 0 is the
		// host_local receipt marker and never legal from a fetch port.
		if obs.Outcome == "success" && obs.HTTPStatus != 200 {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence,
				"fetch success requires http status 200, got %d", obs.HTTPStatus)
		}
		if obs.Outcome == "unchanged" && obs.PriorObservationID == "" {
			return RecordAttemptResult{}, errf(CodeInvalidRequest, "unchanged requires a prior observation")
		}
		bodyHash := canonicalHash(obs.BodyBytes)
		if _, err := tx.Exec(ctx,
			`INSERT INTO content_blobs (hash,bytes) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			bodyHash, obs.BodyBytes); err != nil {
			return RecordAttemptResult{}, storageErr(err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO observations (id,source_id,resource_url,retrieved_at,http_status,outcome,body_hash,headers_bytes,collector_version)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			obs.ObservationID, sourceID, obs.ResourceURL, obs.RetrievedAt, obs.HTTPStatus,
			obs.Outcome, bodyHash, obs.HeadersBytes, params.CollectorVersion); err != nil {
			// A CHECK rejection here is evidence-shaped caller data (outcome/
			// status pairing), not a transient storage fault.
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23514" {
				return RecordAttemptResult{}, errf(CodeInvalidEvidence,
					"observation outcome %q with status %d is not a legal pairing", obs.Outcome, obs.HTTPStatus)
			}
			return RecordAttemptResult{}, storageErr(err)
		}
		if obs.Outcome == "unchanged" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO unchanged_observations (observation_id,prior_observation_id)
				VALUES ($1,$2)`, obs.ObservationID, obs.PriorObservationID); err != nil {
				return RecordAttemptResult{}, storageErr(err)
			}
		}
		observationID = obs.ObservationID
	case "parse", "resolve":
		if params.Observation != nil {
			return RecordAttemptResult{}, errf(CodeInvalidRequest, "%s creates no observation", stage)
		}
		var env struct {
			Stage         string   `json:"stage"`
			ObservationID string   `json:"observation_id"`
			ItemIDs       []string `json:"assertion_ids"`
			FindingIDs    []string `json:"finding_ids"`
		}
		if err := strictDecode(params.OutputBytes, &env); err != nil {
			return RecordAttemptResult{}, err
		}
		if env.Stage != stage {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence, "envelope stage %q does not match job stage", env.Stage)
		}
		var one int
		err := tx.QueryRow(ctx, `SELECT 1 FROM observations WHERE id = $1`,
			env.ObservationID).Scan(&one)
		if err == pgx.ErrNoRows {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence, "contributing observation %s does not exist", env.ObservationID)
		}
		if err != nil {
			return RecordAttemptResult{}, storageErr(err)
		}
		if stage == "parse" {
			if len(env.ItemIDs) != len(params.Assertions) {
				return RecordAttemptResult{}, errf(CodeInvalidEvidence, "parse envelope assertion ids do not match submitted assertions")
			}
			for i, id := range env.ItemIDs {
				if id != assertionID(params.JobID, i) {
					return RecordAttemptResult{}, errf(CodeInvalidEvidence, "parse envelope assertion id %d is not the submitted assertion", i)
				}
			}
		}
		if stage == "resolve" && len(env.FindingIDs) != len(params.CoverageFindings) {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence, "resolve envelope finding ids do not match submitted findings")
		}
		for i, f := range params.CoverageFindings {
			if stage == "resolve" && f.Outcome != "researched" && f.Outcome != "identity_ambiguous" {
				return RecordAttemptResult{}, errf(CodeInvalidEvidence, "resolve findings carry %s or identity_ambiguous only", "researched")
			}
			want := f.SubjectUpstreamID + "|" + f.Channel
			if stage == "resolve" && env.FindingIDs[i] != want {
				return RecordAttemptResult{}, errf(CodeInvalidEvidence, "finding id %d does not match the submitted finding", i)
			}
		}
		observationID = env.ObservationID
	case "compile":
		if params.Observation != nil {
			return RecordAttemptResult{}, errf(CodeInvalidRequest, "compile creates no observation")
		}
		var env compileEnvelope
		if err := strictDecode(params.OutputBytes, &env); err != nil {
			return RecordAttemptResult{}, err
		}
		if env.Stage != "compile" {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence, "envelope stage %q does not match job stage", env.Stage)
		}
		if _, err := decodeHash32(env.ArtifactHash); err != nil {
			return RecordAttemptResult{}, err
		}
	default:
		return RecordAttemptResult{}, errf(CodeInvalidRequest, "unknown stage %q", stage)
	}

	// Assertions: per-assertion validation failures quarantine (skip and
	// count); parser version mismatch is invalid_evidence for the whole call.
	accepted := 0
	quarantined := 0
	for i, a := range params.Assertions {
		if a.ParserVersion != parserVersion {
			return RecordAttemptResult{}, errf(CodeInvalidEvidence,
				"assertion %d parser version %q does not match job parser version", i, a.ParserVersion)
		}
		if !structurallyValidAssertion(a, observationID) {
			quarantined++
			continue
		}
		ok, err := insertAssertion(ctx, tx, params.JobID, i, observationID, a)
		if err != nil {
			return RecordAttemptResult{}, err
		}
		if ok {
			accepted++
		} else {
			quarantined++
		}
	}
	for _, f := range params.CoverageFindings {
		if _, err := tx.Exec(ctx, `
			INSERT INTO coverage_findings (observation_id,subject_upstream_id,channel,outcome,detail)
			VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`,
			observationID, f.SubjectUpstreamID, f.Channel, f.Outcome, f.Detail); err != nil {
			return RecordAttemptResult{}, storageErr(err)
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO stage_outputs (job_id,output_id,output_hash,output_bytes,recorded_at)
		VALUES ($1,$2,$3,$4,$5)`,
		params.JobID, outputID, outputHash, params.OutputBytes, params.RecordedAt); err != nil {
		return RecordAttemptResult{}, storageErr(err)
	}
	var linkObs *string
	if observationID != "" {
		s := observationID
		linkObs = &s
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO stage_output_links (observation_id,job_id,source_id,resource_url,stage,input_hash,parser_version,output_id,output_hash,recorded_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		linkObs, params.JobID, sourceID, resourceURL, stage, inputHash, parserVersion,
		outputID, outputHash, params.RecordedAt); err != nil {
		return RecordAttemptResult{}, storageErr(err)
	}

	// StageOutputRecorded first, then ObservationRecorded for fetch only, in
	// one transaction.
	stagePayload := map[string]any{
		"job_id":         params.JobID,
		"source_id":      sourceID,
		"resource_url":   resourceURL,
		"stage":          stage,
		"input_hash":     hex.EncodeToString(inputHash),
		"parser_version": parserVersion,
		"output_id":      outputID,
		"output_hash":    hex.EncodeToString(outputHash),
	}
	if observationID != "" {
		stagePayload["observation_id"] = observationID
	}
	if err := emitEvent(ctx, tx, firstNonEmpty(observationID, params.JobID),
		"StageOutputRecorded", 1, params.RecordedAt, stagePayload); err != nil {
		return RecordAttemptResult{}, err
	}
	if stage == "fetch" {
		if err := emitEvent(ctx, tx, observationID, "ObservationRecorded", 1,
			params.RecordedAt, map[string]any{
				"observation_id":    observationID,
				"source_id":         sourceID,
				"resource_url":      params.Observation.ResourceURL,
				"body_hash":         hexHash(params.Observation.BodyBytes),
				"outcome":           params.Observation.Outcome,
				"collector_version": params.CollectorVersion,
				"output_id":         outputID,
			}); err != nil {
			return RecordAttemptResult{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return RecordAttemptResult{}, storageErr(err)
	}
	return RecordAttemptResult{
		ObservationID:         observationID,
		AcceptedAssertions:    accepted,
		QuarantinedAssertions: quarantined,
		OutputID:              outputID,
		OutputHash:            outputHash,
	}, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func requireActiveLease(ctx context.Context, tx pgx.Tx, jobID string, token int64, now time.Time) error {
	var expiresAt time.Time
	err := tx.QueryRow(ctx,
		`SELECT expires_at FROM job_leases WHERE job_id = $1 AND token = $2`,
		jobID, token).Scan(&expiresAt)
	if err == pgx.ErrNoRows {
		return errf(CodeStaleLease, "no lease %d for job %s", token, jobID)
	}
	if err != nil {
		return storageErr(err)
	}
	if !expiresAt.After(now) {
		return errf(CodeStaleLease, "lease %d of job %s expired at %s", token, jobID, expiresAt.UTC())
	}
	return nil
}

// countChildren counts the set child payloads. Typed nil pointers must be
// compared per field: boxing them in any would make nil interfaces non-nil.
func countChildren(a AssertionInput) int {
	n := 0
	for _, set := range []bool{
		a.Capability != nil, a.Claim != nil, a.Limit != nil,
		a.ImageFormat != nil, a.ImageSemantic != nil, a.ImageRequirement != nil,
	} {
		if set {
			n++
		}
	}
	return n
}

func inTokens(v string, tokens ...string) bool {
	for _, t := range tokens {
		if v == t {
			return true
		}
	}
	return false
}

// structurallyValidAssertion performs the port-level classification: a
// structurally invalid assertion is quarantined (skipped and counted), never
// fatal. Row-CHECK-level combination violations are classified as quarantine
// by insertAssertion.
func structurallyValidAssertion(a AssertionInput, observationID string) bool {
	switch a.EvidenceClass {
	case "provider_reported", "third_party_reported", "competitor_reported", "host_observed":
	default:
		return false
	}
	if a.SubjectSourceID == "" || a.SubjectUpstreamID == "" || a.Locator == "" ||
		len(a.Verbatim) == 0 || a.ParserVersion == "" || observationID == "" {
		return false
	}
	if countChildren(a) != 1 {
		return false
	}
	operations := []string{"text_assistance", "vision_understanding", "image_generation", "image_editing"}
	switch {
	case a.Capability != nil:
		c := a.Capability
		return c.Capability != "" &&
			inTokens(c.Support, "supported", "unsupported", "unknown") &&
			inTokens(c.Operation, operations...) &&
			(c.Capability == "tool_use" || inTokens(c.Direction, "input", "output", "not_applicable"))
	case a.Claim != nil:
		c := a.Claim
		return c.ClaimText != "" && inTokens(c.Polarity, "strength", "weakness", "operational_limit") &&
			inTokens(c.Dimension, "general", "code", "review", "reasoning", "extraction",
				"writing", "translation", "vision", "image_generation", "image_editing")
	case a.Limit != nil:
		l := a.Limit
		return l.Value >= 0 &&
			inTokens(l.Name, "context_tokens", "input_tokens", "output_tokens", "image_count",
				"reference_image_count", "mask_count", "image_width", "image_height",
				"image_pixels", "image_bytes") &&
			inTokens(l.Unit, "token", "image", "pixel", "pixel_area", "byte") &&
			inTokens(l.Bound, "minimum", "maximum", "exact") &&
			inTokens(l.Scope, "input", "output", "combined") &&
			inTokens(l.Operation, operations...) &&
			inTokens(l.Role, "request", "input_image", "output_image", "reference", "mask")
	case a.ImageFormat != nil:
		f := a.ImageFormat
		return f.Format != "" && inTokens(f.Operation, operations...) && f.Operation != "text_assistance" &&
			inTokens(f.Role, "request", "input_image", "output_image", "reference", "mask")
	case a.ImageSemantic != nil:
		return inTokens(a.ImageSemantic.Semantic,
			"alpha_zero_edits", "alpha_one_edits", "white_edits", "black_edits")
	case a.ImageRequirement != nil:
		r := a.ImageRequirement
		return inTokens(r.Operation, operations...) &&
			inTokens(r.Role, "reference", "mask") &&
			inTokens(r.Requirement, "required", "optional", "forbidden", "unknown") &&
			(r.Role != "mask" || r.Operation == "image_editing")
	}
	return false
}

// assertionID derives the deterministic assertion identity of ordinal i of a
// job, so parse envelopes can reference assertions before the port runs.
func assertionID(jobID string, ordinal int) string {
	return fmt.Sprintf("as-%s-%04d", jobID, ordinal)
}

// insertAssertion stores one assertion and its child inside a savepoint. A
// CHECK rejection rolls the savepoint back and classifies the assertion as
// quarantined (ok=false, nothing stored); anything else is a storage error.
// Content-addressed dedup: assertions are unique on (observation, subject,
// locator, fingerprint) regardless of the minted id, so a re-parse of the
// same observation that yields byte-identical content counts as accepted
// without inserting a duplicate row pair.
func insertAssertion(ctx context.Context, tx pgx.Tx, jobID string, ordinal int,
	observationID string, a AssertionInput) (bool, error) {
	id := assertionID(jobID, ordinal)
	childBytes, err := canonicalize(childPayload(a))
	if err != nil {
		var se *Error
		if errors.As(err, &se) {
			return false, err // canonical-v1 boundary violation: evidence-shaped
		}
		return false, storageErr(err)
	}
	fingerprint := canonicalHash(childBytes)
	if _, err := tx.Exec(ctx, `SAVEPOINT assertion_insert`); err != nil {
		return false, storageErr(err)
	}
	ct, err := tx.Exec(ctx, `
		INSERT INTO assertions (id,observation_id,subject_source_id,subject_upstream_id,locator,verbatim,parser_version,evidence_class,payload_fingerprint)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (observation_id,subject_source_id,subject_upstream_id,locator,payload_fingerprint)
		DO NOTHING`,
		id, observationID, a.SubjectSourceID, a.SubjectUpstreamID, a.Locator,
		a.Verbatim, a.ParserVersion, a.EvidenceClass, fingerprint)
	if err != nil {
		return false, quarantineOrStorage(ctx, tx, err)
	}
	if ct.RowsAffected() == 0 {
		// Identical content already stored under its first id; the child row
		// committed with it. Nothing to add.
		if _, err := tx.Exec(ctx, `RELEASE SAVEPOINT assertion_insert`); err != nil {
			return false, storageErr(err)
		}
		return true, nil
	}
	var execErr error
	switch {
	case a.Capability != nil:
		_, execErr = tx.Exec(ctx, `
			INSERT INTO capability_assertions (assertion_id,capability,support,operation,direction)
			VALUES ($1,$2,$3,$4,$5)`,
			id, a.Capability.Capability, a.Capability.Support, a.Capability.Operation, a.Capability.Direction)
	case a.Claim != nil:
		_, execErr = tx.Exec(ctx, `
			INSERT INTO claim_assertions (assertion_id,dimension,polarity,claim_text)
			VALUES ($1,$2,$3,$4)`,
			id, a.Claim.Dimension, a.Claim.Polarity, a.Claim.ClaimText)
	case a.Limit != nil:
		_, execErr = tx.Exec(ctx, `
			INSERT INTO limit_assertions (assertion_id,limit_name,value,unit,bound,scope,operation,role)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			id, a.Limit.Name, a.Limit.Value, a.Limit.Unit, a.Limit.Bound, a.Limit.Scope,
			a.Limit.Operation, a.Limit.Role)
	case a.ImageFormat != nil:
		_, execErr = tx.Exec(ctx, `
			INSERT INTO image_format_assertions (assertion_id,operation,role,format)
			VALUES ($1,$2,$3,$4)`,
			id, a.ImageFormat.Operation, a.ImageFormat.Role, a.ImageFormat.Format)
	case a.ImageSemantic != nil:
		_, execErr = tx.Exec(ctx, `
			INSERT INTO image_semantic_assertions (assertion_id,operation,role,semantic)
			VALUES ($1,'image_editing','mask',$2)`, id, a.ImageSemantic.Semantic)
	case a.ImageRequirement != nil:
		_, execErr = tx.Exec(ctx, `
			INSERT INTO image_requirement_assertions (assertion_id,operation,role,requirement)
			VALUES ($1,$2,$3,$4)`,
			id, a.ImageRequirement.Operation, a.ImageRequirement.Role, a.ImageRequirement.Requirement)
	}
	if execErr != nil {
		return false, quarantineOrStorage(ctx, tx, execErr)
	}
	if _, err := tx.Exec(ctx, `RELEASE SAVEPOINT assertion_insert`); err != nil {
		return false, storageErr(err)
	}
	return true, nil
}

// quarantineOrStorage maps a failed insert inside the assertion savepoint:
// a CHECK violation is a quarantine and rolls the savepoint back; anything
// else is a storage error that aborts the call.
func quarantineOrStorage(ctx context.Context, tx pgx.Tx, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23514" {
		if _, rbErr := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT assertion_insert`); rbErr != nil {
			return storageErr(rbErr)
		}
		return nil
	}
	return storageErr(err)
}

// childPayload renders the one child payload for the fingerprint.
func childPayload(a AssertionInput) any {
	switch {
	case a.Capability != nil:
		return a.Capability
	case a.Claim != nil:
		return a.Claim
	case a.Limit != nil:
		return a.Limit
	case a.ImageFormat != nil:
		return a.ImageFormat
	case a.ImageSemantic != nil:
		return a.ImageSemantic
	case a.ImageRequirement != nil:
		return a.ImageRequirement
	}
	return nil
}

// DeferJob implements defer_job. deferred requires a future retry_not_before;
// terminal outcomes pin it to observed_at.
func (p *Ports) DeferJob(ctx context.Context, jobID string, token int64,
	outcome, detail string, observedAt, retryNotBefore time.Time) error {
	switch outcome {
	case "deferred":
		if !retryNotBefore.After(observedAt) {
			return errf(CodeInvalidRequest, "deferred requires a future retry_not_before")
		}
	case "quarantined", "failed_terminal":
		if !retryNotBefore.Equal(observedAt) {
			return errf(CodeInvalidRequest, "%s pins retry_not_before to observed_at", outcome)
		}
	default:
		return errf(CodeInvalidRequest, "outcome %q is not a defer token", outcome)
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return storageErr(err)
	}
	defer tx.Rollback(ctx)
	if err := requireActiveLease(ctx, tx, jobID, token, observedAt); err != nil {
		return err
	}
	// A completed job is terminal: no unlisted transitions, no junk attempt
	// evidence on a closed aggregate.
	var one int
	err = tx.QueryRow(ctx, `SELECT 1 FROM job_completions WHERE job_id = $1`, jobID).Scan(&one)
	if err == nil {
		return errf(CodeConflict, "job %s is completed", jobID)
	}
	if err != pgx.ErrNoRows {
		return storageErr(err)
	}
	ct, err := tx.Exec(ctx, `
		INSERT INTO job_attempts (job_id,token,observed_at,outcome,detail,retry_not_before)
		VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`,
		jobID, token, observedAt, outcome, detail, retryNotBefore)
	if err != nil {
		return storageErr(err)
	}
	// Replay of the identical attempt row is a no-op and must not re-emit.
	inserted := ct.RowsAffected()
	if outcome == "deferred" && inserted > 0 {
		var sourceID string
		if err := tx.QueryRow(ctx,
			`SELECT source_id FROM jobs WHERE id = $1`, jobID).Scan(&sourceID); err != nil {
			return storageErr(err)
		}
		if err := emitEvent(ctx, tx, jobID, "JobDeferred", 1, observedAt, map[string]any{
			"job_id":           jobID,
			"source_id":        sourceID,
			"token":            token,
			"retry_not_before": retryNotBefore.Unix(),
			"outcome":          outcome,
			"detail":           detail,
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// CompleteJob implements complete_job: the output must already be durably
// recorded by RecordAttempt for this exact job. Response completion ID.
func (p *Ports) CompleteJob(ctx context.Context, jobID string, token int64,
	outputID string, completedAt time.Time) (string, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return "", storageErr(err)
	}
	defer tx.Rollback(ctx)
	if err := requireActiveLease(ctx, tx, jobID, token, completedAt); err != nil {
		return "", err
	}
	var storedOutputID string
	var stage string
	err = tx.QueryRow(ctx, `
		SELECT o.output_id, j.stage FROM stage_outputs o
		JOIN jobs j ON j.id = o.job_id WHERE o.job_id = $1`, jobID).
		Scan(&storedOutputID, &stage)
	if err == pgx.ErrNoRows {
		return "", errf(CodeConflict, "job %s has no recorded output", jobID)
	}
	if err != nil {
		return "", storageErr(err)
	}
	if storedOutputID != outputID {
		return "", errf(CodeConflict, "output %s is not the recorded output of job %s", outputID, jobID)
	}
	// Replay of an existing completion returns without re-emitting; a
	// completion minted by a different token is a conflict.
	var doneToken int64
	err = tx.QueryRow(ctx, `SELECT token FROM job_completions WHERE job_id = $1`, jobID).Scan(&doneToken)
	if err == nil {
		if doneToken != token {
			return "", errf(CodeConflict, "job %s completed under token %d", jobID, doneToken)
		}
		if err := tx.Commit(ctx); err != nil {
			return "", storageErr(err)
		}
		return jobID, nil
	}
	if err != pgx.ErrNoRows {
		return "", storageErr(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO job_completions (job_id,token,output_id,completed_at)
		VALUES ($1,$2,$3,$4)`,
		jobID, token, outputID, completedAt); err != nil {
		return "", storageErr(err)
	}
	if err := emitEvent(ctx, tx, jobID, "JobCompleted", 1, completedAt, map[string]any{
		"job_id":    jobID,
		"token":     token,
		"output_id": outputID,
		"stage":     stage,
	}); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", storageErr(err)
	}
	return jobID, nil
}

// ---------------------------------------------------------------------------
// Operator
// ---------------------------------------------------------------------------

// LaneState is the set_publication_channel response.
type LaneState struct {
	Channel   string
	Class     string
	RetiredAt *time.Time
}

// RegisterPublicationChannel creates or replays an open publication lane.
// Changing the class of a live lane conflicts; retire and register a new lane.
func (p *Ports) RegisterPublicationChannel(ctx context.Context, channel, class string,
	createdAt time.Time) (LaneState, error) {
	if class != "local" && class != "distributable" {
		return LaneState{}, errf(CodeInvalidRequest, "channel class %q is not a token", class)
	}
	if channel == "" {
		return LaneState{}, errf(CodeInvalidRequest, "channel is required")
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return LaneState{}, storageErr(err)
	}
	defer tx.Rollback(ctx)
	var existingClass string
	err = tx.QueryRow(ctx,
		`SELECT channel_class FROM publication_channels WHERE channel = $1 FOR UPDATE`,
		channel).Scan(&existingClass)
	switch {
	case err == nil:
		var retiredAt *time.Time
		_ = tx.QueryRow(ctx,
			`SELECT retired_at FROM publication_channel_retirements WHERE channel = $1`,
			channel).Scan(&retiredAt)
		if retiredAt != nil {
			return LaneState{}, errf(CodeConflict, "channel %s is retired", channel)
		}
		if existingClass != class {
			return LaneState{}, errf(CodeConflict, "channel %s is live with class %s", channel, existingClass)
		}
		if err := tx.Commit(ctx); err != nil {
			return LaneState{}, storageErr(err)
		}
		return LaneState{Channel: channel, Class: class}, nil
	case err == pgx.ErrNoRows:
		if _, err := tx.Exec(ctx, `
			INSERT INTO publication_channels (channel,channel_class,created_at)
			VALUES ($1,$2,$3)`, channel, class, createdAt); err != nil {
			return LaneState{}, storageErr(err)
		}
		if err := tx.Commit(ctx); err != nil {
			return LaneState{}, storageErr(err)
		}
		return LaneState{Channel: channel, Class: class}, nil
	default:
		return LaneState{}, storageErr(err)
	}
}

// RetirePublicationChannel retires a lane; retiring an already-retired lane
// replays the stored retired_at.
func (p *Ports) RetirePublicationChannel(ctx context.Context, channel string,
	retiredAt time.Time) (LaneState, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return LaneState{}, storageErr(err)
	}
	defer tx.Rollback(ctx)
	var class string
	err = tx.QueryRow(ctx,
		`SELECT channel_class FROM publication_channels WHERE channel = $1 FOR UPDATE`,
		channel).Scan(&class)
	if err == pgx.ErrNoRows {
		return LaneState{}, errf(CodeNotFound, "channel %s", channel)
	}
	if err != nil {
		return LaneState{}, storageErr(err)
	}
	var existing *time.Time
	_ = tx.QueryRow(ctx,
		`SELECT retired_at FROM publication_channel_retirements WHERE channel = $1`,
		channel).Scan(&existing)
	if existing == nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO publication_channel_retirements (channel,retired_at)
			VALUES ($1,$2)`, channel, retiredAt); err != nil {
			return LaneState{}, storageErr(err)
		}
		existing = &retiredAt
	}
	if err := tx.Commit(ctx); err != nil {
		return LaneState{}, storageErr(err)
	}
	return LaneState{Channel: channel, Class: class, RetiredAt: existing}, nil
}

// projectionPolicyDoc is the closed projection policy grammar. The pointer
// and presence checks make every required member mandatory: an omitted
// member is rejected, not defaulted.
type projectionPolicyDoc struct {
	Version                 string            `json:"version"`
	Weights                 map[string]int64  `json:"weights"`
	TaskMap                 map[string]string `json:"task_map"`
	BasisRules              map[string]string `json:"basis_rules"`
	ZeroImageChargeEvidence *bool             `json:"zero_image_charge_evidence"`
}

var taskTokens = map[string]bool{
	"general": true, "code": true, "review": true, "reasoning": true,
	"extraction": true, "writing": true, "translation": true, "vision": true,
	"image_generation": true, "image_editing": true,
}

// ProjectionPolicyState is the register_projection_policy response: the
// stored row, returned identically on registration and on replay.
type ProjectionPolicyState struct {
	Version     string
	ContentHash []byte
	RecordedAt  time.Time
}

// RegisterProjectionPolicy implements register_projection_policy: canonical
// content, closed grammar, version match, immutable and idempotent by version
// plus content hash. Identical replay returns the stored row.
func (p *Ports) RegisterProjectionPolicy(ctx context.Context, version string,
	contentBytes []byte, recordedAt time.Time) (ProjectionPolicyState, error) {
	if version == "" {
		return ProjectionPolicyState{}, errf(CodeInvalidRequest, "version is required")
	}
	if !isCanonical(contentBytes) {
		return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "projection policy bytes are not canonical-v1")
	}
	var doc projectionPolicyDoc
	if err := strictDecode(contentBytes, &doc); err != nil {
		return ProjectionPolicyState{}, err
	}
	if doc.Version != version {
		return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "content version %q does not match %q", doc.Version, version)
	}
	if len(doc.Weights) == 0 {
		return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "weights map is required")
	}
	if doc.TaskMap == nil {
		return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "task_map is required")
	}
	if doc.ZeroImageChargeEvidence == nil {
		return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "zero_image_charge_evidence is required")
	}
	for task, w := range doc.Weights {
		if !taskTokens[task] {
			return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "weight key %q is not a task token", task)
		}
		if w < 1 || w > 10000 {
			return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "weight for %s is outside 1..10000", task)
		}
	}
	for metric, task := range doc.TaskMap {
		if metric == "" || !taskTokens[task] {
			return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "task_map entry %q is not a task mapping", metric)
		}
	}
	permitted := map[string]string{
		"provider_reported":    "publisher_claim",
		"third_party_reported": "published",
		"competitor_reported":  "published",
		"host_observed":        "host_observed",
	}
	if len(doc.BasisRules) != len(permitted) {
		return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "basis_rules must cover the four evidence classes")
	}
	for class, basis := range doc.BasisRules {
		want, ok := permitted[class]
		if !ok || basis != want {
			return ProjectionPolicyState{}, errf(CodeInvalidEvidence, "basis rule %q->%q is not the permitted derivation", class, basis)
		}
	}
	sum := canonicalHash(contentBytes)
	_, err := p.Pool.Exec(ctx, `
		INSERT INTO projection_policies (version,content_hash,content_bytes,recorded_at)
		VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, version, sum, contentBytes, recordedAt)
	if err != nil {
		return ProjectionPolicyState{}, storageErr(err)
	}
	var storedHash []byte
	var storedAt time.Time
	if err := p.Pool.QueryRow(ctx,
		`SELECT content_hash, recorded_at FROM projection_policies WHERE version = $1`, version).
		Scan(&storedHash, &storedAt); err != nil {
		return ProjectionPolicyState{}, storageErr(err)
	}
	if string(storedHash) != string(sum) {
		return ProjectionPolicyState{}, errf(CodeConflict, "version %s already registered with different content", version)
	}
	return ProjectionPolicyState{Version: version, ContentHash: storedHash, RecordedAt: storedAt}, nil
}

// ---------------------------------------------------------------------------
// Host outcome reporter
// ---------------------------------------------------------------------------

var usageComponents = []string{
	"input_tokens", "output_tokens", "input_images", "output_images",
	"references", "searches", "elapsed_millis",
}

var outcomeTokens = map[string]bool{
	"validated_success": true, "validated_failure": true,
	"unvalidated_completion": true, "indeterminate": true, "cancelled": true,
}

// ReceiptUsage is one usage component: known with a value, or explicitly
// unknown. Unknown is never a zero value.
type ReceiptUsage struct {
	Component string
	Known     bool
	Value     int64
}

// RecordReceiptParams is the record_receipt input. ReceiptBytes are the
// canonical-v1 bytes of the receipt document and become the host_local
// observation body.
type RecordReceiptParams struct {
	ReceiptID        string
	RequestID        string
	OfferID          string
	Task             string
	Outcome          string
	ValidatorVersion string
	RecordedAt       time.Time
	Usage            []ReceiptUsage
	ReceiptBytes     []byte
	CollectorVersion string
}

// RecordReceiptResult is the record_receipt response.
type RecordReceiptResult struct {
	ReceiptID     string
	ObservationID string
	UsageHash     []byte
}

// receiptDoc is the closed receipt document grammar pinned by
// host-integration.md Outcome capture. strictDecode rejects any member
// outside it, so task prompts, file contents, images, tool outputs and
// credentials are mechanically refused.
type receiptDoc struct {
	ReceiptID        string          `json:"receipt_id"`
	RequestID        string          `json:"request_id"`
	OfferID          string          `json:"offer_id"`
	Task             string          `json:"task"`
	Outcome          string          `json:"outcome"`
	ValidatorVersion string          `json:"validator_version"`
	Usage            []receiptDocUse `json:"usage"`
}

// receiptDocUse decodes one usage element. Value is a pointer so that a
// missing or null value is distinguishable from an explicit zero: known
// usage must state its measurement, and zero is only ever accepted when the
// document carries it.
type receiptDocUse struct {
	Component string `json:"component"`
	Known     bool   `json:"known"`
	Value     *int64 `json:"value,omitempty"`
}

// toUsage converts a decoded usage element to the port fact. Unknown usage
// keeps its existing semantics: any document value is ignored and the fact
// value stays zero.
func (u receiptDocUse) toUsage() ReceiptUsage {
	value := int64(0)
	if u.Value != nil {
		value = *u.Value
	}
	return ReceiptUsage{Component: u.Component, Known: u.Known, Value: value}
}

// decodeReceipt is the shared receipt-document decoder for both grammar
// paths. strictDecode closes the schema (unknown members rejected, canonical
// framing, 53-bit integers); on top of it this requires known usage to carry
// an explicit value, so no zero is fabricated from absent evidence, and
// re-canonicalizes the decoded document, which enforces the exact required
// member set, exact casing, no duplicates and the null/presence shape
// without a second schema map: encoding/json also accepts case-aliased
// members and missing or null fields that the grammar forbids, but their
// decoded form never re-canonicalizes to the original bytes.
func decodeReceipt(data []byte) (receiptDoc, error) {
	var doc receiptDoc
	if err := strictDecode(data, &doc); err != nil {
		return receiptDoc{}, err
	}
	for _, u := range doc.Usage {
		if u.Known && u.Value == nil {
			return receiptDoc{}, errf(CodeInvalidEvidence,
				"known usage %s carries no explicit value", u.Component)
		}
	}
	round, err := canonicalize(doc)
	if err != nil {
		return receiptDoc{}, err
	}
	if !bytes.Equal(round, data) {
		return receiptDoc{}, errf(CodeInvalidEvidence,
			"receipt document does not match the closed receipt grammar")
	}
	return doc, nil
}

// validateReceipt is the single semantic grammar source for receipt
// submissions, shared by the host boundary (ParseReceipt, upfront, before
// any config or storage is touched) and the port (RecordReceipt, revalidate
// on record). It owns framing (canonical bytes, strict decode), the ID
// requirements, the closed task and outcome tokens, the validator split,
// document/submission coherence and the seven usage components.
// requireCollectorVersion separates the caller fact only the port can
// check: ParseReceipt output does not yet carry one. Error details quote
// submitted content and stay inside the collector; boundaries that print
// diagnostics wrap them (see ParseReceipt).
func validateReceipt(params RecordReceiptParams, requireCollectorVersion bool) (receiptDoc, error) {
	if params.ReceiptID == "" || params.RequestID == "" || params.OfferID == "" {
		return receiptDoc{}, errf(CodeInvalidRequest, "receipt, request and offer ids are required")
	}
	if !taskTokens[params.Task] {
		return receiptDoc{}, errf(CodeInvalidRequest, "task %q is not a token", params.Task)
	}
	if !outcomeTokens[params.Outcome] {
		return receiptDoc{}, errf(CodeInvalidRequest, "outcome %q is not a token", params.Outcome)
	}
	validated := params.Outcome == "validated_success" || params.Outcome == "validated_failure"
	if validated && params.ValidatorVersion == "" {
		return receiptDoc{}, errf(CodeInvalidRequest, "validated outcomes require validator_version")
	}
	if !validated && params.ValidatorVersion != "" {
		return receiptDoc{}, errf(CodeInvalidRequest, "unvalidated outcomes carry no validator_version")
	}
	if len(params.ReceiptBytes) == 0 {
		return receiptDoc{}, errf(CodeInvalidRequest, "receipt bytes are required")
	}
	doc, err := decodeReceipt(params.ReceiptBytes)
	if err != nil {
		return receiptDoc{}, err
	}
	// The stored document must state exactly the same facts the port writes.
	if doc.ReceiptID != params.ReceiptID || doc.RequestID != params.RequestID ||
		doc.OfferID != params.OfferID || doc.Task != params.Task || doc.Outcome != params.Outcome ||
		doc.ValidatorVersion != params.ValidatorVersion {
		return receiptDoc{}, errf(CodeInvalidEvidence,
			"receipt document does not match the submitted receipt facts")
	}
	if len(doc.Usage) != len(params.Usage) {
		return receiptDoc{}, errf(CodeInvalidEvidence,
			"receipt document usage does not match the submitted usage")
	}
	docUsage := map[string]ReceiptUsage{}
	for _, u := range doc.Usage {
		if _, dup := docUsage[u.Component]; dup {
			return receiptDoc{}, errf(CodeInvalidEvidence,
				"receipt document repeats usage component %s", u.Component)
		}
		docUsage[u.Component] = u.toUsage()
	}
	for _, u := range params.Usage {
		d, ok := docUsage[u.Component]
		if !ok || d.Known != u.Known || (u.Known && d.Value != u.Value) {
			return receiptDoc{}, errf(CodeInvalidEvidence,
				"receipt document usage %s does not match the submitted usage", u.Component)
		}
	}
	if requireCollectorVersion && params.CollectorVersion == "" {
		return receiptDoc{}, errf(CodeInvalidRequest, "collector version is required")
	}
	seen := map[string]bool{}
	for _, u := range params.Usage {
		if !usageComponentToken(u.Component) {
			return receiptDoc{}, errf(CodeInvalidRequest, "usage component %q is not a token", u.Component)
		}
		if seen[u.Component] {
			return receiptDoc{}, errf(CodeInvalidRequest, "usage component %s appears twice", u.Component)
		}
		seen[u.Component] = true
		if u.Known && u.Value < 0 {
			return receiptDoc{}, errf(CodeInvalidRequest, "known usage %s is negative", u.Component)
		}
		if u.Known && (u.Value > maxSafeInteger) {
			return receiptDoc{}, errf(CodeInvalidEvidence,
				"known usage %s exceeds the signed 53-bit range", u.Component)
		}
	}
	if len(seen) != len(usageComponents) {
		return receiptDoc{}, errf(CodeInvalidRequest, "usage must cover the seven components exactly once")
	}
	return doc, nil
}

// RecordReceipt implements record_receipt: host_local observation with HTTP
// status 0 (honest absence marker), typed outcome fact with the validation
// split, usage facts split by knowledge and ReceiptRecorded. Idempotent by
// receipt ID plus body hash.
func (p *Ports) RecordReceipt(ctx context.Context, params RecordReceiptParams) (RecordReceiptResult, error) {
	if _, err := validateReceipt(params, true); err != nil {
		return RecordReceiptResult{}, err
	}
	validated := params.Outcome == "validated_success" || params.Outcome == "validated_failure"

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return RecordReceiptResult{}, storageErr(err)
	}
	defer tx.Rollback(ctx)

	// Idempotency by receipt ID plus body hash.
	var priorBodyHash []byte
	err = tx.QueryRow(ctx, `
		SELECT o.body_hash FROM outcome_facts f
		JOIN observations o ON o.id = f.observation_id
		WHERE f.receipt_id = $1`, params.ReceiptID).Scan(&priorBodyHash)
	switch {
	case err == nil:
		bodyHash := canonicalHash(params.ReceiptBytes)
		if string(priorBodyHash) != string(bodyHash) {
			return RecordReceiptResult{}, errf(CodeConflict, "receipt %s exists with different bytes", params.ReceiptID)
		}
		var observationID string
		if err := tx.QueryRow(ctx,
			`SELECT observation_id FROM outcome_facts WHERE receipt_id = $1`,
			params.ReceiptID).Scan(&observationID); err != nil {
			return RecordReceiptResult{}, storageErr(err)
		}
		usageHash, err := receiptUsageHash(ctx, tx, params.ReceiptID)
		if err != nil {
			return RecordReceiptResult{}, err
		}
		return RecordReceiptResult{ReceiptID: params.ReceiptID, ObservationID: observationID, UsageHash: usageHash}, nil
	case err != pgx.ErrNoRows:
		return RecordReceiptResult{}, storageErr(err)
	}

	// The single host_local source owns receipt observations.
	var hostSourceID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM sources WHERE source_kind = 'host_local' LIMIT 1`).Scan(&hostSourceID)
	if err == pgx.ErrNoRows {
		return RecordReceiptResult{}, errf(CodeNotFound, "no host_local source is registered")
	}
	if err != nil {
		return RecordReceiptResult{}, storageErr(err)
	}

	bodyHash := canonicalHash(params.ReceiptBytes)
	if _, err := tx.Exec(ctx,
		`INSERT INTO content_blobs (hash,bytes) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		bodyHash, params.ReceiptBytes); err != nil {
		return RecordReceiptResult{}, storageErr(err)
	}
	observationID := newID("obs")
	resourceURL := "host-local://receipts/" + params.ReceiptID
	if _, err := tx.Exec(ctx, `
		INSERT INTO observations (id,source_id,resource_url,retrieved_at,http_status,outcome,body_hash,headers_bytes,collector_version)
		VALUES ($1,$2,$3,$4,0,'success',$5,$6,$7)`,
		observationID, hostSourceID, resourceURL, params.RecordedAt, bodyHash,
		[]byte{}, params.CollectorVersion); err != nil {
		return RecordReceiptResult{}, storageErr(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO outcome_facts (receipt_id,observation_id,request_id,offer_id,task,outcome,recorded_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		params.ReceiptID, observationID, params.RequestID, params.OfferID,
		params.Task, params.Outcome, params.RecordedAt); err != nil {
		return RecordReceiptResult{}, storageErr(err)
	}
	if validated {
		if _, err := tx.Exec(ctx, `
			INSERT INTO validated_outcomes (receipt_id,validator_version,recorded_at)
			VALUES ($1,$2,$3)`,
			params.ReceiptID, params.ValidatorVersion, params.RecordedAt); err != nil {
			return RecordReceiptResult{}, storageErr(err)
		}
	}
	usageHash := canonicalHash(mustCanonicalize(usageDoc(sortedUsage(params.Usage))))
	for _, u := range params.Usage {
		if u.Known {
			_, err = tx.Exec(ctx, `
				INSERT INTO outcome_usage_known (receipt_id,component,value)
				VALUES ($1,$2,$3)`, params.ReceiptID, u.Component, u.Value)
		} else {
			_, err = tx.Exec(ctx, `
				INSERT INTO outcome_usage_unknown (receipt_id,component)
				VALUES ($1,$2)`, params.ReceiptID, u.Component)
		}
		if err != nil {
			return RecordReceiptResult{}, storageErr(err)
		}
	}
	if err := emitEvent(ctx, tx, params.ReceiptID, "ReceiptRecorded", 1, params.RecordedAt, map[string]any{
		"receipt_id":     params.ReceiptID,
		"observation_id": observationID,
		"offer_id":       params.OfferID,
		"task":           params.Task,
		"outcome":        params.Outcome,
		"usage_hash":     hex.EncodeToString(usageHash),
	}); err != nil {
		return RecordReceiptResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RecordReceiptResult{}, storageErr(err)
	}
	return RecordReceiptResult{ReceiptID: params.ReceiptID, ObservationID: observationID, UsageHash: usageHash}, nil
}

func usageComponentToken(c string) bool {
	for _, t := range usageComponents {
		if t == c {
			return true
		}
	}
	return false
}

// sortedUsage orders the seven components bytewise UTF-8 ascending, the
// canonical usage_hash ordering.
func sortedUsage(in []ReceiptUsage) []ReceiptUsage {
	out := append([]ReceiptUsage(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Component < out[j-1].Component; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func mustCanonicalize(v any) []byte {
	b, err := canonicalize(v)
	if err != nil {
		panic(err)
	}
	return b
}

func receiptUsageHash(ctx context.Context, tx pgx.Tx, receiptID string) ([]byte, error) {
	// Rebuild the canonical usage document from the stored split rows.
	type row struct {
		component string
		known     bool
		value     int64
	}
	rows := map[string]row{}
	knownRows, err := tx.Query(ctx,
		`SELECT component, value FROM outcome_usage_known WHERE receipt_id = $1`, receiptID)
	if err != nil {
		return nil, storageErr(err)
	}
	for knownRows.Next() {
		var r row
		if err := knownRows.Scan(&r.component, &r.value); err != nil {
			return nil, storageErr(err)
		}
		r.known = true
		rows[r.component] = r
	}
	knownRows.Close()
	unknownRows, err := tx.Query(ctx,
		`SELECT component FROM outcome_usage_unknown WHERE receipt_id = $1`, receiptID)
	if err != nil {
		return nil, storageErr(err)
	}
	for unknownRows.Next() {
		var r row
		if err := unknownRows.Scan(&r.component); err != nil {
			return nil, storageErr(err)
		}
		rows[r.component] = r
	}
	unknownRows.Close()
	usage := make([]ReceiptUsage, 0, len(rows))
	for _, r := range rows {
		usage = append(usage, ReceiptUsage{Component: r.component, Known: r.known, Value: r.value})
	}
	return canonicalHash(mustCanonicalize(usageDoc(sortedUsage(usage)))), nil
}

func usageDoc(usage []ReceiptUsage) []map[string]any {
	doc := []map[string]any{}
	for _, u := range usage {
		m := map[string]any{"component": u.Component}
		if u.Known {
			m["value"] = u.Value
		}
		doc = append(doc, m)
	}
	return doc
}

// ---------------------------------------------------------------------------
// Transactional outbox delivery
// ---------------------------------------------------------------------------

// OutboxEvent is one delivered envelope.
type OutboxEvent struct {
	EventID           string
	AggregateID       string
	AggregateSequence int64
	EventName         string
	SchemaVersion     int
	OccurredAt        time.Time
	PayloadHash       []byte
	Payload           []byte
}

// emitEvent writes one outbox row inside the caller's transaction. The
// aggregate sequence is minted under the aggregate's existing rows.
func emitEvent(ctx context.Context, tx pgx.Tx, aggregateID, name string,
	schemaVersion int, occurredAt time.Time, payload any) error {
	payloadBytes, err := canonicalize(payload)
	if err != nil {
		return errf(CodeInvalidEvidence, "event payload encoding failed")
	}
	payloadHash := canonicalHash(payloadBytes)
	// Mint the sequence under a transaction-scoped advisory lock on the
	// aggregate: two jobs contributing to one aggregate hold different job
	// locks and would otherwise race on max(sequence)+1.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, aggregateID); err != nil {
		return storageErr(err)
	}
	var seq int64
	if err := tx.QueryRow(ctx, `
		SELECT coalesce(max(aggregate_sequence),0) + 1 FROM outbox_events
		WHERE aggregate_id = $1`, aggregateID).Scan(&seq); err != nil {
		return storageErr(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO outbox_events (event_id,aggregate_id,aggregate_sequence,event_name,schema_version,occurred_at,payload_hash,payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		newID("evt"), aggregateID, seq, name, schemaVersion, occurredAt, payloadHash, payloadBytes); err != nil {
		return storageErr(err)
	}
	return nil
}

// DeliveryScan returns up to limit undelivered events for one consumer in
// (occurred_at, event_id) order. At-least-once: an event stays visible until
// the consumer records its receipt, so a consumer that cannot process an
// event simply leaves it visible.
func (p *Ports) DeliveryScan(ctx context.Context, consumerID string, limit int) ([]OutboxEvent, error) {
	if consumerID == "" || limit <= 0 {
		return nil, errf(CodeInvalidRequest, "consumer id and positive limit are required")
	}
	rows, err := p.Pool.Query(ctx, `
		SELECT e.event_id, e.aggregate_id, e.aggregate_sequence, e.event_name,
		       e.schema_version, e.occurred_at, e.payload_hash, e.payload
		FROM outbox_events e
		WHERE NOT EXISTS (
			SELECT 1 FROM consumer_receipts r
			WHERE r.consumer_id = $1 AND r.event_id = e.event_id)
		ORDER BY e.occurred_at, e.event_id
		LIMIT $2`, consumerID, limit)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.EventID, &e.AggregateID, &e.AggregateSequence,
			&e.EventName, &e.SchemaVersion, &e.OccurredAt, &e.PayloadHash, &e.Payload); err != nil {
			return nil, storageErr(err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// RecordConsumerReceipt marks one event delivered for one consumer.
// Idempotent by the composite PK.
func (p *Ports) RecordConsumerReceipt(ctx context.Context, consumerID, eventID string,
	receivedAt time.Time) error {
	if _, err := p.Pool.Exec(ctx, `
		INSERT INTO consumer_receipts (consumer_id,event_id,received_at)
		VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, consumerID, eventID, receivedAt); err != nil {
		return storageErr(err)
	}
	return nil
}

// ParseReceipt decodes one canonical-v1 receipt document and validates it
// against the shared receipt grammar (framing, IDs, closed tokens, the
// validator split, the seven usage components) before any config or
// storage is touched, so an invalid document fails at the host boundary
// instead of when postgres opens. ReceiptBytes keeps the exact input
// bytes for the body hash and idempotency; the port re-validates on
// record. Error details are fixed and safe for the host boundary: never
// document content, per cli-adapter.md's safe-detail rule.
func ParseReceipt(data []byte) (RecordReceiptParams, error) {
	doc, err := decodeReceipt(data)
	if err != nil {
		return RecordReceiptParams{}, errf(CodeInvalidRequest,
			"receipt document is not a canonical closed-grammar document")
	}
	params := RecordReceiptParams{
		ReceiptID:        doc.ReceiptID,
		RequestID:        doc.RequestID,
		OfferID:          doc.OfferID,
		Task:             doc.Task,
		Outcome:          doc.Outcome,
		ValidatorVersion: doc.ValidatorVersion,
		ReceiptBytes:     data,
	}
	for _, u := range doc.Usage {
		params.Usage = append(params.Usage, u.toUsage())
	}
	if _, err := validateReceipt(params, false); err != nil {
		return RecordReceiptParams{}, errf(CodeInvalidRequest,
			"receipt document does not satisfy the receipt contract")
	}
	return params, nil
}
