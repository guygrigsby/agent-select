# Phase 6 cross-model extraction review: swebench

Date: 2026-09-30
Source: `data/research/benchmark-sources/_swebench_site.html` (2,392,125 bytes, captured from `https://www.swebench.com/`). The leaderboard data is embedded plainly (not HTML-escaped) in `<script type="application/json" id="leaderboard-data">`.

## What was extracted

- 157 rows in `data/research/swebench/findings.json`, batch `swebench`, format `unresolved_only`.
- Sections: Multilingual 13, Test 13, Verified 83, Lite 27, Multimodal 21 (total 157; splits kept separate, `benchmark_version` carries the split name).
- Only `checked: true` submissions: 166 of 323 source rows carry `checked: false` (not yet verified by the SWE-bench team) and are excluded, recorded as a gap.
- `reported_model` is the record's `tags` `Model:` pin (e.g. `gemini-3-flash`), not the display name; display name, org, agent, agent version, attempts, run date, open-source-system flag and submissions site are pinned in `configuration`.
- Values are byte-exact lexemes of the `"resolved": <lexeme>` field, located section-scoped (folders repeat across sections; global byte search picks the wrong record — caught and fixed during extraction).
- `evidence_class` is `third_party_reported`: the leaderboard is published by the SWE-bench team; rows are submitted agent+model results.
- `identity_status` is `unresolved`; `proposed_offer_ids` is empty (phase 9).

## Verification performed

- The derived file `_swebench_leaderboard_extracted.json` was validated field-by-field against the site-embedded JSON for all 323 rows on every field used (agent, date, folder, model_display, model_org, name, resolved, site, tags, mini-swe-agent_version, os_system, checked). It is not cited as a source; the site HTML is.
- Every emitted lexeme was re-checked against the parsed `resolved` float from the site JSON (0 mismatches after the section-scoped fix; the earlier global-search variant produced 41 wrong-record lexemes and was discarded).
- Split counts recomputed from a fresh parse: 13/13/83/27/21 = 157.

## Gaps and limitations

- 166 `checked: false` rows excluded (gap records them as available for future re-review).
- SWE-bench Pro absent (capture failed); no Pro value inferred from any other split.
- Rows are agent+model systems, not model-only runs; the agent/system pins in `configuration` must never be dropped in comparison.
- Per-run cost, per-instance cost and api-call counts exist in the records but are not extracted as measurements.

## Decision

Batch `swebench` is source-validated for published-measurement provenance only. No benchmark is accepted for routing.
