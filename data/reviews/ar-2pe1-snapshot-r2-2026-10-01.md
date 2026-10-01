# Independent review round 2 — snapshot compile/validate/publish (ar-2pe.1), commit d56c46c

- Commit under review: `d56c46c` "collector: repair snapshot projection per review round 1" (diff vs `57a59a0`; HEAD). In-scope files: `internal/collector/snapshot.go`, `internal/collector/snapshot_test.go`, `docs/specs/delivery-storage.md`.
- Spec: delivery-storage.md **V6**, md5 `5b2e478f1e4be235f0a315896b75119c` (verified at review start and end).
- Method: full read of the fix diff, port code, DDL and validator; 22 live probes against disposable Postgres 16 containers (probe file `internal/collector/probe2_review_test.go`, prefix `probe2_`, deleted after all runs, never committed). All claimed fixes exercised against live Postgres, plus adversarial arms: the full min/exact/max fold interaction matrix, a label-vs-class basis discriminator, a 16-arm signal boundary table, a concurrent-compile race, and the port's exact 23505-handler SQL sequence inside a real aborted transaction.
- Two initial probe failures were probe bugs (illegal limit tuples violating `limit_assertions_check`; `seedSnapshotOffer`'s hardcoded `of1` colliding across attempts; an unparameterized UPDATE), corrected before concluding — same round-2 lesson as last time.

Regression (after probe removal): `gofmt -l .` empty; `go vet ./...` clean; full `AR_REQUIRE_POSTGRES=1 go test ./... -count=1` ok (collector 55.1s, research 29.0s, buildgate 0.22s, routing 0.35s, zero skips). The seven claimed-fix committed tests pass as a block in 7.6s.

## BLOCKERS

### N1-R2 — The 23505 handler's replay re-select is dead code: a concurrent identical compile returns `conflict`, never replaying, contradicting the V6 spec sentence and the code's own comment

`snapshot.go:251-272`: after the `snapshots` INSERT fails with 23505, the handler re-selects by the full parameter tuple **inside the same transaction** to distinguish "concurrent identical compile → replay" from "different parameters, identical bytes → conflict".

Postgres aborts a transaction after any statement error: the re-select can never run. Live proof, both levels:

1. **Mechanism** (the port's exact SQL sequence replayed in a real transaction): loser's replay-select → 0 rows; winner commits the identical-parameter row; loser's INSERT → `23505`; loser's re-select → **PG 25P02 "current transaction is aborted, commands ignored until end of transaction block"** — not `ErrNoRows`, not a row. `err2 != nil` always, so the handler always falls through to `CodeConflict`.
2. **Port-level race**: 8 attempts of two concurrent identical `CompileSnapshot` calls (fresh source per attempt, barrier start): **7 conflict outcomes, 0 conflict-path replays**. The one replayed attempt serialized at the *pre-insert* replay select (the path that works). Every attempt ended with exactly 1 snapshot row and exactly 1 `SnapshotCompiled` event, and a retry after each conflict replayed cleanly — so callers recover on a second call, but the first call returns a permanent-class code.

Divergence, in both directions at once:

- Spec V6 (compile port paragraph): "a concurrent identical compile replays the stored row" — never happens on the 23505 path. The code comment (`snapshot.go:253-255`) claims the same unreachable behavior.
- The conflict detail, "a different parameter set produced identical artifact bytes", is **factually wrong** in exactly the concurrent-identical case it fires for: the parameter sets are identical.

The committed `TestSnapshotIdenticalBytesCollisionConflict` cannot catch this: its conflict arm has genuinely non-matching parameters (the re-select failing is correct there, by accident of the same 25P02), and its replay arm is a *fresh call* that hits the pre-insert select, never the 23505 handler with a matching tuple. The claimed fix's central classification therefore has one live half (conflict) and one dead half (replay).

Fix shape: wrap the insert in a savepoint (pgx v5 nested `tx.Begin`) so the re-select runs in a live transaction, or roll back and re-select in a fresh transaction before classifying. Add a committed concurrent-identical race test; it must assert replay, not merely absence of error.

## Claimed fixes — verified against live Postgres

- **B1 price lexeme — FIXED.** `trimDecimalZeros` trims fractional zeros only. Probe: 10-case unit table (`100`→`100`, `0`→`0`, `100.0`→`100`, `0.0`→`0`, `3.50`→`3.5`, `10.00`→`10`, `0.50`→`0.5`, `.5`→`.5`, `1000.000`→`1000`, `7`→`7`). DB: committed `TestSnapshotPriceLexemeIntegrity` (integral survives, zero expressible, `100` vs `1.0` disagree → omitted); probe: `100` vs `100.0` equal exacts agree on the trimmed lexeme → component carried as `100` (correct: same value after the pinned trim).
- **B2 contributing-source audits — FIXED.** Committed stale-pinned test (0 audits, 0 inputs, artifact validates) plus probe: mixed-source compile (one correct pin, one stale) → audit set is exactly `{contributing source}`, no row and no input for the excluded source, validation `accepted`. `governing[src]` is provably non-nil in the audit loop (audit sources derive from `contributing`, which requires `governing` membership).
- **B3 exact bounds fold both ends — FIXED.** Probe matrix, one resource per arm, all against the live DB: exact-only → 100; min50+exact100+max200 → 100; min150+exact100+max200 → omitted (empty interval); exact100 above max50 → omitted; exact0 with max50 → omitted (zero unexpressible in the DTO); min200+max100 without exact → omitted; duplicate agreeing exacts → 100; regression arm: seeded max-only `context_tokens` still 128000. Matches the spec sentence including the both-ends contribution and the "unknown, never a fabricated intersection" rule.
- **B4 signal range post-round — FIXED.** 16-arm `signalValue` table: `100.005%` → skip, `100.000%` → 10000, `99.995%` → 10000, `100.0049%` → 10000, `NaN` (% and count) → skip, `Infinity`/`inf` → skip, `1e400` (ParseFloat range error) → skip, `0` → 0, `0.00004%` → 0, `-1%` → skip, `-0` → 0, `50.5` count → 51, `0.5%` → 50, int64-max raw → overflow-guard skip. Live arm: an `'Infinity'::numeric` measurement (PG16 stores it) compiles cleanly with the signal skipped. Committed boundary test also green.
- **B5 closed findings schema — FIXED.** Committed tests pin `findings_bytes` as the canonical encoding of `{code, offer_id?, detail}` objects, decode them on replay, and exercise `artifact_hash_mismatch`; the corrupt-bytes path returns findings that survive `strictDecode`. Undecodable stored findings degrade gracefully (outcome only, no error) — round-1 item 10 behavior preserved.
- **B6 basis keyed by evidence class — FIXED.** Query joins `assertions` and keys `basis_rules[evidence_class]`; spec line repaired to match, resolving the round-1 contradiction in favor of the policy grammar (the validator's four-class coverage mandate with pinned per-class values is now coherent with the lookup). Discriminator probe: a measurement whose `benchmark_label` is literally `provider_reported` but whose assertion `evidence_class` is `third_party_reported` → basis `published` (class-keyed); label-keying would have produced `publisher_claim`. Committed test confirms the override (`provider_reported` → `publisher_claim`). Minor note below on the now-dead default branch.
- **N1 collision classification — HALF-FIXED, blocker N1-R2 above.** Sequential identical parameters → replay via the pre-insert select (works). Different parameters, identical bytes → `conflict` (correct, committed test). Concurrent identical parameters → `conflict` (wrong; spec and comment say replay).

## Regression re-verification (round-1 "verified correct" list)

Focused live probes, all green:

- **Rights precedence at created_at**: a newer-but-expired policy row leaves the older non-expired row governing (compile with the older row's pin contributes; the expired row's own pin excludes) — both directions probed.
- **Offer listing tie**: `listed` and `withdrawn` at an identical `observed_at` → `withdrawn` wins, offer excluded.
- **Revision binding fail-closed**: two current bindings at cutoff → revision `""`, zero signals.
- **Validate/publish gates**: via committed `TestValidateAndPublishSnapshot` — publish before validation `conflict`; revocation committed before publication `forbidden`; unknown hash/channel `not_found`; idempotent replays emit exactly one `SnapshotValidated`/`SnapshotPublished`.
- **Replay, prices upper-bound/disagreement, capability/limit taxonomy**: unchanged code paths, covered by the committed suite in the full run.

## Notes

1. The dead `basis = "published"` default in `projectSignals` is now unreachable: the validator mandates all four evidence classes with pinned values, so the lookup always hits. Harmless; recording so nobody mistakes it for a bug.
2. Round-1 N2 (validation's rights-audit check is count-only, no source-set equality or config-hash re-derivation) is unchanged and was not re-litigated as a blocker; still the weakest gate in the file.
3. Round-1 N3 scoping asymmetries (prices offer-global vs limits observation-scoped; no `observed_at <= knowledge_cutoff` filter) and N4 (duplicated routing vocabularies in the collector) are unchanged.
4. Test-infrastructure hazard worth knowing: `seedSnapshotOffer` hardcodes single-use IDs (`of1`, `as-cap`, `rev1`, ...), so it can run once per container; the round-2 committed tests work around this correctly.

## Scope integrity

Probe file `internal/collector/probe2_review_test.go` deleted before the final regression run; `git status` shows only the concurrent session's untracked `docs/implementation.md`. No in-scope file touched.

## Verdict

Five of six round-1 blockers are genuinely fixed and live-proven, the fold arithmetic is now correct across its whole interaction space, and the regression surface is clean. The one remaining defect is in the claimed N1 repair itself: the concurrent-identical replay branch is dead code (25P02 in the aborted transaction), so the port returns a permanent `conflict` with a factually wrong detail for exactly the case the freshly repaired V6 spec sentence promises as replay. Mechanical fix (savepoint or post-rollback re-select), but it is a pinned-contract divergence with a misleading error, so the round fails until it lands.

VERDICT: FAIL
