# Phase 9 qualification review (2026-10-01)

Two independent review rounds over the schema-5 qualification layer (spec §12, RunGate7, derived tables).

Round 1 verdict: FAIL. Blockers: (1) gate 7 checked only the count of settings rows, not membership in the twelve common components — a renamed component passed; (2) docs/research/qualification.md miscounted token monetary interpretations as 50 (46 monetary + 4 thresholds). Also flagged: changelog claimed a §5 gate-7 entry that did not exist; the joined-model-to-unjoined flip tamper passed; measurement_settings kinds string mis-typed the knowledge column.

Round 2 verdict: PASS. All fixes verified against the diff and empirically (rename and flip tampers re-executed against a copy of the real artifact; both now fail gate 7; two fresh rebuilds reproduce the published sqlite hash fd5cd174... byte-identically; build, vet, full research tests and make check green). Recorded non-blocking notes: limit direction/operation/accounting vocabularies and the dead token_other_monetary enum value live in code only; gate 7 coverage-checks binding observed_at/basis and limit content rather than recomputing them, as §12.6 specifies.

Review transcripts: /tmp/ar-phase9/review.md (round 1), /tmp/ar-phase9/review2.md (round 2).
