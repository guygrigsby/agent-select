# Phase 6 gap-only review: anthropic

Date: 2026-09-30
Scope: `data/research/anthropic/findings.json` gap expansion (4 to 16 entries) in commit `fd4821f`. The batch has zero measurement rows; this review covers the checked-sources outcomes only.

## What changed

Twelve new gap entries, each a bounded negative outcome naming the inspected file and why it yields no numeric benchmark rows (docs pages with only spec/pricing/feature tables, recipe/link indexes, news/research indexes, HF model-list JSON without scores).

## Verification performed

- Spot-checked claims against the cited bytes: `anthropic_cookbook_readme.md` is a recipe/link index with no score tables; the docs model pages carry context/pricing/latency tables and no benchmark tables; `hf_anthropic_models.json` and `hf_claude_models.json` are model listings without benchmark fields.
- Every named file exists under `data/research/anthropic/sources/` and is pinned in the capture inventory.
- No measurement rows exist in this batch; nothing reaches `benchmark_measurements`; no value is asserted anywhere.

## Decision

Batch `anthropic` records a v1 `published_measurements` decision covering the findings file as gap-only (no rows, no acceptance). SWE-bench-style verified numeric extraction for anthropic remains open until a primary numeric source is located; the gaps record where it was looked for.
