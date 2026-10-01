# Host integration contract

Version 2. Pre-implementation normalization (ar-9r1): runtime generation redefined over the full sorted runtime projection (mutable registry metadata now changes the generation), the native registry read and spawn-handoff wire schemas are closed below, and the CLI adapter contract now exists in [CLI adapter](cli-adapter.md). Normative for this delivery under ADR 0003. Owns the host adapter boundary: live registry discovery, request construction, routing invocation, spawn handoff, outcome capture and snapshot adoption. The routing domain API, request validation and selection algorithm are owned by [educated routing](educated-routing.md). Storage ports for receipts are owned by [delivery storage](delivery-storage.md). Research export contracts are unchanged.

## Registry discovery

The host adapter reads the host process's live model registry through its native read-only API and translates entries into `RuntimeOfferData`. The registry is the only candidate source; no hardcoded model list exists in this delivery. Adapter rules:

- Runtime IDs are `Provider + "/" + UpstreamID` with internal upstream slashes retained, exactly as [educated routing](educated-routing.md) requires. The adapter never synthesizes IDs from display names.
- `Eligible` is true only for entries the host can currently execute: credentials present, account entitled, no host-side disable. Catalog listing is not eligibility.
- Empty capability, limit or price lists mean unknown, never empty support. Host-supplied facts override the same provider/upstream snapshot components only; another provider's quote is never borrowed.
- One registry read produces one `RuntimeGeneration` string: the `canonical_hash` (64 lowercase hex characters) of the canonical-v1 encoding of an array with one member per runtime entry, each member the closed object `{id, provider, upstream_id, revision, eligible, capabilities, unsupported_capabilities, limits, prices, region, privacy}` mirroring `RuntimeOfferData` exactly (limits as `{resource, maximum}` sorted by resource, prices as `{component, usd_per_million}` sorted by component, capability lists deduplicated and sorted). Entries sort by `(provider, upstream_id)` bytewise ascending; this hashing order is an encoding detail of the generation string only and is independent of the domain's stable-ID iteration order. Same generation string means identical registry content — not merely identical IDs — so any change to a runtime offer's metadata changes the generation. The adapter recomputes the generation per read and never mutates it.

### Native registry read wire boundary

The adapter consumes only this closed projection of each host registry entry and maps it one-to-one onto `RuntimeOfferData`; fields the native API does not expose map to their unknown sentinels, never to invented values:

- `provider`, `upstream_id`: nonempty identity strings from the host's own naming; the adapter derives `id` as `provider + "/" + upstream_id` and never from a display name. Display names and marketing labels are not consumed for identity, ranking or generation.
- `revision`: host revision assertion string, empty when the host exposes none.
- `eligible`: boolean the adapter computes from host state (credentials present, account entitled, not disabled), not from catalog listing.
- `capabilities`, `unsupported_capabilities`: string lists over the closed capability vocabulary. Unknown host capability tokens map to neither list — an unrecognized capability is not a negative fact — and the adapter records nothing about it.
- `limits`: list of `{resource, maximum}` over the closed resource vocabulary; a resource absent from the list is unknown.
- `prices`: list of `{component, usd_per_million}` over the closed component vocabulary; absent components are unknown and an empty list never means free.
- `region`, `privacy`: bounded strings, empty means the host asserts nothing.

Anything else in the native registry entry (pricing metadata outside the vocabulary, rate limits, distillation notices, icon URLs) is outside the boundary: not read, not hashed, not routed on.
- The adapter holds no credentials inside routing data and passes none into the domain. Credential handling stays in the host's provider clients.

## Request construction

The host adapter builds `RequestData` from the calling task's structured context. It owns token and image quantity estimation, the effective UTC timestamp and the runtime generation. Construction rules:

- `EffectiveAt` is the host clock at construction, explicit UTC. Replay of a stored request reproduces identical routing because the domain never reads the clock.
- Quantities are estimates from tool context: input size, expected output bound, image counts and dimensions. Zero image quantities with `ImageOperation none` is valid for text-only calls. Unknown quantities are estimated conservatively, never negative.
- `StrictCost` is set when the caller declares a hard spend bound; `StrictLatency` when a real deadline exists. `MaxAgeSeconds` reflects the caller's freshness requirement; zero disables it.
- `AllowUnscored` and `MinimumEstimate` come from the caller's quality policy, not the adapter. The adapter never raises them to manufacture a selection.
- The adapter rejects construction it cannot fill honestly with `invalid_request` rather than inventing values. It does not classify task difficulty from prose; it accepts the caller's declared task family.

## Routing invocation

The adapter calls `NewRequest` then `Engine.Route`. The call is local, pure and offline: no network, database read or inference happens on the route path. The adapter handles the closed error codes as follows: `invalid_request` is a caller bug surfaced with its safe detail; `invalid_snapshot`, `schema_incompatible` and `snapshot_unavailable` are adoption failures handled below; `no_admissible_offer` returns the complete `Decision` (nil `Selected`, full rejections) and the error together, per [educated routing](educated-routing.md) — the adapter surfaces both, never a hidden fallback model. The adapter never relaxes hard constraints to produce a candidate. Decision warnings are surfaced verbatim to the caller; the adapter neither filters nor summarizes them away.

## Spawn handoff

The host owns spawning, tools, retries, cancellation and evaluation. On a selected decision the adapter resolves the decision's `Provider` and `UpstreamID` back to the live registry entry and hands the caller the host's own provider client and configuration. The adapter adds no retry policy and no fallback spawn. If the registry entry vanished between generation read and spawn, the host reports a cancelled outcome and may re-route with a fresh generation. Parallel selections use host-owned admission reservations; routing never starts a device or GPU job itself.

### Spawn-handoff wire boundary

The handoff to the host spawner is one closed object: `request_id` (correlation with the decision), `provider`, `upstream_id` (must resolve to a registry entry present in the same generation read that produced the decision; a vanished entry is the cancelled-and-re-route path above, never an automatic alternate target), `task_family` (the request task, for host-side telemetry only — never used to pick a different model), `deadline_unix_ms` (derived: `effective_at` plus `deadline_millis` as UTC epoch milliseconds — the domain deadline is a latency bound, so the handoff converts it to the absolute instant it implies; 0 when `deadline_millis` is 0, and the host owns what it does with it) and `model_ref` (the host's own client configuration key when it differs from `upstream_id`, empty otherwise; it names configuration, never another model). The spawner receives no snapshot bytes, no runtime generation and no ranking data. Provider credentials stay inside the host's provider clients; the handoff carries no secret material.

## Outcome capture

After execution the host adapter records a receipt through `record_receipt`. Receipt contract:

- `receipt_id`: host correlation string, unique per attempt.
- `request_id`: the routed request's ID, correlating the outcome with its decision.
- `offer_id`, `task`: the selected offer and the request task family. `offer_id` is the runtime offer identity exactly as it appears in the decision; an unmapped live offer selected through unscored fallback records its runtime ID and requires no catalog mapping.
- `outcome`: `validated_success` or `validated_failure` only when a mechanical validator (real tests, schema check, host-side checker) produced the verdict and `validator_version` names it. Provider-reported completion is recorded as `unvalidated_completion`. Abandoned or interrupted work is `cancelled`; work whose validator could not run is `indeterminate`.
- Usage components: `input_tokens`, `output_tokens`, `input_images`, `output_images`, `references`, `searches`, `elapsed_millis`, each recorded as known with a value or explicitly unknown. Unknown usage is never zero. Billed usage on failed or cancelled image operations is still recorded.
- Forbidden payload: task prompts, file contents, images, tool outputs and credentials are never stored. The receipt carries classifications and measurements only.
- Capture happens from existing normal usage. This delivery launches no additional paid evaluation campaign; outcome volume grows from real host work.

Receipts feed `outcome_facts` and later snapshot signals with basis `host_observed`. A receipt never rewrites an in-flight route; feedback lands in a successor snapshot.

## Snapshot adoption

The host adapter owns local artifact reads. It polls or watches the configured local publication directory, verifies and adopts new generations through `adopt_snapshot` semantics: canonical byte check, hash recomputation, schema 1 and policy `educated-v1` check, bound checks, atomic swap. A failed adoption retains the prior generation and logs the safe rejection reason. `local` distribution artifacts are the default; adopting a `distributable` artifact on the host is permitted and changes nothing about route behavior. Remote artifact fetch and signature verification are outside this delivery. The adapter exposes the active artifact hash for correlation with decisions.

## Security and trust boundaries

- Host assertions (registry, eligibility, offer-specific prices and limits) are trusted local facts inside the owning process. They never become public research approval and never alter source review or permissions.
- Source text and model cards cannot alter routing policy; adapters sanitize content crossing into any extraction lane, per [collection pipeline](collection-pipeline.md#source-rights-and-operating-limits).
- Error details crossing the boundary contain no source bytes, credentials or task content.
- The CLI is a thin testing and early integration adapter that imports the same public routing API and performs no ranking, policy or storage logic. Its commands, flags, strict wire schemas, errors and exit codes are specified in [CLI adapter](cli-adapter.md) and precede its implementation.

## Verification cases

Before the host integration slice is accepted: drive the real host registry and spawn path, not a synthetic ranked list; exercise text-only, vision-input, generation and edit requests; verify nil-selection propagation with reasons on `no_admissible_offer`; verify receipt idempotency under duplicate submission and conflict on same ID with different bytes; verify an `unvalidated_completion` cannot be stored as validated; verify adoption retains the prior generation on corrupted or oversized artifacts; verify two requests sharing a runtime generation produce identical candidate sets; verify a registry metadata change (price, limit, eligibility flip) with unchanged IDs produces a different runtime generation; and verify no network access occurs during `Route`.
