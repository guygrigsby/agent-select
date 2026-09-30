# Second source-accuracy review

Compared the working-tree artifacts with `git HEAD` and `/tmp/agent-routing-fixes/review-evidence.md`. No edits, network access, or git changes.

## Meta/Mistral — PASS

- Saved source files exist and contain the cited cells.
- Llama 3.1 70B MBPP++ is now correctly tied to the **Base pretrained models** table, the `Llama 3.1 70B` column, and raw `86.0`; it is not mislabeled as Instruct. The separate MMLU base row is likewise identified as pretrained.
- Llama 3.3 MBPP EvalPlus `87.6` retains the source's `(base)` label while explicitly recording that base/instruct identity is unresolved in the surrounding comparison table; it is not forced into an instruct interpretation.
- Model/benchmark/configuration/metric fields preserve the table semantics and provider-reported evidence. Remaining catalog mappings or gaps are outside this review.

## GenEval — PASS

- Exactly 42 measurements are present: six historical rows × seven columns.
- Every value is stored as the exact source lexeme (including `0.00`, `0.01`, etc.), with the model name and metric/column label.
- Locators preserve the table header, selected model row, selected column, and raw cell; the saved README contains the corresponding header and rows.
- Configuration is explicitly unknown where the README does not specify generation settings, evaluator commit, dates, or sample count; the SD v1.5 four-image example is not incorrectly generalized.
- Historical labels remain `identity_status: unresolved` with empty `proposed_offer_ids`; no offer mapping is inferred. CLIP retrieval is explicitly retained as a reference baseline, not an image generator.
- Source metadata distinguishes these as independent benchmark evidence, separate from vendor-reported findings. This is not a coverage, calibration, or eligibility approval.

## Status guard edits — PASS

- Anthropic has no benchmark rows and no model marked `found` without exported numeric evidence.
- MiniMax/Kimi `found` statuses have corresponding exported numeric benchmark rows with saved source files, locators, model identity, metric, and value. Models without such rows remain not-found/gap states.
- This does not assert that no unpublished or unchecked scores exist; it only verifies that `found` is not used without an exported numeric row.
