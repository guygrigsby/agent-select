# 5. Mechanical host outcomes before model judging

Status: Accepted

## Context

Host adapter draft accepted model-supplied usage and verdicts, never invoked the native spawner and lost failed receipt submissions. Those receipts cannot support routing feedback. User chose an algorithmic judge for v1 and Jev for v2.

## Decision

Reuse pi's native subagent RPC behind one routed execution tool. Pass the exact registry Model object so agent-file defaults and fuzzy string resolution cannot replace the decision. Keep the Go algorithm as routing authority. Host derives attempt facts and runs only operator-configured mechanical validators. A model cannot write receipts, claim validated success or select validator commands. Without a mechanical validator, provider completion is unvalidated.

Filter native registry candidates by current authentication and explicit operator entitlement. Omit unqualified native pricing rather than mistake synthesized zero for a free quote. Persist secret-free attempt intents and canonical completion receipts before submission; retries use identical bytes and a stable attempt ID. Hold a native crash-released OS lock across execution and completion so another host cannot recover a live attempt. Delivery remains the existing Postgres receipt port over ADR 0004's one-shot boundary.

Jev and other model judges remain disabled until a separate v2 contract defines their provenance, failure semantics and calibration eligibility. No new evaluation campaign is authorized.

## Consequences

Trusted mechanical checks are narrower than semantic judges. Unsupported tasks remain unvalidated instead of gaining a fabricated success rate. Native spawning is reused rather than reimplemented. Image execution, GPU admission and native workflow interception stay blocked until their host paths are supported and verified. Strict spend admission still requires qualified prices; no adapter claim turns an estimate into a cumulative-spend guarantee.
