# Independent review round 3 — snapshot compile/validate/publish (ar-2pe.1), commit 882d188

- Commit under review: `882d188` "collector: repair fetch runtime per runtime and snapshot repairs" (diff vs `d56c46c`). My scope: the snapshot fix only — `internal/collector/snapshot.go` (23505 handler savepoint) and `internal/collector/snapshot_test.go` (new `TestSnapshotConcurrentIdenticalCompileReplays`). The runtime files in the same commit belong to the concurrent runtime reviewer.
- Spec: `docs/specs/delivery-storage.md` is now **V7** in this commit (the V6→V7 delta is the runtime reviewer's scope; the V6 snapshot sentences are unchanged). md5 at review end `9934fad3a74ca09b749320950f963930`.
- Method: full read of the in-scope diff and the surrounding `CompileSnapshot` flow; five live probes against disposable Postgres 16 containers (probe file `internal/collector/probe3_review_test.go`, prefix `probe3_`, deleted after all runs, never committed). Two initial probe failures were probe bugs (a fix-section pre-select that reused the tuple the negative control had already committed; a scan-destination typo), corrected before concluding — same discipline as rounds 1-2.

## The claimed fix — verified

`snapshot.go:246-284`: the `snapshots` INSERT now runs inside a savepoint (`tx.Begin` nested transaction, pgx v5). On a 23505 the savepoint rolls back (`_ = sp.Rollback`), the outer transaction stays live, and the parameter-tuple re-select (`input_manifest_hash, distribution, projection_policy_version, created_at, knowledge_cutoff`) executes for real: matching tuple → replay result carrying the stored hash/locator/coverage digest; non-matching → `CodeConflict` unchanged. On success `sp.Commit` releases the savepoint and the rest of the compile (inputs, audits, event, outer commit) proceeds on the outer transaction. `defer tx.Rollback` at the function top covers the replay return path (nothing written in the outer tx; no leakage).

### Live proof, both levels

1. **Mechanism, deterministic** (`TestProbe3SavepointMechanism`): the port's exact SQL sequence replayed at pool level. Negative control first — bare insert, no savepoint: pre-select misses, winner commits an identical-tuple row, loser's INSERT → `23505`, loser's re-select → **`25P02`** (the round-2 dead-code behavior, reproduced). Then the same sequence with a savepoint: INSERT → `23505` → savepoint rollback → re-select **returns the winner's row in the still-live transaction**. The fix is exactly the round-2 prescribed shape and it works.
2. **Conflict arm, deterministic** (`TestProbe3DeterministicConflictViaHandler`): two projection policies `pp-1`/`pp-2` with identical grammar (identical artifact bytes) but different parameter tuples. First compile succeeds; second deterministically misses the pre-insert select, collides at the INSERT, and gets `CodeConflict` with the colliding hash named in the detail (which is now factually true — the parameter sets really do differ). Zero partial rows (snapshots=1, inputs=1, audits=1, events=1), pool healthy afterwards, and a `pp-1` replay after the conflict returns the winner's hash and digest intact.
3. **Port-level races**:
   - 2-way identical (`TestProbe3RaceIdenticalPair`, `-count=5`): 5/5 both callers succeed, equal hash and coverage digest, exactly one snapshots/input/audit/`SnapshotCompiled` row.
   - Committed `TestSnapshotConcurrentIdenticalCompileReplays`, `-count=5`: 5/5 green with the same invariants.
   - 3-way identical (`TestProbe3RaceThreeWay`, `-count=5`): 5/5 all three succeed, one row set.
   - Mixed identical+distinct (`TestProbe3RaceMixedIdenticalAndDistinct`, `-count=6`): two `pp-1` compiles race one `pp-2` at a barrier. Every run: exactly one snapshots row; callers whose tuple matches the stored row succeed, the others get `conflict`; child counts exactly 1/1/1. The `pp-2` row won the insert in all 6 runs (a goroutine-scheduling bias, not a correctness issue — the `pp-1` losers provably reached the INSERT and exercised the handler's conflict arm, while the 2-way/3-way identical races exercise the replay side). Both classification outcomes are live-proven.
4. **Savepoint rollback leaves no partial rows**: verified in (2) and (4) — no inputs/audits/events for the losing attempts, and the loser tx's deferred rollback returns the connection healthy.

## Blockers

None. Round-2 N1-R2 is genuinely fixed: the replay half of the 23505 classification is live code, proven at mechanism level and under port-level contention at `-count=5..6`. The V6 spec sentence "a concurrent identical compile replays the stored row" is now true on the path it names, and the conflict detail is no longer factually wrong in the concurrent-identical case (that case replays instead).

## Notes (non-blocking)

1. `_ = sp.Rollback(ctx)` swallows the rollback error. If the rollback itself failed (broken connection), the re-select would fail and the handler would fall through to `CodeConflict` — a storage failure masquerading as a permanent-class error. Same swallowing shape that existed before; low likelihood; recording so it isn't mistaken for intentional classification.
2. The savepoint wraps only the `snapshots` INSERT, and no writes precede it in the compile transaction (all earlier work is reads), so a savepoint rollback can never discard earlier writes. Placement is correct; worth preserving if writes ever move ahead of the insert.
3. The committed race test pins the contract (both callers succeed, one row set) but cannot tell which internal path served the loser — pre-insert select serialization vs 23505 replay. Fine for a contract pin; the mechanism probe covers the distinction at review level.
4. Replay returns the stored row's `coverage_digest`, not the loser's recomputed one — correct (they must be identical by construction, and the race asserts digest equality).
5. `codeOfErr(nil)` returns `temporarily_unavailable` — a runtime-helper quirk my probe hit while logging; not a product bug, out of scope.

## Regression (after probe removal)

- `gofmt -l .`: empty. `go vet ./...`: clean.
- Full `AR_REQUIRE_POSTGRES=1 go test ./... -count=1`: ok — buildgate 0.26s, **collector 58.1s**, research 29.1s, routing 0.79s, zero skips.
- Scope integrity: `git diff 882d188 -- internal/collector/snapshot.go internal/collector/snapshot_test.go` empty; tree clean except the concurrent session's untracked `docs/implementation.md`; their `probeR2_test.go` appeared and was removed by them mid-review, never touched by me.

## Verdict

The one round-2 blocker is mechanically and port-level fixed, adversarially probed across 2-way/3-way/mixed contention, with both classification arms live-proven and no partial-row leakage. Regression surface clean.

VERDICT: PASS
