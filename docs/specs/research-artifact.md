# Research artifact tool contract

Status: Proposed (phase 2 gate; implementation of the exporter is blocked until
this contract passes independent review).

Normative contract for the Go research tool that replaces the recovered Python
exporter (`data/research-inputs/recovery/build-db.py`). Scope is the local
research artifact only: `data/research.sqlite`, `data/source-manifest.tsv`,
`data/source-hashes.sha256` and derived coverage output. Nothing here authorizes
production collection, routing admission or network access.

## 1. Commands

Binary `agent-routing` (module path `github.com/guygrigsby/agent-routing`,
provisional until the repository remote exists). All subcommands are local
OS-user operations: no network, no provider calls, no production mutations, no
inferred approvals, no execution of fetched content.

| Command | Behavior |
|---|---|
| `agent-routing research check` | Validate durable inputs, review pins and the published artifact set without modifying anything, under a shared lock (§3). Reports each violation with file/row diagnostics. Reports an interrupted publication read-only as exit 3; recovery is rebuild's job — a deliberate deviation from the plan text ("check/rebuild first recover or roll back") so the CI gate stays side-effect-free. |
| `agent-routing research rebuild` | Read the same durable inputs and emit a complete replacement artifact set, or fail while preserving the last good set. Performs §3.1 recovery before staging. An identical successful rebuild (same inputs, same tool version) is a no-op: byte-identical outputs, exit 0. |
| `agent-routing research coverage` | Derive the coverage inventory (machine-readable TSV on stdout, human summary on stderr) from the same facts the export uses. Columns: `model_id`, `kind`, `batches` (`;`-joined sorted batch names with a `model_research` row), `benchmark_coverage` (the schema 3 coverage view value), `measurements`, `unresolved_measurements`, `located_claims`, `unlocated_claims` (counts). Rows sorted by `model_id`. Applies the same load-time validation as
`check` (path rules, inventory completeness, pin gates): on invalid or stale
inputs it exits per §1 without emitting the TSV. Never parses the published DB for facts the inputs already carry. Shared lock. |

Locking: one lock file `data/.research-rebuild.lock`. `rebuild` takes an
exclusive `flock` (bounded retry, then exit 4). `check` and `coverage` take a
shared `flock`, so they never observe a publication mid-rename or a journal
mid-write.

Common flags: `--root <dir>` (repository root; default cwd), `--json` (emit a
structured result object: input fingerprint, schema/artifact versions, driver
and Go toolchain versions, output hashes, diagnostics). stdout carries the
requested result; stderr carries safe diagnostics. Logs never contain
credentials or file contents, only paths, hashes and row identifiers.

Exit codes:

| Code | Meaning |
|---|---|
| 0 | Success or verified no-op |
| 1 | Invalid inputs: malformed JSON, undeclared findings shape, path-rule violation, input hash/size mismatch against the capture inventory, invalid or non-finite decimal, duplicate identifier, dangling reference, uninventoried read path |
| 2 | Stale review: a reviewed input's bytes differ from its pinned hash and no newer decision covers the new bytes |
| 3 | Inconsistent export: staged or published output fails schema, gate or cross-artifact validation; mixed generation; driver/toolchain mismatch; absent artifact set under `check` |
| 4 | IO failure (unreadable file, disk full, lock unavailable after timeout) |
| 64 | Usage error |

## 2. Durable inputs (versioned in Git)

Input files divide into evidence inputs and control files:

- Evidence inputs project into `source_files`/`source-manifest.tsv` and must
  each have a capture-inventory row (§2.1).
- Control files are `data/research-inputs/capture-inventory.tsv` itself,
  `data/research-inputs/build.json`, `data/research-inputs/fetch-logs/*.log`
  (durable redacted fetch-command records backing `recovery_basis`
  references; verbatim commands only, credential-scanned, no transcript
  prose), and `data/reviews/decisions/*.json`. They
  do not project into `source_files` (the frozen manifest's 340 rows are
  unchanged by their introduction) and are exempt from inventory membership.
  An inventory row naming a control file is exit 1.
  Their integrity is covered by `source-hashes.sha256` (§3) and by Git.
- Review reports (`data/reviews/*.md`) are evidence inputs, inventoried and
  projected, exactly as in the frozen manifest.

### 2.1 Capture inventory: `data/research-inputs/capture-inventory.tsv`

One row per evidence input file. TSV, header row, LF line endings, UTF-8,
sorted by path. Columns:

| Column | Rule |
|---|---|
| `path` | Repository-relative, `/`-separated. Must satisfy §6 path rules. |
| `origin` | `captured_remote` (bytes fetched from `url`), `local_derived` (produced locally: assignments, findings, reports), `header_snapshot` (response headers saved beside a body). |
| `url` | Recoverable remote URL for `captured_remote`/`header_snapshot`; empty for `local_derived`. Missing URLs on captured files are a provenance gap recorded in `recovery_basis`, never invented. |
| `sha256` / `bytes` | Expected content identity. Rebuild fails exit 1 on mismatch. |
| `observed_at` / `timestamp_basis` | Retained historical timestamp and its basis (e.g. `local_file_mtime_not_http_retrieval`). Never regenerated from current mtimes. |
| `license` / `publication_permission` | Per-file rights facts. `unknown` is valid and common; the tool never upgrades them. |
| `recovery_basis` | For provenance recovered after capture: short statement naming the evidence the URL/time was recovered from, or empty when the mapping is original and nothing extra corroborates it (a row whose url is inherited unchanged from the frozen manifest may still carry a basis, e.g. header corroboration for a capture with no retained fetch command). Fetch-log references use the exact grammar `fetch-log:<file> command <N> url+filename match` (`url+filename match` may read `filename match` when only the output filename ties them); every other basis is free text. Every `captured_remote` row with empty `url` must carry a non-empty `recovery_basis` (§5 gate 5). |

Every evidence file the exporter reads must appear in the inventory (exit 1
naming the unexpected path); every inventory row must exist on disk (exit 1).
Generated outputs (§3), `data/research-inputs/recovery/` reference copies
and `data/research-inputs/fetch-logs/` control records
must not appear.

### 2.2 Build config: `data/research-inputs/build.json`

Single JSON object, exact decimal literals preserved lexically where numeric:

- `artifact_version`, `schema_version` (currently `0.1.2` / `3`).
- `phase` (`research_only_not_routing_authority`), `publication_permission`.
- `baseline_url`, `pricing_crossref_url`, `collection_date_utc`.
- `scope`: rules determining which catalog kinds are in research scope
  (language output or image output) — the rule that yields the current 302.
- `expected`: fixture expectations compared against derived counts
  (`catalog_models` 395, `exact_catalog_joins` 394, `model_research` distinct
  models 302, `price_facts` 31659, `pricing_disagreements` 12). Mismatch is exit 3. These are expectations
  about the frozen inputs, not limits on future inputs; changing inputs updates
  them in the same reviewed change. `expected` flattens to the three
  `expected_*` metadata keys exactly as enumerated in §7.2.
- `publication_permission_vocabulary`: `unknown|allowed|denied`, validated at
  input load (exit 1), not mid-insert.
- `driver`: pinned SQLite driver module path and version, plus pinned Go
  toolchain minor version. At startup the tool compares the linked driver
  (from build info) and toolchain against these pins; mismatch is exit 3. Byte
  reproducibility claims hold for the pinned driver + toolchain on the same
  architecture; architecture is recorded in `--json` output but not pinned
  (SQLite bytes are architecture-independent for the pinned driver).
- `batch_formats`: explicit map of batch directory name under `data/research/`
  to adapter (§2.4). A directory under `data/research/` without a mapping is
  exit 1.

### 2.3 Review decisions: `data/reviews/decisions/*.json`

Append-only decision records; the exporter contains no approved-batch list.
One JSON object per file:

- `batch` (must name an existing `data/research/` directory with a
  `batch_formats` mapping; a phantom key or unmapped name is exit 1), `scope`,
  `state`, `reason` (required text; projects to `review_results.reason` for
  acceptance decisions). Scope binds to adapter at load (exit 1 on mismatch):
  `published_measurements` only for batches whose findings carry
  measurement rows, `catalog_sources` only for a batch using the
  `catalog_sources` adapter (the catalog projection batch), not the broader
  claims-only class.
- `review_path`: repository-relative path to the review report (the existing
  `data/reviews/*.md` files).
- `pins`: array of `{path, sha256}`. For `published_measurements` the pin set
  is exactly: the review report (`review_path`), the batch findings file
  (`data/research/<batch>/findings.json`), and the `source_path` of every
  benchmark and unresolved measurement row the decision covers; claim
  evidence files are not pinned (matching the frozen `evidence_review_inputs`
  set and the make-check needed-set). For `catalog_sources` the pin set
  additionally includes every capture file the findings reference in their
  `evidence`/`headers` blocks (`data/raw/vercel-models.json`,
  `data/raw/vercel.headers`, `data/raw/models-dev.json`,
  `data/raw/models-dev.headers` in the frozen corpus) — the acceptance asserts
  a comparison over exactly those bytes, so changing any of them makes the
  decision stale (exit 2) until a new decision exists; the rebuild never
  repins. Pins naming control files, nonexistent paths, or anything under
  `data/research-staging/` are exit 1.
- `decision_version` (monotone integer per batch+scope), `decided_at` (UTC,
  from the review event, not rebuild time).

Supersession is a total order: by `(decided_at, decision_version, file path)`;
the maximum record per `(batch, scope)` is current. Older records stay in the
ledger. Decision records are validated on load: unknown scope, a scope/state
binding outside the table below, a pin naming a control file or nonexistent
evidence path, or a `batch` without an adapter mapping is exit 1 naming the
decision file and field. The tool cannot detect a forged or edited decision
record; ledger integrity is anchored in Git history and human review. That is
the stated trust root of this contract.

Scope semantics (closed set; new scopes require a spec change):

| Scope | Authorizes | Projects into |
|---|---|---|
| `published_measurements` (state `source_validated`) | Source-accuracy validation of the batch's measurement rows | `evidence_reviews` + `evidence_review_inputs` (schema 3 admits exactly this scope) |
| `catalog_sources` (state `accepted`) | Acceptance of the catalog capture batch | `review_results.grade='accepted'` with the decision's recorded reason, for the named batch only |
| `revoked` (either scope) | Withdrawal of a prior decision without deleting ledger history | Reverts the batch's projection to the frozen default quarantined reason below — the revoking record's `reason` is ledger documentation only and never exports; a revoked `published_measurements` batch still hits the §2.3.1 hard gate, forcing re-review or demotion to the staging lane rather than a quarantined measurement export |

Batches with no current decision still get a `review_results` row (the
`unresolved_measurements.batch` foreign key requires it): `grade='quarantined'`
with the frozen reason string `Provisional research evidence. Composite review
failed or batch not separately approved; no routing admission.` unless a
decision record supplies another reason. `evidence_reviews` rows exist only
for a current `published_measurements` decision.

### 2.3.1 Canonical vs unreviewed staging

`data/research/` is the canonical lane: a batch whose findings carry
`benchmarks` or `measurements` rows must have a current
`published_measurements` decision covering them, or rebuild fails exit 3
naming the batch and instructing the operator to either record the review
decision or move the batch to the staging lane. New raw captures and
extractions land in `data/research-staging/<batch>/` with the same findings
shape; the exporter does not read that directory, and no schema 3 table
projects from it — that is the entire quarantine boundary. Promotion is a
reviewed commit that moves the batch under `data/research/`, adds the
`batch_formats` mapping and appends the decision record, all together. The
canonical export's gates therefore never depend on unreviewed evidence.

Enforcement is structural, not declarative: a capture-inventory row, an
adapter-resolved `source_file`, or any other evidence path under
`data/research-staging/` is exit 1. Recorded deviation from the plan's
"new evidence stays unreviewed until a decision exists": a promoted batch
whose findings carry only `models[]`/claims (no measurement rows) exports
quarantined `model_research`, claims and gaps without a decision, matching
the frozen corpus (claims from unreviewed batches were always exported
quarantined); only measurement rows are decision-gated, because only they
feed the source-validated equality views.

### 2.4 Findings adapters

Batch findings shapes are heterogeneous; each is read by a typed adapter
declared in `build.json.batch_formats` for directories under `data/research/`:

| Adapter | Directories (current) | Shape summary |
|---|---|---|
| `model_batch` | anthropic, deepseek-zai, google, images, mistral-meta, mixedbread, moonshot-minimax, openai, other-text, qwen, xai | `models[]` (id, benchmark_status, notes, strengths[], weaknesses[]), `benchmarks[]`, `sources[]`, `gaps[]` |
| `unresolved_only` | unresolved | `measurements[]`, `gaps[]` |
| `coverage_followup` | coverage-followup | `measurements[]`, `sources[]`, `gaps[]` |
| `benchmark_source_inventory` | benchmark-sources | `assigned_models`, `baseline`, `benchmark_rows`, `note`, `stage`; emits no measurements; rows are source-outcome facts only |
| `catalog_sources` | catalog-sources | Top-level keys exactly: `research_title`, `collected`, `evidence`, `headers`, `counts`, `units`, `pricing_comparison`, `vercel_priced_but_models_dev_cost_missing`, `image_pricing`, `vercel_pricing_keys`, `licensing`, `notes`. Feeds only its `review_results` row and evidence pins; metadata comes from `build.json` and catalog tables from the `data/raw/` captures |

Adapters are closed over keys: an unknown top-level key or item field is exit 1
with the path and key named (no silent drops). The normative item-field sets
(frozen corpus):

| Adapter | List | Item fields |
|---|---|---|
| `model_batch` | `models` | `id`, `strengths`, `weaknesses`, `benchmark_status`, `notes` |
| `model_batch` | `benchmarks` | `model_id`, `benchmark`, `benchmark_version`, `metric`, `value`, `unit`, `configuration`, `source_url`, `source_file`, `locator`, `evidence_class`, `identity_basis` |
| `model_batch` / `coverage_followup` | `sources` | `url`, `file`, `status`, `purpose`, `license` (`coverage_followup` adds `sha256`, `commit`) |
| `unresolved_only` / `coverage_followup` | `measurements` | `reported_model`, `benchmark`, `benchmark_version`, `metric`, `value`, `unit`, `configuration`, `source_url`, `source_file`, `locator`, `evidence_class`, `identity_status`, `proposed_offer_ids` (projects `unresolved_identity_proposals`) |
| `benchmark_source_inventory` | `benchmark_rows` | as defined by the phase 5 inventory findings contract; absent until then |
| all adapters | `gaps` | plain strings (not objects) |
| all adapters | claim items (`models[].strengths[]`/`weaknesses[]`) | `text`, `basis`, `source_url`, `source_file`, `locator` |

`benchmark-sources`, `catalog-sources` and `coverage-followup` also carry
descriptive top-level keys: `benchmark-sources` has `assigned_models`,
`baseline`, `benchmark_rows`, `note` (singular) and `stage`;
`catalog-sources` has the 12 enumerated keys (no `gaps`); `coverage-followup`
has `gaps`, which projects to `research_gaps` like any adapter's `gaps` list.
`gaps` is a valid key only for adapters whose enumerated top-level key set
includes it — a `gaps` list on `catalog-sources` or `benchmark-sources` is
exit 1. Nested object keys inside the
`catalog-sources` descriptive block (`evidence`, `headers`, `counts`,
`units`, `pricing_comparison`, `image_pricing`, `licensing`) are parsed as
data, not closed over: their exact shape is pinned by the frozen findings
round-trip oracle, and any mutation is caught by the findings.json decision
pin. Byte-affecting defaults are
normative, not implementation detail: `models[].notes` defaults to `''`;
`sources[]` rows project to `source_access` with defaults `file=''`,
`purpose=''`, `license='unknown'`, deduplicated `INSERT OR REPLACE` on
`(batch, url, path)` with the last occurrence in list order winning; `gaps[]`
ordinals are the 0-based list position; a duplicate value inside one
`modalities` direction list is exit 1;
`measurements[].identity_status` is `unresolved`. The round-trip oracle is the
frozen DB plus the §9 fixtures.

Source reference normalization is defined once and shared by every adapter that
resolves `source_file`/`source_url` pairs (`model_batch` claims and
benchmarks, `unresolved_only` and `coverage_followup` measurements):

1. Resolution precedence matches the frozen exporter: a `source_file` that
   exists as a repository-relative path resolves as-is; otherwise it resolves
   against the batch directory; otherwise the reference is unresolvable.
2. `basis='catalog_claim'` claims resolve to the baseline catalog capture
   (`data/raw/vercel-models.json`) and its URL, and the locator is rewritten
   to `data entry id=<model_id>; <original locator>`.
3. An empty/missing `source_file` claim exports to `unlocated_claims` with the
   frozen reason `No saved evidence file supplied; quarantined` (no trailing
   period — byte-exact frozen literal). An
   unresolvable non-empty `source_file` claim exports to `unlocated_claims`
   with the same reason. A measurement with an unresolvable `source_file` or
   empty `source_url` is exit 1 (the schema requires both).
4. A claim whose file resolves but whose `source_url` is empty keeps the empty
   URL: no synthesized fallback URLs, ever. Such a claim exports to
   `unlocated_claims` (reason `no_source_url`). An unlocated claim otherwise
   keeps whatever non-empty URL the findings carry (the URL is provenance
   about the source, not about the bytes); the frozen corpus's 48
   `unlocated_claims` rows carry the synthesized `'unknown'` and export with
   `source_url=''` per the documented delta. Phase 3
   repairs the affected findings files through review so the exported rows
   carry true recovered URLs; the delta against the frozen DB is exactly:
   8 located claims move to `unlocated_claims` (`no_source_url`), and 48
   `unlocated_claims` rows previously carrying the synthesized `'unknown'`
   URL export with empty URL — 56 changed rows (81 affected input rows; the
   remaining 25 are `catalog_claim`s whose exported values
   never depended on the findings URL). Phase 3 also repairs, through the same
reviewed findings corrections: the four inverse-status `model_research` rows
(gate 2 fires on `minimax/minimax-m2`, `moonshotai/kimi-k2`, `kimi-k2.5`,
`kimi-k3`), the one false `not_found` `source_access` row
(`data/research/benchmark-sources/_evalplus_results.json`: parseable content
JSON with 125 reported model labels, repaired to `retrieved`), the
`sources[].url` byte-provenance repair (68 rows across nine batches, each set
to the capture-inventory URL of the same file; pointer text, `unknown` and
page-vs-raw disagreements all resolve to the URL the retained fetch command
actually fetched), and any note/wording corrections the review records; those exported
row changes belong to this delta list with counts pinned at the phase 3
review. Four `not_found` rows over API-error-envelope JSON captures
(GitHub/Hugging Face error bodies) keep their status as honest outcomes
under the §8 gate-3 error-envelope exception; they are not repairs. The
phase 3 review report itself (`data/reviews/phase3-provenance-2026-09-30.md`)
joins the evidence inputs as manifest row 341. `source-hashes.sha256` also differs deliberately: the frozen file
appends `data/research.sqlite` and `data/source-manifest.tsv` after the 340
path-sorted evidence rows, while the rebuilt manifest is fully path-sorted
and adds control-file lines — a documented artifact-level delta, not a row
divergence. The delta is a documented intentional
   correction, listed in the phase 3 review.

## 3. Outputs and publication

Outputs: `data/research.sqlite`, `data/source-manifest.tsv`,
`data/source-hashes.sha256`. Treat all three as one generation.

Publication protocol (rebuild):

1. Acquire the exclusive lock (§1).
2. Recover first (§3.1) if a journal exists.
3. Back up: copy (not move) the current generation, if any, into
   `data/.research-backup/` and write `data/.research-backup/manifest.json`
   listing each file's old sha256; fsync. First-generation rebuilds record an
   empty backup manifest.
4. Stage the complete generation into `data/.research-staging/` on the same
   filesystem: build SQLite from scratch, generate the TSV, close the DB,
   then hash the completed DB to generate the hash manifest, fsync every
   staged file.
5. Validate the staged set: schema version, all §5 gates, foreign keys,
   integrity, cross-artifact hash agreement, repeat-build determinism of the
   staged DB bytes (second in-memory build; implemented as a full second
   staged build byte-compared against the first, a strictly stronger check).
6. Journal the publication: write `data/.research-rebuild.journal` containing,
   for each of the three outputs, the staged path, destination path, new
   sha256, and the pre-publication old sha256 from the backup manifest (empty
   for first generation). Fsync the journal.
7. `rename()` each staged file into place, in TSV, hash manifest, DB order
   (cheapest-to-regenerate first: the TSV derives from inventory strings
   alone, the hash manifest additionally hashes every input, the DB requires
   the full build).
8. Validate in place: hash the three published files; all must match the
   journal's new hashes; a mismatch exits 3 leaving the journal in place for
   §3.1 recovery (the run never rolls back renames — a successor, once
   published, is not unlinked). Clear the journal by `unlink()` (never
   truncate),
   prune the backup to exactly one prior generation, fsync the directory,
   release the lock.

The current database is never unlinked before a successor exists. Three renames
are not one transaction: the journal (old and new hashes) plus
validation-on-recovery is the consistency mechanism. `check` under its shared
lock never sees a half-published set. External SQLite inspection is documented
as supported only after a completed build.

No-op rebuild: if every staged output hashes equal to the published set,
`rebuild` prunes the staging area and reports ok without writing the journal
or renaming anything (step 3's backup still runs and is idempotent; it holds
the still-current generation). No-op detection compares all three outputs —
partial identity is not a no-op.

Hash acyclicity: `source-manifest.tsv` records input identities from the
capture inventory (not its own hash, not a DB hash). `source-hashes.sha256`
covers every evidence input file, every control file, the completed DB and the
generated TSV, and excludes itself, the journal, lock, staging and backup
areas.

Artifact byte formats are normative. `source-manifest.tsv`: one header row
then one row per capture-inventory entry, tab-separated, 8 columns in order
`path`, `url`, `sha256`, `bytes`, `observed_at`, `timestamp_basis`, `license`,
`publication_permission`; sorted by `path` (UTF-8 byte order); LF line
endings, trailing newline. `source-hashes.sha256`: one line per hashed file,
sha256sum-compatible `<hex sha256><space><space><path>`, path-sorted
(UTF-8 byte order), LF endings, trailing newline, no header.

### 3.1 Crash recovery (rebuild entry)

If the journal exists, process entries in order; the journal is complete
because it is written and fsynced before any rename:

- Destination missing, staged source present with the journaled new hash:
  re-issue the rename (idempotent completion).
- Destination present with the journaled new hash: entry complete.
- Any other state — destination present with a different hash, source missing
  and destination missing, hash mismatch — is a corrupt generation: roll back
  by restoring every output named in the backup manifest from
  `data/.research-backup/` (verify backup bytes against the manifest's old
  hashes first), delete leftover staged files, clear the journal. With an empty
  backup manifest (first generation), rollback deletes the partial outputs.
  If backup bytes fail verification, exit 3 naming the unrecoverable file;
  never publish on top of an unresolved mixed generation.

A crash after step 8's in-place validation but before journal clear is
recovered as "all entries complete" and the journal is cleared.

A torn or unparseable journal can only exist if the crash happened during
step 6, before the journal fsync and therefore before any rename: delete the
journal, delete the staging directory, and treat the run as never started
(after verifying the published set still hashes against the backup manifest;
mismatch is exit 3). When rollback itself is unrecoverable (backup bytes fail
verification), the operator procedure is: preserve `data/.research-backup/`
for forensics, delete the three outputs plus journal and staging, and rerun
`rebuild`, which proceeds as a first-generation build.

## 4. Determinism

Identical inputs and pinned toolchain produce byte-identical outputs. Banned
from all derived bytes: current time, current mtime, absolute paths, user name,
hostname, map iteration order, directory enumeration order. Tables are
inserted in schema order; within a table, rows are sorted by the tuple of all
non-generated columns — TEXT columns compare as UTF-8 bytes, INTEGER columns
compare numerically — with column order = schema order, and generated integer
IDs are assigned after sorting. Text is
written as-is (no NFC normalization; captured bytes and lexemes are sacred).
SQLite pragmas that affect bytes (`page_size`, `journal_mode=OFF`,
`encoding='UTF-8'`) are pinned in code. Two clean-directory builds must
compare equal; canonical relational content (ordered dumps) is a diagnostic if
the pinned driver version changes, never a silent waiver.

## 5. Gates

Hash-basis precedence, stated once: pins compare against the capture
inventory, and disk files compare against the capture inventory at load —
load-time disk-vs-inventory runs first (exit 1), then decision-pin-vs-inventory
(gate 4, exit 2). A tampered evidence file with an honest inventory is exit 1;
a pin disagreeing with the inventory is exit 2.

Gates 4 and 5 run before staging (fail fast). Gates 1-3 run on the staged
output and on the published set under `check`:

1. Existing SQL gates: integrity, foreign keys, expected counts, scope
   coverage, found-status-has-measurement, no-note/measurement contradiction,
   all measurements in source-validated views, every review input pinned.
   `accepted_benchmarks` is unconditionally empty; admitting benchmark
   acceptance requires a new decision scope plus an explicit gate change in
   this spec, never a silent waiver.
2. Inverse gate (exit 3): a model with measurements for batch B must not have
   `benchmark_status='unresearched'` for B (the four MiniMax/Kimi defects).
3. Access-status gate (exit 3): a `source_access` row whose `status` is in
   the failure vocabulary — `blocked`, `blocked_403`, `blocked_404`,
   `blocked_non_score`, `blocked_redirect`, `not_found`,
   `redirect_or_not_found` — for path P contradicts bytes
   that parse as the success format implied by the file extension: `.json`
   must parse as a JSON object or array that is not an API error envelope;
   `.csv` must have a multi-column
   header plus at least one data row. An API error envelope is a JSON object
   whose key set is a subset of `{message, documentation_url, status, error}`
   carrying an HTTP error `status` value, a non-empty `error` string, or a
   `message` naming a known API authentication failure (`Invalid username or
   password.`) —
   genuine API failure responses (GitHub `404` bodies, the Hugging Face
   `"Invalid username or password."` body) are honest failure outcomes and
   are not flagged. `retrieved` and `deferred` are not
   failure statuses (`deferred` defers retrieval, asserting nothing about the
   bytes). Literal failure bodies (`404: Not Found`
   text under a `.json` name) fail these parses and are not flagged. Bound:
   only `.json`/`.csv` captures are checked; a false-failure declaration over
   an `.html`/`.md` capture is undetectable by this gate and stays a review
   concern. Findings repairs are made in findings files through review, never
   by the exporter.
4. Review pin gate (exit 2, pre-staging): every current decision's pins match
   the capture inventory's hashes; mismatch names the decision file and
   changed path. A pin naming an existing path absent from the inventory is
   exit 1 (uninventoried read path).
5. Input completeness (exit 1, pre-staging): every read evidence path is
   inventoried; every inventoried path exists; every `captured_remote` or
   `header_snapshot` row has
   a URL or a non-empty `recovery_basis`.

## 6. Path and resource safety

All input paths: repository-relative only; reject absolute paths, `..`
segments, symlinks resolving outside the root, and any path not matching
`^data/[A-Za-z0-9._/-]+$` (control files additionally
`^data/(research-inputs/(capture-inventory\.tsv|build\.json|fetch-logs/[A-Za-z0-9._-]+\.log)|reviews/decisions/[A-Za-z0-9._-]+\.json)$`).
Batch directories live under `data/research/`. File size bounds: per-file
64 MiB, total 2 GiB (current set ~40 MB); JSON nesting depth bound 64.
Bound violations are exit 1 naming the path and bound. Fetched
content is parsed as data only; the tool executes nothing found in inputs.
Failure names path and row, never file contents.

## 7. Schema and version policy

Research schema lives in the Go research package as versioned DDL; it is not a
production migration. Schema 3 is reconstructed exactly (all tables, views,
indexes, checks from the recovered exporter) except the documented phase 3
intentional corrections: the §2.4 URL rule (81 claim rows) and `source_files`
URL recovery per §2.1 (256 frozen `url` values change in total: 227
`unknown`/log-pointer values become recovered URLs or empty-with-`recovery_basis` — 159 recovered, 32 `local_derived` files to empty, 36 pointer text — plus 29 page-vs-raw corrections where both values are real URLs and the retained fetch command backs the new one, e.g. repo page → raw README, leaderboard page → `api.github.com` contents, and one hostname case fix). Two further delta classes exist because the frozen
database predates the phase 3 findings repairs (frozen bytes are history, not
the oracle, for these rows): `model_research` — 4 rows (`minimax/minimax-m2`,
`moonshotai/kimi-k2`, `moonshotai/kimi-k2.5`, `moonshotai/kimi-k3`) whose
frozen `benchmark_status='unresearched'` contradicted their recorded
measurements and whose repaired findings carry `found` — and `source_access`
— 69 rows: 68 URL changes (43 frozen log-pointer placeholders (e.g. `(see
anthropic.log for recovered URL mapping)`), `saved source URL recovered
from …` and `unknown` values becoming the recovered URLs, plus 25
http-to-http corrections applying the fetch-URL-over-page-URL rule) and one
EvalPlus `_evalplus_results.json` row whose URL is unchanged while `status`
moves `not_found`→`retrieved` and `purpose` is rewritten (counts from a
primary-key `(batch, url, path)` diff of frozen vs rebuilt; a plain join on
`(batch, path)` overcounts because groups holding both page and fetch URLs
multiply their unchanged siblings).
Disk findings are authoritative over frozen rows for both classes. Further
delta classes follow from normative rules rather than repairs: the frozen
`claims`/`unlocated_claims` row split shifts by 8 rows
(427/48 → 419/56) under the §2.4 URL rule; generated integer ids in
`claims` (419 rows), `benchmark_measurements` (112) and
`unresolved_measurements` (252) are reassigned by §4's sorted-insert rule,
so they need not equal the frozen document-order ids (non-generated tuples
are identical except for the documented row classes). `review_results.reason`
for an accepted batch derives from the current decision record
(`catalog-sources` differs from the frozen hardcoded text; quarantined
batches keep the frozen default reason). The v3 decision records (superseding
the v2 records, which still pinned the per-batch review documents after
`review_path` moved) repoint the five phase-3-corrected batches:
`evidence_reviews.review_path` for those 5
batches moves from the frozen per-batch review documents to
`data/reviews/phase3-provenance-2026-09-30.md`, and their `evidence_review_inputs`
pins are repointed to the corrected findings files (10 input rows differ:
five findings.json pins move to the corrected hashes, five old review-doc
pins are replaced by the provenance-document pin).
The nine findings files corrected by the phase 3 review also change their `source_files` `sha256`, `bytes` and `observed_at` rows (observed_at = the correction event). Future changes append a
`schema_version` bump plus a documented delta in this spec's changelog; the
tool refuses to rebuild when `build.json.schema_version` differs from the DDL
it carries (exit 3). No production Postgres DDL may be derived from this
section. Model association in `unresolved_identity_proposals` remains
provisional identity, never proven revision identity.

### 7.1 Catalog projection rules (normative for byte-identical output)

From `data/raw/vercel-models.json`, for each entry in `data[]`:

- Numeric leaf walk over the entry's `pricing` object only (never the whole
  model object): dict/list recursion building `field_path` rooted at the
  walked object with a leading `/` and `/`-joined segments, list indices as
  segments. A leaf is numeric when its textual form parses as an exact
  decimal and it is not a JSON boolean; JSON string leaves whose decoded
  content parses as an exact decimal are numeric leaves (2,765 of the frozen
  rows) and `raw_decimal` is the decoded string content. An absent `pricing`
  object, absent `cost` object or absent provider `models` map is an empty
  walk (421 of 8,323 models.dev entries carry no `cost`), never an error.
  Verbatim means the JSON token as
  written — no float canonicalization; `2.50` stays `2.50`, and
  `0.024999999999999998` stays unrounded.
- Each numeric leaf emits one `price_facts` row: `source_path=
  data/raw/vercel-models.json`, `provider='vercel'`, `upstream_id=<model id>`,
  `field_path`, `raw_decimal` (verbatim lexeme), `unit`. Unit map by final
  path segment: `input`, `output`, `input_cache_read`, `input_cache_write`,
  `audio_input_token_cost`, `audio_output_token_cost` → `USD_per_token`;
  `image` → `USD_per_image`; anything else → `source_specific_unresolved`.
- `catalog_models` row per entry (`id`, `name`, `owned_by`, `type`,
  `description`, source `data/raw/vercel-models.json`);
  `catalog_modalities` from `modalities` (direction keys input/output);
  `catalog_limits` only for `context_window` and `max_tokens` when present;
  `catalog_parameters` from `supported_parameters` (duplicates ignored);
  `catalog_tags` from `tags` (duplicates ignored).

From `data/raw/models-dev.json`, for each `<provider>.models.<uid>`:
`provider_offers` row (`provider`, `uid`, `name` defaulting to `uid`, source
`data/raw/models-dev.json`); `canonical_labels` from `canonical_model_id`
when present; numeric leaf walk over the entry's `cost` object only, emitting
`price_facts` rows with `source_path=data/raw/models-dev.json`, the walk
root's `provider`/`uid`, `field_path` rooted at `cost` with leading `/`, and
unit map `input`, `output`, `cache_read`, `cache_write` →
`USD_per_million_tokens`, else `source_specific_unresolved`.

`exact_catalog_joins`: one row per catalog id present in
`models.dev['vercel'].models`. `pricing_disagreements`: for each join and
each of the four pairs (Vercel name → models.dev name: `input`→`input`,
`output`→`output`, `input_cache_read`→`cache_read`,
`input_cache_write`→`cache_write`), when both sides carry the component,
compare Vercel decimal × 1000000 against the models.dev decimal with
exact-decimal arithmetic (scaled integers or `math/big`; never binary
floats); insert on inequality with `component` = the Vercel-side name, the
models.dev side stored exactly as written, and the product rendered as a
plain decimal preserving the Vercel lexeme's fraction-digit count (trailing
zeros kept: `0.000000025` × 1000000 renders `0.025000000`, never `0.025`).

### 7.2 Metadata rows

The `metadata` table has exactly these rows, sourced from `build.json`:
`artifact_version`, `schema_version`, `phase`, `baseline_url`,
`pricing_crossref_url`, `collection_date_utc`, `publication_permission`,
`expected_catalog_models`, `expected_exact_joins`,
`expected_text_and_image_research_models`. No other keys; no tool-generated
values.

## 8. Events

None. Pure local transformation over versioned inputs. Automatic invocation on
input change belongs to the future collection coordinator (production slices)
and must call this same CLI; this contract defines no daemon, watcher or
scheduler.

## 9. Verification

Implemented in phase 4: corpus and fixture suites in `internal/research`
(`corpus_test.go`, `gates_test.go`, `recovery_test.go`, adapter loader tests),
`make check` consuming `agent-routing research check`, and real CLI runs of
all three subcommands. Crash injection at journal boundaries is realized as
construction of every §3.1 on-disk state (equivalent to and more
deterministic than SIGKILL mid-publication): complete, rename-needed,
corrupt-with-rollback, corrupt-first-generation, torn-journal each with their
consistency verdicts.

## Changelog

- 2026-09-30: proposed (phase 2). Replaces the recovered Python exporter's
  implicit behavior; no prior version of this contract existed. Independent
  review round 1 findings addressed: control-file coverage and self-reference
  (evidence/control split), journal old+new hashes with backup manifest and
  staged hash manifest inside the journaled set, mixedbread adapter, decision
  scope table and total supersession order, reader shared-lock, deliberate
  read-only `check` deviation, executable access-status predicate, total
  sort-key rule, no-synthesized-URL rule with phase 3 repair, runtime
  driver/toolchain pin enforcement, path regex fix, gate exit-code mapping,
  first-generation and absent-output edges, decision-ledger trust-root
  statement.
- 2026-09-30 round 2: added §2.3.1 canonical-vs-unreviewed staging lane
  (data/research-staging/ excluded from export; measurement-bearing canonical
  batches require a current decision); default quarantined review_results rule;
  empty/unresolvable claim source rules with frozen reasons; root-first
  resolution precedence; close-DB-then-hash ordering; §7.1 normative catalog
  projection rules; catalog_sources shape enumeration and projection fix;
  source_files.url recovery added to documented deltas; torn-journal rule and
  unrecoverable-rollback operator procedure; gate 3 extension bound; decision
  validation exit codes.
- 2026-09-30 round 3: §7.1 price_facts walk scoped to `pricing`/`cost` objects
  with exact column population, leading-slash field paths, verbatim-lexeme
  definition, Vercel-side `component` names; catalog_claim locator rewrite;
  decision `reason` field; §7.2 exact metadata row set; structural staging-lane
  enforcement; exact pin set; unconditional `accepted_benchmarks` emptiness;
  normative adapter defaults; recorded claims-only promotion deviation;
  phantom batch key rejection; permission vocabulary load-time validation;
  coverage TSV column contract.
- 2026-09-30 round 4: §7.1 disagreement product rendering pinned (preserve
  Vercel fraction-digit count) and `pricing_disagreements` 12 added to
  expected; normative TSV/hash-manifest byte formats (§3); string-leaf numeric
  rule; per-adapter item-field vocabulary table; §7 delta corrected to 56
  changed rows with `unlocated_claims` URL rule; journal clear is `unlink()`;
  §4 numeric ordering for INTEGER columns; scope↔adapter load-time binding.
- 2026-09-30 round 5: `catalog_sources` decisions pin the capture files named
  by the findings evidence/headers blocks (mutation makes the decision stale,
  never repinned) and the scope binds to the `catalog_sources` adapter; gate 3
  failure-status vocabulary enumerated; claim item fields added to the
  vocabulary table; absent walk objects are empty walks; `source_access`
  dedup winner and duplicate-modality rule stated; resource bounds map to
  exit 1; `revoked` decision state added; unlocated-claim URL retention rule
  and 56/81 delta counts fixed.
- 2026-09-30 round 6: unlocated-claim frozen reason literal corrected (no
  trailing period); revocation restores the frozen default reason and never
  exports the revoking reason; gate 5 covers `header_snapshot` rows;
  hash-basis precedence stated (disk-vs-inventory exit 1 before
  pin-vs-inventory exit 2); phase-3 delta list extended with gate-2/gate-3
  repair classes; catalog-sources nested descriptive keys explicitly
  oracle-pinned rather than closed over.
- 2026-09-30 round 7 (PASS + fixes applied): per-batch descriptive key
  attribution corrected and `gaps` restricted to enumerated adapters;
  `coverage` runs the same load-time validation as `check`; inventory rows
  naming control files are exit 1; uninventoried pin paths are exit 1; step-8
  validation failure exits 3 leaving the journal; source-hashes ordering
  delta documented.
- 2026-09-30 phase 3 inputs: `data/research-inputs/fetch-logs/*.log` added as
  control records (durable redacted fetch-command provenance backing
  `recovery_basis` references; §2 control list, §2.1 exclusion and the §6
  control-path regex extended). Gate 3 gains the API-error-envelope
  exception (error-envelope JSON bodies are honest failure outcomes); the
  phase-3 gate-3 repair count is pinned at one row (EvalPlus results.json)
  and the 68-row `sources[].url` byte-provenance repair is added to the
  delta list. `build.json` seeds `driver` pins as placeholders that phase 4
  replaces with the linked module/toolchain versions before the first Go
  build; a placeholder value at load time is not a mismatch (there is
  nothing to compare against yet).
- 2026-09-30 phase 3 review (PASS + fixes applied): EvalPlus `purpose`
  corrected alongside the status repair (stale 404 wording over the JSON
  scoreboard); §7 URL-delta count corrected to 256 (227 unknown/pointer + 29
  page-vs-raw) and the nine findings sha/bytes/observed_at deltas listed;
  `recovery_basis` semantics and fetch-log reference grammar pinned; manifest
  row 341 recorded.
- 2026-09-30 phase 4 implementation: driver pins filled and enforced
  (`modernc.org/sqlite` v1.60.1, `go1.26`; placeholder values never mismatch);
  step-5 determinism realized as a full second staged build byte-compare;
  no-op rebuild semantics recorded (§3); §7 gains the two post-freeze delta
  classes (4 `model_research` rows, 69 `source_access` rows: 68 URL changes
  plus one status/purpose-only EvalPlus row); hash
  manifest generation is 386 lines (341 evidence + 24 control + 19 decision
  records + DB + TSV) and grows as the decision ledger appends; §9 records
  the implemented verification posture (journal-boundary state construction);
  `make check` now runs `agent-routing research check`, retiring the
  shell-ported SQL gates (queries preserved verbatim in the tool).
- 2026-09-30 phase 4 review (fixes applied): gate 5 closed for batch
  findings reads (uninventoried `data/research/<batch>/findings.json` is now
  exit 1; previously the only evidence-read class with no inventory
  backstop, live-reproduced); §7 source_access delta corrected to verified
  counts (69 changed rows: 68 URL changes = 43 placeholder recoveries + 25
  fetch-URL corrections, plus 1 status/purpose-only EvalPlus row; derived
  from a primary-key `(batch, url, path)` diff — the interim 140/97 figures
  came from a `(batch, path)` join that multiplies unchanged sibling rows in
  multi-URL groups);
  §7 enumerates the structural delta classes (generated-id reassignment,
  accepted-batch review reason derivation, the 8-row claims/unlocated split);
  §3 step 7 rename order corrected to TSV, hash manifest, DB
  (cheapest-to-regenerate first, matching the parenthetical rationale);
  gate 3 envelope definition gains the authentication-failure `message` case
  the code implements; the phase-3 seed literal is a recognized placeholder;
  control directories reject non-conforming names and subdirectories
  symmetrically; the 2 GiB total input bound is enforced in inventory
  verification; tests cover the findings-inventory gate, the
  measurement-without-decision hard gate, backup-verification failure and
  lock-timeout exit 4.
