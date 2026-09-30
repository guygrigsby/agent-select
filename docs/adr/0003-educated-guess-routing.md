# 3. Route using educated estimates

Status: Accepted

## Context

User wants an educated guess about a suitable model, not proof of the best model or a guaranteed success rate. Requiring calibration before useful routing blocked delivery. Published measurements, advertised capabilities and normal host outcomes can inform a choice while remaining uncertain.

## Decision

Rank with versioned heuristic estimates and disclose their evidence, identity assumptions and missing information. Estimates are relative scores, never probabilities. Source-checked measurements may inform the heuristic without satisfying the scientific comparison gate. Mixing evaluation settings adds a warning, not an invented assertion that conditions match. Explicitly different revisions never inherit one another's scores. Models without useful signals remain unscored candidates under an explicit fallback policy.

Keep credentials, hard capability requirements, budgets, privacy, region and device admission as hard constraints. Uncertainty about quality never relaxes them. Use the host's live registry and preserve provider-specific prices. Capture outcomes from existing normal usage only. Do not launch additional paid evaluation campaigns. A controlled campaign or probability calibration is optional future work, not a prerequisite.

`docs/specs/educated-routing.md`, `delivery-storage.md` and `host-integration.md` define this delivery. They supersede the calibrated-quality prerequisite and conceptual service/data surface in earlier production drafts for the implemented slice. They do not alter the research artifact's review ledger, source rights or raw measurements. ADR 0002 still owns the Go language decision.

## Consequences

A decision can be useful without being scientifically certain. Explanations must distinguish estimates, published measurements, publisher claims and observed outcomes. Unknown quality is not zero quality. Conservative cost and capability checks remain independent of heuristic ranking. Exact comparison and calibration lanes can be added later without changing published facts.
