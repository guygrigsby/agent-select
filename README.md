# Agent routing

Research artifact for a future local Go subagent routing library. No router or collection service is implemented. A CLI is a proposed thin testing/early integration adapter that invokes the same Go public API.

Open `data/research.sqlite` for captured catalogs, provider prices, model claims, benchmark measurements, unresolved model labels, source hashes and coverage gaps. Vercel is the baseline; models.dev is cross-referenced by exact offer ID. `make check` validates artifact integrity, evidence hashes and catalog coverage, not model quality.

Only catalog-source research passed review. Benchmark/characteristic batches are quarantined in `review_results`; `accepted_benchmarks` is empty. `benchmark_measurements` and `unresolved_measurements` preserve provisional source values, not approved routing scores. Claims without saved evidence are isolated in `unlocated_claims`. Missing scores do not establish weakness. Raw pricing paths preserve source dimensions; `source_specific_unresolved` units cannot be used for cost admission.

`data/source-manifest.tsv` records byte hashes and recoverable URLs. File timestamps are labelled local mtime, not asserted HTTP retrieval times. JSON files are interchange/source snapshots, not database JSON columns. Rights remain unknown; keep this material local. The MIT license covers original project material only, not third-party evidence or datasets.

Design drafts live in `docs/specs/`. Start with `collection-pipeline.md`. They have not passed the contract gate: image-limit direction, live identity binding derivation and cohort/configuration representation remain blocked in `ar-yl1`. Benchmark coverage and evidence approval remain tracked in `ar-zdr`. Do not begin implementation from these drafts.

Architecture decisions recorded in `docs/adr/`. See [0001-local-routing-and-evidence-collection.md](docs/adr/0001-local-routing-and-evidence-collection.md) (prior Rust direction) and [0002-go-library-and-cli-adapter.md](docs/adr/0002-go-library-and-cli-adapter.md) (current language direction).