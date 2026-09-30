# Qwen research

## Coverage
The assigned catalog contains 32 exact Vercel IDs. `data/research/qwen/findings.json` has one model row for each; unfamiliar and future-looking revisions are retained without inheriting predecessor measurements.

## Sources and findings
Saved fetching-stage bytes are preserved under `data/research/qwen/sources/`, including Qwen3, Qwen3-Coder, Qwen3-VL and Hugging Face pages. Catalog descriptions are recorded only as `catalog_claim` strengths. The checked saved source tables did not yield a numeric benchmark row with a sufficiently clear exact-model identity and locator, so `benchmarks` is intentionally empty rather than importing remembered or inferred values. No empirical weaknesses are asserted.

## Automation path
The extraction workflow reads `assignment.json`, keeps exact IDs, inspects saved source bytes and records evidence in the required interchange schema. Future runs should parse primary markdown/HTML tables, retain the verbatim value-and-label locator, record benchmark version/configuration, and distinguish provider-reported from independent results.

## Limitations
This is a saved-bytes-only extraction: no network requests were made in this stage. Some fetched pages are unavailable/404 or do not expose usable benchmark tables, and URL mapping for a few saved artifacts is uncertain. Catalog metadata is baseline capability context, not benchmark proof. Absence of a row means no verified measurement was found in the checked evidence, not that the model lacks capability.
