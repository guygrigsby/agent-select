# Phase 6 cross-model extraction review: aider

Date: 2026-09-30
Source: `data/research/benchmark-sources/_aider_leaderboard.html` (captured from `https://aider.chat/docs/leaderboards/`).

## What was extracted

- 138 rows in `data/research/aider/findings.json`, batch `aider`, format `unresolved_only` (69 model rows x 2 metrics: percent correct, percent well-formed edits).
- Table structure: 139 `tr` = 1 header + 69 model rows + 69 interleaved detail rows; detail rows carry run date, aider version, 225-case aggregation, per-run outcomes. Model row and its detail row are paired positionally.
- `configuration` pins aider version, edit format, full command, and run date per row (e.g. `aider 0.86.2.dev; edit format diff; command aider --model openai/gpt-5; run date 2025-08-23`).
- `evidence_class` is `third_party_reported`: the leaderboard is published by the aider team; runs are self-submitted.
- `identity_status` is `unresolved`; `proposed_offer_ids` is empty (phase 9).

## Verification performed

- Row count: 69 model rows parsed with 7 cells each; 0 missing dates, 0 missing versions after detail pairing.
- All 69 model labels byte-verified in the raw HTML; all percent cells verified as stripped-integer/decimal strings with `%` suffix in source.
- Values are the exact table lexemes with the `%` suffix stripped (`88.0`, not `88`); the locator cites row label and column name.

## Gaps and limitations

- Dollar cost per run and per-case outcomes exist in the source and are not extracted (suite-run cost is not a benchmark score).
- Model labels are display labels with reasoning-effort suffixes (e.g. `gpt-5 (high)`); identity resolution deferred to phase 9.
- Runs are model+edit-format+command systems; the pins in `configuration` must never be dropped in comparison.

## Decision

Batch `aider` is source-validated for published-measurement provenance only. No benchmark is accepted for routing.
