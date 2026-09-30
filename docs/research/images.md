# Image Generation Models Research

## Scope
Research covers 41 image generation and editing models across 9 providers listed in `data/research/images/assignment.json`. All 41 rows are retained. No numeric benchmark result was extracted for any row, so this report does not infer comparative quality from model size, speed, variant naming or provider marketing.

## Source collection
Saved source files and their status are recorded in the machine-readable source ledger in `data/research/images/findings.json`. The ledger has 23 `retrieved` entries and 9 blocked or error entries: 6 `blocked_404`, 2 `blocked_403` and 1 `blocked_redirect`. Blocked responses are not evidence for model claims. Catalog entries in `assignment.json` are identified as `catalog_claim` where they are used.

The checked sources include provider documentation, product pages, model repositories and benchmark project pages. Public APIs and documentation should be used only as documented and allowed. No access restriction should be bypassed.

## Evidence and limitations
Provider catalog claims are distinguished from verified primary claims in the findings ledger. A primary claim requires a retrieved source file with a specific locator. A blocked or redirected page does not support a claim.

No numeric benchmark evidence was found for any of the 41 models in the checked sources. This is a reporting gap, not evidence of weak quality. The capture checked static copies of the listed pages, while some pages may expose additional content dynamically. It does not establish that every dynamic table is absent.

Some future or unavailable catalog entries are retained because they are in the assignment. Their wording is limited to catalog metadata or availability status when no retrieved primary source supports more. The report does not assert that Imagen was shut down, that all Recraft models are wrappers, or that GPT-Image-1.5 superseded another model with measured improvements.

## Safety and operational notes
Where directly documented, FLUX.2 safety tolerance and Google Gemini SynthID output marking are recorded as provider or primary claims in the findings ledger. Gemini 3.1 Flash Lite Image is documented as not optimized for multiple reference images. These are operational descriptions, not quality rankings.

## Machine ledger
For complete per-model claims, bases, locators, benchmark status and durable source URL/file mappings, use `data/research/images/findings.json`. The ledger records `benchmark_status: "not_found_in_checked_sources"` for all 41 rows.
