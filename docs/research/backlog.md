# Research backlog

Generated from the assembled build tables and `data/research-inputs/backlog-tasks.json` (never the published DB). Regenerate with `make research-backlog`; `agent-routing research backlog --check docs/research/backlog.md` gates drift in `make check`.

- catalog models: 395 (in scope 302, out of scope 93)
- kinds: embedding 26, evaluation 2, image 34, language 268, realtime 9, reranking 5, speech 8, transcription 8, video 35
- image-output models: 41, all assigned to the images batch (34 image-kind, 7 language-kind with image output; the two sets are gated equal)
- exact-ID price joins: 394 models joined, 1 unjoined
- benchmark coverage: scores_found 39, no_scores_in_checked_sources 246, unresearched 17, out_of_research_scope 93

## Cross-model gaps

- unlocated claims: 56 total (48 at the frozen baseline; the §2.4 URL rule moved 8 formerly located claims here)
  - google: 4
  - images: 39
  - mistral-meta: 13
- price leaves: 5108 numeric with 2198 carrying source_specific_unresolved units across catalog-joined offers; all provider offers carry 31659 numeric leaves with 3109 unresolved (unjoined offers are models.dev inventory beyond catalog scope)
- alias/revision gaps: 252 unresolved measurements, 0 identity proposals recorded
- image-accounting gaps: 41 of 41 image-output models have no image-specific limit facts (format, mask, alpha, dimensions, count accounting)
- protocol metadata: cells missing evaluated-revision, harness or setting components stay unresolved rather than comparison-qualified; `accepted_benchmarks` remains empty
- rights: 302 in-scope models have publication_permission unknown, 0 have a decided permission
- access states over in-scope models: blocked 83, blocked_403 41, blocked_404 41, blocked_redirect 41, not_found 134, redirect_or_not_found 19, retrieved 302

## Acquisition and extraction tasks

### Batch acquisition

| task | batches | models | evidence files | closure condition |
|---|---|---|---|---|
| anthropic | anthropic | 19 | data/research/anthropic/findings.json | Every assigned ID has either a source-verified numeric cell with revision identity or a bounded checked-sources outcome naming the last inspected primary source; the 13 log-pointer source URLs are recovered or recorded unrecoverable. |
| openai | openai | 64 | data/research/openai/findings.json | Unresolved simple-evals cells keep exact lexemes and footnotes; every OSS model card with numeric tables is parsed; no turbo alias mapping without dated identity evidence. |
| google | google | 19 | data/research/google/findings.json | Reviewed 30 Gemma cells preserved byte-exact; Gemini benchmark reports located with version identity or each of the 13 unresearched rows carries a bounded outcome; failed blog provenance resolved or recorded. |
| deepseek-zai | deepseek-zai | 28 | data/research/deepseek-zai/findings.json | All unresolved V3.1/R1 cells preserved; headers, footnotes, thinking modes and evaluated revisions normalized with evidence; GLM researched from its own release material; V3.2 has verifiable table data or figure extraction stays explicitly unresolved. |
| qwen | qwen | 32 | data/research/qwen/findings.json | Uncertain URL mappings recovered or marked unrecoverable; captured cards inspected for raw numeric tables; unresolved current-offer identity never blocks raw cell preservation. |
| xai | xai | 13 | data/research/xai/findings.json | Captured model configuration parsed as capability evidence only; independent or primary numeric evaluations located per ID or bounded outcome recorded; unmatched 4.1 fast variants resolved or listed. |
| mistral-meta | mistral-meta | 19 | data/research/mistral-meta/findings.json | Reviewed 62 cells preserved; current mistral-small alias resolved with dated evidence or stays unbound; base/instruct and Llama 3.3 ambiguity retained; gated Codestral/Ministral recorded as access-blocked without auth bypass; empty Large tables and missing Muse/Medium evidence each carry an explicit outcome. |
| moonshot-minimax | moonshot-minimax | 16 | data/research/moonshot-minimax/findings.json | Status contradictions stay repaired under reviewed decisions; each variant researched from its own tables without inheriting neighboring scores; exact effort/tool/harness settings retained per cell. |
| other-text | other-text | 55 | data/research/other-text/findings.json | Step 3.5/3.7 and Trinity repairs preserved; the remaining 52 IDs each have a per-model outcome from their publisher channel, bounded to checked sources. |
| images | images | 41 | data/research/images/findings.json | Each of the 41 assigned IDs has generation/editing evaluation coverage and capability facts (format, mask, alpha, dimensions, combined accounting) or a bounded outcome; GenEval historical labels never substitute; image prices distinguished from token prices. |
| mixedbread | mixedbread | 1 | data/research/mixedbread/findings.json | mixedbread/toast-1 researched from its exact primary identity and capability sources with an explicit outcome; never silently excluded. |
| unresolved-historical | unresolved;coverage-followup | 0 | data/research/unresolved/findings.json; data/research/coverage-followup/findings.json | All 252 unresolved labels and lexemes across the unresolved (210) and coverage-followup (42) batches are preserved; immutable evaluated revision, protocol and current binding researched separately per cell; unresolved stays a first-class state. |

### Source acquisition

| task | evidence files | desired fields | closure condition |
|---|---|---|---|
| vercel-modelsdev | data/raw/vercel-models.json | full raw schemas; exact-ID join state; permitted collection/extraction terms; scoped price interpretation | Join invariants hold (394/395, jev unmatched documented); units and price interpretation scoped per dimension; permitted collection terms recorded before any production fetch; conditional-support claims limited to demonstrated headers. |
| provider-apis | (none captured yet) | documented URLs per family; commit/release identifiers; exact source labels; numeric cells; limits | Each assigned family has an inventoried documented URL set; captures follow actual release links; no guessed endpoint loops. |
| swebench | data/research/benchmark-sources/_swebench_leaderboard_extracted.json; data/research/benchmark-sources/_swebench_verified.json; data/research/benchmark-sources/_swebench_pro.json | split records; agent/scaffold/harness details | Derived JSON verified against its captured source or quarantined; splits kept separate; no Pro result inferred from the failed fetch. |
| evalplus | data/research/benchmark-sources/_evalplus_results.json | HumanEval/MBPP vs plus variants; pass@k; prompted/base variants; harness/decoding metadata | Genuine result JSON parsed with nulls retained as missing; provenance repaired under review; release/harness metadata located through official links. |
| aider | (none captured yet) | leaderboard HTML; edit format; scaffold; settings | Saved leaderboard HTML inspected; a documented structured source used if evidenced; guessed JSON/YAML paths never asserted as API. |
| livecodebench | data/research/benchmark-sources/_livecodebench_results.json | tagged result release; problem window; sampling; evaluated identity | A documented tagged result release located with window/release/sampling/identity or the source stays a recorded blocked outcome; repository listings never treated as results. |
| helm | (none captured yet) | official run/scenario artifacts; adaptation; release; evaluator metadata | Official machine-readable run artifacts located or HELM stays a documented unusable-source outcome; SPA/error pages never claimed as tables. |
| hf-open-llm | (none captured yet) | dataset/card/release revisions; licenses; harness/precision/task metadata; evaluated revision | Documented dataset/card/release revisions and licenses captured with harness metadata, or bounded outcome; landing pages never treated as scores. |
| arena | data/research/benchmark-sources/_lmarena_elo.csv | permitted export/API; candidate pool; category; votes; rating method; uncertainty; evaluation window | Genuine permitted export/API established with full methodology or Arena stays blocked; the saved 404 CSV never presented as evidence; anonymous models stay unresolved. |
| artificial-analysis | (none captured yet) | documented API; entitlement; terms; permitted storage | Documented API, entitlement, terms and storage permission verified before enabling; saved HTML-under-JSON-name never treated as API evidence; no subscription assumed. |
| geneval | data/research/images/findings.json | commit-pinned source bytes; prompt set; evaluator; settings | 42 reviewed cells preserved; captured main bytes compared with a commit-pinned source before claiming the pin; dataset/output rights recorded separately from repository MIT. |
| ella-dpg-t2i | (none captured yet) | official tables; release artifacts; settings | Official tables/release artifacts located with settings; figure-only results remain unreviewed proposals requiring source-cell verification; no invented OCR numbers. |
| image-evaluators | (none captured yet) | evaluator definitions vs result tables; published results for assigned offers | ImageReward/PickScore/Artifactor/FID/CLIP/aesthetic/text-rendering evaluations distinguished as definitions vs results; published results located for current assigned offers or dynamic/blocked sources documented. |
| mmmu-mathvista | (none captured yet) | official split/version data; model/input-resolution data | Official split/version/model/resolution data captured for vision-understanding coverage, kept distinct from generation/editing evidence. |
| host-outcomes | (none captured yet) | authorized task-success corpus; cost; latency; failure facts | Remains an explicitly authorized later collection requirement (phase 15); public benchmarks never substitute for host outcomes; not a blocker for raw evidence preservation. |
| host-registry-discovery | (none captured yet) | host registry API; spawn API; eligibility surface | Read-only discovery recorded in docs/research/host-registry.md with verified host behavior for live-family prioritization and phase 14 contracts; no integration implemented before its contract. |

## Per-model records

The machine-readable backlog is `agent-routing research backlog` (stdout TSV, 395 rows, one per catalog model). Columns: kind, in scope, batches, benchmark coverage, measurements, unresolved measurements, located claims, unlocated claims, image output, offers, price leaves, price unresolved unit, limit facts, modality facts, parameter facts, access states, licenses, publication permission

Every in-scope model has an explicit `benchmark_status` and channel assignment; out-of-scope kinds (embedding, evaluation, realtime, reranking, speech, transcription, video) are retained with their kind and no research status.
