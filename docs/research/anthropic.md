# Anthropic research

## Findings
The assigned catalog contains 19 exact Vercel IDs, including future/unfamiliar revisions. Saved Anthropic documentation was inspected for model identity and capability claims. Exact pages were available for Fable 5.1, Haiku 4.5, Opus 4, Opus 5.5, Sonnet 4, and Sonnet 5.5; claims are labeled provider claims in `findings.json`. No benchmark row is included because the inspected saved source tables did not establish a numeric value with a complete configuration.

## Automation path
The fetching stage saved raw bytes under `data/research/anthropic/sources/`; URL-to-file mapping is recovered from `anthropic.log`. A future run should parse provider release/model cards and only emit a benchmark when the source explicitly identifies the exact model and reports a numeric metric plus configuration.

## Limitations and coverage
Missing evidence is explicit per model. Future revisions and fast variants were not assigned predecessor scores. Catalog capability metadata is not treated as benchmark evidence. No broad completeness claim is made.
