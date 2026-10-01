# Independent Review: Collector Fetch Worker Runtime & Daemon

Commit under review: `57a59a0` ("collector: add fetch worker runtime and collect daemon")
Baseline: `046cf25`. Repo: `/Users/guygrigsby/projects/agent-routing`, HEAD at `57a59a0`, clean tree at review start.
Method: every behavioral claim probed against live Postgres 16 containers (`AR_REQUIRE_POSTGRES=1`), temporary probe file `internal/collector/probe_rev3runtime_test.go` (tests `TestRevR3*`), written, run, then deleted. Nothing committed. The other reviewer's files (`TestProbe*`) never touched.

## VERDICT: FAIL

Three blockers, each reproduced live. Two are correctness contracts (the daemon cannot fetch anything by hostname; `recover_job` is not idempotent); one is a livelock (stranded-output refetch loop with zero durable progress).

---

## Blockers

### B1. Production dial guard refuses every DNS hostname (fetch path is broken end to end)

`httpfetch.go: guardedDial` applies `isPrivateAddr` to the *host string* before resolution, and `isPrivateAddr` returns `true` when `netip.ParseAddr` fails. Any non-IP-literal host is therefore classified "private" and refused. The daemon wires only `NewHTTPFetcher` (grep-verified; `NewLocalTestFetcher` is test-only), so the production collector can only fetch literal-IP HTTPS URLs. The `LookupIPAddr` validation loop is dead code for hostnames.

Live probe (`TestRevR3GuardedDialHostname`):

```
hostname target refused: collector: refusing private address vercel.com
localhost refused: collector: refusing private address localhost
strict fetch hostname URL: Get "https://vercel.com/x": collector: refusing private address vercel.com
```

Consequence: every real source (provider APIs, vercel catalog, models.dev) fails at dial before any I/O, classified `deferred` as a transport error, retried forever at MaxBackoff. The committed suite cannot see this because every test injects `NewLocalTestFetcher` against `httptest` (127.0.0.1) — there is no test that drives the production fetcher at a hostname. That coverage gap is how this shipped green.

Fix direction: resolve first (`LookupIPAddr`), validate the resolved IPs (all of them), then dial the validated IP with `Host`/SNI preserved; or parse-IP the host and only run the private check on literal IPs + resolved addresses. Fold in N3/N6 at the same time.

### B2. `recover_job` is not idempotent by expired token; reconciliation re-adopts the same lease every sweep and inflates `attempts` indefinitely

contracts.md: "`recover_job` … Idempotent by expired token or new work tuple." Implementation (`ports.go RecoverJob`): unconditional `UPDATE jobs SET attempts = attempts + 1`, no replay guard, no durable trace — and `runtime.go ExpiredLeases` lists a lease when the job has "no attempt recorded under it". Since `RecoverJob` writes no attempt row and `fencing_token` never changes on recovery, the identical (job, token) pair re-qualifies on every sweep.

Live probe (`TestRevR3ClaimableAndRecovery`):

```
recovery of the SAME expired token: attempts 2 -> 3 -> 4, attempt rows still 0
BLOCKER CONFIRMED: replaying recover_job on the same expired token advanced attempts 3 -> 4
  (contracts.md: idempotent by expired token); attempt rows=0 so ExpiredLeases re-lists the lease every sweep
```

Amplification: the daemon sweeps on a ticker (default 60s, `daemon.go:141`, `reconcileEvery = time.Minute`). A stranded lease with no worker to claim it gains +1 `attempts` per minute, forever. `deferRetry` computes backoff as `w.backoff(job.Attempts)` (`runtime.go:422`), so inflated attempts pin retries at MaxBackoff immediately. (The counter's semantics are "take-ups": `LeaseJob` also increments it, `ports.go:164` — see N7.)

Fix direction: make recovery replay-safe — e.g. record the recovery (attempt-style row keyed by token, or a `recovered_at`/`recovered_token` column on the lease/job) so a second sweep with the same expired token is a no-op returning the same derived state.

### B3. Stranded-output refetch loop: crash between RecordAttempt and CompleteJob with changed bytes refetches forever with zero durable progress

Crash window: lease → `RecordAttempt` (stage_outputs + observation committed) → process dies before `CompleteJob`. No attempt row exists, so `ExpiredLeases` recovers; the job re-enters the claimable scan. On refetch with a *changed* body, `RecordAttempt` returns `conflict` (different canonical output for the same work tuple), which `runtime.go processOne` swallows (`_ = err`). Result: the job is immediately claimable again — no backoff (the conflict path never reaches `deferRetry`), no dead-letter, no attempt row, no event — and every scan refetches.

Live probe (`TestRevR3StrandedOutputLoop`), scripted source changed B1→B2 after the crash:

```
after refetch: fetches=1 attempts=0 completions=0 events=1 (first loop)
after second scan: fetches=2 attempts=0 completions=0 events=1
CONFIRMED: refetch loop with zero durable progress (attempts=0 completions=0 events still=1)
```

With the default 1s scan cadence this is an unbounded refetch loop against the upstream — the exact behavior collection-pipeline.md's operating-limits section exists to prevent. Note the contrast: with an *unchanged* body the adopt path completes the job correctly (verified separately).

Fix direction: a conflict on an already-stranded output is a distinct condition. Minimum: defer with backoff and emit evidence (attempt row or event) so the loop is visible and rate-limited; the resolution (quarantine for manual disposition vs. force-adopt newest vs. supersede the stale output) is a spec decision — delivery-storage.md's adopt-path sentence does not cover the changed-bytes-after-crash case.

---

## Notes (non-blocking)

- N1. Retry-After is `max(backoff, header)`, not a strict override: a header *smaller* than the computed backoff is ignored. Both header forms are honored — delta-seconds and HTTP-date, proven over the wire: `retry_not_before - observed_at = 2h0m0s (header 2h, backoff 1h)`; HTTP-date end-to-end landed within 1s of the parsed instant. Defensible (conservative), but not what "Retry-After takes precedence" would promise.
- N2. 451 is classified `deferred` (transport-error family), yet the observations CHECK pairs `source_blocked` with 451 — a pair the worker can never produce since 451 writes no observation. Latent inconsistency; currently unreachable.
- N3. `isPrivateAddr` misses CGNAT 100.64/10 and benchmarking 198.18/15. Irrelevant while B1 stands; matters the moment it's fixed.
- N4. Jobs of disabled sources remain claimable — `ClaimableJobs` does not join `sources.enabled`. Registering a source toggled off does not stop its in-flight work.
- N5. `HostGate` is per-process. "Shared per-host operating limit across all workers" (collection-pipeline.md) holds only within one daemon; N daemons each send 1 request per interval. Single-daemon deployment assumption not stated anywhere.
- N6. `guardedDial` has a DNS-rebinding TOCTOU window (resolve → validate → dial re-resolves). Moot while B1 stands; the B1 fix should dial the validated IP directly.
- N7. `jobs.attempts` counts take-ups (leases + recoveries), not fetch attempts; `job_attempts` rows count recorded outcomes. Two counters, one name. B2's inflation lands on the former and distorts `deferRetry`'s backoff exponent.
- N8. `DeferJob` `ON CONFLICT DO NOTHING` silently swallows a replay with a different detail/retry_not_before (no error, no event) — asymmetric with `RecordAttempt`, which conflicts on different output. Replay-with-divergence is silently dropped.
- N9. Test-quality: suite greenliness masks B1 (see there). Also `daemon_test.go` lifecycle coverage is good (exit-on-cancel, startup reconcile evidence, exit 64) but never exercises the production fetcher.

## Verified correct (clean hunts)

All probed live unless noted:

- **Status classification end-to-end** through the real worker loop: 200 → completed + `success` observation; 503/429 → deferred (`status 503` / `status 429`), re-claimable after deadline; 404/410 → `failed_terminal` (`resource gone: …`), never claimable; 401/403 → `quarantined` (`access blocked: …`); transport error → deferred with the error as detail.
- **304 chain**: with prior success observation → `unchanged` observation, `unchanged_observations` row, `JobCompleted`, exactly one `StageOutputRecorded` + one `ObservationRecorded` event; envelope honest (`body_hash = sha256("") = e3b0c442…`, `http_status 304`); empty-body blob row exists (NOT NULL satisfied) exactly once. Without prior → deferred (`no prior observation for unchanged`), zero observations written.
- **Claimable scan adversarial edges**: expired lease + deferred-until-future → not claimable before deadline, claimable after; `LeaseJob` before deadline → conflict (`job … is deferred until …`).
- **ExpiredLeases selection**: expired lease without attempts → listed; with attempt evidence under the token → skipped; completed → skipped; superseded (non-latest) lease → not listed.
- **Quarantine dead-letter**: quarantined job's attempts frozen across sweeps, never claimable again, never re-adopted.
- **Concurrency**: MaxInFlight bounded (peak exactly 2 with limit 2 over 6 jobs); host gate spacing ≥80ms same-host with cross-host traffic immediate (+3ms); ctx-cancel mid-fetch → `Run` returns, no attempt row, lease expires, reconciliation adopts, job reclaimable (no leak).
- **PumpOutbox at-least-once under handler failure mid-batch**: 2 receipted, failure on 3rd stops the batch, next pump redelivers exactly the unreceipted 7, event table unmutated.
- **Daemon**: malformed TOML propagates (`toml: line 1 (last key "dsn"): expected value but found "not" instead`); re-`Migrate` on a migrated DB is a no-op (8 migrations, no error); no-DSN error names all three sources (`no collector DSN: pass --dsn, set AGENT_ROUTING_DSN or set dsn in config.toml`); `collect` with no DSN exits **64** through the real binary (`go run ./cmd/agent-routing collect` → `exit status 64`).
- **HTTP fetcher guards** (live server): User-Agent on the wire; body cap enforced on both the Content-Length path (rejected before read) and the streamed path (LimitReader); `http://` refused in production mode, allowed only test mode; literal private IP refused at validation.

## Probe transcripts (key lines)

See the live lines quoted inline in B1/B2/B3 and the note lines above. Full run of the probe suite (before deletion): 8/8 `TestRevR3*` tests PASS, two of them logging the blocker confirmations.

## Regression (verbatim, clean tree after probe deletion)

```
$ gofmt -l .
(no output)

$ go vet ./...
(no output)

$ AR_REQUIRE_POSTGRES=1 go test ./... -skip 'TestProbe' -count=1
?   	github.com/guygrigsby/agent-routing/cmd/agent-routing	[no test files]
ok  	github.com/guygrigsby/agent-routing/internal/buildgate	0.180s
ok  	github.com/guygrigsby/agent-routing/internal/collector	55.306s
ok  	github.com/guygrigsby/agent-routing/internal/research	29.317s
ok  	github.com/guygrigsby/agent-routing/routing	0.611s
```

Probe-file deletion verified (`ls internal/collector/probe_*` → no such file; `git status` shows only the concurrent session's untracked files). Concurrent-session churn noted: `probe_review3_test.go` absent at baseline, a `probe2_review_test.go` appeared and disappeared mid-review; this review never touched either.

## Recommendation

Do not ship this slice as-is. B1 must land before any real source registration (the collector is inert against hostnames today). B2 and B3 are contract/livelock defects that get worse with deployment time (60s and 1s cadences respectively). All three have concrete, small fix surfaces; B3 needs a one-line spec decision in delivery-storage.md first (what a changed-output conflict on a stranded job means).
