# Host registry and spawn discovery (phase 5, read-only)

Verified surface of the target host (the pi coding agent,
`@earendil-works/pi-coding-agent`) as of the installed package. Source of
truth: `dist/core/model-registry.d.ts`, `dist/core/provider-composer.d.ts`
and `docs/{extensions,sdk,custom-provider,virtual-models,models}.md` of the
installed package. No integration implemented here; phase 14 contracts
consume this.

## Registry surface (`ctx.modelRegistry`, class `ModelRegistry`)

- `getAll(): Model<Api>[]` — every known model across composed layers
  (built-in, `models.json`, extension-registered).
- `getAvailable(): Model<Api>[]` — models whose provider has working
  credentials. Availability is credential-based, not capability-based.
- `find(provider, modelId)` / `getModelOfType(type, provider, modelId)` —
  exact identity lookup. Identity at the host is `(provider, modelId)`
  strings; this is the seam where provider/upstream identity mapping lives.
- `getModelsOfType(type, provider?)` and
  `getAvailableOfType(type, provider?, options)` — typed families: `chat`,
  `image`, `classifier`. `getAvailableOfType` is async (auth check).
- `hasConfiguredAuth(model)`, `isUsingOAuth(model)`,
  `getProviderAuthStatus(provider)` — synchronous auth/entitlement signals;
  `getProviderAuth(provider)` / `getApiKeyForProvider(provider)` are async.
- `refresh(options?): Promise<ModelsRefreshResult>` — reloads `models.json`;
  docs state callers must await refresh before synchronous registry reads.
  The registry is a synchronous compatibility facade over `ModelRuntime`.
- `registerProvider(provider | name, config: ProviderConfigInput)` /
  `unregisterProvider`, `registerVirtualModel(definition)` /
  `unregisterVirtualModel` — runtime layer changes; registrations are not
  persisted across host restarts.

`ProviderConfigInput` (provider-composer.d.ts): `name`, `baseUrl`, `apiKey`,
`api`, `streamSimple`, `images`, `classifiers`, `headers`, `authHeader`,
`oauth`, `models: ProviderModelConfig[]` and an optional
`refreshModels(context)` hook the host can call to re-enumerate a provider's
models dynamically. Auth status sources observed in the type:
`stored | runtime | environment | fallback | models_json_key |
models_json_command`.

## Request/spawn surface

- Nested model calls from extensions go through
  `ctx.modelRegistry.streamSimple(model, context, options)` /
  `complete(...)` — provider-neutral, request-time authentication. No
  network, retry or credential handling belongs to a router on top of this.
- Programmatic sessions: `createAgentSession()` (SDK, `docs/sdk.md`) owns
  one conversation, model and tools; `AgentSessionRuntime` adds
  `newSession/switchSession/fork/importFromJsonl`.
- No documented subagent-spawn tool exists inside pi itself. Spawn-style
  orchestration observed in this repo's workflow comes from the driving
  harness, not a pi API. Phase 14 must therefore define its own process
  boundary (framing, timeouts, generation handoff) and measure one-shot
  process overhead before choosing a long-lived transport; nothing here
  implies a daemon.

## Router-relevant consequences

1. The live registry is dynamic: `refresh()` plus `refreshModels` hooks mean
   snapshot adoption must tolerate registry membership changing between
   refresh and route. Route intersects the snapshot with the registry read
   at routing time; unmapped models surface as diagnostics.
2. Credential truth is per provider, not per model: eligibility checks map
   `hasConfiguredAuth`/`getProviderAuthStatus` onto offers. OAuth
   (`isUsingOAuth`) is a distinct credential class (recorded constraint:
   agent-facing testing uses OAuth, not API keys).
3. Model identity at the host is `(provider, modelId)`; the research
   catalog's provider offers must map onto exactly this pair shape. The 394
   exact-ID joins are the current evidence base for that mapping.
4. Type families (`chat`/`image`/`classifier`) are host-native operation
   classes; image *output* research batches map to the `image` family, and
   the images/language separation in the backlog mirrors the type split
   rather than a modality flag.
5. The registry API is in-process TypeScript. A Go router reaches it through
   a process boundary whose contract is intentionally unwritten until phase
   14 measures it.

## Unknowns left to phase 14

- Exact `Model<Api>` field surface the adapter may read (context limits,
  tool support, reasoning options) — enumerate from `@earendil-works/pi-ai`
   types when the adapter contract is written.
- Whether `models.json` layering can express every credential class the
  production host uses, or an extension-registered provider is required.
- Measured cost of one-shot `agent-routing` process startup versus a
  persistent boundary.
