# 2. Go library and proposed CLI adapter

Status: Accepted for language direction; CLI role and protocol proposed

## Context

The prior ADR recorded a Rust choice for the local routing boundary. The project now needs a Go library as the primary integration surface. A command-line process may be useful for testing and early integration, including from non-Go harnesses.

## Decision

Use a Go library for the embedded router. The library owns the existing public routing API, ranking and policy behavior. A CLI may provide a thin testing and early integration adapter that invokes that same Go API. It must not duplicate ranking or policy logic. Non-Go harnesses may initially use the proposed process boundary rather than importing Go directly.

The CLI role, commands, flags and wire protocol remain PROPOSED and unspecified until the contract pass. No HTTP server or daemon is implied. Process startup, decoding and encoding are measured separately from warm route latency. Use a stdlib-first Go direction, including `encoding/json` and `sync/atomic` where appropriate, with exact money represented by scaled integers or `math/big`. No dependencies are added and no speed claim is made.

## Consequences

The library remains offline, deterministic and independent of collection. The same API prevents behavior drift between embedded and CLI use. Existing live-registry, snapshot, evidence, quarantine and contract gates remain in force. This ADR replaces only the original Rust language choice and does not approve the remaining proposed architecture or implementation.
