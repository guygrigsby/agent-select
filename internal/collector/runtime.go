package collector

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	mathrand "math/rand/v2"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// codeOfErr extracts the storage error code; non-storage errors map to
// temporarily_unavailable so callers retry.
func codeOfErr(err error) Code {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return CodeTemporarilyUnavailable
}

// ClaimableJob is one row of the claimable scan: work whose retry deadline
// has arrived, holding no live lease, not completed and not dead-lettered.
// Retry deadlines are the only timer-owned work (collection-pipeline.md
// Event-driven collection); this scan is the deadline sensor, not a poller
// releasing stages.
type ClaimableJob struct {
	JobID       string
	SourceID    string
	ResourceURL string
	Stage       string
	Attempts    int
}

// ClaimableJobs returns up to limit jobs that may be leased now. Read-only:
// fencing still happens in LeaseJob.
func (p *Ports) ClaimableJobs(ctx context.Context, now time.Time, limit int) ([]ClaimableJob, error) {
	if limit <= 0 {
		return nil, errf(CodeInvalidRequest, "positive limit required")
	}
	rows, err := p.Pool.Query(ctx, `
		SELECT j.id, j.source_id, j.resource_url, j.stage, j.attempts
		FROM jobs j
		JOIN sources s ON s.id = j.source_id AND s.enabled
		LEFT JOIN LATERAL (
			SELECT a.outcome, a.retry_not_before
			FROM job_attempts a
			WHERE a.job_id = j.id
			ORDER BY a.observed_at DESC, a.token DESC
			LIMIT 1
		) last ON true
		WHERE NOT EXISTS (
			SELECT 1 FROM job_completions c WHERE c.job_id = j.id)
		  AND NOT EXISTS (
			SELECT 1 FROM job_leases l
			WHERE l.job_id = j.id AND l.expires_at > $1)
		  AND COALESCE(last.outcome, '') NOT IN ('quarantined','failed_terminal')
		  AND COALESCE(last.retry_not_before, $1) <= $1
		ORDER BY j.created_at, j.id
		LIMIT $2`, now, limit)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	var jobs []ClaimableJob
	for rows.Next() {
		var j ClaimableJob
		if err := rows.Scan(&j.JobID, &j.SourceID, &j.ResourceURL, &j.Stage, &j.Attempts); err != nil {
			return nil, storageErr(err)
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// PriorObservationID returns the latest successful observation of a source's
// resource, or "" when none exists. It backs the unchanged (304) path.
func (p *Ports) PriorObservationID(ctx context.Context, sourceID, resourceURL string) (string, error) {
	var id string
	err := p.Pool.QueryRow(ctx, `
		SELECT id FROM observations
		WHERE source_id = $1 AND resource_url = $2 AND outcome = 'success'
		ORDER BY retrieved_at DESC, id DESC
		LIMIT 1`, sourceID, resourceURL).Scan(&id)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", storageErr(err)
	}
	return id, nil
}

// ExpiredLeases returns expired leases needing recovery: the job's latest
// lease (not superseded), not completed, not already recovered, with no
// attempt recorded under it. A lease that already produced attempt evidence
// needs no recovery: the claimable scan re-enters that work by retry
// deadline, and the durable output link lets a fresh lease adopt the same
// work tuple. A lease already marked recovered stays marked: reconciliation
// replays derive state without re-adopting the lease.
func (p *Ports) ExpiredLeases(ctx context.Context, now time.Time, limit int) ([]ExpiredLease, error) {
	if limit <= 0 {
		return nil, errf(CodeInvalidRequest, "positive limit required")
	}
	rows, err := p.Pool.Query(ctx, `
		SELECT l.job_id, l.token
		FROM job_leases l
		JOIN jobs j ON j.id = l.job_id
		WHERE l.expires_at <= $1
		  AND l.recovered_at IS NULL
		  AND j.fencing_token = l.token
		  AND NOT EXISTS (
			SELECT 1 FROM job_completions c WHERE c.job_id = l.job_id)
		  AND NOT EXISTS (
			SELECT 1 FROM job_attempts a
			WHERE a.job_id = l.job_id AND a.token = l.token)
		ORDER BY l.expires_at, l.job_id, l.token
		LIMIT $2`, now, limit)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	var out []ExpiredLease
	for rows.Next() {
		var e ExpiredLease
		if err := rows.Scan(&e.JobID, &e.Token); err != nil {
			return nil, storageErr(err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

type ExpiredLease struct {
	JobID string
	Token int64
}

// FetchResult is one completed HTTP exchange from a Fetcher.
type FetchResult struct {
	HTTPStatus int
	Body       []byte
	Headers    []byte // canonical bytes; hashed into the envelope
	RetryAfter time.Duration
}

// Fetcher executes one fetch for one source resource. The production
// implementation is HTTPFetcher; tests inject fakes.
type Fetcher interface {
	Fetch(ctx context.Context, sourceID, resourceURL string) (FetchResult, error)
}

// HostGate enforces the shared per-host operating limit
// (collection-pipeline.md Sources and operating limits): at least
// MinInterval between requests to the same host across all workers.
type HostGate struct {
	MinInterval time.Duration

	mu   sync.Mutex
	next map[string]time.Time
}

func NewHostGate(minInterval time.Duration) *HostGate {
	if minInterval <= 0 {
		minInterval = time.Second
	}
	return &HostGate{MinInterval: minInterval, next: make(map[string]time.Time)}
}

func (g *HostGate) Wait(ctx context.Context, host string) error {
	for {
		g.mu.Lock()
		now := time.Now()
		at := g.next[host]
		if !at.After(now) {
			g.next[host] = now.Add(g.MinInterval)
			g.mu.Unlock()
			return nil
		}
		g.mu.Unlock()
		wait := at.Sub(now)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

// WorkerConfig bounds one collector worker.
type WorkerConfig struct {
	WorkerID         string
	CollectorVersion string
	LeaseDuration    time.Duration
	MinHostInterval  time.Duration
	MaxInFlight      int // total concurrency bound across hosts
	ScanLimit        int
	PollIdle         time.Duration // sleep when a scan finds nothing
	BaseBackoff      time.Duration // first retry deferral (before jitter)
	MaxBackoff       time.Duration
}

func (c WorkerConfig) withDefaults() WorkerConfig {
	if c.WorkerID == "" {
		c.WorkerID = "worker-" + hex.EncodeToString(randBytes8())
	}
	if c.CollectorVersion == "" {
		c.CollectorVersion = "collector-dev"
	}
	if c.LeaseDuration <= 0 {
		c.LeaseDuration = time.Minute
	}
	if c.MinHostInterval <= 0 {
		c.MinHostInterval = time.Second
	}
	if c.MaxInFlight <= 0 {
		c.MaxInFlight = 4
	}
	if c.ScanLimit <= 0 {
		c.ScanLimit = 32
	}
	if c.PollIdle <= 0 {
		c.PollIdle = 5 * time.Second
	}
	if c.BaseBackoff <= 0 {
		c.BaseBackoff = 30 * time.Second
	}
	if c.MaxBackoff <= 0 {
		c.MaxBackoff = time.Hour
	}
	return c
}

// Worker drives the durable fetch loop: claimable scan, fenced lease, fetch
// through the shared host gate, then RecordAttempt or DeferJob. It never
// aborts on one job's failure; storage outages back off and retry.
type Worker struct {
	p       *Ports
	fetch   Fetcher
	cfg     WorkerConfig
	gate    *HostGate
	now     func() time.Time
	Fetches func() int // test hook: completed fetch exchanges
}

func NewWorker(p *Ports, f Fetcher, cfg WorkerConfig) *Worker {
	cfg = cfg.withDefaults()
	return &Worker{p: p, fetch: f, cfg: cfg, gate: NewHostGate(cfg.MinHostInterval), now: time.Now}
}

// Run blocks until ctx is done. Idle scans sleep PollIdle; storage errors
// sleep and retry (durable by default: accepted output is never lost, the
// scan is the re-entry point).
func (w *Worker) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		n, err := w.runOnce(ctx)
		if err != nil && ctx.Err() == nil {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(w.cfg.PollIdle):
			}
			continue
		}
		if n == 0 {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(w.cfg.PollIdle):
			}
		}
	}
}

// runOnce scans once and processes the claimable set with bounded
// concurrency. Returns the number of jobs picked up.
func (w *Worker) runOnce(ctx context.Context) (int, error) {
	jobs, err := w.p.ClaimableJobs(ctx, w.now(), w.cfg.ScanLimit)
	if err != nil {
		return 0, err
	}
	if len(jobs) == 0 {
		return 0, nil
	}
	sem := make(chan struct{}, w.cfg.MaxInFlight)
	var wg sync.WaitGroup
	for _, j := range jobs {
		j := j
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			w.processOne(ctx, j)
		}()
	}
	wg.Wait()
	return len(jobs), nil
}

// processOne runs one job to a terminal attempt state or a deferral. Errors
// are logged by the caller surface (the daemon); the loop never dies on one
// job.
func (w *Worker) processOne(ctx context.Context, job ClaimableJob) {
	host := hostOf(job.ResourceURL)
	if err := w.gate.Wait(ctx, host); err != nil {
		return
	}
	now := w.now()
	lease, err := w.p.LeaseJob(ctx, job.JobID, w.cfg.WorkerID, w.cfg.LeaseDuration, now)
	if err != nil {
		return // contended, completed or dead-lettered meanwhile
	}
	res, ferr := w.fetch.Fetch(ctx, job.SourceID, job.ResourceURL)
	if w.Fetches != nil {
		w.Fetches()
	}
	if ferr != nil || res.HTTPStatus >= 500 || res.HTTPStatus == 429 {
		w.deferRetry(ctx, job, lease.Token, res, ferr)
		return
	}
	switch res.HTTPStatus {
	case 200:
		obs := &AttemptObservation{
			ObservationID: newID("obs"),
			ResourceURL:   job.ResourceURL,
			RetrievedAt:   w.now(),
			HTTPStatus:    200,
			Outcome:       "success",
			BodyBytes:     res.Body,
			HeadersBytes:  res.Headers,
		}
		env, err := fetchEnvelopeBytes(obs)
		if err != nil {
			w.deferRetry(ctx, job, lease.Token, FetchResult{}, err)
			return
		}
		rec, err := w.p.RecordAttempt(ctx, RecordAttemptParams{
			JobID: job.JobID, Token: lease.Token,
			CollectorVersion: w.cfg.CollectorVersion,
			RecordedAt:       w.now(), Observation: obs, OutputBytes: env,
		})
		if err != nil {
			if codeOfErr(err) == CodeTemporarilyUnavailable {
				// Observation not committed: the job stays claimable and
				// will refetch. Nothing durable was lost.
				return
			}
			if codeOfErr(err) == CodeConflict {
				// The stranded output under this work tuple disagrees with
				// this fetch: upstream bytes changed between the crash and
				// the refetch. Not adoptable (identity differs), not a
				// transport retry. Defer with backoff so the disagreement is
				// durable, visible and rate-limited pending disposition;
				// without this the job refetches on every scan with zero
				// durable progress.
				w.deferRetry(ctx, job, lease.Token, FetchResult{},
					fmt.Errorf("conflicting output for work tuple: upstream bytes changed under a stranded output; awaiting disposition"))
			}
			return
		}
		// The durable output exists; completion closes the aggregate and
		// emits JobCompleted. Idempotent by output identity.
		if _, err := w.p.CompleteJob(ctx, job.JobID, lease.Token, rec.OutputID, w.now()); err != nil &&
			codeOfErr(err) != CodeTemporarilyUnavailable {
			// A crash after RecordAttempt leaves the output link; the
			// reconciliation sweep or a retry lease adopts it for the same
			// work tuple, so non-transient failures need no re-fetch.
			_ = err
		}
	case 304:
		prior, err := w.p.PriorObservationID(ctx, job.SourceID, job.ResourceURL)
		if err != nil || prior == "" {
			w.deferRetry(ctx, job, lease.Token, FetchResult{}, fmt.Errorf("no prior observation for unchanged"))
			return
		}
		body := res.Body
		if body == nil {
			body = []byte{} // content_blobs.bytes is NOT NULL; an unchanged exchange carries no body
		}
		obs := &AttemptObservation{
			ObservationID:      newID("obs"),
			ResourceURL:        job.ResourceURL,
			RetrievedAt:        w.now(),
			HTTPStatus:         304,
			Outcome:            "unchanged",
			PriorObservationID: prior,
			BodyBytes:          body,
			HeadersBytes:       res.Headers,
		}
		env, err := fetchEnvelopeBytes(obs)
		if err == nil {
			rec, rerr := w.p.RecordAttempt(ctx, RecordAttemptParams{
				JobID: job.JobID, Token: lease.Token,
				CollectorVersion: w.cfg.CollectorVersion,
				RecordedAt:       w.now(), Observation: obs, OutputBytes: env,
			})
			if rerr == nil {
				_, _ = w.p.CompleteJob(ctx, job.JobID, lease.Token, rec.OutputID, w.now())
			} else if codeOfErr(rerr) == CodeConflict {
				// Same stranded-output disagreement as the success path:
				// defer durably instead of refetching on every scan.
				w.deferRetry(ctx, job, lease.Token, FetchResult{},
					fmt.Errorf("conflicting output for work tuple: upstream bytes changed under a stranded output; awaiting disposition"))
			}
		}
	case 401, 403:
		// Access-blocked: visible dead letter; re-enters only on a changed
		// work tuple (source/parser/policy change).
		at := w.now()
		w.p.DeferJob(ctx, job.JobID, lease.Token, "quarantined",
			fmt.Sprintf("access blocked: status %d", res.HTTPStatus), at, at)
	case 404, 410:
		at := w.now()
		w.p.DeferJob(ctx, job.JobID, lease.Token, "failed_terminal",
			fmt.Sprintf("resource gone: status %d", res.HTTPStatus), at, at)
	default:
		w.deferRetry(ctx, job, lease.Token, res, nil)
	}
}

// deferRetry schedules a bounded-jitter exponential deferral. Retry-After
// (already parsed by the fetcher) overrides the computed backoff.
func (w *Worker) deferRetry(ctx context.Context, job ClaimableJob, token int64,
	res FetchResult, ferr error) {
	now := w.now()
	detail := "fetch retry"
	if ferr != nil {
		detail = ferr.Error()
	} else {
		detail = fmt.Sprintf("status %d", res.HTTPStatus)
	}
	wait := w.backoff(job.Attempts)
	if res.RetryAfter > wait {
		wait = res.RetryAfter
	}
	w.p.DeferJob(ctx, job.JobID, token, "deferred", detail, now, now.Add(wait))
}

func (w *Worker) backoff(attempts int) time.Duration {
	d := w.cfg.BaseBackoff
	for range min(attempts, 12) {
		d *= 2
		if d >= w.cfg.MaxBackoff {
			d = w.cfg.MaxBackoff
			break
		}
	}
	if d > w.cfg.MaxBackoff {
		d = w.cfg.MaxBackoff
	}
	// jitter in [0.8, 1.2]
	j := 0.8 + 0.4*mathrand.Float64()
	return time.Duration(float64(d) * j)
}

// ReconcileExpiredLeases adopts every expired non-completed lease. Run at
// startup and periodically; RecoverJob advances the attempt counter and
// leaves the job claimable after its retry deadline.
func ReconcileExpiredLeases(ctx context.Context, p *Ports, now time.Time, limit int) (int, error) {
	expired, err := p.ExpiredLeases(ctx, now, limit)
	if err != nil {
		return 0, err
	}
	adopted := 0
	for _, e := range expired {
		_, err := p.RecoverJob(ctx, e.JobID, e.Token, now)
		if err != nil {
			if codeOfErr(err) == CodeStaleLease || codeOfErr(err) == CodeNotFound {
				continue // renewed or superseded meanwhile
			}
			return adopted, err
		}
		adopted++
	}
	return adopted, nil
}

// PumpOutbox delivers up to limit undelivered events to one consumer.
// Delivery is at least once: the handler runs first and the receipt is
// recorded after, so a crash between them redelivers. A handler error stops
// the pump before its event's receipt.
func (p *Ports) PumpOutbox(ctx context.Context, consumerID string, limit int,
	handle func(OutboxEvent) error) (int, error) {
	events, err := p.DeliveryScan(ctx, consumerID, limit)
	if err != nil {
		return 0, err
	}
	delivered := 0
	for _, e := range events {
		if err := handle(e); err != nil {
			return delivered, err
		}
		if err := p.RecordConsumerReceipt(ctx, consumerID, e.EventID, time.Now().UTC()); err != nil {
			return delivered, err
		}
		delivered++
	}
	return delivered, nil
}

// fetchEnvelopeBytes builds the closed fetch envelope canonical bytes for an
// observation.
func fetchEnvelopeBytes(obs *AttemptObservation) ([]byte, error) {
	return canonicalize(map[string]any{
		"stage":          "fetch",
		"observation_id": obs.ObservationID,
		"resource_url":   obs.ResourceURL,
		"body_hash":      hexHash(obs.BodyBytes),
		"http_status":    int64(obs.HTTPStatus),
		"headers_hash":   hexHash(obs.HeadersBytes),
	})
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return strings.ToLower(raw)
	}
	return strings.ToLower(u.Host)
}

func randUint64() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0
	}
	return binary.BigEndian.Uint64(b[:])
}

func randBytes8() []byte {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return b[:]
}

// extraBlockedRanges are non-routable or special-purpose ranges beyond the
// netip primitives: CGNAT 100.64/10 and the benchmarking range 198.18/15.
var extraBlockedRanges = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("198.18.0.0/15"),
}

// isPrivateIP reports whether an address is private, loopback, link-local,
// multicast, unspecified or otherwise non-routable. Hostname strings are
// resolved and validated by guardedDial before dialing; this predicate only
// ever sees parsed addresses.
func isPrivateIP(a netip.Addr) bool {
	if a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() ||
		a.IsLinkLocalMulticast() || a.IsUnspecified() || a.IsMulticast() {
		return true
	}
	for _, p := range extraBlockedRanges {
		if p.Contains(a) {
			return true
		}
	}
	return false
}
