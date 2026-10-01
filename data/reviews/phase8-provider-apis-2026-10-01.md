# Phase 8 provider-apis bounded acquisition review (2026-10-01)

Scope: close the provider-apis source-acquisition task. Audit question: does every
assigned provider family (in-scope text/image catalog owners) have an inventoried
documented URL set?

## Audit

Policy URL inventory was joined against the 41 distinct `catalog_models.owner`
values for in-scope text/image models. 13 owners had no URL whose domain matched
the owner; manual reconciliation showed bfl, mixedbread and spacexai already
covered (blackforestlabs.ai/FLUX model cards, mixedbread.ai and
ai-gateway.vercel.sh, docs.x.ai). True gaps: zai, xiaomi, tencent, meituan,
poolside, inference-net, stealth.

## Acquisitions (curl, UA agent-routing-research/0.2, 1.2s spacing)

| Family | Documented URL | Outcome | License bytes | Posture |
|---|---|---|---|---|
| zai | huggingface.co/zai-org/GLM-5.3 raw README | 200, 14209 B | card declares `license: other`, `license_name: glm-5.3` | unknown x3 (custom license not in allowlist) |
| xiaomi | raw.githubusercontent.com/XiaomiMiMo/MiMo README + LICENSE | 200 both | Apache-2.0 retained | allowed x3 |
| meituan | raw.githubusercontent.com/meituan-longcat/LongCat-2.0 README + LICENSE | 200 both | MIT retained | allowed x3 |
| tencent | api.github.com search org:Tencent-Hunyuan (stars) | 200, 29033 B | none; all NOASSERTION | unknown x3; bounded: no public LLM repo for hy3/hy4-preview/hy-mt2, Hunyuan-MT search returned 0 |
| poolside | huggingface.co/poolside/laguna-s-2.1 raw README (via resolve redirect) | 200, 8650 B | card declares openmdw-1.1 (allowlisted) | allowed x3 |
| inference-net | inference.net/ | 200, 203131 B | none retained | unknown x3 |
| stealth | stealth.ai/ | 200, 219 B placeholder shell | none | unknown x3; no public documentation exists |

## Verdict

Every assigned family now has a source outcome: seven documented URL captures
plus two bounded negative outcomes (tencent hy* has no locatable public model
documentation; stealth publishes none). All license quotes are raw byte
substrings of retained evidence. Permissions follow the conservative posture
rule: no retained permissive license bytes, no `allowed`. Repo/model-card
licenses cover documentation only; benchmark-data rights remain separately
tracked. No guessed endpoints were used; GitHub search API and org repos were
the documented discovery path.

Reviewer: author self-review, byte-verified via `sub()` substring assertions at
record construction; `make check` gates re-run after rebuild.
