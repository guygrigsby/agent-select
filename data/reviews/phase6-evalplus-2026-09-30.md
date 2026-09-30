# Phase 6 cross-model extraction review: evalplus

Date: 2026-09-30
Source: `data/research/benchmark-sources/_evalplus_results.json` (34,305 bytes, captured from `https://evalplus.github.io/results.json`, the leaderboard's own results endpoint; the rendered results pages were 404 HTML so this JSON is primary, not derived).

## What was extracted

- 428 rows in `data/research/evalplus/findings.json`, batch `evalplus`, format `unresolved_only`.
- 125 leaderboard labels x 4 suites (HumanEval, HumanEval+, MBPP, MBPP+) pass@1 cells; 89 labels carry MBPP/MBPP+ values, 36 report null MBPP cells (retained as gaps, listed in the findings `gaps`).
- Values are byte-exact lexemes from the source JSON (regex on the label's object span, not re-serialized floats). One integer lexeme exists (Command-R+ humaneval `64`); it is preserved as `64`, not `64.0`.
- `configuration` carries the per-entry `prompted`, `open-data`, `size` and submitter `link` fields verbatim.
- `evidence_class` is `third_party_reported` for every row: the file is published by the EvalPlus team but entries are submitted by model authors without per-entry re-verification; decoding settings and harness versions are not stated per entry.
- `identity_status` is `unresolved`; `proposed_offer_ids` is empty (identity resolution is deferred to phase 9).

## Verification performed

- Row count: 428 = 125 humaneval + 125 humaneval+ + 89 mbpp + 89 mbpp+ non-null cells; recomputed independently from a fresh parse.
- Every value string was verified against the raw source bytes: the locator records the absolute byte offset of the `"key": value` pair, and the exact value lexeme was confirmed to start at offset + len(prefix) for all 428 rows. (Review round 1 found the first emission double-added the label offset into the recorded number; all 428 offsets were recomputed and byte-verified.)
- Null cells: 36 labels with null MBPP/MBPP+ confirmed by direct count over the parsed file.
- No label lacks a humaneval cell (checked; the earlier reconnaissance note of 124 float + 1 int humaneval cells is consistent: 125 total).

## Gaps and limitations

- Null MBPP cells (36 labels) are recorded as a gap, not as zeros.
- Entries are author-submitted; EvalPlus publishes without per-entry independent re-verification and without per-entry settings disclosure, so cross-model comparability is limited; this bars any future acceptance into `accepted_benchmarks` without a separate comparability decision.
- Labels are leaderboard display labels; mapping to catalog offers is deferred (phase 9).

## Decision

Batch `evalplus` is source-validated for published-measurement provenance only. No benchmark is accepted for routing.
