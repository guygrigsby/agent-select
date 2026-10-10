# 4. One-shot process boundary for host routing

Status: Accepted

## Context

The host (pi, TypeScript) needs routing decisions from the Go domain API. ADR 0002 forbids a second ranker, so the boundary must cross a process. The implementation plan required measuring one-shot process overhead before choosing any persistent transport, and no daemon or HTTP API is implied by the CLI.

## Decision

Host routing crosses as a one-shot `agent-routing route` subprocess per decision: canonical request bytes in, canonical decision bytes out, closed exit-code table (`docs/specs/cli-adapter.md`). Measured 2026-10-01 on the development machine: about 6.9 ms per one-shot route (N=50, process startup plus artifact read plus decode dominating) versus about 0.001 ms for a warm in-process `Engine.Route` (N=100k). The gap is pure process boundary. 6.9 ms is acceptable next to a subagent spawn that costs seconds, so no persistent transport, long-lived subprocess or daemon is introduced. Any future workload routing many times per second gets its own ADR before a transport change.

## Consequences

- Routing latency budget is ~7 ms per decision; startup dominates it, so micro-optimizing the domain is pointless until the boundary changes.
- No protocol state to negotiate or version: stdin/file in, stdout bytes, exit codes and the one-line stderr error form are the whole contract.
- Cancellation is process termination; retries, timeouts and spawn admission stay host-owned per `docs/specs/host-integration.md`.
- Snapshot handoff is the filesystem store: the host adapter resolves the store's CURRENT pointer to `snapshots/<hex>.bin` and passes that path to each invocation; adoption happens through `agent-routing snapshot adopt`.
- Outcome capture crosses the same way: `agent-routing receipt record` against the collector DSN.
