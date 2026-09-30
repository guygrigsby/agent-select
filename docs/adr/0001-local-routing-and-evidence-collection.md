# 1. Separate local routing from evidence collection

Status: Proposed

## Context

User requested a fast Rust subagent router, Vercel discovery, models.dev pricing cross-reference, automated characteristics collection and text/image support. Research data may live in SQLite in this repository. Routing cannot depend on cloud lookup, an inference call or a live database. Published evidence is incomplete and evaluation configurations differ.

## Decision

Embed Rust routing against an immutable local snapshot and intersect its offers with the host's live registry. Keep collection and evidence reconciliation outside the request path. Use SQLite only for research and Postgres for future collection authority. Preserve byte-exact source evidence and normalized relational facts. Model revisions, mutable aliases and provider offers remain distinct.

Use deterministic collectors first. Allow bounded change-triggered extraction for unstructured sources, with exact source-span validation and quarantine. Derive characteristics from applicable evidence rather than maintaining hand-authored model rankings. Track missingness, rights, identity ambiguity and per-field freshness. Publish a successor snapshot only after deterministic validation.

Keep source adapters behind anti-corruption boundaries. See [domain model](../specs/domain-model.md), [pipeline](../specs/collection-pipeline.md) and [router](../specs/router.md).

## Consequences

Warm routes have no network or datastore dependency. New models appear through source events and live host registration, not code edits. Incomplete or stale evidence can reduce eligible candidates and cannot be hidden by guessed scores. Aggregator licensing, quality calibration and remote snapshot trust remain decisions before those lanes are enabled.

Rust is a user choice; its speed advantage is unmeasured. Starting with production SQLite would turn a temporary research exception into architecture. A Go daemon template does not fit the embedded Rust hot-path boundary. No daemon or CLI-client service is being implemented in this phase.
