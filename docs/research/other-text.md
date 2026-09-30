# Other-text research

## Corrected scope

All 55 assigned Vercel model IDs retain model-specific rows. Numeric evidence is limited to saved provider/model-card tables for StepFun Step 3.5 Flash, Step 3.7 Flash, and Arcee Trinity-Large-Thinking; missingness is not upgraded from adjacent revisions.

Step 3.5 values are explicitly from the provider README's **Detailed Benchmarks** table and exact **Step 3.5 Flash** column: **AIME 2025** 97.3 (Reasoning), **SWE-bench Verified** 74.4, and **Terminal-Bench 2.0** 51.0 (Coding). The source notes reproduced/asterisked comparisons; exact harness, shots, and tool settings are not stated, so they remain unknown. Trinity values are from its exact Trinity-Large-Thinking column; the README says all models were evaluated in `mini-swe-agent-v2`, with the SWE-bench row marked `*`.

All numeric evidence is provider-reported. Source rights are unknown; repository source paths are retained in findings. Catalog-only and unavailable models remain explicit missingness.

## Evidence repair note
The Step 3.5 Flash provider README has three exact-model rows (AIME 2025 97.3; SWE-bench Verified 74.4; Terminal-Bench 2.0 51.0). Its model note now reflects those rows and explicitly limits the claim; harness, shots, and tools remain unstated. The Step 3.7 benchmark source-map entry was corrected to the saved Step 3.7 Hugging Face page provenance; catalog/source gaps remain gaps.
