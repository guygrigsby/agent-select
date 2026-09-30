# Image capability contract

Proposed contract overall, but the relational subset referenced by `contracts.md` and the delivery DDL (capability, format, semantic, requirement and limit assertion child tables) is normative for this delivery under ADR 0003. Owns the image and numeric-limit payload variants and child tables referenced by `contracts.md`. No implementation or complete model coverage is claimed. Every field below is required unless represented by absence of a separate fact row. Missing evidence is unknown, not unsupported or unlimited.

## Facts and admission

Operation is `text_assistance`, `vision_understanding`, `image_generation` or `image_editing`. Role identifies what a fact constrains: `request`, `input_image`, `output_image`, `reference` or `mask`. A reference is an input image used for conditioning or editing. A mask is a distinct input plane defining edited regions. Neither counts as a generated output.

Capability, accepted format, image semantics, requirement and numeric limit are separate immutable Assertion payload variants. Each references the assertion's source evidence and offer/revision subject. Adapters translate explicit publisher statements into scoped facts; model kind or a family name never establishes image support. Input image support does not establish output generation, editing or mask support.

Capability fields: `capability`, `support: supported|unsupported|unknown`, `operation` and `direction: input|output|not_applicable`. Modality assertions require input/output direction. Tool use may use not_applicable. Image-format fields: `operation`, `role`, `format`. Distinct accepted formats form a set, not conflicting categorical answers. Image-semantic fields: `operation`, `role`, `semantic`; known semantics include `alpha_zero_edits`, `alpha_one_edits`, `white_edits` and `black_edits`. An unknown semantic has no known semantic row. Mask interpretation is never inferred from an alpha channel's presence.

Requirement fields: `operation`, `role: reference|mask` and `requirement: required|optional|forbidden|unknown`. Formats, alpha support and mask interpretation are separate assertions, not columns whose meaning changes with requirement state. Required/optional facts do not establish a format or dimension by themselves.

Limit fields: `limit_name`, `value: nonnegative integer`, `unit`, `bound: minimum|maximum|exact`, `scope: input|output|combined`, `operation` and `role`. Canonical legal combinations:

| Name | Unit | Scope | Role |
|---|---|---|---|
| `context_tokens` | token | input, output or combined as explicitly documented | request |
| `input_tokens` | token | input | request |
| `output_tokens` | token | output | request |
| `image_count` | image | input or output | input_image or output_image respectively |
| `reference_image_count` | image | input | reference |
| `mask_count` | image | input | mask |
| `image_width`, `image_height` | pixel | input or output | matching image, reference or mask role |
| `image_pixels` | pixel_area | input or output | matching image, reference or mask role |
| `image_bytes` | byte | input or output | matching image, reference or mask role |

Token limits apply to the named operation and role=request. Vision/text assistance have no image output role. Generation/editing may have output images. Reference or ordinary image inputs to generation require explicitly advertised conditioned-generation support. Mask is legal only for editing. Reference/mask roles always have input scope; output_image always has output scope. No numeric limit has not_applicable scope.

A context window is not universally input-only. Source evidence must distinguish input-only, output-only and combined accounting. For a combined bound, admission sums conservative input and output token bounds under the documented accounting rule, including image tokens when charged to context. Host-estimated image token usage needs an approved estimator or a proven upper bound. Unknown required accounting fails strict admission; no generic pixel-to-token conversion is assumed.

Select current verified facts matching subject, operation, scope and role. Within a numeric scope intersect all supported bounds: use the highest minimum, lowest maximum and consistent exact values. An empty intersection rejects. Different source values need not conflict if their intersection is usable. Categorical contradictions such as supported and explicitly unsupported remain conflicts; unknown assertions do not prove either side. Missing required formats, bounds or semantics reject as unknown. Retractions and stale evidence cannot satisfy requirements. Identity-ambiguous revision facts are excluded; proven offer-scoped facts remain usable under the router's explicit fallback policy.

Check references and masks separately from outputs. The request supplies each image's format, width, height and byte upper bound, plus counts by role. Width times height is pixel_area. A required mask must be present with evidenced format, dimensions and interpretation. Optional mask absence is valid. Supplying a forbidden or unknown-support mask does not pass. Request alpha requirements require an independent supported alpha capability. No format fact alone proves alpha semantics.

## Relational contract

All columns are NOT NULL. All child tables are owned by Assertion and have `assertion_id TEXT PRIMARY KEY REFERENCES assertions(id)` with restrictive FK. That PK supports assertion lookup. Exactly one payload subtype per assertion is enforced by a deferred subtype-completeness constraint and publication validation. All rows are append-only; retraction uses the common assertion-retraction fact.

- `capability_assertions` is defined authoritatively in [contracts](contracts.md#data-layer); this contract owns its legal modality/tool direction combinations and names the capability-admission index `(operation,direction,capability,support)` there. It remains one of the five payload subtypes (capability, format, semantic, requirement, limit); no separate image capability table exists.
- `image_format_assertions`: common PK plus `operation TEXT`, `role TEXT`, `format TEXT`. CHECK legal image operation/role and a validated advertised format name. Index `(operation,role,format)` for accepted-format membership.
- `image_semantic_assertions`: common PK plus `operation TEXT`, `role TEXT`, `semantic TEXT`. CHECK operation=image_editing, role=mask and the declared semantic enum. Index `(operation,role,semantic)` for mask interpretation.
- `image_requirement_assertions`: common PK plus `operation TEXT`, `role TEXT`, `requirement TEXT`. CHECK role reference/mask, requirement enum and legal operation/role. Index `(operation,role,requirement)` for reference/mask admission.
- `limit_assertions`: common PK plus `limit_name TEXT`, `value BIGINT CHECK(value>=0)`, `unit TEXT`, `bound TEXT`, `scope TEXT`, `operation TEXT`, `role TEXT`. CHECK every legal name/unit/scope/operation/role combination above and the bound enum. Index `(operation,scope,role,limit_name)` for bound intersection. No duplicate mutable effective-limit columns are stored.

## Service and event cross-check

`record_attempt` accepts these complete typed variants in its assertion list. Collector-worker identity, source assignment, lease fencing and idempotency remain its existing contract. Invalid combinations or missing evidence fail `invalid_evidence`; unknown knowledge stays explicit and does not create a numeric row. SourceObservation.Record owns immutable output creation. `ObservationRecorded`/`StageOutputRecorded` identify the durable observation/output link; consumers retrieve typed children by assertion ID. No extra per-field lifecycle or aggregate is introduced.

`route` receives tagged operation and role-separated image specifications through RouteRequest. Its response includes `rejections` with offer ID, requirement scope/role, reason and evidence IDs. Reasons are unknown, unsupported, conflicting, stale or exceeded. Public error taxonomy remains unchanged; no admissible offer is the aggregate result of these candidate rejections. HostReceipt uses separate usage components with category, input/output direction and image role; known quantities have bounds, unknown quantities are explicitly unknown. Image bytes are included as a usage category without assuming they are a billed unit.

## Verification cases

1. Verified JPEG vision input within its input-image bounds admits; no output-image assertion is needed.
2. A multimodal language label without evidenced image input rejects as unknown, not unsupported.
3. Two advertised output formats PNG and JPEG are both usable; neither invalidates the other.
4. Editing references within bounds do not establish mask format/interpretation. Required but unknown mask semantics reject.
5. Two outputs above output-image maximum reject even when input-image counts permit two.
6. A combined context limit rejects input plus output above the bound; the same input alone fitting cannot admit it.
7. Mask dimension limits do not constrain generated-output dimensions. Inconsistent minimum/maximum or exact bounds reject their own scope.
