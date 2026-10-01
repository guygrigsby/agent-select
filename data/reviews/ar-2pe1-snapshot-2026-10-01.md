# Adversarial review — commit 046cf25 (ar-2pe.1 Phase 12, snapshot compile/validate/publish)

- Commit: `046cf257fa530e4f94bcc9b93e48cff33b8cc1da` (HEAD; working tree identical to the commit for all three in-scope files)
- Scope: `internal/collector/snapshot.go`, `internal/collector/snapshot_test.go`, delivery-storage.md V5 section "Snapshot projection grammar (compile_snapshot)" (md5 `3cb0bf719f91598a6351549b8a8fb73c`, verified)
- Method: full read of spec section + port code + tests; 13 live probes against disposable Postgres 16 containers (scratch file `probe_review3_test.go`, deleted after runs, never committed). Two initial probe "failures" (listing tie, expired policy) were probe-script bugs — timestamp skew and a wrong expectation — corrected and re-run; the round-2 lesson applies to reviewers too.
- Out of scope (not re-litigated): migration CHECKs, non-snapshot ports, declined round-1/2 notes.

Regression: `gofmt -l .` empty; `go vet ./...` clean; `AR_REQUIRE_POSTGRES=1 go test ./internal/collector/ -count=1 -run 'TestCompile|TestValidate'` ok (7.5s); full collector suite ok (46.4s) after probe removal.

## BLOCKERS

### B1 — Price trailing-zero trim corrupts integer lexemes and merges disagreeing exacts
`snapshot.go:703`: `lex := strings.TrimRight(strings.TrimRight(amount, "0"), ".")`

Spec (line 91): "The amount is carried as its exact decimal lexeme with trailing zeros trimmed".

The only conformant reading trims *fractional* zeros (`3.50`→`3.5`). The code trims the whole lexeme:
- exact `100` projects as `"1"` (live-probed; the artifact carries a price 100x wrong, silently)
- exact `0` (free model) trims to `""` → `PriceData{USDPerMillion:""}` → `routing.NewSnapshot` rejects → whole compile fails `invalid_evidence` (live-probed)
- `100` and `1.0` both trim to `"1"`, so `exactLexemes` (snapshot.go:711) sees agreement where the evidence disagrees — the omission rule for disagreeing exacts cannot fire (live-probed)

### B2 — Rights audits are written for sources with no contributing inputs; validation then rejects the artifact compile just staged
`snapshot.go:260-271`: audits inserted for every source in `governing` (admissible policy), regardless of whether any of its inputs survived the pinned-version check.

Spec (line 88): "Contributing sources get one `snapshot_rights_audits` row each ...; excluded sources get no row and no inputs."

Live probe: compile with `policy_version="stale"` for a source whose governing policy is admissible → `snapshot_rights_audits` = 1 row, `snapshot_inputs` = 0 rows → `ValidateSnapshot` returns **rejected** with finding "rights audit covers 1 of 0 contributing sources". The port pair is internally inconsistent: such artifacts can never be validated, so their publish gate is unreachable. Deterministic for any compile where all of a source's inputs are excluded by pinned-version mismatch.

### B3 — Exact bounds are folded into the maximum end only; exact-above-maximum projects the maximum instead of omitting
`snapshot.go:655-671`: `maxima = append(maxima, bd.exacts...)` (line 658) then min over maxima; the omit check (line 667, `bd.hasMin && bd.min > maximum`) only sees literal `minimum` bounds — an exact bound never reaches the lower end.

Spec (line 92): "`maximum` is the minimum over all maximum and exact bounds (**an exact bound contributes its value to both ends**), subject to agreement of exact values; conflicting exacts, **or a minimum above the derived maximum, omit the resource**".

Live probe: `input_tokens` exact=100 + maximum=50 → artifact carries `Maximum:50`. Per spec, the exact's lower end (100) exceeds the derived maximum (50) → empty interval → the resource must be omitted ("unknown, never a fabricated intersection"). The code fabricates 50 — a limit the evidence contradicts.

### B4 — Signal range check runs pre-round with a 10000.5 threshold; values that round outside 0..10000 are emitted, not skipped, and kill the whole compile
`snapshot.go:832-844`: `signalValue` guards `if f > 10000.5 { return 0, false }` (line 841) **before** `int(f + 0.5)` rounding.

Spec (line 94): "Value is `round(raw_value * 100)` for unit `'%'` ... **values outside 0..10000 are skipped**".

Live probe: raw `100.005` `'%'` → emitted `10001` → `routing.NewSnapshot` rejects ("signal value 10001 outside 0..10000") → compile fails `invalid_evidence`. Spec: that value is skipped and the compile succeeds with the remaining signal. Same hole for `NaN` raw values: `ParseFloat` accepts them, both comparisons are false, and the garbage conversion flows into the artifact path instead of a skip.

### B5 — `findings_bytes` holds `[]string`, not the pinned closed finding schema
`snapshot.go:855` (`Findings []string`), `:924` (findings appended as free-text strings), `:931` (`canonicalize(findings)`).

Spec (line 123): "`findings_bytes BYTEA` holding the canonical-v1 encoding of the closed findings array; **each finding is `{code: string, offer_id: string, detail: string}`** with `offer_id` omitted for artifact-global findings."

Live-probed findings are bare strings ("rights audit covers 1 of 0 contributing sources"). Consumers of `snapshot_validations.findings_bytes` and of `SnapshotValidated.findings_hash` (spec line 67) cannot decode the pinned shape. Mechanical fix, but it is the written data contract of a column and an event payload.

### B6 — `basis_rules` keying is incoherent across grammar, validator and tests; basis derivation is dead configuration
Spec (line 94, the section under review): "`Basis` is `basis_rules[benchmark_label]` when present, else `published`." — keyed by **benchmark label**.

The port's own validator demands the opposite: `RegisterProjectionPolicy` rejects a policy whose `basis_rules` don't "cover the four evidence classes" (live-probed rejection of `basis_rules: {}`; the fixture `validProjectionPolicy`, ports_test.go:747-757, keys `provider_reported`/`third_party_reported`/`competitor_reported`/`host_observed`).

Since a benchmark label is never an evidence class, every mandated basis value is unreachable: the lookup key can never match a mandated key except by string collision. Live probe: `provider_reported` measurement + `basis_rules{provider_reported: publisher_claim}` → `Basis: "published"` (label `swe_verified` missed, default fired). Because the fixture keys evidence classes, the label lookup never hits anywhere in the suite — basis is `"published"` in every test, and the override mechanism is untested *and* untestable as pinned. The section also contradicts line 47's policy ownership sentence ("owns ... basis derivation") as implemented by the validator. Needs a spec repair decision (one keying), then validator/code/fixture alignment.

## NOTES

### N1 — Compile idempotency deviates from its own pin; the pin is unsatisfiable as written
`snapshot.go:178-190` comments "Idempotent replay by the complete parameter set" but keys on `input_manifest_hash` — the **contributing** input manifest — plus distribution, policy version and times.

Spec (line 47): "Idempotent by the complete ordered parameter set ... **any parameter difference compiles a distinct snapshot row**."

Live probes, both directions:
- Adding an input whose pinned version excludes it → replayed the earlier row (1 snapshot row, spec demands a distinct row).
- Two policies differing only off-artifact (`zero_image_charge_evidence`) over identical inputs/times → identical artifact bytes → `snapshots_pkey` 23505 → `temporarily_unavailable` — a transient code for a permanent condition; a retry loop spins forever.

The "distinct snapshot row" sentence cannot hold under `snapshots` PK = `artifact_hash` whenever a parameter difference does not reach the bytes. Spec needs repair (drop the sentence, or key the table on a params digest); independently, the code should classify the identical-bytes collision as a replay or `conflict`, not `temporarily_unavailable`. Same collision surface exists for two concurrent first-time compiles (no `ON CONFLICT` on the artifact insert).

### N2 — Validation's rights-audit check is count-only
`snapshot.go:910-925`: compares the audit row count against the distinct source count of `snapshot_inputs`. No set equality (which sources), no re-derivation of each source's governing hash against the pinned `config_hash`, no cross-check of the artifact's declared inputs against `snapshot_inputs` rows. B2's phantom row defeats it only incidentally (count mismatch); an audit row for the *wrong* source pair passes.

### N3 — Scoping asymmetries (spec-conformant, recorded)
- Limits are contributing-observation-scoped; prices are offer-global across all observations (live: a non-contributing source's price `9.99` projected into the artifact; its limit did not). Spec line 91 does not scope prices to contributing observations — conformant, but surprising.
- The offer-set listing subquery (snapshot.go:477-483) reads `offer_listing_facts` regardless of which observation recorded them. Spec line 87 is verbatim conformant ("whose latest `offer_listing_facts` row").
- The grammar's opening sentence ("evaluated at `knowledge_cutoff` except where stated") is never operationalized for caps/limits/prices/signals — no `observed_at <= cutoff` filter exists anywhere in the projection. Spec is silent, so no divergence; flagging the underdetermination.

### N4 — Duplicated routing vocabulary in the collector
`routingPriceComponents`, the capability vocabulary and the decimal-syntax rules are re-declared in collector code instead of imported from `routing` — drift risk the repo's own drift-guard policy calls out. Dedupe candidate, no behavior divergence found.

## Verified correct (per the 13 items)

1. **Rights gate at created_at** — snapshot.go:309-331 filters `effective_at <= created_at AND expires_at > created_at` before ordering, so an expired newest row leaves the older valid row governing (live-verified; spec line 88 verbatim), a newer `denied` row is never bypassed (live-verified; spec line 47), same-time conflict excludes (committed test, snapshot_test.go:315-336), missing/unknown excludes.
2. **Offer set tie-break** — `ORDER BY f.observed_at DESC, f.action DESC` (snapshot.go:477-483); live probe with a true observed_at tie: `withdrawn` wins, offer excluded.
3. **Revision binding fail-closed** — zero or multiple current bindings at cutoff → empty `Revision` and no signals (live probe with two current bindings: revision `""`, signals `[]`); binding current-ness anti-joins match spec line 89 (snapshot.go:422ff, 778ff).
4. **Limits** — conservative intersection otherwise correct: conflicting exacts omit (live: 3.0 vs 5.0 → component absent), greatest admissible quote as upper bound (live: exact 3 + proven_upper_bound 7 → 7), `mask_count`/`image_pixels` not projected. Exception: B3.
5. **Prices** — upper-bound semantics and disagreement omission verified (modulo B1's lexeme corruption); `3.50`→`"3.5"` lexeme preserved; non-USD/non-million_tokens rows excluded (snapshot.go:682-689).
6. **Signals** — bp conversion verified (86.0% → 8600), `task_map`/`weights` applied, revision-binding precondition enforced, per-signal identity by assertion ID. Exceptions: B4 (skip), B6 (basis).
7. **snapshot_inputs** — one row per ordered input with ordinals following the artifact's sorted input order (snapshot.go:248-256; spec line 95).
8. **Locator** — `snapshot://` + lowercase hex of the hash (snapshot.go:295), used for staged and published artifacts alike (snapshot.go:1046).
9. **Replay** — identical call replays identical hash/coverage without a second `SnapshotCompiled` (committed test, snapshot_test.go:181-193); deviations are N1.
10. **ValidateSnapshot** — idempotent per `(artifact_hash, validator_version)` (snapshot.go:875-884, survives undecodable stored findings), unknown hash → `not_found` (:894), re-hash of staged bytes (:900), decode contract (:905), one `SnapshotValidated` per validation, findings_hash over findings_bytes (:936-944). Coverage check shallowness: N2; findings shape: B5; audit-row production: B2.
11. **PublishSnapshot** — idempotent by `(artifact_hash, channel)` (:976-982), unknown snapshot `not_found` (:990), requires an accepted validation (`conflict`, :995-1001), `channel_class` read from the registry never the caller (:1004-1010), retired channel `conflict` with the migration-0007 BEFORE INSERT trigger as backstop (:1011-1013; spec line 124), `local` artifact → non-local channel `forbidden` (:1014-1015), in-transaction rights re-evaluation at `now()` for every audited source with any revocation failing the gate (:1036-1058), publication row + `rights_decision='allowed'` + `SnapshotPublished` (payload matches spec line 68) in one transaction; insert-race 23505 → `conflict` (:1063-1066).
12. **Error taxonomy / atomicity** — every write path is a single transaction with deferred rollback (no partial writes on any failure return); typed codes throughout (`invalid_request`, `invalid_evidence`, `not_found`, `conflict`, `forbidden`, `temporarily_unavailable`), no byte or credential leakage in detail strings (checked across probes and code); caller-data CHECK rejections classify per errors.go.
13. **Test coverage** — five committed test functions; untested gates named below.

## Untested gates (item 13)

All of the following have no committed test; most were live-probed here:
- policy_version mismatch exclusion (B2's trigger), phantom-audit consequence
- expired-newest-row precedence and newer-denial precedence (verified correct live, but regressable)
- listed/withdrawn observed_at tie
- zero-current-binding and multi-current-binding identity (no identity tests at all)
- limit intersection edges: conflicting exacts, minimum-above-maximum, exact-above-maximum (B3), greatest-upper-bound
- price edges: integer lexeme (B1), zero price (B1), exact disagreement, upper-bound semantics
- signal clamp boundary skip (B4), NaN raw values (B4)
- basis override (untestable as pinned, B6)
- distinct-parameter compiles: excluded-input addition, off-artifact policy difference, concurrent first-compile race (N1)
- `findings_bytes` closed schema (B5)
- publish: in-tx revocation race, channel-class mismatch, unregistered channel, concurrent publish race

## Verdict

Six blockers, all live-proven against real Postgres: B1-B4 are data-corruption or false-compile paths in the projection arithmetic; B2 makes compile and validate mutually inconsistent; B5/B6 are pinned-contract violations (findings schema, basis keying) that need a spec-side decision plus code alignment. The gate scaffolding (rights precedence, tie-breaks, identity fail-closed, publish gates, atomicity, taxonomy) is solid and matches the spec closely; the defects are concentrated in the arithmetic folds and the two contract mismatches.

VERDICT: FAIL
