# OpenAI research findings

## Coverage
The assigned catalog contains 64 OpenAI IDs, and `data/research/openai/findings.json` contains one row for each. Catalog descriptions and context-window metadata are recorded as catalog claims/operational limits only; they are not benchmark results. Fast variants were kept separate and were not given predecessor scores.

## Sources and findings
Saved evidence includes the OpenAI developer model page, GPT-4 Technical Report (arXiv 2303.08774), OpenAI `simple-evals`, and Hugging Face cards for GPT OSS 20B/120B. The GPT-4 report describes GPT-4, but does not establish identity for the assigned `gpt-4-turbo` ID, so no score was transferred. The GPT OSS cards were inspected; no numeric table was accepted without an explicit exact-model value and label. OpenAI blog responses were saved but were JS/challenge shells and were not treated as evidence.

## Automation path
The fetching stage saves raw responses under `data/research/openai/sources/` and logs URL-to-file mapping in `data/source-manifest.tsv`. Extraction can parse `assignment.json`, inspect saved HTML/Markdown directly, and emit the strict interchange shape in `findings.json`. Benchmark rows should be added only when a saved source contains a numeric value, label, configuration, and exact identity/verified alias.

## Limitations
This is a bounded review of saved bytes, not a completeness claim. Many catalog IDs are revisions, speed aliases, or unavailable/future-looking releases with no exact primary evidence in the checked files. Missing benchmark evidence is represented explicitly; no historical number is inherited by a later revision.
