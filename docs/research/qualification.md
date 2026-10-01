# Evidence qualification

Stored benchmark cells have passed source-accuracy review. This checks the published label, table row/column, metric, unit, number and recorded settings. It does not establish a current provider binding, equal evaluation conditions or the chance of succeeding on a new task. Claims and routing-quality admission remain quarantined.

`evidence_reviews` records the reviewed batches. `evidence_review_inputs` pins each findings file, review report and cited source file by path and SHA-256. Its composite foreign key references `source_files`; `make test` rejects missing pins and missing source reviews. `make check` also checks the files against the hash manifest. Changed evidence needs a new review, not an inherited batch approval.

Read the source-checked values and the separate routing gate:

```sql
SELECT * FROM source_validated_benchmarks;
SELECT * FROM source_validated_unresolved;
SELECT * FROM accepted_benchmarks;
```

Google and Moonshot/MiniMax source checks are recorded in [their review](../../data/reviews/google-moonshot-source-validation.md). Earlier checks are recorded in [the first review](../../data/reviews/source-evidence-2026-09-30.md) and [its corrections](../../data/reviews/source-evidence-corrections-2026-09-30.md). HTML-escaped `evalResults` records in the Gemma source are real data. An audit that called them nonexistent was rejected after parsing the saved attributes.

Fair routing comparisons still require the evaluated revision, its relationship to a live offer and the critical settings defined by [benchmark comparability](../specs/benchmark-comparability.md). Shared table headings alone do not prove those settings match. Reported-label tables remain useful evidence without becoming a ranking of today's offers. GenEval also contains a CLIP retrieval baseline, not an image generator.

Benchmark points are not success probabilities. The implementation plan can proceed from this source-checked research while requiring identity and comparison checks before performance ranking. A numeric success floor needs separately validated host outcomes; until then it must remain disabled. Public evidence cannot supply missing observations of the user's work. Source rights remain unknown for publication and must be settled before enabling production collection or distribution.

## Phase 9: derived qualification (schema 5)

The rebuild now derives bounded qualification tables (research-artifact.md §12) instead of leaving identity, settings, limits and price dimensions as unexamined assumptions. Every row is recomputed from assembled inputs at each rebuild; manual edits cannot survive. Gate 7 verifies coverage and recomputation.

Current derived state, honestly stated:

- Offer bindings: 394 of 395 catalog models carry an exact-id join to a models.dev offer; `typesafe-ai/jev` is unjoined. Name similarity never creates a binding.
- Measurement settings: 1,562 measurements × 12 common critical components = 18,744 settings rows. `benchmark_release` is concrete for 220 measurements (the `unknown` sentinel stays knowledge-unknown); `tool_policy` is documented for 82; `output_compute_budget` for 4; `reasoning_effort` for 8. Every other component is `not documented in retained source bytes`.
- Comparison eligibility: zero of 1,562 measurements are eligible, because no measurement documents all twelve common components. This is the recorded baseline, not a gate failure; eligibility must be earned through documented settings, never hand-entered.
- Limits: all 758 catalog limit rows qualify with direction, operation-scope and accounting unknown; the vercel catalog carries none of those semantics.
- Price dimensions: 228 distinct unit/path pairs — 46 map to token monetary classes and 4 to count thresholds (50 classified); 178 stay unresolved. Values stay verbatim (Mercury's `0.024999999999999998` is exact) and disagreements stay exact-value.
