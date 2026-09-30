# Independent contract review: current dirty-tree documents

## Scope and gate assessment

This is a design-contract review of the current `docs/specs/*.md` and `docs/adr/*.md`, limited to the first offline library/collector slices. It is not broad product approval. Calibration is treated as explicitly disabled: no empirical calibration dataset is required merely to plan. The later HostReceipt outcome/calibration tables are also treated as a known future gate, not as a passed contract and not as a blocker for unrelated raw collection.

The stated “API + event + DDL all complete before implementing the corresponding slice” gate has **not passed** for either the collector core or an end-to-end routing library slice. The problems below are concrete contradictions or missing executable contracts in the current documents.

## Resolved / internally consistent areas

- The current language direction is Go, with an embedded library as the primary boundary and only a proposed thin CLI. ADR 0002, `router.md`, and `contracts.md` agree on this direction.
- Collection authority (future Postgres), temporary research export (SQLite), immutable offline snapshots, and an in-memory route path are separated.
- Revision, provider offer, and mutable alias/binding are distinct. Identity uses append-only bindings, confirmations, ends, and supersessions; measurements do not follow the current alias.
- Raw benchmark capture is separated from cohort comparison and from calibration. Unresolved raw measurements are retained, unknown critical metadata does not compare equal, model/run IDs are excluded from cohort keys, and preference/latency windows—not static accuracy generally—require calendar alignment.
- Image input, output generation, editing reference, and mask roles are separated. Accepted formats are set-valued facts; mask semantics and requirements are separate; combined context limits are not silently treated as input-only.
- Source rights distinguish research, extraction, and redistribution. Local compilation requires research and extraction; distributable compilation additionally requires redistribution. Unknown, absent, expired, or conflicting permission fails closed in the stated decision rules.
- Collection choreography correctly separates durable observation/output recording from job completion, uses fencing tokens, and releases dependencies only after `JobCompleted`.
- Host runtime eligibility is not inferred from catalog listing, and route-time network/database/inference access is excluded.

## Exact remaining blockers

### 1. Snapshot distribution class is required by the API but absent from the authoritative DDL

`compile_snapshot(..., distribution_class)` says the class is pinned into the artifact, and `publish_snapshot` re-evaluates rights against the artifact's pinned class. `snapshots` has no `distribution_class` column. Nor does an authoritative artifact format define where that pinned value lives. Consequently validation/publication cannot prove the class from relational authority rather than trusting opaque `projection_bytes`.

**Needed artifact changes:** add `snapshots.distribution_class` with the exact `local|distributable` check (or an equally explicit normalized fact), include it in snapshot hash/validation/event payload contracts, and define the publication check against that field. Also state whether `channel_class` must equal the configured class of `channel`; currently a caller supplies both but no channel-to-class authority is defined.

### 2. Snapshot rights decisions are not reproducible from the manifest

Compilation requires the current non-conflicting source policy for every input, and publication rechecks current policy. But `snapshot_inputs` pins only observation IDs. Source-policy rows are versioned by `(source_id, config_hash)` and are not observations, and no snapshot relation records which policy fact authorized each input at compile time. Thus the claimed reproducibility of a snapshot from its manifest cannot reproduce the compile-time rights decision. The publication check also has no declared persisted audit result identifying the policies evaluated.

**Needed artifact changes:** define a snapshot-rights input/audit relation keyed by snapshot and source/input that pins the authorizing `source_policies.config_hash`, permission uses, decision time, and result; include its ordered representation in the content hash. Specify publication's current-policy query and persisted decision atomically with `snapshot_publications`.

### 3. The snapshot byte/hash contract needed by the offline library does not exist

`adopt_snapshot(artifact_bytes, expected_schema)` promises hash, integrity, schema, freshness, and atomic-adoption behavior, but no document defines the artifact envelope, canonical byte encoding, section layout, hash domain, manifest encoding/order, compatibility rules, or how `content_hash` is excluded/included while hashing. `projection_bytes` is opaque. A Go library cannot independently produce a conforming reader or verify reproducibility from these contracts.

**Needed artifact changes:** add a versioned snapshot artifact specification with canonical encoding, required sections and field types, deterministic ordering, hash algorithm/domain, schema compatibility matrix, limits, and validation procedure. Then give `adopt_snapshot` concrete Go request/result types against it.

### 4. The public Go library surface is still conceptual, not an implementable API contract

`route(RouteRequest)` and `adopt_snapshot(...)` name methods, but the supposedly complete `RouteRequest`, `RoutingDecision`, `Selection`, price, latency, configuration, rejection, warning, freshness, and identity-evidence structures remain prose. Important values such as privacy expressions, reasoning configuration, billing regimes, strictness, relaxation reasons, configuration, and evidence-tier vocabularies have no closed Go sum/type definitions. There is no constructor signature or versioning/compatibility behavior. CLI incompleteness is acceptable while the CLI is deferred; library incompleteness is not.

**Needed artifact changes:** publish the concrete exported Go types and method signatures, all closed enum values/tagged variants, units/bounds, validation errors, and compatibility/version behavior. Keep CLI wire details deferred if the CLI slice is not implemented.

### 5. Disabled calibration leaves performance-mode ranking undefined

The documents correctly forbid treating benchmark scores as calibrated success. However, performance mode is defined as maximizing a calibrated quality lower bound, while the current lane explicitly has no calibration. The router also promises an “honest ... evidence-based ranking” and deterministic ordering, but does not define which comparable raw measurements can rank across dimensions/cohorts or what performance mode does when every candidate is uncalibrated. “Unknown candidates are a separately labelled fallback class” does not specify whether any are admitted absent a numeric floor or how the primary/fallback pools are ordered.

This does **not** require empirical data now; it requires a disabled-lane rule.

**Needed artifact changes:** specify the exact uncalibrated state machine for both modes: admission with/without a numeric quality floor, whether performance mode is rejected or reduced to a named deterministic policy, primary versus fallback pool selection, and complete tie ordering. Do not introduce cross-cohort arithmetic.

### 6. Collector output idempotency names an `output_hash` that no request or DDL records

`record_attempt` is idempotent by full work tuple, `output_hash`, and parser version. Its request has `observation, assertions, coverage_findings`, but no named output hash; `stage_output_links` stores only `work_input_hash` and opaque `output_id`; `job_completions` also stores only opaque `output_id`. There is no typed stage-output relation or declared derivation of a canonical output hash. The API response's “durable-output-accepted” therefore cannot be implemented with the stated key, and crash adoption cannot byte-compare retried output.

**Needed artifact changes:** define a canonical stage-output envelope and hash calculation, add the hash (and stage/output type if needed) to the API, link/completion DDL, and `StageOutputRecorded`, and state exact same-key/same-hash versus same-key/different-hash behavior.

### 7. `StageOutputRecorded` does not carry the exact work identity it claims to validate

The durable work identity is `(source_id, resource_url, input_hash, parser_version, stage)`. The event payload lists observation ID, job ID, work-input hash, output ID, and source ID. `stage_output_links` likewise lacks resource URL, parser version, and stage. A consumer can currently recover those by dereferencing mutable job counters plus an immutable-looking job row, but that dependency and validation algorithm are not the declared event contract, and the event is not independently sufficient for the asserted “exact-work validation.”

**Needed artifact changes:** either put the complete work tuple and output hash in the link/event, or explicitly make `job_id` the immutable tuple authority and define the transactional consumer comparison, including which fields are immutable and what mismatch outcome is persisted.

### 8. Collection job state and attempt vocabulary are not defined in DDL/API/events

The aggregate has queued, leased, deferred, completed, quarantined, and lease-expiry transitions. `jobs` has no state; the documents do not provide a deterministic state derivation from leases/attempts/completions. `job_attempts.outcome` says only “CHECK classified outcome,” with no enum. `defer_job` accepts `outcome` without a closed vocabulary. Quarantine has a state transition but no named event or exact persisted outcome/re-entry derivation. `retry_not_before` is NOT NULL for every attempt although the contract says attempts include terminal outcomes and does not define a sentinel/deadline rule for those.

**Needed artifact changes:** provide either an explicit locked job-state column and transition constraints or a complete deterministic derivation/query; enumerate attempt/defer/quarantine outcomes; define retry timestamp semantics for every outcome; and align transitions, events, API errors, and re-entry prerequisites.

### 9. Observation/parser fields conflict across object, event, API, and DDL

`SourceObservation` has `collector_version` and `body_locator`; `observations` has `collector_version` but no `body_locator` (apparently relying on `body_hash -> content_blobs`, without saying that this replaces the object field). `ObservationRecorded` names `parser_version`, which is not an observation field or column. `record_attempt` accepts an `observation` plus assertions, while the method-level idempotency also names parser version without defining whether it is the job parser version, every assertion's parser version, or the event field.

**Needed artifact changes:** choose one authoritative field model. Remove `body_locator` or map it to a defined content-store locator relation; correct the event to `collector_version` or add a real observation parser version; define and validate the relationship among job parser version, assertion parser version, and collector version.

### 10. Core collector enums and cross-field checks are placeholders, not exact DDL contracts

The data section claims explicit CHECK constraints, but several first-slice tables say only “CHECK source kind,” “CHECK outcome,” “CHECK stage/hash,” “CHECK classified outcome,” and “CHECK evidence enum.” Domain labels contain human-readable forms such as “catalog API” while no storage tokens are fixed. Observation status/outcome compatibility is only asserted in prose. This prevents schema generation and permits adapter drift in a security-sensitive ingestion boundary.

**Needed artifact changes:** enumerate exact storage tokens and SQL CHECK expressions for source kind, observation outcome, job stage, attempt outcome, evidence class, coverage channel inventory (or a referenced channel-policy table), and status/outcome compatibility. Name deferred trigger/function contracts for invariants that cannot be plain CHECKs.

### 11. Image capability vocabulary and legal matrices are still incomplete

`capability_assertions` promises closed capability and legal modality/tool-direction checks, but no document enumerates the capability tokens (including the alpha support later required by routing) or the complete capability/operation/direction matrix. Image format and limit DDL similarly says “legal image operation/role” or “every legal ... combination above,” while the prose leaves cases such as conditioned generation inputs to an additional advertised capability without naming that capability. `image-capabilities.md` also says capability is “one of the four payload subtypes” although it defines capability, format, semantic, requirement, and limit—five variants.

**Needed artifact changes:** add exhaustive token tables/matrices for capability, operation, direction, role, format grammar, and every limit combination; name the conditioned-generation and alpha capabilities; turn them into exact CHECK/validation contracts; correct the subtype count.

### 12. Typed benchmark metadata is not concretely replayable yet

`benchmark_cohort_rules.canonical_grammar` is free `TEXT`. The contract says it carries ordered enum members, bounds/ranges, or instant-pair form, but defines neither a meta-grammar nor canonical escaping/normalization. `canonical_value TEXT` likewise lacks per-type canonical lexical rules. Implementations can hash semantically identical integers/decimals/instants/enums differently. In addition, registration must reject “incomplete critical coverage,” but there is no exact family-to-required-component registry or algorithm resolving the repeated “where applicable” conditions.

This blocks deterministic cohort hashes; it does not call for calibration data.

**Needed artifact changes:** define a versioned canonicalization specification for every `value_type` (Unicode/bytes, decimal and integer lexical form and bounds, enum escaping/order, artifact IDs, UTC instant-pair precision), a machine-parseable grammar schema rather than unconstrained text, and an exhaustive family/component requirement matrix with explicit applicability/not-applicable rules. Include exact ordered serialization used for policy and cohort hashes.

### 13. Several event payloads are not byte-level replay contracts

The common envelope says payloads use “a versioned boundary serializer,” but names no encoding or canonicalization. Binding events are described collectively (“with ... IDs ... times ... versions”) rather than as exact per-event schemas. `SnapshotPublished` omits identity authority/freshness versions, comparison-policy versions, knowledge cutoff, distribution class, channel, and publication ID even though host adoption/audit depends on these. `SourcePolicyChanged` says “permission decisions” without exact fields and omits effective/expiry/terms evidence needed to interpret the decision if the policy row is unavailable.

Events may legitimately be notification events whose consumers load authoritative rows, but that must be explicit and replay behavior must then define required row retention and transaction visibility. The present claim of byte-exact typed payloads is insufficient.

**Needed artifact changes:** specify canonical payload encoding and exact versioned schema for every event; classify each field as self-contained fact or authoritative-row reference; add the missing snapshot publication/channel/class/version identifiers; and define replay failure behavior when referenced rows/schema readers are unavailable.

### 14. ADR status leaves a live Rust decision alongside the Go decision

ADR 0002 says it replaces the original Rust language choice, but ADR 0001 remains `Status: Proposed` and its Decision still says “Embed Rust.” Cross-references do not mark it superseded. A reader enumerating current ADR decisions sees two language decisions.

**Needed artifact changes:** mark ADR 0001 superseded in its status/header and link ADR 0002, or amend its decision text while preserving history. No code change is needed.

## Gate conclusion

Planning may continue, including planning raw benchmark collection without calibration. Implementation should remain sliced behind contract acceptance:

- **Collector core:** blocked by items 6–10 and 13; rights-aware snapshot publication additionally by 1–2.
- **Offline routing library:** blocked by items 1, 3–5, 11–12 (to the extent those facts are admitted), and 13.
- **Calibration/HostReceipt outcome processing:** intentionally later and still gated by the already acknowledged missing outcome/calibration tables; this review does not promote that known gate into a requirement for the raw collector or disabled-calibration routing plan.
