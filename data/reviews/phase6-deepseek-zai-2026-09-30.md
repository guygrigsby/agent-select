# Phase 6 provider extraction: deepseek-zai

Date: 2026-09-30. Scope: `published_measurements` rows extracted from
saved deepseek-zai batch sources into
`data/research/deepseek-zai/findings.json`. Review judgment covers source
accuracy of extracted cells only — byte-presence of each value in the
cited saved file at the stated locator. It does not approve revision
identity, comparability, calibration or routing use; `review_results`
state for this batch remains quarantined.

## Sources

- `saved_sources/deepseek-r1-paper.html` — DeepSeek-R1 paper page (21
  rows).
- `saved_sources/hf-deepseek-v31-readme.md` — HF model card (25 rows).
- `saved_sources/hf-deepseek-v4-flash-readme.md` — HF model card (61
  rows).
- `saved_sources/hf-deepseek-v4-pro-readme.md` — HF model card (78 rows).
- `saved_sources/hf-deepseek-v41-flash-readme.md` — HF model card (47
  rows).

All paths are relative to `data/research/deepseek-zai/`.

## Rows

232 total. Provider rows 202 (`provider_reported`): deepseek-r1 21,
deepseek-v3.1 25, deepseek-v4-flash 61, deepseek-v4-pro 61,
deepseek-v4.1-flash 34. Competitor rows 30 (`competitor_reported`):
zai/glm-5.1 17 and zai/glm-5.3 13, extracted from GLM comparison columns
in DeepSeek's own model-card tables. Reporting party decides the class:
DeepSeek's numbers for DeepSeek models are provider-reported; DeepSeek's
page reporting Zhipu GLM columns is competitor-reported.

## Verification performed

Independent re-check (not the extraction worker's own): every row's
`value` lexeme and identifier tokens re-searched in the cited file bytes;
all 232 present, including long-decimal lexemes preserved exactly. GLM
rows confirmed to sit in DeepSeek's comparison tables (not Zhipu pages —
no Zhipu source is in this batch's capture).

## Gaps

22 gap records in findings name precise per-file blockers (missing
tables, auth walls, placeholder pages). No Zhipu-authored evidence exists
in this batch; `zai/*` provider-reported rows require the zai batch's own
sources, which are not captured.

## Limitations

Identity basis ties rows to saved-page labels; revision identity and
current-offer mapping are phase 9. Comparability (cohort and protocol) and
calibration are out of scope for this review.
