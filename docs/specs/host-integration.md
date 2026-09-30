# Host integration contract

Version 1. Normative for this delivery under ADR 0003. Owns the host adapter boundary: live registry discovery, request construction, routing invocation, spawn handoff, outcome capture and snapshot adoption. The routing domain API, request validation and selection algorithm are owned by [educated routing](educated-routing.md). Storage ports for receipts are owned by [delivery storage](delivery-storage.md). Research export contracts are unchanged.

## Registry discovery

The host adapter reads the host process's live model registry through its native read-only API and translates entries into `RuntimeOfferData`. The registry is the only candidate source; no hardcoded model list exists in this delivery. Adapter rules:

- Runtime IDs are `Provider + "/" + UpstreamID` with internal upstream slashes retained, exactly as [educated routing](educated-routing.md) requires. The adapter never synthesizes IDs from display names.
- `Eligible` is true only for entries the host can currently execute: credentials present, account entitled, no host-side disable. Catalog listing is not eligibility.
- Empty capability, limit or price lists mean unknown, never empty support. Host-supplied facts override the same provider/upstream snapshot components only; another provider's quote is never borrowed.
- One registry read produces one `RuntimeGeneration` string, derived deterministically from the sorted runtime IDs. Two requests with the same generation string saw the same registry. The adapter recomputes the generation per read and never mutates it.
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
- The CLI is a thin testing and early integration adapter that imports the same public routing API and performs no ranking, policy or storage logic. Its commands, flags and wire schemas are specified when that slice is implemented, not before.

## Verification cases

Before the host integration slice is accepted: drive the real host registry and spawn path, not a synthetic ranked list; exercise text-only, vision-input, generation and edit requests; verify nil-selection propagation with reasons on `no_admissible_offer`; verify receipt idempotency under duplicate submission and conflict on same ID with different bytes; verify an `unvalidated_completion` cannot be stored as validated; verify adoption retains the prior generation on corrupted or oversized artifacts; verify two requests sharing a runtime generation produce identical candidate sets; and verify no network access occurs during `Route`.
