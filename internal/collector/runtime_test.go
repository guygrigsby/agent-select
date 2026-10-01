package collector

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// scriptedFetcher returns queued results in order, recording call times.
type scriptedFetcher struct {
	mu       sync.Mutex
	results  []FetchResult
	errs     []error
	calls    []time.Time
	realTime bool
}

func (f *scriptedFetcher) Fetch(ctx context.Context, sourceID, resourceURL string) (FetchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, time.Now())
	if len(f.errs) > 0 {
		err := f.errs[0]
		if len(f.results) > 0 {
			f.results = f.results[1:]
		}
		if err != nil {
			f.errs = f.errs[1:]
			return FetchResult{}, err
		}
	}
	if len(f.results) == 0 {
		return FetchResult{}, errors.New("scriptedFetcher exhausted")
	}
	r := f.results[0]
	f.results = f.results[1:]
	if len(f.errs) > 0 {
		f.errs = f.errs[1:]
	}
	return r, nil
}

func (f *scriptedFetcher) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func newTestWorker(t *testing.T, p *Ports, f Fetcher, cfg WorkerConfig) (*Worker, *func(time.Time)) {
	t.Helper()
	cfg = cfg.withDefaults()
	cfg.MinHostInterval = time.Millisecond
	cfg.LeaseDuration = 500 * time.Millisecond
	cfg.BaseBackoff = time.Second
	cfg.MaxBackoff = 2 * time.Second
	w := NewWorker(p, f, cfg)
	current := w.now
	set := func(t2 time.Time) {
		w.now = func() time.Time { return t2 }
	}
	_ = current
	return w, &set
}

func parseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func seedFetchJob(t *testing.T, p *Ports, ctx context.Context, sourceID string) string {
	t.Helper()
	registerSource(t, p, sourceID, "provider_api", true)
	r, err := p.EnqueueJob(ctx, sourceID, "https://"+sourceID+"/r", "fetch",
		hash32([]byte("w-"+sourceID)), "p1", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	return r.JobID
}

func TestWorkerFetchSuccessLifecycle(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID := seedFetchJob(t, p, ctx, "wsrc1")
	f := &scriptedFetcher{results: []FetchResult{{
		HTTPStatus: 200, Body: []byte(`{"ok":true}`), Headers: []byte("etag: x\n"),
	}}}
	w, _ := newTestWorker(t, p, f, WorkerConfig{})
	if n, err := w.runOnce(ctx); err != nil || n != 1 {
		t.Fatalf("runOnce n=%d err=%v", n, err)
	}
	state, err := p.JobState(ctx, jobID, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if state != "completed" {
		t.Fatalf("job state %s", state)
	}
	if got := countEventsName(t, p, "", "JobCompleted"); got != 1 {
		t.Fatalf("JobCompleted events = %d", got)
	}
	if got := countEventsName(t, p, "", "StageOutputRecorded"); got != 1 {
		t.Fatalf("StageOutputRecorded events = %d", got)
	}
	// Completed work is never rescanned.
	jobs, err := p.ClaimableJobs(ctx, time.Now().UTC(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("completed job still claimable: %+v", jobs)
	}
	// Second run does not re-fetch.
	if n, err := w.runOnce(ctx); err != nil || n != 0 {
		t.Fatalf("second runOnce n=%d err=%v", n, err)
	}
	if f.callCount() != 1 {
		t.Fatalf("fetches = %d", f.callCount())
	}
}

func TestWorkerDeferAndRetry(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID := seedFetchJob(t, p, ctx, "wsrc2")
	f := &scriptedFetcher{results: []FetchResult{
		{HTTPStatus: 503, Body: nil, Headers: []byte("retry-after: 0\n")},
		{HTTPStatus: 200, Body: []byte("second"), Headers: []byte("etag: y\n")},
	}}
	w, setClock := newTestWorker(t, p, f, WorkerConfig{})
	base := time.Now().UTC().Truncate(time.Millisecond)
	w.now = func() time.Time { return base }
	_ = setClock
	if n, err := w.runOnce(ctx); err != nil || n != 1 {
		t.Fatalf("runOnce n=%d err=%v", n, err)
	}
	if state, _ := p.JobState(ctx, jobID, base); state != "deferred" {
		t.Fatalf("state after 503: %s", state)
	}
	// Before the retry deadline the job is not claimable.
	jobs, err := p.ClaimableJobs(ctx, base.Add(500*time.Millisecond), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("prematurely claimable: %+v", jobs)
	}
	// After the deadline the same job re-enters and completes.
	w.now = func() time.Time { return base.Add(3 * time.Second) }
	if n, err := w.runOnce(ctx); err != nil || n != 1 {
		t.Fatalf("retry runOnce n=%d err=%v", n, err)
	}
	if state, _ := p.JobState(ctx, jobID, base.Add(3*time.Second)); state != "completed" {
		t.Fatalf("state after retry: %s", state)
	}
	// The failed attempt left durable defer evidence; success is evidenced
	// by the recorded observation (attempt rows carry deferrals only).
	var outcome, detail string
	if err := p.Pool.QueryRow(ctx, `
		SELECT outcome, detail FROM job_attempts WHERE job_id = $1
		ORDER BY observed_at DESC, token DESC LIMIT 1`, jobID).Scan(&outcome, &detail); err != nil {
		t.Fatal(err)
	}
	if outcome != "deferred" || detail != "status 503" {
		t.Fatalf("attempt evidence %s (%s)", outcome, detail)
	}
	var obsCount int
	if err := p.Pool.QueryRow(ctx, `
		SELECT count(*) FROM observations o
		JOIN stage_output_links l ON l.observation_id = o.id
		WHERE l.job_id = $1 AND o.outcome = 'success'`, jobID).Scan(&obsCount); err != nil {
		t.Fatal(err)
	}
	if obsCount != 1 {
		t.Fatalf("success observations linked to job = %d", obsCount)
	}
	var deferDetail string
	if err := p.Pool.QueryRow(ctx, `
		SELECT detail FROM job_attempts WHERE job_id = $1 AND outcome = 'deferred'`,
		jobID).Scan(&deferDetail); err != nil {
		t.Fatal(err)
	}
	if deferDetail != "status 503" {
		t.Fatalf("defer detail %q", deferDetail)
	}
}

func TestWorkerRetryAfterHeader(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	seedFetchJob(t, p, ctx, "wsrc3")
	f := &scriptedFetcher{results: []FetchResult{
		{HTTPStatus: 429, RetryAfter: 2 * time.Hour},
	}}
	w, _ := newTestWorker(t, p, f, WorkerConfig{})
	base := time.Now().UTC().Truncate(time.Millisecond)
	w.now = func() time.Time { return base }
	if _, err := w.runOnce(ctx); err != nil {
		t.Fatal(err)
	}
	var retryNotBefore time.Time
	if err := p.Pool.QueryRow(ctx, `
		SELECT retry_not_before FROM job_attempts
		WHERE outcome = 'deferred' ORDER BY observed_at DESC LIMIT 1`).
		Scan(&retryNotBefore); err != nil {
		t.Fatal(err)
	}
	if retryNotBefore.Before(base.Add(2 * time.Hour)) {
		t.Fatalf("retry_not_before %v ignores Retry-After 2h", retryNotBefore)
	}
	// Even past the jitter window the job stays parked until Retry-After.
	jobs, err := p.ClaimableJobs(ctx, base.Add(3*time.Second), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("claimable before Retry-After: %+v", jobs)
	}
}

func TestWorkerAccessBlockedAndReentry(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID := seedFetchJob(t, p, ctx, "wsrc4")
	f := &scriptedFetcher{results: []FetchResult{{HTTPStatus: 403}}}
	w, _ := newTestWorker(t, p, f, WorkerConfig{})
	if _, err := w.runOnce(ctx); err != nil {
		t.Fatal(err)
	}
	state, err := p.JobState(ctx, jobID, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if state != "quarantined" {
		t.Fatalf("state after 403: %s", state)
	}
	jobs, err := p.ClaimableJobs(ctx, time.Now().UTC().Add(time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("quarantined job claimable: %+v", jobs)
	}
	// Reentry requires a changed work tuple: a new job row.
	if _, err := p.ReenterJob(ctx, jobID, hash32([]byte("w-changed")), "p2", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	jobs, err = p.ClaimableJobs(ctx, time.Now().UTC(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].JobID == jobID {
		t.Fatalf("reentry did not produce a claimable job: %+v", jobs)
	}
}

func TestWorkerUnchanged304(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	sourceID := "wsrc5"
	// Seed the source and a prior successful observation directly.
	now := time.Now().UTC()
	registerSource(t, p, sourceID, "provider_api", true)
	priorBody := []byte(`{"v":1}`)
	mustExec(t, p.Pool, `INSERT INTO content_blobs (hash,bytes) VALUES ($1,$2)`,
		hash32(priorBody), priorBody)
	mustExec(t, p.Pool, `
		INSERT INTO observations (id,source_id,resource_url,retrieved_at,http_status,outcome,
		  body_hash,headers_bytes,collector_version)
		VALUES ('obs-prior',$1,$2,$3,200,'success',$4,'','cv')`,
		sourceID, "https://"+sourceID+"/r", now.Add(-time.Hour), hash32(priorBody))
	r, err := p.EnqueueJob(ctx, sourceID, "https://"+sourceID+"/r", "fetch",
		hash32([]byte("w-"+sourceID)), "p1", now)
	if err != nil {
		t.Fatal(err)
	}
	_ = r
	f := &scriptedFetcher{results: []FetchResult{{
		HTTPStatus: 304, Body: nil, Headers: []byte("etag: x\n"),
	}}}
	w, _ := newTestWorker(t, p, f, WorkerConfig{})
	if n, err := w.runOnce(ctx); err != nil || n != 1 {
		t.Fatalf("runOnce n=%d err=%v", n, err)
	}
	var outcome string
	var prior string
	if err := p.Pool.QueryRow(ctx, `
		SELECT outcome, prior_observation_id FROM unchanged_observations u
		JOIN observations o ON o.id = u.observation_id
		ORDER BY o.retrieved_at DESC LIMIT 1`).Scan(&outcome, &prior); err != nil {
		t.Fatal(err)
	}
	if outcome != "unchanged" || prior != "obs-prior" {
		t.Fatalf("unchanged chain %s -> %s", outcome, prior)
	}
}

func TestReconcileExpiredLeases(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID := seedFetchJob(t, p, ctx, "wsrc6")
	now := time.Now().UTC()
	if _, err := p.LeaseJob(ctx, jobID, "w-dead", time.Second, now); err != nil {
		t.Fatal(err)
	}
	// Live lease: not claimable, not recoverable.
	if _, err := ReconcileExpiredLeases(ctx, p, now, 10); err != nil {
		t.Fatal(err)
	}
	jobs, err := p.ClaimableJobs(ctx, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("live lease claimable: %+v", jobs)
	}
	later := now.Add(2 * time.Second)
	adopted, err := ReconcileExpiredLeases(ctx, p, later, 10)
	if err != nil {
		t.Fatal(err)
	}
	if adopted != 1 {
		t.Fatalf("adopted = %d", adopted)
	}
	jobs, err = p.ClaimableJobs(ctx, later, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].JobID != jobID {
		t.Fatalf("reconciled job not claimable: %+v", jobs)
	}
	// Once the job completes, the sweep has nothing left to adopt.
	f := &scriptedFetcher{results: []FetchResult{{
		HTTPStatus: 200, Body: []byte("b"), Headers: []byte("etag: e\n"),
	}}}
	w, _ := newTestWorker(t, p, f, WorkerConfig{})
	w.now = func() time.Time { return later }
	if _, err := w.runOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if s, _ := p.JobState(ctx, jobID, later); s != "completed" {
		t.Fatalf("post-reconcile state %s", s)
	}
	if adopted, err := ReconcileExpiredLeases(ctx, p, later, 10); err != nil || adopted != 0 {
		t.Fatalf("second reconcile adopted=%d err=%v", adopted, err)
	}
}

func TestPumpOutboxAtLeastOnce(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	// Two completed fetch jobs produce four events.
	for i := range 2 {
		src := fmt.Sprintf("wsrc7-%d", i)
		jobID := seedFetchJob(t, p, ctx, src)
		f := &scriptedFetcher{results: []FetchResult{{
			HTTPStatus: 200, Body: []byte("b"), Headers: []byte("etag: e\n"),
		}}}
		w, _ := newTestWorker(t, p, f, WorkerConfig{})
		if _, err := w.runOnce(ctx); err != nil {
			t.Fatal(err)
		}
		if s, _ := p.JobState(ctx, jobID, time.Now().UTC()); s != "completed" {
			t.Fatalf("job %s state %s", jobID, s)
		}
	}
	total := countEvents(t, p)
	if total < 4 {
		t.Fatalf("events = %d", total)
	}
	// A failing handler receives the first event but records no receipt.
	calls := 0
	if _, err := p.PumpOutbox(ctx, "c1", 10, func(OutboxEvent) error {
		calls++
		return errors.New("boom")
	}); err == nil {
		t.Fatal("handler error swallowed")
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d", calls)
	}
	// The failed event redelivers on the next pump; all events then receipt.
	var seen []string
	delivered, err := p.PumpOutbox(ctx, "c1", 10, func(e OutboxEvent) error {
		seen = append(seen, e.EventName)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if delivered != total {
		t.Fatalf("delivered = %d want %d", delivered, total)
	}
	// A second pump delivers nothing new.
	if delivered, err := p.PumpOutbox(ctx, "c1", 10, func(OutboxEvent) error {
		t.Fatal("unexpected delivery")
		return nil
	}); err != nil || delivered != 0 {
		t.Fatalf("second pump delivered=%d err=%v", delivered, err)
	}
	// A distinct consumer still receives everything (broadcast semantics).
	delivered, err = p.PumpOutbox(ctx, "c2", 10, func(OutboxEvent) error { return nil })
	if err != nil || delivered != total {
		t.Fatalf("c2 delivered=%d want %d err=%v", delivered, total, err)
	}
}

func TestWorkerHostGateSpacing(t *testing.T) {
	gate := NewHostGate(50 * time.Millisecond)
	ctx := context.Background()
	start := time.Now()
	for range 3 {
		if err := gate.Wait(ctx, "example.com"); err != nil {
			t.Fatal(err)
		}
	}
	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Fatalf("three waits on one host took %v, want >= 100ms", elapsed)
	}
	// A different host is not serialized behind the first.
	if err := gate.Wait(ctx, "other.example"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerTerminal404(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID := seedFetchJob(t, p, ctx, "wsrc8")
	f := &scriptedFetcher{results: []FetchResult{{HTTPStatus: 404}}}
	w, _ := newTestWorker(t, p, f, WorkerConfig{})
	if _, err := w.runOnce(ctx); err != nil {
		t.Fatal(err)
	}
	var outcome string
	if err := p.Pool.QueryRow(ctx, `
		SELECT outcome FROM job_attempts WHERE job_id = $1
		ORDER BY observed_at DESC LIMIT 1`, jobID).Scan(&outcome); err != nil {
		t.Fatal(err)
	}
	if outcome != "failed_terminal" {
		t.Fatalf("attempt outcome after 404: %s", outcome)
	}
	state, err := p.JobState(ctx, jobID, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if state != "quarantined" {
		t.Fatalf("state after 404: %s", state)
	}
	jobs, err := p.ClaimableJobs(ctx, time.Now().UTC().Add(time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("terminal job claimable: %+v", jobs)
	}
}

func TestHTTPFetcherGuards(t *testing.T) {
	// Retry-After parsing, both forms.
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	if d := parseRetryAfter("120", now); d != 2*time.Minute {
		t.Fatalf("seconds form: %v", d)
	}
	if d := parseRetryAfter("Wed, 21 Oct 2026 07:28:00 GMT", now); d != 475*time.Hour+28*time.Minute {
		t.Fatalf("date form: %v", d)
	}
	if d := parseRetryAfter("garbage", now); d != 0 {
		t.Fatalf("garbage: %v", d)
	}

	// Production fetcher refuses http and private hosts before any dial.
	strict := NewHTTPFetcher("test-agent")
	if _, err := strict.Fetch(context.Background(), "s", "http://example.com/r"); err == nil {
		t.Fatal("http scheme accepted")
	}
	if _, err := strict.Fetch(context.Background(), "s", "https://127.0.0.1/r"); err == nil {
		t.Fatal("loopback accepted")
	}
	if _, err := strict.Fetch(context.Background(), "s", "https://10.0.0.1/r"); err == nil {
		t.Fatal("private accepted")
	}
	for ip, want := range map[string]bool{
		"::1": true, "169.254.1.1": true, "93.184.216.34": false,
		"100.64.1.1": true, "198.18.5.5": true, "198.16.0.1": false, "100.63.0.1": false,
		"192.0.2.1": false, "fe80::1": true, "fd00::1": true,
	} {
		a := netip.MustParseAddr(ip)
		if got := isPrivateIP(a); got != want {
			t.Fatalf("isPrivateIP(%s) = %v, want %v", ip, got, want)
		}
	}

	// Body cap and live fetch in test mode.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("empty User-Agent")
		}
		w.Write([]byte("payload"))
	}))
	defer srv.Close()
	loose := NewHTTPFetcher("test-agent")
	loose.InsecureAllowLocal = true
	// NewHTTPFetcher built the client before the flag flip; rebuild transport.
	loose.Client = srv.Client()
	loose.Client.Timeout = loose.Timeout
	res, err := loose.Fetch(context.Background(), "s", srv.URL+"/r")
	if err != nil {
		t.Fatal(err)
	}
	if res.HTTPStatus != 200 || string(res.Body) != "payload" {
		t.Fatalf("live fetch %+v", res)
	}
	loose.MaxBody = 3
	res, err = loose.Fetch(context.Background(), "s", srv.URL+"/r")
	if err == nil || res.HTTPStatus != 200 {
		t.Fatalf("oversized body accepted: %+v %v", res, err)
	}

	// Redirect validation: strict mode rejects a private redirect target.
	if err := strict.checkRedirect(&http.Request{
		URL: parseURL("https://127.0.0.1/x"),
	}, nil); err == nil {
		t.Fatal("private redirect accepted")
	}
	if err := strict.checkRedirect(&http.Request{
		URL: parseURL("https://example.com/x"),
	}, nil); err != nil {
		t.Fatalf("public redirect rejected: %v", err)
	}
}

// TestGuardedDialResolvesHostnames pins the production dial guard's hostname
// path: a hostname is resolved and every resolved address validated (review
// round 1 shipped a guard that refused all hostnames as unparseable), the
// dial goes to the validated address itself (no re-resolution window), and
// CGNAT/benchmarking ranges are blocked alongside the classic private ones.
func TestGuardedDialResolvesHostnames(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name     string
		resolve  []string
		wantDial bool
		wantErr  string
	}{
		{"public hostname resolves and dials", []string{"93.184.216.34"}, true, ""},
		{"hostname to loopback refused", []string{"127.0.0.1"}, false, "resolves to private address"},
		{"hostname to cgnat refused", []string{"100.64.1.1"}, false, "resolves to private address"},
		{"hostname to benchmarking refused", []string{"198.18.0.7"}, false, "resolves to private address"},
		{"mixed resolution with one private refused", []string{"93.184.216.34", "10.0.0.5"}, false, "resolves to private address"},
		{"empty resolution refused", nil, false, "no addresses"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := NewHTTPFetcher("test-agent")
			addrs := make([]netip.Addr, 0, len(tc.resolve))
			for _, s := range tc.resolve {
				addrs = append(addrs, netip.MustParseAddr(s))
			}
			f.resolve = func(context.Context, string) ([]netip.Addr, error) {
				return addrs, nil
			}
			var dialed []string
			f.dial = func(_ context.Context, _, addr string) (net.Conn, error) {
				dialed = append(dialed, addr)
				return nil, fmt.Errorf("stub dial")
			}
			_, err := f.guardedDial(ctx, "tcp", "catalog.example:443")
			if tc.wantDial {
				if err == nil || err.Error() != "stub dial" {
					t.Fatalf("expected stub dial error, got %v", err)
				}
				if len(dialed) != 1 || dialed[0] != net.JoinHostPort(tc.resolve[0], "443") {
					t.Fatalf("dialed %v, want the validated address %s:443", dialed, tc.resolve[0])
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected refusal containing %q, got %v", tc.wantErr, err)
			}
			if len(dialed) != 0 {
				t.Fatalf("dialed despite refusal: %v", dialed)
			}
		})
	}

	// Literal IPs bypass resolution and are validated directly.
	f := NewHTTPFetcher("test-agent")
	f.resolve = func(_ context.Context, host string) ([]netip.Addr, error) {
		if ip, err := netip.ParseAddr(host); err == nil {
			return []netip.Addr{ip.Unmap()}, nil // same passthrough as resolveHost
		}
		return nil, fmt.Errorf("unexpected hostname resolution %q", host)
	}
	f.dial = func(_ context.Context, _, addr string) (net.Conn, error) {
		return nil, fmt.Errorf("dialled %s", addr)
	}
	if _, err := f.guardedDial(ctx, "tcp", "10.1.2.3:443"); err == nil ||
		!strings.Contains(err.Error(), "resolves to private address") {
		t.Fatalf("literal private IP not refused: %v", err)
	}
	_, err := f.guardedDial(ctx, "tcp", "93.184.216.34:443")
	if err == nil || !strings.Contains(err.Error(), "93.184.216.34:443") {
		t.Fatalf("literal public IP not dialed: %v", err)
	}
}

// TestWorkerStrandedOutputChangedBytesDefers pins the changed-bytes-after-crash
// path: a stranded stage output (crash between RecordAttempt and CompleteJob)
// followed by an upstream change conflicts on output identity. The worker must
// defer durably with backoff (round 1 shipped an unbounded refetch loop with
// zero durable progress: no attempt row, no event, immediately claimable).
func TestWorkerStrandedOutputChangedBytesDefers(t *testing.T) {
	p := mustPort(t)
	ctx := context.Background()
	jobID, token, obs, output := leaseFetchJob(t, p, "stranded")
	if _, err := p.RecordAttempt(ctx, RecordAttemptParams{
		JobID: jobID, Token: token, CollectorVersion: "cv-test",
		RecordedAt: obs.RetrievedAt, Observation: obs, OutputBytes: output,
	}); err != nil {
		t.Fatal(err)
	}
	// Crash: no CompleteJob. The lease expires; the job re-enters the
	// claimable scan with the stranded output link in place.
	later := obs.RetrievedAt.Add(2 * time.Minute)
	f := &scriptedFetcher{results: []FetchResult{{
		HTTPStatus: 200, Body: []byte("body-v2-upstream-changed"), Headers: []byte("h2"),
	}}}
	w, setNow := newTestWorker(t, p, f, WorkerConfig{})
	(*setNow)(later)
	if n, err := w.runOnce(ctx); err != nil || n != 1 {
		t.Fatalf("runOnce n=%d err=%v", n, err)
	}
	if f.callCount() != 1 {
		t.Fatalf("fetches = %d", f.callCount())
	}
	// The disagreement is durable and visible: a deferred attempt row whose
	// detail names the conflict.
	var outcome, detail string
	var retryNotBefore time.Time
	if err := p.Pool.QueryRow(ctx, `
		SELECT a.outcome, a.detail, a.retry_not_before
		FROM job_attempts a WHERE a.job_id = $1
		ORDER BY a.observed_at DESC, a.token DESC LIMIT 1`, jobID).
		Scan(&outcome, &detail, &retryNotBefore); err != nil {
		t.Fatal(err)
	}
	if outcome != "deferred" || !strings.Contains(detail, "conflicting output") {
		t.Fatalf("latest attempt %q %q", outcome, detail)
	}
	if !retryNotBefore.After(later) {
		t.Fatalf("no backoff: retry_not_before %v not after %v", retryNotBefore, later)
	}
	// Rate-limited: scans before the deadline do not refetch.
	(*setNow)(later.Add(500 * time.Millisecond))
	if n, err := w.runOnce(ctx); err != nil || n != 0 {
		t.Fatalf("pre-deadline runOnce n=%d err=%v (refetch loop)", n, err)
	}
	if f.callCount() != 1 {
		t.Fatalf("refetched before deadline: %d", f.callCount())
	}
	// After the deadline the job is claimable again: retries proceed under
	// backoff rather than a 1 Hz loop.
	(*setNow)(retryNotBefore.Add(time.Second))
	jobs, err := p.ClaimableJobs(ctx, retryNotBefore.Add(time.Second), 10)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, j := range jobs {
		if j.JobID == jobID {
			found = true
		}
	}
	if !found {
		t.Fatal("job not claimable after retry deadline")
	}
}
