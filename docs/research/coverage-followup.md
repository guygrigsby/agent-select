# Coverage follow-up

## GenEval structured image-generation results

The fetched upstream GenEval README has a complete six-row, seven-column “Main results” table. Its 42 numeric cells are transcribed individually in `data/research/coverage-followup/findings.json`, retaining exact historical labels, fraction units, row/column locators, and independent-benchmark evidence type. The table’s available per-model generation configurations are not fully specified; the README’s example command (four images per prompt) is not generalized to all rows. These are not provider-reported scores.

All historical model identities remain unresolved and `proposed_offer_ids` are empty. Historical labels are usable collection-source evidence even without a current catalog baseline; they do not establish current catalog coverage, and image-model coverage is not complete. The CLIP retrieval baseline is retained as reported, not recast as a model offer.

Verified collection source: `https://raw.githubusercontent.com/djghosh13/geneval/main/README.md`, saved bytes at `data/research/coverage-followup/sources/geneval-readme.md` (SHA-256 recorded in findings). GitHub API main lookup returned commit `af4902f24d3ca90ebbb446dd9891a59e0f82725f`; repository LICENSE was retrieved and saved. Its MIT terms apply to the repository as stated in that file, not automatically to benchmark datasets, model outputs, images, or third-party components. ELLA README was also fetched and preserved; it contains benchmark figures, but no numerical values have been invented or transcribed from images.

## Collection path and boundaries

Use event-driven collection on an upstream repository commit or README-content hash change: fetch the documented raw README, preserve exact bytes and commit provenance, then deterministically parse the named Markdown table, validating row count, columns, numeric cells, and locators before ingest. Keep provider-reported values separate from independent benchmark results and preserve exact source labels pending identity resolution. Do not guess organizational repositories or offer IDs.

This establishes an allowed structured image-generation collection lane, not full model benchmark coverage. Access-denied/challenge/error responses are blocked/non-score, never benchmark values. Prior saved response placeholders remain non-score. Keep arena preference evidence separate from image scoring; verify actual saved bytes before calling any arena CSV retrieval successful. Respect source terms and do not infer underlying data rights from repository licensing.
