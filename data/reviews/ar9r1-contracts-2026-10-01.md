# ar-9r1 contract normalization review (2 rounds, spec-only)

Reviewer: fresh-context GLM-5.3-flash via aperture-responses, round 1 2026-10-01, round 2 same day. Scope: uncommitted spec changes at HEAD e41ea39 (delivery-storage.md, host-integration.md, new cli-adapter.md).

Round 1 verdict: FAIL. Blockers B1 (decision wire nil-selected vs canonical-v1 no-null vs all-members-present) and B2 (spawn-handoff deadline_unix_ms carried a duration as an epoch instant); notes N1-N8. Both blockers and N1-N7 fixed in the working tree before round 2.

Round 2 verdict: PASS. All changes-since-R1 verified closed against routing/route.go, routing/decision.go, routing/request.go and routing/dto.go; at-risk round-1 checks re-run and still hold. Non-blocking notes: N-R2-1 (educated-routing blanket presence rule needed the nil-selected omission clause, applied before commit), N-R2-2 (cosmetic boolean annotations), N-R2-3 (trigger UPDATE arm is belt-and-suspenders given append-only outcome_facts).

Basis: full-text read of all three changed files plus educated-routing.md and contracts.md; wire field fidelity checked member-by-member against Go JSON tags; all cross-reference anchors resolved; preserved constraints verified intact (calibration lane untouched, ordinary-usage-only capture, research-artifact contracts unchanged, no JSON columns, no hardcoded model lists).
