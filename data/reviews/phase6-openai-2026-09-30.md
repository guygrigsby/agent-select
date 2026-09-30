# Phase 6 provider extraction: openai

Date: 2026-09-30. Scope: `published_measurements` rows extracted from saved
openai-batch sources into `data/research/openai/findings.json`. Review
judgment covers source accuracy of extracted cells only — byte-presence of
each value in the cited saved file at the stated locator. It does not
approve revision identity, comparability, calibration or routing use;
`review_results` state for this batch remains quarantined.

## Sources

- `data/research/openai/sources/github-simple-evals-readme.md` — the
  openai/simple-evals README results table, matched to catalog models by
  dated snapshot labels in the table (111 rows).
- `data/research/openai/sources/huggingface-gpt-oss-120b.html` — saved HF
  model repo page; html-escaped `evalResults` metadata (41 rows).
- `data/research/openai/sources/huggingface-gpt-oss-20b.html` — saved HF
  model repo page; html-escaped `evalResults` metadata (40 rows).

## Rows

152 total. Provider rows 141 (`provider_reported`): o3 21, o4-mini 21,
o3-mini 21, o1 6, gpt-4.1 7, gpt-4.1-mini 7, gpt-4.1-nano 7, gpt-4o 7,
gpt-4o-mini 7, gpt-4-turbo 7, gpt-oss-120b 15, gpt-oss-20b 15.
Third-party rows 11 (`third_party_reported`): gpt-oss-120b 6 (EvalEval
GPQA Diamond 80.8080808081, EvalEval MMLU-Pro 80.8, Scale SWE-bench Pro
16.2, SwissLegalEvals LEXam-hard 37.55, LEXam Leaderboard open_question
51.74 and mcq_4_choices 47.71), gpt-oss-20b 5 (EvalEval GPQA Diamond
58.5858585859, EvalEval MMLU-Pro 73.6, Liquid AI IFStruct v1.0 91.95,
LEXam Leaderboard open_question 32.12 and mcq_4_choices 40.78).

Evidence class is decided per entry from the `source.name` recorded in the
saved bytes: entries naming `GPT-OSS Model Card` are provider-reported
(mirror PRs noted in configuration), entries naming external evaluators
are third-party. Each `evalResults` array appears 4x in the page's SSR
payloads; every unique entry appears once in findings and locators cite
the first copy. Long-decimal lexemes (80.8080808081,
58.5858585859) are preserved exactly.

## Verification performed

Independent re-check (not the extraction worker's own): every row's
`value` lexeme and identifier tokens re-searched in the cited file bytes;
all 152 present. gpt-oss provider values cross-checked against the parsed
`evalResults` JSON: 60 card entries per page are 15 unique entries x 4
identical SSR copies and all 30 unique entries are emitted. README rows
re-checked against the dated-snapshot table rows.

## Gaps

26 gap records in findings name precise per-file blockers (404/stub pages,
auth-walled results, tables without exact-model cells). Third-party
evaluator rows withheld elsewhere (ScaleAI headline page rows citing
non-openai models) belong to cross-model batches, not this batch.

## Limitations

Identity basis ties rows to saved-page labels; revision identity and
current-offer mapping are phase 9. Comparability (cohort and protocol) and
calibration are out of scope for this review.
