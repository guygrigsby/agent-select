# xAI / SpaceXAI research

## Findings
The saved official documentation page (`data/research/xai/sources/xai_docs_overview.html`) embeds a model configuration object. It explicitly identifies assigned aliases for Grok 4.20 reasoning/non-reasoning and multi-agent beta variants, Grok 4.3, 4.5, 4.6, 4.7, and Grok Build 0.1. The configuration reports modalities, function calling, structured outputs, reasoning, effort options, prompt limits, and rate limits. These are catalog/provider capability claims, not benchmark measurements.

No numeric benchmark table with an exact assigned model identity was found in the checked saved sources, so `benchmarks` is intentionally empty. No predecessor or family scores were copied. The 4.1 fast variants have no matching evidence in the checked official configuration.

## Automation path
The fetching stage saves raw responses under `data/research/xai/sources/`. The embedded `globalThis.__XAI_PUBLIC_MODELS__` JSON can be extracted from the official HTML with a standard-library parser, then aliases can be matched to the exact Vercel IDs in `assignment.json`. Numeric benchmark rows should only be emitted when a primary or independent source contains a value, label, exact model identity, and configuration.

## Limitations and coverage
Coverage is limited to the saved provider docs, Hugging Face organization page, GitHub organization page, and Grok 3 release-path files. Several retrieved pages are dynamic or unrelated to the assigned revisions. No empirical weakness is asserted; operational limits are recorded separately where directly documented. Unknown future revisions remain unresearched.
