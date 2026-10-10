# Pi host adapter contract review

Pass 1 returned FAIL before implementation. Reviewed installed pi 1.1.0 and pi-subagents 0.20.0 against `docs/specs/pi-host-adapter.md`, ADR 0005 and the existing delivery contracts. No tests or implementation changes came from this review.

Confirmed contract gaps:

- Native `Agent` tool frontmatter could override model, background mode, turn limit and isolation. Comparing its model argument did not prove exact spawning.
- Startup recovery could classify another process's live intent as abandoned. Completion needed atomic no-replace publication and acknowledgement needed intent-first durable cleanup.
- Native workflow spawning bypassed the Agent tool hook. Unsupported spawning paths needed explicit model-facing refusal.
- Same-UID tools could modify operator files. Filesystem permissions did not establish a sandbox against the owning OS user.
- Terminal outcomes, safe error content, closed operator configuration and response/error fields were incomplete.
- V1 needed explicit scope for refused images, CURRENT reads and receipt submission's network/config/clock exception.

Pass 2 changes the handoff to documented `subagents:rpc:spawn` with an exact native Model object and explicit turn limit. Its installed runner prioritizes those options over agent defaults. Isolated children cannot delegate. Model-facing Agent, workflow and steering bypasses are refused. Attempt execution, recovery and delivery share a crash-released OS lock; completion publication is no-replace and acknowledgement removes intent before completion. Config, terminal outcomes and the same-UID limitation are explicit. Mechanical validators remain v1; Jev remains disabled v2.

Pass 2 returned FAIL. [Full verdict](pi-host-contract-pass2-2026-10-10.txt) retained. Reviewer ran no tests. Parent confirmed these findings against installed pi 1.1.0 and pi-subagents 0.20.0:

- `agent-manager.ts` invokes `onStart` before `onSpawned`; the callback does not bind identity before the started event. Pre-reply events need correlation buffering. An already-aborted queued spawn can become stopped without a terminal event, so cancellation needs an independent settlement path.
- Native background notifications include raw `record.error`. Safe wrapper responses do not suppress them. Owned terminal events must synchronously consume the native notification before any validation or delivery await.
- RPC spend accumulates in `PendingUsagePool`. Only native usage-reporting tool results drain it into parent accounting. The wrapper is not automatically wrapped; attaching correlated usage independently risks later double counting.
- Native started precedes `runAgent` and session startup. The proposed `startup_millis` measures admission or queue delay, not actual startup.

Status at pass 2: host contract and implementation stopped after two failures. Further contract repair needed user authorization. No replacement host code or receipt-lock implementation was started.

User then authorized one additional bounded repair and review. Contract pass 3 buffers early lifecycle events, binds identity through queue/spawn callbacks and the reply, settles silent cancellation, consumes owned notifications synchronously and records first-settlement receipts without waiting forever. Parent accounting is explicitly not integrated, tracked in `ar-r8z`; admission timing is no longer called startup. Host execution remains an adapter over the existing native API with no new service or Knowledge schema.

Pass 3 review `ff64f471-921e-407` pending. Implementation remains gated; another failure stops this task again.
