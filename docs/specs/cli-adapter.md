# CLI adapter contract

Version 2. Adds the `receipt record` host-reporter command so the host adapter's process boundary covers [host integration](host-integration.md) outcome capture; everything else is version 1 unchanged. Version 1. Normative under ADR 0002 and ADR 0003 once the host integration slice (phase 14) is implemented; specified now so the adapter is coded against a closed contract. The CLI is a thin testing and early integration adapter over the public Go routing API ([educated routing](educated-routing.md)): it parses one request, invokes `NewRequest` and `Engine.Route`, and prints the decision. It contains no ranking, policy, weighting, storage or collection logic, re-decides nothing and adds no fallback. Where this document and the Go API disagree, the Go API and [educated routing](educated-routing.md) win and this document is fixed.

## Commands and flags

- `agent-routing route --snapshot PATH --request PATH`
- `agent-routing snapshot check --artifact PATH`
- `agent-routing snapshot adopt --artifact PATH --store PATH`
- `agent-routing receipt record --receipt PATH [--dsn DSN]`
- `agent-routing receipt lock --path PATH`

`receipt record` is the host reporter's outcome-capture boundary: it reads one canonical-v1 receipt document (the closed grammar of [host integration](host-integration.md) outcome capture — members exactly `receipt_id`, `request_id`, `offer_id`, `task`, `outcome`, `validator_version`, `usage`), opens Postgres exactly as `collect` does (`--dsn`, then `AGENT_ROUTING_DSN`, then config.toml) and calls the `record_receipt` port. It stores nothing else and re-decides nothing. A missing `host_local` source is the port's `not_found`: the operator registers it, this command never invents authority. `recorded_at` is the host clock at submission; `collector_version` comes from config with the `collect` default. Exit 0 writes one line to stdout: the canonical-v1 object `{receipt_id, observation_id, usage_hash}` (hex usage hash); port errors keep their port code on the one-line stderr form with exit 4, invalid documents exit 2 with `invalid_request`.

`receipt lock` is an offline Host Execution transport primitive defined completely in [pi host adapter](pi-host-adapter.md#durable-delivery-and-events): acquire a nonblocking crash-released exclusive OS file lock, emit `{locked:true}` plus LF, hold until stdin EOF and release. It has no Postgres or routing logic. Lock contention exits 4 with fixed `lock_busy`; other acquisition failures use fixed `lock_unavailable`. It never prints a path. A locked attempt cannot be recovered by another host. Lock files contain no domain state.

`route` is the whole point: one process, one route, exit. `--request -` reads the request from stdin. `--snapshot` is a local artifact path (the same artifact bytes `snapshot adopt` stores); the CLI never fetches, never takes a URL and holds no credentials. `snapshot check` verifies artifact bytes (canonical check, hash recomputation, schema and policy pins, bound checks) without changing any state and reports the artifact hash. `snapshot adopt` verifies the same way then atomically installs the artifact into the local store directory (temp file, rename) exactly as [host integration](host-integration.md#snapshot-adoption) requires; a failed adoption keeps the previous generation and exits nonzero with the safe reason. Unknown flag, unknown subcommand or missing required flag is a usage error: usage text on stderr, exit 2, nothing on stdout.

Process boundary honesty: each invocation pays process startup, artifact read and decode before routing. Any latency claim about the adapter measures a warm in-process route separately; the one-shot cost is the process boundary, per ADR 0002.

## Request wire schema

The request input is the canonical-v1 encoding (UTF-8, no BOM, LF newlines, exactly one trailing newline) of one closed object, snake_case mirror of `RequestData`. Unknown members, duplicate members or invalid encoding are `invalid_request_wire` (exit 2). Members:

`id`, `task`, `mode`, `effective_at` (RFC 3339 UTC with milliseconds), `runtime_generation`, `runtime` (array of runtime offer objects), `required_capabilities` (string array), `input_tokens`, `output_tokens`, `image_operation`, `image_count`, `image_width`, `image_height`, `image_bytes`, `reference_count`, `mask`, `mask_alpha` (booleans), `searches`, `budget_known`, `budget_micros`, `strict_cost`, `deadline_millis`, `strict_latency`, `max_age_seconds`, `allowed_providers` (string array), `region`, `privacy`, `allow_unscored`, `minimum_estimate`.

A runtime offer object uses the same closed grammar as the native registry read projection in [host integration](host-integration.md#native-registry-read-wire-boundary): `id`, `provider`, `upstream_id`, `revision`, `eligible`, `capabilities`, `unsupported_capabilities`, `limits` (array of `{resource, maximum}`), `prices` (array of `{component, usd_per_million}`), `region`, `privacy`. One grammar, two transports; the CLI adds no fields of its own. All domain validation (bounds, vocabularies, exclusivities, runtime uniqueness) is owned by `NewRequest` and reported as `invalid_request`.

## Decision wire schema

On a completed route the CLI writes the canonical-v1 encoding of the `Decision` value as returned by `Engine.Route` — the Go struct's own JSON field set (`request_id`, `selected`, `alternatives`, `rejections`, `snapshot_hash`, `policy`, `runtime_generation`, `warnings`) — to stdout, exactly one trailing newline. Encoding rule for the nullable member: `selected` is omitted exactly when `Route` returns no selection; every other member is always present; empty arrays encode as `[]` and never as `null`, so the exit-3 decision never contains a `null` value. The CLI projects, never edits: warnings, rejections and alternatives pass through verbatim. When `Route` returns `no_admissible_offer`, the decision still carries the full rejection set; the CLI writes that decision to stdout unchanged and exits 3.

## Errors and exit codes

Exit codes are closed and stable. 0: route completed with a selection (decision on stdout). 3: route completed with no admissible offer (decision on stdout). 2: usage error or `invalid_request` / `invalid_request_wire` (the wire-only code for a malformed request document; it is owned by this adapter and never raised by the Go API). 4: snapshot adoption failure (`invalid_snapshot`, `schema_incompatible`, `snapshot_unavailable`) — for `route`, an unreadable or missing artifact file is `snapshot_unavailable`. 5: unexpected internal error. Every failure exit (2, 4 or 5) writes exactly one line to stderr: `agent-routing: <code>: <safe detail>`, where safe detail carries no source bytes, no task content and no credentials; stdout stays empty for exits 2, 4 and 5. Usage errors use the single usage line described above. Exit 0 and exit 3 never write to stderr.

## Determinism and isolation

Same request bytes plus same artifact bytes produce byte-identical stdout and the same exit code on every run. `route` and `snapshot check` perform no network access, read no wall clock (the request carries `effective_at`) and read no environment beyond flags. `snapshot adopt` additionally writes the validated local generation and atomic pointer. These commands remain replayable offline. `receipt record` is the explicit exception: it resolves collector configuration and credentials, reads the submission clock and connects to Postgres. Its persistence and events are owned by the receipt port, not the routing API.

## Verification cases

Before the CLI slice is accepted: golden tests replaying fixed request and artifact fixtures produce byte-identical stdout and stable exit codes across runs; the exit-code table holds for each error class, including a no-admissible-offer fixture whose stdout decision lists the full rejection set; an unknown flag, an unknown JSON member and a truncated artifact each produce their specified exit code with a one-line stderr message; `snapshot adopt` on a corrupted artifact retains the previous generation; and a network-disabled run of `route`, `snapshot check` and `snapshot adopt` succeeds unchanged. `receipt record` requires its configured Postgres connection and fails safely when unavailable. Before the receipt slice is accepted: a receipt document with an unknown member is rejected, an idempotent replay of identical bytes returns the same observation, and the same receipt ID with different bytes conflicts.
