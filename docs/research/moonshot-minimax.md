# Moonshot AI / MiniMax research

## Source findings

The saved primary sources provide benchmark tables for MiniMax-M2, Kimi K2 Instruct, Kimi K2.5, and Kimi K3. MiniMax-M2 reports coding/agentic results (including SWE-bench Verified 69.4) and Artificial Analysis-aligned intelligence rows. Kimi K2 reports LiveCodeBench v6 and SWE-bench results; Kimi K2.5 reports reasoning, vision, and agentic results; Kimi K3 reports reasoning, coding, agentic, and multimodal results. These are provider/model-card reported measurements, not independent replications.

Capability claims include MiniMax-M2's 230B total/10B active MoE design and MiniMax-M3's native multimodality, 1M context, and approximately 428B/23B parameter counts. Claims are kept separate from measured benchmark rows in `findings.json`.

## Automation path

The fetching stage saved raw API and README bodies under `data/research/moonshot-minimax/sources/`. Extraction reads those files directly, records the exact source URL/file and a short table locator, and emits the interchange JSON at `data/research/moonshot-minimax/findings.json`. Model IDs come from the assignment and are not inferred from family names.

## Limitations and coverage

All 16 assigned IDs have a model row. No scores were copied across revisions or speed variants. The checked evidence did not contain numeric tables for several later/unfamiliar releases (including M2.1 Lightning, M2.7 variants, K2.6, K2.7 Code variants, and K3 Fast); those gaps remain explicit. Benchmark configurations are only as detailed as the saved source, and independent validation was not established. Future revisions remain unresearched rather than inheriting historical numbers.
