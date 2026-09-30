# DeepSeek / ZAI research

## Coverage
Catalog baseline rows cover every assigned model. Saved primary evidence inspected includes the DeepSeek-R1 and DeepSeek-V3 arXiv pages and their repository metadata. Later revisions and GLM variants were not assigned predecessor scores.

## Findings
No numeric benchmark row is included unless the saved source directly identifies the exact Vercel model and exposes a value with a reliable locator. Catalog descriptions are retained only as provider/catalog claims; context windows are operational limits, not quality measurements.

## Automation path
`findings.json` is generated/consumed as interchange JSON. Raw evidence is preserved under `data/research/deepseek-zai/saved_sources/`; future extraction should parse table rows, verify model identity and configuration, then add one benchmark object per exact variant.

## Limitations
The checked sources do not establish exact-model numeric results for the later DeepSeek revisions or Zhipu GLM releases. Unknown future revisions remain unresearched.
