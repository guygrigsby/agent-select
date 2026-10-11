# 1. Lean core with pluggable scorers

Status: Accepted
Date: 2026-10-10

## Context

The first design made evidence the foundation: snapshots with canonical JSON and hashes, source provenance, a Postgres collector, a research artifact and contract reviews before each slice. It never ranked anything in a host, because the foundation (admitted benchmark evidence) was not finished, and the contracts made every change expensive. The first host, autophage, needed a live-registry filter and cost order, then a way to say which model is better at a task.

## Decision

One package, `route`: candidates and a task in, an ordered result with reasons out. Quality comes from a `Scorer` interface. Ship `Table` (operator-declared), `History` (recorded outcomes) and `Chain`. Everything else from the first design moved to the `archive` branch, including the collector and benchmark pipeline, which would return as a `Scorer` if wanted. Cost is a float estimate for ordering and budgets, not billing.

## Consequences

- Breaking: `routing` is gone. Hosts migrate to `route` (v0.2.0).
- No snapshot hash or provenance: a result replays from its inputs, but is not tied to a signed evidence artifact.
- Scores from different Scorers are not comparable, so `Chain` is first-hit, not a blend.
