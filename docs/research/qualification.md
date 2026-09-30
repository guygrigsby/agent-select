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
