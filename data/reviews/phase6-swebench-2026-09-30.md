# Phase 6 cross-model extraction review: swebench

Date: 2026-09-30 (revised after review round 1)
Source: `data/research/benchmark-sources/_swebench_site.html` (2,392,125 bytes, captured from `https://www.swebench.com/`). The leaderboard data is embedded plainly (not HTML-escaped) in `<script type="application/json" id="leaderboard-data">`.

## What was extracted

- 131 rows in `data/research/swebench/findings.json`, batch `swebench`, format `unresolved_only`.
- Sections (checked=true only): Multilingual 12, Test 13, Verified 60, Lite 25, Multimodal 21 = 131. Splits kept separate; `benchmark_version` carries the split name.
- Inclusion filter is exactly `checked is true` (the boolean). The source's 323 records break down as: 166 `checked: false`, 18 `checked: null` (first-party mini-swe-agent submissions), 8 carrying the string `"false (See README.md for info on how to get your results verified)"`, 131 `true`. All 192 non-true records are excluded and recorded as a gap with this breakdown.
- `reported_model` is the record's `tags` `Model:` pin (e.g. `gemini-3-flash`), not the display name; display name, org, agent, agent version, attempts, run date, open-source-system flag and submissions site are pinned in `configuration`.
- Values are byte-exact lexemes of the `"resolved": <lexeme>` field, located section-scoped (folders repeat across sections; a global byte search picks the wrong record — caught during extraction).
- `evidence_class` is `third_party_reported`: the leaderboard is published by the SWE-bench team; rows are submitted agent+model results.
- `identity_status` is `unresolved`; `proposed_offer_ids` is empty (phase 9).

## Verification performed

- The derived file `_swebench_leaderboard_extracted.json` was validated field-by-field against the site-embedded JSON for all 323 rows on every field used (agent, date, folder, model_display, model_org, name, resolved, site, tags, mini-swe-agent_version, os_system, checked). It is not cited as a source; the site HTML is.
- Every emitted lexeme was re-checked against the parsed `resolved` float for that section's record (0 mismatches).
- The checked distribution (131 true / 166 false / 18 null / 8 string-false) was recomputed from a fresh parse; the 131 emitted rows are exactly the true records.

## Review round 1 correction

The first emitted version (157 rows) used the filter "checked is not boolean false", silently including 18 null-checked and 8 string-false records, while this document and the decision claimed "only checked=true". Round 1 caught it; the filter is now literally `checked is true` and every stating document records the real distribution.

## Gaps and limitations

- 192 non-true records excluded (gap records the breakdown; available for future re-review if the leaderboard verifies them).
- SWE-bench Pro absent: `_swebench_pro.json` is a literal 404 response; no Pro value is inferred from any other split.
- Rows are agent+model systems, not model-only runs; the agent/system pins in `configuration` must never be dropped in comparison.
- Per-run cost, per-instance cost and api-call counts exist in the records but are not extracted as measurements.

## Decision

Batch `swebench` is source-validated for published-measurement provenance only. No benchmark is accepted for routing.
