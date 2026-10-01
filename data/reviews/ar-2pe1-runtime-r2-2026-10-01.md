# Independent review round 2: collector fetch worker runtime and daemon (ar-2pe.1)

Commit under review: `882d188` "collector: repair fetch runtime per runtime and snapshot reviews" (diff vs `57a59a0`). Round-1 report: `data/reviews/ar-2pe1-runtime-2026-10-01.md`. Normative spec: `docs/specs/delivery-storage.md` md5 `9934fad3a74ca09b749320950f963930` verified at review start and at report time (V7 changelog records all claimed fixes; the `recover_job`, `record_attempt`/stage-output, `jobs`/`job_leases` and job-state sections all match the implemented behavior).

Method: live Postgres 16 containers via testcontainers (`AR_REQUIRE_POSTGRES=1`). Probe file `probeR2_test.go` (tests `TestProbeR2*`), written, run, deleted; never committed. Because the concurrent reviewer's probe file had a compile error blocking the shared tree's package build, the probe runs and the regression ran in an isolated `git worktree` at exactly `882d188`; the shared tree was never written beyond a briefly present copy of my own probe file (removed).

## VERDICT: PASS

All three round-1 blockers are fixed and hold under adversarial live probes. Two non-blocking notes carried forward / newly observed, neither a defect introduced by this commit.

---

## Claimed fixes, verified

### B1. Hostname fetch through the production dial guard — FIXED

`guardedDial` is now `(*HTTPFetcher).guardedDial`: resolve via the injected resolver (`resolveHost` in production — literal IPs pass through, everything else through `net.DefaultResolver.LookupIPAddr`), validate EVERY resolved address with `isPrivateIP` (netip primitives plus the new CGNAT `100.64/10` and benchmarking `198.18/15` prefixes), then dial `net.JoinHostPort(addrs[0], port)` — the validated address itself, so the transport cannot re-resolve between check and connect. Empty and unparseable resolutions refuse. SNI and Host stay URL-derived (real-TLS evidence below: the certificate validated against the hostname while the socket went to the resolved IP).

Live probes (production fetcher, no injected resolve/dial, real DNS + real TLS):

```
example.com exchange: status=200 bodyLen=713                       <- hostname dial works end to end
localhost refused: Get "https://localhost/": collector: localhost resolves to private address ::1
::ffff:127.0.0.1 refused: collector: refusing private host ::ffff:127.0.0.1   <- IPv4-mapped unmap closes the bypass
redirect into private hostname refused: Get "https://priv.example/x": collector: priv.example resolves to private address 127.0.0.1
```

The redirect probe ran a real HTTPS first leg (local TLS server reached under a fake public name via stub resolver + trusted test cert; everything else production) and a 302 into a hostname resolving private — refused at the redirect leg's dial, exactly the round-1 `vercel.com` failure shape inverted. Committed `TestGuardedDialResolvesHostnames` (stub resolver + recording dialer) additionally pins: public hostname resolves and dials the validated IP only, loopback/CGNAT/benchmarking/mixed/empty resolutions refuse before any dial, literal IPs bypass resolution and are validated directly. Round-1's B1 probe shape (any hostname refused) is gone: the production fetcher fetched a real hostname over the real network.

### B2. `recover_job` idempotency by expired token — FIXED

Migration `0009_lease_recovery_marks.sql` adds `job_leases.recovered_at TIMESTAMPTZ`; `RecoverJob` marks it with a conditional `UPDATE ... WHERE recovered_at IS NULL` and advances `jobs.attempts` only when that UPDATE affected a row; `ExpiredLeases` excludes `recovered_at IS NOT NULL`.

Live race probe — 8 concurrent `RecoverJob` calls on one expired token:

```
after 8 concurrent recoveries: attempts=2 recovered_rows=1 (want 2/1)
```

All 8 derived `queued` with no error; the conditional UPDATE is the serialization point (exactly one winner). The lease is no longer listed by `ExpiredLeases` after the mark, so the 60s reconcile ticker cannot re-adopt it — the round-1 probe shape (attempts 2→3→4 with 0 attempt rows) is gone. Fresh-token recovery counts once (committed extension to `TestRecoverJobRequiresExpiredLease`, green: lease1 + recovery + lease2 + recovery2 = 4). Round 1's "verified correct" `ExpiredLeases` selection edges still hold in the suite.

### B3. Stranded-output refetch loop — FIXED on both legs

`processOne` classifies `RecordAttempt` `CodeConflict` as a durable deferral on both the success (200) path and the 304 path: `deferRetry` with detail `conflicting output for work tuple: upstream bytes changed under a stranded output; awaiting disposition` — a `deferred` attempt row naming the conflict, exponential backoff from take-ups, claimable only after `retry_not_before`.

- Success leg: committed `TestWorkerStrandedOutputChangedBytesDefers` (green live): exactly one fetch, deferred attempt row with conflict detail, no pre-deadline refetch, claimable after deadline.
- 304 leg (my probe, not covered by a committed test): stranded success output (crash before CompleteJob), refetch answers 304 — the stranded record itself provides the prior observation, so the unchanged envelope conflicts with the stranded success envelope:

```
304-stranded deferral: outcome="deferred" detail="conflicting output for work tuple: upstream bytes changed under a stranded output; awaiting disposition" retry_not_before=<after crash+backoff>
pre-deadline scan: 0 jobs, fetch count stays 1; claimable again after the deadline
```

Round-1's loop signature (fetches increasing with attempts=0, completions=0, events frozen) no longer reproduces on either leg; the disagreement is now durable, visible and rate-limited pending disposition, per spec V7's stage-output paragraph.

### N4. Disabled-source dispatch — FIXED

`ClaimableJobs` joins `sources.enabled`. Live probe: transport-deferred job, source disabled, past the retry deadline — not in the claimable set, worker scan fetches nothing (fetch count unchanged); re-enable restores dispatch. `ExpiredLeases` is deliberately NOT gated by enabled (see R2-N2) — spec V7 scopes the gate to the claimable scan.

### N3/N6 — FIXED

`isPrivateIP` covers the netip primitives plus CGNAT and benchmarking ranges (committed matrix green, real-dial refusals above); dialing the validated address closes the rebinding TOCTOU.

## Round-1 "verified correct" list re-checked

Full committed suite green in the clean worktree (see regression). Targeted live re-probes stayed green: quarantine dead-letter freeze (401 → attempts frozen across a 1h-later scan), 304-without-prior → deferred `no prior observation for unchanged` with zero observations written, `TestDaemon*` lifecycle and recover tests green, `cmd/` untouched by the diff (exit-64 path unchanged from round 1's verified state). Status classification, claimable edges, MaxInFlight/host-gate, PumpOutbox at-least-once and the HTTP-fetcher guards are covered by the green suite.

## Regression (worktree at 882d188, probe file deleted)

```
$ gofmt -l .
(no output)

$ go vet ./...
(no output)

$ AR_REQUIRE_POSTGRES=1 go test ./... -count=1
ok  github.com/guygrigsby/agent-routing/internal/collector   54.746s
ok  github.com/guygrigsby/agent-routing/internal/research    27.288s
ok  github.com/guygrigsby/agent-routing/routing               0.679s
```

One earlier full-suite run failed on `TestOfferAdvertisedKindVocabulary` with `postgres container unavailable: ... mapped port: ... context deadline exceeded` — testcontainers Docker contention while the concurrent reviewer was also spinning containers; an immediate rerun was fully green. Not a code failure.

## Notes (non-blocking)

- **R2-N1. `RecoverJob` lacks the supersession guard its own listing has.** Recovering a superseded-but-expired token (a newer lease holds `jobs.fencing_token`) still marks the old lease and advances attempts — probed live: `superseded-token recovery: attempts 2 -> 3`. `ExpiredLeases` requires `j.fencing_token = l.token`, so the runtime can never trigger this; only a direct port caller could, and that caller double-counts a take-up. The idempotency contract (replay of the same token is a no-op) holds. Suggest aligning the port with the listing's discipline in a later pass.
- **R2-N2. `ExpiredLeases` lists expired leases of disabled sources** (recovery bookkeeping proceeds; dispatch stays gated by the claimable scan). Matches spec V7's sentence, which scopes the enabled gate to the claimable scan. Recorded so nobody reads it as drift.
- **R2-N3. The conflict-deferral's durability depends on `DeferJob` succeeding; its error is ignored** (`deferRetry`, pre-existing). If the deferral write hits a transient failure, the job resurfaces claimable at lease-expiry cadence and re-conflicts until a deferral lands — bounded and self-healing, but the "awaiting disposition" evidence row can be briefly absent. Same shape as the pre-existing `CodeTemporarilyUnavailable` path; unchanged by this commit.
- **R2-N4 (carried, N7).** `jobs.attempts` counts take-ups (leases + first recoveries), `job_attempts` rows count recorded outcomes. With B2 fixed the two are no longer divergent under reconciliation; the naming overlap remains.

## Probe hygiene

Probe file `probeR2_test.go` (6 tests, prefix `TestProbeR2*`) deleted from the worktree before the final regression; worktree removed afterward; repo tree clean except the concurrent session's untracked `docs/implementation.md`. The concurrent reviewer's files were never touched; their mid-review `probe3_review_test.go` compile error is why the worktree detour exists (round-1 lesson applied both ways: their file, not mine, was left alone).

## Recommendation

Ship it. All three round-1 blockers are closed with live evidence, the spec V7 changelog matches the implementation, and the remaining notes are pre-existing or bookkeeping-level.
