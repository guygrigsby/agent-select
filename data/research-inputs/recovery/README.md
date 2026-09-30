# Recovery references

Recovered 2026-09-30 from volatile `/tmp` scratch dirs before they disappeared.
These are read-only references for designing the Go replacement. They are never
build inputs and must never be executed as part of any pipeline.

- `build-db.py` — the 142-line Python exporter that produced `data/research.sqlite`
  schema 3. Known failure modes to replace, not port: deletes the DB before
  rebuilding; reconstructs URLs from temporary fetch logs; records current mtimes;
  repins current hashes for hardcoded approved batches on every run.
- `review-contracts.md` — independent contract review listing the first-slice
  contract blockers (snapshot envelope/hash domain, distribution class in DDL,
  rights pins, fallback ordering, collector work/output identity, event fields,
  capability matrices, benchmark grammar). Phase 7 must resolve each finding.

The remaining `/tmp/ar-qualification` and `/tmp/agent-routing-research` contents
(export logs, fetch logs, audit scratch) stay un-recovered deliberately: they are
session scratch. Any provenance they hold that the ledger needs is recovered into
`data/research-inputs/` as typed inventory records with an explicit recovery
basis during phase 3, never by copying logs wholesale.
