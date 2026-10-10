# Build the Go research pipeline and offline agent router

## Scope

Plan the whole project, not only an export script: reproducible research artifacts, missing-source acquisition, evidence qualification, complete contracts, Go routing library, durable Postgres collector, actual host integration and an explicitly gated outcomes/calibration lane.

Planning only. No implementation, source acquisition campaign, paid evaluation or publication is authorized by this document. Keep this plan outside the working tree at `~/.agents/plans/agent-routing/implementation.md`. Normative contracts and ADRs belong in the repository when their work starts.

Baseline inspected at Git HEAD `e95f894` plus the current dirty working tree. `make check` passes. The earlier nine-row failure is stale. Passing these checks establishes artifact consistency, not source completeness, current model identity, comparison eligibility or routing quality. Changes already present in the working tree belong to ongoing qualification work and must not be overwritten or silently committed with unrelated work. That concurrent qualification work subsequently landed as `3cba095` during planning. Recheck the execution baseline before phase 1. This planning pass changed no repository files.

## Goals

- Rebuild `data/research.sqlite`, `data/source-manifest.tsv` and `data/source-hashes.sha256` using checked-in Go tooling and durable inputs only. Eliminate dependencies on `/tmp`, Python, filesystem modification times and the prior output database.
- Preserve exact captured bytes, published numeric lexemes, reported model labels, source locations, unknowns and review decisions. Rebuilding must never approve changed evidence.
- Account for every captured catalog entry and every applicable model/source channel. Collect all discoverable published numeric cells within an explicit versioned source inventory, not an unprovable claim about the entire web.
- Repair missing provenance and incomplete extraction before fetching replacements. Keep restricted, unavailable and unresolved evidence visible without inventing values or predecessor mappings.
- Separate source accuracy, historical revision identity, live-offer binding, comparison eligibility, calibration and rights. Passing one does not pass the others.
- Complete and independently review API, event and data contracts for each implementation slice before coding it.
- Deliver a deterministic offline Go router driven by the host's live eligible registry. No network, database or inference on the route path. CLI is only an adapter over the same library.
- Deliver automated collection with Postgres authority, bounded concurrency, durable work, event-driven invalidation and recovery without manual kickoff.
- Verify the actual host registry-to-spawn path, restart behavior, strict constraints, evidence explanations and measured performance. Keep calibrated quality claims disabled until authorized outcomes support them.

## Verified baseline and gaps

### Repository and tracking

There is no Go module, application source, checked-in exporter, migration directory or CI workflow. `Makefile` currently checks the SQLite artifact and hashes; `install` and `redeploy` intentionally fail. `CLAUDE.md` permits SQLite only for the research artifact and blocks router/collector implementation before contract and source-access acceptance.

Seven proposed specifications cover domain objects, service/event/data contracts, collection, routing, image capabilities, temporal identity and benchmark comparability. ADR 0002 accepts Go as the language and leaves CLI protocol proposed. ADR 0001 is historical; its Rust decision must not be rewritten to satisfy a mistaken review finding.

Existing work: `ar-7xd` tracks first-slice contract gaps; `ar-zdr` tracks evidence qualification; `ar-orn` tracks broader source coverage. `ar-yl1` closed a bounded reconciliation task, not the entire contract gate. `ar-1rs` and `ar-2pe` remain blocked parent work. Reconcile their dependencies during execution, not by treating the earlier closure as blanket approval.

Current dirty state includes `Makefile`, `README.md`, the SQLite artifact, both manifests, a Google/Moonshot review and research reports. Preserve it. Some earlier reports and temporary reviews contradict current bytes.

### Verified research baseline

Counts below describe this inspected input set, not constants to bake into collectors:

| Artifact | Observed state |
|---|---|
| Metadata | Research schema 3, artifact 0.1.2, research-only phase, publication permission unknown |
| Vercel catalog | 395 models; 394 exact joins to models.dev's Vercel provider; `typesafe-ai/jev` is unmatched |
| models.dev | 8,323 provider offers; other providers are distinct offers, not interchangeable Vercel prices |
| Research scope | 302 distinct language/image model IDs; 93 catalog entries outside this research scope |
| Catalog-associated measurements | 112 cells across 18 model IDs: Google 30/2, Mistral/Meta 62/9, Moonshot/MiniMax 12/4 and other-text 8/3 |
| Unresolved measurements | 252 cells: 210 OpenAI/DeepSeek cells and 42 GenEval cells |
| Source checks | All 364 stored cells appear in source-validated views, with six batch reviews and 32 review-input pins |
| Routing admission | `accepted_benchmarks` contains zero rows; source checks do not approve comparison or current-offer identity |
| Coverage view | 18 scores-found, 267 no-scores-in-checked-sources, 17 unresearched and 93 out-of-scope |
| Claims | 427 located claims and 48 unlocated claims; neither set is automatically empirical quality |
| Pricing | 31,659 numeric leaves; 3,109 carry unresolved source-specific units; 12 exact disagreements including cache components |
| Source inventory | 340 file records; all publication permissions unknown; 227 lack an HTTPS URL, including 198 captured-source paths and 29 local/derived paths |

Local/derived files do not require invented remote URLs. Missing URLs on captured source files are a provenance gap. Existing timestamps are explicitly local file mtimes, not proven HTTP retrieval times; preserve that distinction.

### Rebuild code recovered, not available to a clean checkout

`/tmp/ar-qualification/build-db.py` contains a 142-line exporter. An earlier variant and fetch logs exist under `/tmp/agent-routing-research/`. They are recovery references only, never future build inputs. The later script:

1. Deletes the database before rebuilding.
2. Reconstructs URLs partly by inspecting temporary fetch logs.
3. Enumerates files and records current mtimes, making clean-clone output differ.
4. Hardcodes review-approved batches and repins their current hashes on every run. This can bless changed evidence without a new review.
5. Embeds schema and some counts in imperative code, accepts heterogeneous findings shapes and silently normalizes some claim paths.
6. Leaves the final aggregate hash-manifest workflow outside its complete reproducible input contract.

Do not mechanically port these failure modes to Go. Use the existing export and scripts to identify required behavior, then replace their trust and publication boundaries.

### Concrete data defects and contradictions still present

- Four Moonshot/MiniMax models have measurements but remain `unresearched`: `minimax/minimax-m2`, `moonshotai/kimi-k2`, `moonshotai/kimi-k2.5` and `moonshotai/kimi-k3`. The existing gate checks only the reverse direction. A batch gap still says those exported rows do not exist. Repair statuses and notes through reviewed source changes, not exporter guesswork.
- `coverage-followup/findings.json` calls `_evalplus_results.json` a 404 HTML response. The current file is valid JSON with 125 reported model labels and pass@1 fields. Verify its provenance and correct the stale access record. Some values are null and must remain unpublished, not zero.
- `_lmarena_elo.csv`, `_swebench_verified.json`, `_swebench_pro.json` and `_livecodebench_results.json` contain `404: Not Found`, not scores. Report text claiming usable exports must be corrected against bytes.
- `_swebench_leaderboard_extracted.json` is real structured data with Multilingual, Test, Verified, Lite and Multimodal sections. It is a derived artifact; establish its exact derivation from the captured page/JavaScript before using it. Verified has 180 records in the inspected copy. Do not infer a Pro result from the failed Pro fetch.
- Google Gemma HTML contains genuine HTML-escaped `evalResults` in `data-props`. A raw substring audit previously called them fabricated; source review disproved that. Parser tests must exercise HTML decoding and table column attribution.
- OpenAI simple-evals footnotes and DeepSeek paper settings are not fully normalized. DeepSeek V3.2 and ELLA contain image-based results that were not transcribed. Source checks do not fill those gaps.
- Mistral Small's historical 3.1 measurements are not proof of the current `mistral/mistral-small` alias. Llama base/instruct distinctions and the explicitly unresolved Llama 3.3 base label must remain intact.
- Price documentation describes five input/output model disagreements, while the database has 12 component disagreements including cache read. Explain the different scopes rather than forcing counts to agree. models.dev literally supplies Mercury-2 cache read as `0.024999999999999998`; do not silently round that source lexeme into `0.025`.
- Several reports assert unreleased status, absence of all published scores or capability conclusions more broadly than the saved evidence supports. Replace unsupported global claims with bounded checked-source outcomes.
- Narrative benchmark research has no complete durable source ledger: `benchmark-sources/findings.json` uses a separate shape with zero model assignments. All specialized findings shapes must be handled explicitly rather than silently ignored.

### Missing evidence by model batch

| Batch | Scope and present scores | Required work |
|---|---|---|
| Anthropic | 19 IDs, zero cells | Recover 13 source URLs currently pointing readers at temporary logs; inspect saved model/release material for exact tables; follow documented primary release/model-card links. Keep six incompletely researched revisions explicit. |
| OpenAI | 64 IDs, zero catalog-associated cells | Preserve unresolved simple-evals results, normalize row-specific footnotes, inspect OSS model cards and documented reports; establish dated-model/alias evidence before mapping any result. Challenge shells are not model cards. |
| Google | 19 IDs, 30 Gemma cells on two IDs | Preserve reviewed Gemma values and exact 26B/31B columns; seek Gemini benchmark reports and version identity; resolve failed blog provenance. Treat current 13 unresearched rows as work, not proof of nonexistence. |
| DeepSeek/ZAI | 28 IDs, zero catalog-associated cells | Preserve unresolved V3.1/R1 cells; normalize headers, footnotes, thinking modes and evaluated revisions; inspect GLM release/model cards separately. Obtain verifiable V3.2 table data or leave figure extraction unresolved. |
| Qwen | 32 IDs, zero cells | Recover uncertain URL mappings; inspect captured Qwen3/Coder/VL cards and official release links for raw numeric tables even when current-offer identity is unresolved. |
| xAI | 13 IDs, zero cells | Parse the captured model configuration as capability evidence only; locate independent/primary numeric evaluations and alias statements, including unmatched 4.1 fast variants. |
| Mistral/Meta | 19 IDs, 62 cells on nine IDs | Preserve reviewed tables; investigate current Small alias, base/instruct labels, gated Codestral/Ministral sources, unpopulated Large tables and missing Muse/Medium evidence. No auth bypass. |
| Moonshot/MiniMax | 16 IDs, 12 cells on four IDs | Repair status contradictions; retain exact effort/tool/harness settings; research the other variants without inheriting neighboring scores. |
| Other-text | 55 IDs, eight cells on three IDs | Preserve Step 3.5, Step 3.7 and Trinity repairs; research the remaining 52 IDs through their publisher channels, with explicit per-model outcomes. |
| Images | 41 assigned IDs, zero associated cells | Collect generation/editing evaluations and independent capability, format, mask, alpha, dimension, accounting and price facts. GenEval historical labels do not cover these 41 offers. |
| Mixedbread | One ID, zero cells | Research `mixedbread/toast-1` from exact primary identity/capability sources; do not silently exclude it. |
| Historical/unresolved | 252 cells | Preserve all labels and numeric lexemes. Seek immutable evaluated revision, protocol and current binding separately. CLIP retrieval remains a baseline, not a generator. |

### Missing source coverage and source-specific work

1. Vercel and models.dev: retain full raw schemas and exact-ID joins; establish permitted collection/extraction, units and scoped price interpretation. Vercel's captured response has no demonstrated ETag/Last-Modified; models.dev has an ETag. Do not invent conditional support.
2. Provider APIs, system/model cards, release repositories and papers: inventory documented URLs per assigned family. Store commit/release identifiers, exact source labels, numeric cells and limits; follow actual release links rather than constructing endpoint guesses.
3. SWE-bench: verify saved page/JS extraction, then parse actual split records and agent/scaffold/harness details. Keep Verified, Lite, Multilingual, Multimodal, Test and any later verified Pro source separate.
4. EvalPlus: parse saved genuine result JSON after URL/status repair; distinguish HumanEval/MBPP from their plus variants, pass@k, prompted/base variants and missing cells. Find release/harness/decoding metadata through official project links.
5. Aider: inspect saved leaderboard HTML, benchmark release, edit format, scaffold and settings. Guessed JSON/YAML paths did not establish an API. Use a documented structured source if one is evidenced.
6. LiveCodeBench: saved repository listing is not result evidence; result attempts include 404 and invalid-credential responses. Locate a documented tagged result release with problem window, release, sampling and evaluated identity.
7. HELM: captured SPA/error pages do not establish a usable machine table. Locate official run/scenario artifacts with adaptation, release and evaluator metadata.
8. HF Open LLM Leaderboard: obtain documented dataset/card/release revisions and licenses, preserving harness, precision, task and evaluated model revision. A saved landing page is not scores.
9. Arena: establish a genuine permitted export/API and methodology; the saved CSV is a 404. Capture candidate pool, category, votes, rating method, uncertainty and exact evaluation window. Anonymous models stay unresolved.
10. Artificial Analysis: verify documented API, entitlement, terms and permitted storage before enabling it. Saved HTML under a JSON filename is not API evidence. No subscription or paid access is assumed.
11. GenEval: preserve 42 reviewed cells and the repository commit evidence; compare captured `main` bytes with a commit-pinned source before claiming the commit pins those exact bytes. Collect prompt set/evaluator/settings separately. Repository MIT does not establish dataset/output rights.
12. ELLA, DPG-Bench and T2I-CompBench: find official tables/release artifacts and their settings. If only figures exist, treat transcription as an unreviewed proposal requiring source-cell verification. No invented OCR numbers.
13. ImageReward, PickScore, Artifactor, FID/CLIP/aesthetic and text-rendering/editing evaluations: distinguish an evaluator's definition from model result tables. Locate actual published results for current assigned offers. Document dynamic/blocked sources instead of claiming none exist.
14. MMMU and MathVista: obtain official split/version/model/input-resolution data. Vision understanding is not image generation or editing.
15. Host outcomes: no representative task-success, cost, latency or failure corpus is available here. Public benchmarks cannot replace it. This is a later authorized collection requirement, not a blocker for raw evidence preservation.

### Contract blockers verified against current specifications

The independent review at `/tmp/ar-qualification/review-contracts.md` identifies real missing first-slice contracts. Preserve those findings in durable tracked work, not a dependency on this temporary file:

- Snapshot byte envelope, canonical hash domain, section ordering, resource limits, compatibility and concrete Go API types are unspecified.
- Compile's distribution class is absent from snapshot authority DDL/object fields; channel classification is caller-supplied without a configured authority. Compile idempotency also needs distribution class.
- Snapshot manifests do not pin compile-time rights policies; publication lacks persisted current-rights audit and a defined isolation/locking protocol against revocation.
- Uncalibrated cost/performance admission and complete fallback ordering are not executable rules.
- Collector output hash, exact work-tuple authority, job state derivation, terminal retry semantics and outcome enums are underspecified.
- Observation body locator, collector version, parser version and event fields disagree.
- Capability/operation/direction/role matrices are incomplete; conditioned generation and alpha tokens are not fully enumerated. Image subtype prose incorrectly says four where five variants are listed, and the parent delegation omits `image_semantic_assertions`.
- Benchmark canonical grammar and exact lexical rules are unspecified; family/component applicability is not a complete registry. Hardware applicability currently conflicts with the closed family vocabulary for static accuracy.
- Event encoding and whether payloads carry facts or retained authority-row references remain incomplete.
- Outcome vocabulary, privacy/retention, task validators, runtime policy persistence and calibration tables remain future contracts. Their absence blocks those slices, not all unrelated collection planning.

## Approach

### Boundaries and ownership

Reuse the existing two bounded contexts. Knowledge is the core domain; evidence collection, research export, identity and projection are groupings within it, not new bounded contexts. Host Execution owns credentials, eligibility, spawning, retries, resource reservations and validators. Vendor sources sit behind source-specific anti-corruption adapters. Postgres, SQLite export and transport types stay at the edges.

Retain Source, CollectionJob, SourceObservation, ModelRevision, ProviderOffer and RoutingSnapshot as the proposed aggregate roots, with constructors and methods enforcing their invariants. Assertions, bindings, source runs, measurement metadata, receipts and review decisions are immutable value facts owned or referenced as specified. Route remains a domain service. Do not invent an Agent, ExportJob or mutable model-ranking aggregate.

Store evidence and explicit decisions. Derive coverage, source-checked projections, cohort membership, current identity, cost admission and rankings. Every new object must have complete fields, classification, ownership/cardinality and lifecycle before implementation. Every transition must trace through API, event and DDL contracts or state why a pure/local operation has no event/table.

### Dependency choices and reuse

No project dependencies exist yet. Use Go stdlib for CLI flags, file IO, hashing, JSON/TSV interchange, HTTP, exact arithmetic and atomic in-memory snapshot swaps. Decode upstream JSON numbers with `json.Number`, retain original lexemes and use scaled integers or `math/big` for arithmetic; no float64 money path.

Proposed dependency choices, pinned and license-audited when their slice begins:

- `modernc.org/sqlite` through `database/sql`, confined to the explicitly authorized research export adapter. Prefer an in-process Go rebuild over shelling out to `sqlite3` or requiring CGO. Inspect transitive size and pin compatible versions before adding it. Registry: https://pkg.go.dev/modernc.org/sqlite.
- `github.com/jackc/pgx/v5` for future Postgres authority. Reuse the migration mechanism provided by rookery if applicable rather than writing another runner. Registry: https://pkg.go.dev/github.com/jackc/pgx/v5.
- `golang.org/x/net/html` for source DOM parsing and `github.com/yuin/goldmark` with table support when Markdown parsing is required. Do not use regex as a general HTML/Markdown parser. Registry: https://pkg.go.dev/github.com/yuin/goldmark.
- `testcontainers-go` Postgres module for authority integration tests: https://golang.testcontainers.org/modules/postgres/.
- Prefer native supervised file-change facilities, with `fsnotify` only if needed for the actual host. A file-event loss must not strand work: initial reconciliation and durable input fingerprints supplement events.

No dependencies are installed in this planning pass. Maintenance, exact versions and transitive licenses remain execution checks. No new SQL engine, queue product, vector store, model registry or distributed service is justified. Scaffold the collector's service wrapper from `github.com/guygrigsby/rookery` on `perch` when that service slice is accepted; do not impose daemon scaffolding on an embedded library or one-shot research exporter. Preserve canonical-repo source ownership.

### Go rebuild tool: required replacement for the missing script

Proposed command family: `agent-routing research check`, `agent-routing research rebuild` and `agent-routing research coverage`, implemented by `cmd/agent-routing` over reusable internal research packages. Names and wire schemas become normative in a small research-tool contract before code. These are local OS-user operations with no network, no provider calls, no production mutations and no inferred approvals.

`research check` validates the durable inputs, review pins and generated artifacts without modifying them. `research rebuild` reads the same inputs and emits a complete replacement artifact set, or fails while preserving the last good set. `research coverage` derives a machine-readable inventory and a human report from the same facts. Structured results report input fingerprint, schema/artifact versions, output hashes and diagnostics; stdout contains the requested result, stderr contains safe diagnostics. Distinct documented exit codes cover invalid inputs, stale review, inconsistent export and IO failure. Repeating an identical successful rebuild is a no-op.

Required durable inputs:

- Existing raw catalog bodies/headers, assignments, findings files, captured sources and review reports.
- A checked-in capture inventory, proposed under `data/research-inputs/`, holding explicit local-versus-captured origin, recoverable URL, retained timestamp and timestamp basis, expected byte hash/size, source/license references and provenance limitations. Seed it from the inspected manifest and defensibly recovered logs; never rederive historical retrieval times from a fresh clone's mtime.
- A versioned research schema and build configuration with artifact version, scope rules and pinned source revisions. Derive catalog/join counts from the frozen input catalog while checking assignment completeness independently; historical audit counts are fixture expectations, not future model limits.
- Immutable review decisions plus a separate path/hash pin list under `data/reviews/`. Each decision names its scope, batch, reviewer/report, reviewed input hashes and decision version/time. Existing report evidence must support the seeded scope; no hardcoded list of approved batches inside the exporter.
- An explicit adapter/format version for heterogeneous findings shapes, including catalog research, benchmark-source inventory, ordinary model batches and unresolved-only batches.

Generated outputs remain the three familiar paths. Treat the database and both manifests as a recoverable generation: stage files on the same filesystem, finish and close SQLite, validate the complete staged set, sync it and publish under a shared lock with a small durable replacement journal and backups. Check/rebuild first recover or roll back an interrupted publication. Never unlink the current database first. Three renames are not one atomic transaction; make all project readers use the shared lock/recovery path, document that external SQLite inspection occurs after a completed build and prove crash recovery. If the contract review prefers a single directory pointer instead, settle that once before implementation and preserve documented compatibility paths.

Hashing must be acyclic. Source inventory records input bytes. The generated TSV records those input identities, not its own hash or a database hash. The final hash manifest covers inputs, the completed DB and generated TSV but excludes itself, temporary files, journal and backups. No `time.Now`, current mtime, absolute path, map iteration or directory-enumeration order may affect deterministic output. Sort rows, assign stable identities, pin schema/driver versions and validate repeat builds in separate clean directories. SQLite byte reproducibility is required within the pinned toolchain; compare canonical relational content as a diagnostic if engine versions change, never silently waive differences.

Review pins are inputs, not regenerated endorsements. A changed findings file, report or cited source invalidates the old decision. The tool may report precisely what changed, but cannot replace reviewed hashes, mark a batch source-validated or relax the acceptance view. New raw proposals remain in an explicit unreviewed staging lane until source review; the canonical source-validated export must continue satisfying its gate. Newly reviewed evidence gets an append-only decision with a new pin set.

Reconstruct every current table/view, including claims quarantine, source access, gaps, exact joins, raw price paths, unresolved identities and review inputs. Preserve stable IDs for existing rows or provide a reviewed migration/crosswalk. Do not silently drop unknown findings keys, rewrite input evidence or convert missing values into zero. Treat model association in the legacy schema as provisional, not proven revision identity.

Validate input paths before opening them: repository-relative only, no traversal, symlink escape or arbitrary external file reads. Parse fetched content as data and never execute JavaScript, shell snippets or model-card instructions. Bound file sizes and parser resources. Seed an unsupported source locator as an explicit gap, not a fabricated match.

`make build` and `make check` verify that inputs and outputs agree without rewriting tracked artifacts. `make research-rebuild` is the developer recovery entry point, not the production trigger. Production collection reacts to committed source/parser/policy changes and uses the same pure transformations. CI always runs the freshness comparison, so a missing rebuild cannot pass unnoticed.

### Evidence and missing-data policy

Maintain separate classifications for captured bytes, located source cell, source-reviewed value, exact historical revision, current-offer binding, comparable cohort, calibrated task estimate and permission for each use. Show all relevant reasons, not a single accepted flag.

Raw extraction does not require current-offer resolution or a complete evaluation protocol. If a table contains a genuine result for an unresolved historical label, preserve it in the unresolved lane. Missing critical metadata prevents comparison, not evidence capture. Samples, uncertainty, dates, evaluator versions, quantization, image settings, scaffold and tool policy must be source-backed; no inferred defaults.

Generate per-model/per-channel missingness with checked URLs, exact source files, parser version, outcome, blocker, next acquisition target and latest attempt evidence. Every inventory channel gets a result even when blocked. Source-wide restrictions defer that source without aborting other work. Research reports become generated or drift-checked views, not an independent hand-maintained catalog.

Perform independent source-cell review, then identity/protocol review, then rights admission. Existing source reviews are retained but cannot authorize current aliases. Recheck every finding against actual source bytes, especially escaped embedded data and multi-column tables. After two failed reviews of one work unit, stop that unit and report its blocker; a new session does not reset the count.

### Proposed operational limits and unresolved decisions

Proceed without paid APIs or active evaluation by default. Optional inference extraction stays disabled. Remote snapshot distribution stays disabled until a signing/trust and redistribution contract is approved. Source permissions remain unknown until evidence-backed policy is recorded; unauthorized lanes stay blocked while synthetic fixtures and permitted lanes proceed.

Retain current proposed refresh ceilings and per-host politeness as proposals until source-specific policy approval. Confirm the target host integration, service deployment owner and billing regime before enabling their corresponding lanes. Choose module path and release packaging from the repository's actual remote and conventions at execution time. Those decisions are tracked prerequisites, not reasons to omit the rest of the plan.

## Implementation sequence

1. **Freeze the execution baseline and reconcile work ownership.**
   - Dependencies: plan approval only.
   - Re-read `git status`, logs, beads and the current DB. Record input hashes before editing. Preserve concurrent qualification changes and identify which are already reviewed.
   - Retain the current green gate as a baseline. Check in only coherent, independently reviewed existing research changes in their own unit when execution begins.
   - Convert this plan's phases into linked beads beneath `ar-7xd`, `ar-zdr` and `ar-orn`; link the blocked parents to remaining work. Preserve historical closures and review limits.
   - Recover relevant temporary scripts, source-log provenance and review findings into durable, typed inputs or cited reports. Do not copy temporary agent logs wholesale: inspect for credentials and irrelevant instructions first.
   - Exit: execution inventory identifies authoritative files, ownership, review scope and dependencies without relying on session memory.

2. **Close the research-tool contract and design its trust boundary.**
   - Dependencies: 1. Can run alongside source inventory audit and production contract work.
   - Specify the three Go research subcommands, all inputs/outputs, path rules, diagnostics/exit codes, no-network behavior, exact numeric handling, schema/version policy, deterministic ordering and crash-recovery publication.
   - Define the capture inventory, review decisions/pins and legacy findings adapters in `docs/specs/research-artifact.md`. Include explicit treatment of local files, unknown remote URLs, historical mtimes, unsupported fields and source-label-only measurements.
   - Define research artifact DDL/version migrations separately from future Postgres DDL. Document no domain events for pure local checking/export; the input-change coordinator owns automatic invocation later.
   - Specify review invalidation and canonical-versus-unreviewed staging so new captures do not force unreviewed scores into the approved export.
   - Review the contract independently, including malicious input paths, approval laundering and mixed-generation publication. No exporter code before this gate passes.

3. **Build a durable provenance and review ledger.**
   - Dependencies: 2 for the data shape; source audit can start after 1.
   - Seed inventory from all 340 current file records, separating actual fetched evidence from assignments, findings and reports. Preserve existing bytes and historical timestamp basis.
   - Recover the 198 captured-source URL gaps from retained fetch evidence where possible. Commit only verified URL/path relationships with their recovery basis. Ambiguous or absent mappings remain explicit and block affected evidence promotion, not all unrelated research.
   - Seed six existing source-review decisions and 32 pins from the reviewed bytes/reports; cross-check every pin independently. Keep catalog acceptance and measurement source checks distinct.
   - Fix the stale EvalPlus access claim, impossible successful-export claims and bounded-coverage wording through reviewed evidence corrections. Do not overwrite original fetched responses.
   - Correct the four MiniMax/Kimi research statuses and contradictory batch note, then obtain a new review covering the changed findings bytes. Do not automatically carry over an old whole-file approval.
   - Exit: all information required for a clean-checkout rebuild is versioned; unrecoverable provenance is listed rather than guessed.

4. **Implement the Go rebuild/check/coverage tool test-first.**
   - Dependencies: 2 and 3.
   - Create `go.mod`, the shared command entry point, minimal research adapters and embedded/versioned SQLite schema. Inspect dependency licenses and use the chosen in-process SQLite driver only here.
   - Add failing tests for the actual gaps: inverse status mismatch, stale access classification, changed reviewed source, missing review pin, temporary-log dependency, mtime-sensitive output and destructive interrupted rebuild.
   - Implement typed readers for each existing findings shape, catalog flattening, exact decimal prices, claim quarantine, unresolved rows, source access and review projections. Reject malformed/duplicate identifiers and non-finite numeric values with file/row diagnostics.
   - Rebuild all tables/views from durable inputs. Compare every table and source lexeme with the frozen current export. Document intentional corrections rather than treating the current DB as infallible.
   - Add deterministic hashing, stable row IDs, staged validation, publication locking/journal, fsync and recovery. Parallelize independent decoding/hash work with bounded workers; keep one sorted SQLite write transaction because it is a single-writer export.
   - Wire `make build`, `test`, `lint`, `check` and `research-rebuild`; add CI using those exact targets. Keep install/redeploy explicitly unavailable for products not built yet.
   - Run the real CLI from a clean checkout with the old DB/manifests absent, `/tmp` unavailable and input mtimes changed. Two builds must match; a failed one must preserve the previous usable generation.
   - Exit: reviewed Go replacement exists, current gates stay strong and there is no Python/shell exporter dependency.

5. **Generate the complete source and missing-data backlog.**
   - Dependencies: 3; coverage tooling from 4 automates it when available.
   - Materialize one record per catalog model and applicable channel, including exact-ID association, historical scores, capability evidence, price completeness, checked primary sources, independent benchmarks, access state and rights.
   - Cover all 302 in-scope IDs, retain all 395 catalog entries and explicitly identify out-of-scope types. Keep 41 image assignments separate from all language models with image output.
   - Expand every item in the batch/source inventories above into an assigned acquisition/extraction task with file/URL evidence, exact desired fields and a closure condition. Begin read-only discovery of the target host registry/spawn API here so live-family prioritization and phase 14 contracts use verified host behavior, not an assumed integration.
   - Mark the 48 unlocated claims, 3,109 unresolved price leaves, image-accounting gaps, alias/revision gaps and incomplete protocol metadata separately. A task closes with verified evidence or a bounded documented unavailable outcome, not a promise that the web has been exhausted.
   - Generate the report from the same records that feed the export; add drift guards for report counts, paths, statuses and examples.
   - Exit: no discovered in-scope model silently lacks a research status/channel assignment.

6. **Extract usable saved evidence before collecting more.**
   - Dependencies: 2/3; 4 supplies the reproducible output path.
   - Fan out independent source-family workers: provider tables, SWE-bench/EvalPlus, historical OpenAI/DeepSeek and image evaluations. Shared adapters and file ownership are assigned before workers edit.
   - Build deterministic parsers for saved Markdown tables, HTML tables, escaped `data-props` JSON and documented structured records. Preserve exact raw lexemes and row/column/header locators. Parse captured JS data without executing it.
   - Validate the SWE-bench derived JSON against its captured source, retaining split/scaffold fields. Extract genuine EvalPlus cells while retaining nulls as missing. Revisit provider files previously rejected merely because current-offer identity or full configuration was unknown.
   - Add regression fixtures for Gemma column/escaped-data handling, Meta base/instruct distinctions, StepFun/Trinity configuration and all 42 GenEval cells. Normalize OpenAI and DeepSeek footnotes only where the source states them.
   - Preserve proposals from figure/OCR extraction separately and require exact source-cell review before ingest. Do not make a vision inference call without an accepted spend policy.
   - Independently review each changed measurement set against the bytes, append decisions/pins and rebuild. Existing 364 reviewed cells must not silently vanish or acquire stronger status.
   - Exit: saved usable evidence is parsed or has a precise documented extraction blocker; more data does not automatically mean more comparable models.

7. **Complete production contracts in independent slices.**
   - Dependencies: 1. Parallel with 2-6; no production implementation before the relevant contract slice is accepted.
   - Snapshot/library slice: define canonical encoding and hash domains, concrete Go request/result/error types, numeric/time units, constructor invariants, compatibility, resource limits, immutable generations and adoption failures. Define no-calibration behavior for both routing modes. Proposed conservative baseline: a numeric quality floor rejects without calibration; `allow_fallback` admits only hard-constraint-compliant uncalibrated candidates, orders them by an explicitly named deterministic cost/latency/ID policy and reports no performance claim. No cross-cohort arithmetic.
   - Source-access slice: close `docs/specs/source-access.md` against the existing Source/source_policies model. Define who may register permission decisions, what terms/license evidence supports each separate research/extraction/redistribution decision, the exact policy identity/hash, effective and expiry handling, source scope, same-time conflict detection and revocation semantics. Unknown, missing, expired or conflicting permission denies the affected production use. Define retained evidence and audit requirements, approved metadata-fetch scope and access-blocked coverage/re-entry. Independently review this contract before phase 8 records enabling policies or performs production acquisition.
   - Rights/publication slice: persist snapshot distribution class and compile-time policy pins, include them in hashes and idempotency, define authoritative channel configuration and publication audits. Serialize policy revision and publication through a stated Postgres isolation/locking protocol; re-read policy after lock acquisition. A transaction boundary alone does not remove the revocation race.
   - Evidence slice: enumerate complete capability/operation/role/direction matrices, conditioned-generation/alpha semantics, image count/byte/pixel limits and accounting. Fix subtype/delegation inconsistencies. Define the benchmark grammar's syntax, exact lexical canonicalization, required-family registry, applicability, interval precision and hash serialization.
   - Collector slice: define immutable work identity, typed stage-output canonical hash, same-key/same-output versus conflicting-output behavior, explicit lifecycle derivation and attempt enums, quarantine/re-entry and terminal retry facts. Reconcile body locator, collector version and parser version across object/API/event/DDL. Enumerate all storage tokens and exact cross-row trigger contracts.
   - Event slice: define each versioned envelope/payload, fact-versus-reference semantics, row retention, replay behavior, schema mismatch handling, ordering and consumer deduplication. Add publication/channel/class identifiers where consumers need them.
   - Produce executable examples, exact DDL/Go schemas and a transition-to-API/event/table cross-check. Render diagrams. Update domain model and append new ADRs for newly accepted decisions; do not rewrite ADR 0001.
   - Independent contract review plus fresh security review for identity, rights, ingestion and trust boundaries. Verify findings before accepting them. Record slice-specific acceptance; keep outcome/calibration and remote-distribution contracts explicitly unopened until their phase.
   - Exit: `ar-7xd` closes only when its full stated first-slice gaps are resolved and reviewed.

8. **Settle source access and collect missing primary/independent evidence.**
   - Dependencies: 5 and source-policy contract in 7. Saved-byte inspection may continue without enabling a production fetcher.
   - For each source above, record terms URL/text evidence, checked/effective/expiry times and separate research, extraction and redistribution decisions. Model/code license alone does not establish data rights. Unknown permission disables the corresponding production use.
   - Acquire allowed documented sources in parallel across hosts, with a shared per-host request limit, initial one-second spacing or stricter published rules, real User-Agent, body limits and both Retry-After formats. Persist original bytes, status, redirect chain, source URL, observation time and revision metadata.
   - Prefer official releases, exports and documented APIs. Revalidate dynamic-page claims against actual serialized data; use browser/rendered extraction only where permitted and needed, with exact source retention. No guessed endpoint loops, captcha bypass or credential workarounds.
   - Work all provider batches and independent benchmark families in the source register. Prioritize current live-registry families once the host mapping is available, but retain per-channel outcomes for every assigned ID.
   - Pin GitHub/HF releases where documented; verify mutable `main` captures against commit bytes before claiming an immutable revision. New inaccessible sources become explicit coverage facts with retry conditions.
   - Exit: versioned inventory has a source outcome for every requested channel; access/budget/rights blockers remain individually tracked and cannot be hidden by an aggregate PASS.

9. **Qualify identities, protocols, prices and capability facts.**
   - Dependencies: 6-8 and accepted evidence contracts.
   - Construct exact historical revision facts from publisher-supported identity, then separately bind live provider offers with effective time, recorded time, authority and freshness. Candidate name similarity never approves a binding.
   - Normalize benchmark settings into known/unknown/not-applicable components with evidence. Derive cohorts only after completeness checks; maintain unresolved scores for display. Keep provider, independent, preference and host-observed evidence distinct.
   - Populate capability and limit facts per offer, operation, direction and image role. Obtain mask/reference/alpha/format and combined-context accounting independently. No broad multimodal label substitutes for these facts.
   - Interpret pricing paths by documented dimensions, not final key name alone. Distinguish numeric thresholds/counts from monetary values, flat image prices from tier matrices, zero from unknown and token prices from per-image/megapixel/request/second prices. Retain every original path/value even if normalization cannot yet interpret it.
   - Keep exact disagreement reporting separate from policy tolerances. Preserve Mercury's source decimal and all real provider disagreements; no averaging across offers or unsupported rounding.
   - Produce field-level qualification/rejection reasons with source IDs, policy versions and freshness. Derive only named within-cohort descriptive strengths/weaknesses; missing scores never imply weakness.
   - Exit: usable facts are explicitly bounded by identity, protocol and rights. Zero comparison-qualified facts is an acceptable truthful result, not permission to weaken the contract.

10. **Implement the Go domain and offline snapshot library.**
   - Dependencies: accepted snapshot/library/evidence contracts from 7. Synthetic owned fixtures permit development while real-source rights remain blocked.
   - Implement minimal domain objects with constructor-enforced invariants and aggregate methods; keep string IDs, vendor types and database types in their proper layers. Reuse pure parsers/normalizers from research where semantics genuinely match.
   - Implement canonical snapshot encoding/validation, referential closure, policy pins, excluded-fact explanations and bounded decoding. Separate creation timestamps from semantic determinism exactly as the contract specifies.
   - Implement load-once immutable indexes and `sync/atomic` generation swaps. Active readers keep their old generation; a failed adoption retains the last integrity-valid one. Freshness is checked against the current routing instant, not reset by loading.
   - Compile synthetic snapshots first, then only rights-admissible real inputs. Do not promote the research SQLite DB into authority or make route depend on it.
   - Exit: byte and replay tests pass; corrupt/oversized/incompatible snapshots cannot replace a valid generation; race tests cover simultaneous readers/reloads.

11. **Implement admission, cost and deterministic routing.**
   - Dependencies: 10 and accepted no-calibration policy from 7. Cost and capability algorithms can be developed in parallel after their shared types are fixed.
   - Intersect the snapshot with the current eligible host registry, preserving diagnostics for unmapped models. Enforce credentials/entitlements, devices, privacy, region, operation, modalities, tools, structured output, context and image constraints before ranking.
   - Calculate conservative bounded costs from applicable exact-decimal components; default to zero cache reuse unless proven. Incomplete price or accounting fails strict budget admission. Unknown latency fails strict deadlines.
   - Apply explicit known/fallback pools and complete deterministic ordering. Strict numeric quality floors fail without accepted calibration. Performance mode cannot quietly synthesize probabilities or average incompatible benchmark scales.
   - Return offer/configuration, admissible alternatives, rejected reasons, explicit relaxations, identity cutoffs, evidence IDs, price completeness, freshness and pinned versions. Selection itself performs no spawn, network call or persistence.
   - Benchmark warm route p50/p95/p99 with explanations and concurrent snapshots at 1,000 and 10,000 offers; record memory/allocations and cold load separately. The proposed p99 under 1 ms is a target to measure, not a guaranteed claim.
   - Exit: library passes replay/property/fuzz/race tests with no hard-constraint escape and no route-time IO.

12. **Implement Postgres authority and durable collection.**
   - Dependencies: collector/event/rights contracts from 7; source adapters from 6/8 can be reused after boundary review. Can run alongside 10/11.
   - Scaffold only the collector service wrapper from rookery after inspecting its existing facilities. Add numbered Postgres migrations, restrictive FKs, exact CHECKs, 5NF child tables and documented cross-row deferred constraints. No JSON/JSONB and no SQLite test substitute.
   - Implement byte-exact observations, 304 references, typed assertions, coverage and append-only corrections. Implement jobs, fenced leases, attempt/completion facts, durable outbox and idempotent consumers according to the accepted work/output contract.
   - Persist output before completion. Crash recovery must adopt a validated output only for the identical work tuple under a current lease. Duplicate events cannot release downstream work early. Source/parser/policy changes re-enter quarantined work automatically.
   - Drive collection from upstream changes; use durable conditional-revalidation deadlines only where upstream events are unavailable. Run startup reconciliation, persist retry state/backoff and handle Postgres/network outages without losing accepted output.
   - Share host rate limits across workers, parallelize independent hosts/stages, honor cache headers and keep failures source-local. No generic queue abstraction or scheduler product unless the existing Postgres/native facilities cannot meet a concrete requirement.
   - Exit: real-Postgres tests prove restart, stale-worker, duplicate-event and partial-outage behavior; no operator button is required after installation.

13. **Compile, validate and publish rights-admissible snapshot successors automatically.**
   - Dependencies: 9, 10 and 12 plus rights/publication acceptance.
   - Invalidate only affected projections on changed evidence, rights, binding, parser or policy events. Unchanged content may confirm freshness without redundant extraction.
   - Compile from pinned observations and authorizing policies; independent validation checks hashes, schema, identity, field freshness, comparison eligibility and distribution class.
   - Recheck current rights under the agreed transaction/lock protocol, persist publication audit and emit the exact publication event. Commit publication intent durably and reconcile filesystem artifact/pointer changes after crashes; do not imply SQL and file rename share one transaction.
   - Retain at least two validated generations and adopt atomically. Rollback restores bytes, never freshens stale facts or resurrects denied permissions. Define and test what a post-publication rights revocation does to subsequent adoption and existing local use according to the accepted rights contract.
   - Keep remote publication disabled pending signing, key lifecycle, replay protection and redistribution approval. Local filesystem ownership remains a trust boundary.
   - Exit: source change reaches a validated local successor without manual kickoff; failed builds leave the prior generation intact while per-fact freshness still applies.

14. **Define and implement thin CLI and real host integration.**
   - Dependencies: 10/11; use synthetic snapshots before licensed real ones are available. Protocol contract precedes this adapter.
   - Specify local process request/response framing, schema negotiation, timeouts, cancellation, stderr discipline, errors and snapshot/runtime generation handoff. Measure one-shot process overhead before choosing any persistent transport. No daemon or HTTP API is implied by a CLI.
   - Implement CLI commands as adapters over the Go API, not another ranker. For a non-Go host, decide whether measured process startup is acceptable before selecting a long-lived process boundary; any new service decision gets its own contract/ADR and template review.
   - Read the target host's actual registry and spawn contracts. Map provider/upstream identities at the host adapter; never hardcode model lists. Implement host-owned attempts, accumulated spend, health updates and resource reservation enforcement across restarts.
   - Drive actual subagent spawning with returned eligible offers. Exercise text, vision, generation and editing independently when the host supports them. Missing host operation support remains a documented blocked test, not simulated success.
   - Record privacy-filtered operational receipts only under the accepted receipt subset. Do not start paid trials without an explicit spend policy; if none exists, use an already authorized execution or report that the paid end-to-end path remains unverified.
   - Exit: actual live-registry-to-spawn evidence exists, with startup, routing and provider latency measured separately.

15. **Define and validate outcomes/calibration as a separate, gated delivery.**
   - Dependencies: accepted host outcome/privacy/spend contracts, representative authorized tasks and 14. This phase does not block an honest uncalibrated release.
   - Specify typed outcome/usage/validator tables and events, task-family/version identity, success/failure/cancelled/refused categories, retry/cost attribution, retention, consent and deletion policy. Byte-exact receipts are not a substitute for queryable typed facts.
   - Build representative local evaluation tasks and mechanically verifiable validators. Include failures, retries and abstentions; provider completion and judge opinion are not interchangeable with validated task success.
   - Define sample eligibility, configuration/cohort matching, train/holdout separation, leakage controls, uncertainty method, minimum support and calibration expiry. No arbitrary sample-count threshold or probability mapping is invented in this plan.
   - Use only authorized budgets. Local GPU work requires the host's admission checks and `gputex`; no overlapping unmanaged workloads.
   - Validate held-out estimates per task/cohort and publish a versioned calibration only after independent review. Unsupported task/configurations remain uncalibrated. Compare costs against a baseline policy rather than claiming routing savings from list prices alone.
   - Exit: numeric quality admission remains disabled until this evidence and all three contracts pass; track unavailable data explicitly.

16. **Operationalize, release and close only evidenced work.**
   - Dependencies: releasable slices above, their reviews and actual-path verification.
   - Provide `build`, `test`, `lint`, `check`, `install` and `redeploy` through the Makefile; CI calls the same targets. Add race, fuzz, fixture, provenance, policy and restart checks where each applies.
   - Install the collector under a durable supervisor with persistent state: systemd user service plus linger on a Linux deployment, or the accepted native macOS supervisor if deployed there. Source changes flow through this canonical repo, commits and pulls, never one-off remote edits.
   - Add source lag, blocked-permission, parse quarantine, unresolved identity, empty cohort, outbox backlog, stale lease, publication rejection and per-field freshness diagnostics. Logs contain no credentials, prompts or source-body secrets.
   - Test database backup/restore, content-store restoration and snapshot rollback; retain manifests/review pins so provenance survives recovery. Ensure automatic recovery reconciles missed events on restart.
   - Audit release contents for third-party restrictions. OSS code may use MIT; the local research corpus is not automatically distributable. Publish code/owned fixtures only when evidence rights remain unknown. Pushing, releases and PRs require an explicit request.
   - Bump distributed artifact/code versions for content changes and schema versions for contract changes. Commit each coherent tested/reviewed unit separately, excluding plans and unrelated concurrent work.
   - Close source-coverage work only relative to the declared inventory; close routing delivery only with actual host-path proof. Leave blocked rights, sources, remote distribution and calibration as linked issues with exact enabling conditions.

## Files and areas

All paths below are proposed execution targets, not files created by this planning task.

- Existing policy/docs: `CLAUDE.md`, `README.md`, `Makefile`, all affected `docs/specs/*.md` and `docs/research/*.md`. Replace unsupported claims and generate or guard duplicated facts. Do not rewrite historical ADRs.
- New contracts: `docs/specs/research-artifact.md`, `snapshot-format.md`, `go-api.md`, `source-access.md`, `host-adapter.md` and later `host-outcomes.md`/`calibration.md`, only where separate ownership prevents duplication. Fold into existing specs where that is clearer.
- New decisions: next available numbered `docs/adr/` records for accepted artifact/review trust, production snapshot publication and any new process/deployment boundary.
- Durable research inputs: proposed `data/research-inputs/` build metadata/capture inventory; `data/reviews/` versioned decisions and input pins; existing source review reports and `data/research/*/assignment.json`/`findings.json`.
- Evidence: `data/raw/`, `data/research/*/sources/`, `saved_sources/` and benchmark-source captures. Preserve originals; additions/corrections retain their own provenance. Research SQLite schema SQL belongs in the research adapter, not production migrations.
- Generated compatibility artifacts: `data/research.sqlite`, `data/source-manifest.tsv`, `data/source-hashes.sha256` and source-derived coverage reports. No manual SQL repair as the normal update path.
- Go foundation: `go.mod`, `go.sum`, `cmd/agent-routing/`, a public routing package and small internal packages for research, source adapters, exact facts, snapshot encoding, Postgres persistence and collection. Final package split follows dependencies, not one package per noun.
- Production persistence: numbered `migrations/` after the accepted Postgres contract, with real integration tests and owned fixtures in `testdata/`.
- Host integration: the canonical host adapter repository if it lives elsewhere; identify its path and ownership before editing. No edits to installed npm packages or deployed source copies.
- Automation: `.github/workflows/` or the repository's chosen CI, generated command/docs guards and supervisor packaging from the accepted collector template.
- Tracking: `.beads/` only for coherent issue updates during execution. Plan stays at `~/.agents/plans/agent-routing/implementation.md` and is never committed.

## Risks and edge cases

1. **Green checks with semantic drift.** Current one-way status tests pass despite scores marked unresearched. Add bidirectional and source-shape checks; still require semantic review rather than trusting a count.
2. **Approval laundering.** Recomputing source hashes alongside a hardcoded accepted-batch list makes mutated evidence look reviewed. Immutable review pins are authoritative inputs; rebuild cannot update them.
3. **Missing original provenance.** Temporary logs may disappear and mtimes change on clone. Recover defensible mappings now; retain unknowns and original timestamp basis when recovery is impossible.
4. **Concurrent agent work.** Dirty files and review outcomes may change during execution. Recheck status and input hashes before publishing/committing; use separate file ownership and do not reset other work.
5. **Exact decimals and source semantics.** JSON floats, percent/fraction conversions, rounded cache prices, numeric threshold leaves and overlapping price bands can produce misleading costs. Preserve raw values and validate dimensions before normalization.
6. **Label/identity leakage.** Reviewed table cells do not prove live alias bindings, base/instruct equivalence or provider equality. Retain historical revisions, ambiguous facts and unresolved labels. Alias freshness cannot move old benchmark evidence.
7. **Comparability without calibration.** Shared rows or equal unknown metadata are not a cohort. Raw benchmark scores cannot meet a numeric success floor. Do not weaken the comparison policy merely to populate rankings.
8. **Image evidence conflation.** Vision, generation, editing, references, masks, alpha and output formats require separate evidence. Static GenEval baselines cannot fill current image-offer coverage.
9. **Rights and trust races.** Unknown technical access is not permission. Caller-supplied channel class and unlocked current-policy reads can allow inappropriate publication. Require authoritative configuration, transactional audit and tested synchronization.
10. **Filesystem/database crash windows.** Deleting the old DB or renaming multiple files without recovery exposes a mixed generation. Research publishing needs shared recovery semantics; production SQL publication and filesystem adoption need durable reconciliation rather than a false atomicity claim.
11. **Untrusted fetch/parse surfaces.** Revalidate redirects and resolved IPs, block private/link-local targets, bound decoded sizes, reject path escapes and never execute fetched content. Security review includes source-induced prompt injection into optional extraction.
12. **Source restrictions and cost.** Paid aggregators, gated cards, unavailable exports and figures may remain blocked. Record each missing source and enabling condition; no spend or bypass is implied by this plan.
13. **Durable recovery and concurrency.** Stale leases, duplicate events, restart after output commit and multiple workers must not lose work or acknowledge the wrong tuple. Share rate limits by host and use deadlines only for retries/revalidation.
14. **Performance promises.** Warm embedded route timing says nothing about CLI startup, host IPC or provider latency. Measure each separately; add no service solely to improve an unmeasured number.
15. **Scope interpretation.** Complete coverage means every declared model/channel has evidence or a recorded outcome within the versioned inventory. It cannot guarantee scores exist for every model or that all of the web has been searched.
16. **Review availability.** Do not silently substitute self-review for mandatory independent/security review. A read-only fresh-context reviewer can run through an available CLI if no Agent tool exists; lack of tooling leaves an explicit gate, not invented approval.

## Verification

### Research tool and evidence

- Seed failing tests for the four inverse-status contradictions, stale EvalPlus classification and missing/stale review pins before fixing them.
- Rebuild without the existing DB, generated manifests, temporary scripts/logs or original mtimes. Compare every table/view, original decimal lexeme, source path and review scope against the frozen baseline plus reviewed corrections.
- Rebuild twice in separate directories and with shuffled enumeration/worker completion order. Require identical outputs under the pinned toolchain; no absolute paths or current timestamps in derived bytes.
- Mutate a reviewed raw file, findings file and report independently. Each must fail with a diagnostic naming the affected review and required new decision. Export must never repin automatically.
- Test malformed JSON, duplicate IDs/keys, unsupported format versions, path traversal/symlink escape, dangling source files, invalid decimals, empty arrays, unresolved identities, explicit zero, null cells and source-shape mismatch.
- Test publication interruption at every durable boundary, disk-full, lock contention and SIGKILL. The next check/rebuild must recover or reject a mixed generation and retain the prior usable export.
- Replay real saved parsers for escaped Gemma data, exact Meta/StepFun columns, GenEval table cells, EvalPlus nulls and SWE split/scaffold records. Negative fixtures use the actual 404/HTML-under-JSON failures.
- Run the real `research rebuild`, `check` and `coverage` commands, not only package tests. CI runs the same Make targets and checks that generated reports/outputs match inputs without silently changing the checkout.

### Contracts and production facts

- Machine-check API/event/table cross-references, enum vocabularies, sum-type completeness, documented CLI examples, relative links and version constants. Render all diagrams.
- Use real Postgres containers for numbered migrations, FKs, hash checks, deferred constraints, immutable records, temporal identity and rights concurrency. Tests run independently with isolated databases/ports and parallel execution where safe.
- Exercise confirmation expiry, late-arriving corrections, close/retract/supersede, same-revision corroboration, conflicting equal-time transitions, cycles and historical cutoff replay.
- Verify metadata canonicalization test vectors across all types, unknown-versus-unknown exclusion, period linkage, same protocol across different models, distinct pass@k/splits and no cohort from incomplete identity.
- Exhaust image role/operation matrices and price component/band boundaries. Unknown accounting, mask interpretation or applicable price components must fail strict admission.

### Collector and publication

- Use local HTTP fixtures for ETag/304, Last-Modified, redirect validation, both Retry-After formats, timeouts, 401/403, captcha, 429, 5xx, oversized/decompression bodies and schema changes. Use authorized real sources only for a separately gated live smoke test.
- Kill/restart before and after fetch persistence, assertion commit, output link, outbox dispatch, job completion, snapshot validation, publication and adoption. Replay duplicates and expired workers. Confirm one failure never aborts unrelated sources.
- Run an adversarial rights-revocation/publication race and persist which policy authorized each successful output. Verify unknown/expired/conflicting rights and caller-misclassified channels fail closed.
- Confirm source/parser/policy events cause work automatically, unchanged content avoids redundant extraction and startup reconciliation recovers missed notifications.

### Router and host

- Table/property tests for all hard constraints, exact costs, uncalibrated admission, deterministic tie order, allowed relaxations, no-calibration quality-floor rejection and runtime model removal.
- Fuzz requests/snapshots/metadata, run `go test -race`, exercise concurrent swaps/rollback and assert retained snapshots do not refresh old prices or identities.
- Prove no route-time network, database or inference calls. Test with those resources unavailable.
- Run the actual host's registry-to-selection-to-spawn path with authorized models and capture bounded outcomes. Test auth changes, context failures, rate limits, cancellation, cumulative spend and device reservations. Never report unsupported image execution as tested.
- Measure warm p50/p95/p99, allocations and memory at the stated catalog sizes, plus cold load, reload, CLI process/encoding overhead and provider end-to-end latency separately.
- Later calibration tests use held-out representative work, real validators, failures/cancellations, uncertainty checks and explicit spend/retention policy. No calibrated release from public benchmark percentages alone.

Every bug follows red test, fix and green real-path verification. Each coherent change is self-reviewed for duplicated logic, independently reviewed and security-reviewed where it crosses trust boundaries before its commit is accepted.

## Acceptance

1. Plan covers research recovery, missing sources/data, tooling, contract closure, production persistence/collection, offline routing, host integration and later calibration with explicit dependencies. No implementation occurred during this planning task.
2. A clean checkout can rebuild and check the research artifact using Go and durable versioned inputs only. Existing input bytes and numeric lexemes survive. The tool cannot approve changed evidence, weaken quarantine or depend on a previous output DB.
3. All current catalog entries and declared in-scope model/source channels have generated coverage records. Missing sources, blocked access, unknown rights, unlocated claims, unresolved prices, identities and critical configuration are named and independently actionable.
4. Saved usable evidence is extracted and reviewed without requiring current-offer identity; unusable files are classified from bytes, not extension or stale report prose. New evidence remains unreviewed until a new pinned decision exists.
5. Relevant API, event and DDL slices pass independent review before implementation. Snapshot encoding, rights pins, channel authority, output hashes, enums, capability matrices and cohort canonicalization are executable contracts, not prose placeholders.
6. Production authority uses real Postgres with numbered migrations. Research SQLite stays an export-only exception. Route reads an immutable in-memory snapshot and the host's live eligibility, with no route-time IO or hardcoded model list.
7. Collector operates automatically under a supervisor, survives crashes/transient outages, respects source rights/rate limits and publishes only independently validated successors. Unknown or restricted sources cannot leak into distributable artifacts.
8. Router enforces hard requirements, exact conservative budgets, freshness and explicit unknown-quality fallback. Strict quality floors remain unavailable without accepted calibration. Same request, effective time, runtime generation and snapshot produce the same decision.
9. Actual host spawning, snapshot recovery and measured latency have evidence. Unit tests and synthetic ranked lists alone do not satisfy delivery.
10. Calibration, paid evaluation, optional inference extraction and remote distribution stay disabled until their own data, contracts, spend and trust gates pass. Report these as blocked deliverables rather than silently omitting them or claiming the full system complete.
11. Makefile and CI enforce tests, source/review pins, generated-output freshness and documentation drift. Plans remain outside Git; unresolved work remains linked in beads; no push, public data release or PR is implied.
