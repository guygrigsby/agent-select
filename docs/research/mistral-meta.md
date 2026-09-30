# Mistral / Meta Research Summary

## Overview

This document covers research findings for 19 models assigned under the `data/research/mistral-meta/` directory: 11 Meta models and 8 Mistral models from the Vercel model catalog.

## Source Findings

### Meta Models with Published Benchmarks

**Llama 3.1 (70B & 8B)**: Extensive benchmark tables sourced from Meta's official GitHub model card (`meta_llama_eval_github.html`). Includes base pretrained and instruction-tuned evaluations across general knowledge (MMLU, MMLU-Pro), reasoning (GPQA, ARC-C), code (HumanEval, MBPP++), math (GSM-8K, MATH), tool use (BFCL, API-Bank, Nexus), and multilingual (MGSM). All scores are provider-reported using Meta's internal evaluation library.

**Llama 3.3 (70B)**: Benchmark tables from `meta_llama3_3_github.html` showing improvements over Llama 3.1 70B on IFEval (92.1), HumanEval (88.4), MATH (77.0), and MGSM (91.1). The model card compares directly against Llama 3.1 70B and 405B.

**Llama 4 (Maverick & Scout)**: Comprehensive benchmark tables from `meta_llama4_github.html` covering both pretrained and instruction-tuned evaluations. Includes vision benchmarks (MMMU, ChartQA, DocVQA, MathVista), long context (MTOB), and LiveCodeBench. Both models are natively multimodal MoE architectures.

### Mistral Models with Published Benchmarks

**Mistral Nemo**: Benchmark table from `hf_mistral_nemo.html` (Hugging Face model card) with 8 standard benchmarks plus multilingual MMLU across 8 languages. Trained jointly with NVIDIA, Apache 2.0 license.

**Ministral 3B & 8B**: Benchmarks from `hf_ministral_8b.html` (Hugging Face model card). Both base and instruct model benchmarks across knowledge/comprehension, code/math, multilingual, and chat quality (MTBench, Arena Hard). The model card covers both sizes.

**Mistral Small 3.1 (24B)**: Extensive benchmarks from `hf_mistral_small3.html` with text, vision, multilingual, and long-context evaluations. Used as representative for `mistral/mistral-small` (latest) - note potential version mismatch.

### Models Without Published Benchmarks

**Muse Spark models (1.1, 1.2, 1.3) and Muse Glimmer 30B**: These are very recent Meta models with no published benchmark tables found. Catalog descriptions indicate multimodal capabilities (text, image, video, audio, PDF) and agentic workflows but no numeric performance data is yet publicly available.

**Mistral Large 3**: The Hugging Face model card has benchmark section headers but no populated tables. Blog pages are JavaScript-rendered. No numeric scores extracted.

**Mistral Codestral**: Hugging Face model card is gated (401). Blog content rendered via JavaScript. No benchmarks extracted from static sources.

**Mistral Medium 3.5**: No accessible model card or benchmark data found.

**Ministral 14B**: Hugging Face model card is gated (401).

## Automation Path

1. **Meta models**: GitHub raw model cards (`raw.githubusercontent.com/meta-llama/llama-models/main/models/`) provided the most reliable and complete data. Hugging Face model cards require authentication for gated access.
2. **Mistral models**: Hugging Face model cards (`huggingface.co/mistralai/*/raw/main/README.md`) were the primary source. Some are gated (Codestral, Ministral 14B). Blog pages at `mistral.ai/news/` are JavaScript-rendered and not suitable for static extraction.
3. **Benchmark tables** were extracted as markdown tables from the README files (Meta) or HTML tables (Meta GitHub cards). Values were verified against the raw source text.

## Limitations

- **Gated access**: Meta's Hugging Face models and some Mistral models (Codestral, Ministral 14B) require authentication. GitHub mirrors of model cards were used for Meta models.
- **JavaScript-rendered content**: Mistral AI and Meta blog posts deliver benchmark data dynamically. Static curl downloads capture only the shell HTML.
- **Model version ambiguity**: `mistral/mistral-small` maps to "Mistral Small (latest)" in the catalog. Benchmarks are from Mistral Small 3.1 (2503); the actual latest model on Vercel may differ.
- **Muse models**: No benchmarks exist in any checked source. These are frontier models with proprietary performance claims only.
- **Mistral Large 3**: Model card has placeholder sections for benchmarks that were never populated with actual data.

## Coverage

| Model | Benchmarks | Strengths | Weaknesses |
|-------|-----------|-----------|------------|
| meta/llama-3.1-70b | Found (15+ metrics) | From model card | Operational (text-only) |
| meta/llama-3.1-8b | Found (15+ metrics) | From model card | Operational (text-only) |
| meta/llama-3.3-70b | Found (8+ metrics) | From model card | Operational (text-only) |
| meta/llama-4-maverick | Found (15+ metrics) | From model card | Operational (text+image output only) |
| meta/llama-4-scout | Found (15+ metrics) | From model card | Operational (text+image output only) |
| meta/muse-glimmer-30b | Not found | From catalog | None established |
| meta/muse-spark-1.1 | Not found | From catalog | None established |
| meta/muse-spark-1.2 | Not found | From catalog | None established |
| meta/muse-spark-1.2-contributor | Not found | From catalog | None established |
| meta/muse-spark-1.3 | Not found | From catalog | None established |
| meta/muse-spark-1.3-contributor | Not found | From catalog | None established |
| mistral/codestral | Not found | From catalog | Operational (text-only) |
| mistral/ministral-14b | Not found | From catalog | None established |
| mistral/ministral-3b | Found (7+ metrics) | From model card | None established |
| mistral/ministral-8b | Found (7+ metrics) | From model card | None established |
| mistral/mistral-large-3 | Not found | From HF card | Provider stated limitations |
| mistral/mistral-medium-3.5 | Not found | From catalog | None established |
| mistral/mistral-nemo | Found (16 metrics) | From model card | Operational (text-only) |
| mistral/mistral-small | Found (15+ metrics) | From model card | None established |

All evidence classes are `provider_reported` (no independent benchmarks found). Sources are saved under `data/research/mistral-meta/sources/`.
## Provenance correction
Finding source paths are repository-relative (`data/research/mistral-meta/...`). Base and instruct benchmark labels remain distinct where source tables support the distinction. Llama 3.1 70B MBPP++ 86.0 is specifically located in the source's **Base pretrained models** table under the Llama 3.1 70B column; it is not the Instruct value. The Llama 3.3 MBPP EvalPlus (base) row retains an explicit identity gap because the source table context conflicts with its base label; no base/instruct attribution is asserted. Missing/gated source bytes were not reconstructed.
