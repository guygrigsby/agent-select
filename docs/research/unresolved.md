# Unresolved published benchmark evidence

`data/research/unresolved/findings.json` preserves numeric cells from the saved provider tables whose published labels/revisions could not be safely mapped to current Vercel aliases.

- OpenAI Simple Evals: provider model rows, including dated GPT-4.1/GPT-4o/GPT-4.5 labels and reasoning-effort rows; table footnotes and prompt configuration remain in each configuration/locator.
- DeepSeek-V3.1 README: DeepSeek V3.1 non-thinking and thinking columns plus DeepSeek R1 0528; internal search/agent framework context is retained.
- DeepSeek-R1 paper: DeepSeek-labelled comparison-column cells only, with verbatim extracted row evidence.

All records are `provider_reported` and `identity_status: unresolved`; proposed IDs are intentionally empty unless a plausible current alias was unambiguous. These records are evidence preservation only and must not be used as route eligibility. No DeepSeek-V3.2 image values were invented or transcribed because the saved README contains no parseable numeric table. Other vendors' columns in the OpenAI and DeepSeek comparison tables were excluded.

## Locator and configuration caveats
DeepSeek V3.1 records identify the table header, exact row, and applicable column; row-specific internal search/code-agent/HLE notes are included where relevant, while unspecified settings are explicitly unknown. Historical reported model labels remain unchanged. DeepSeek R1 paper locators preserve row/cell identity, but paper footnote/configuration validation remains a gap. OpenAI row-specific footnote implications are not fully normalized and remain unresolved.
