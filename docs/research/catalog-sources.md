# Catalog and pricing sources

**Collected:** 2026-09-30, about 01:33 UTC
**Evidence:** `data/raw/vercel-models.json`, `data/raw/models-dev.json`, their saved headers, and `data/research/catalog-sources/source-evidence.md`.

## Sources

- Vercel AI Gateway: `https://ai-gateway.vercel.sh/v1/models`. The response is JSON with `data`; saved headers show 200, `server: Vercel`, and `x-matched-path: /v1/models`. The headers do not prove any other URL.
- Models.dev: `https://models.dev/api.json`. The response is provider keyed. The `vercel` provider has 394 models and no `api` field.
- Verified documentation: `https://ai-sdk.dev/providers/ai-sdk-providers/ai-gateway` returned 200. The Vercel provider path pinned in the data was checked at GitHub and returned 200. A lookup of `vercel-labs/models.dev` `main/LICENSE` returned 404. Neither JSON structure contains a license field, so licensing is unknown.

## Structure and joins

The Vercel JSON has 395 records and no `total` field; IDs are unique. Models.dev has 225 top-level providers. Joining `data[].id` to `vercel.models` gives 394 common IDs, one Vercel-only ID (`typesafe-ai/jev`), and no models.dev-only IDs. These are computed counts, not declared totals.

Vercel `pricing.input` and `pricing.output` are decimal strings whose values are per token where applicable. Models.dev `cost.input` and `cost.output` are numbers in dollars per million tokens. Multiply a Vercel token value by 1,000,000 before comparison. Other Vercel pricing dimensions are not interchangeable token prices: image, audio token, speech character, transcription duration, video duration/token, realtime, web search, maps search, and tier or regional fields are represented separately.

## Pricing comparison

For input, 327 Vercel records have `pricing.input`, 286 models.dev records have `cost.input`, and 286 joined records have both. 281 are within 1%; 5 disagree. Output has 287 Vercel records, 286 models.dev records, 286 joined records, 281 within 1%, and 5 disagreements.

The five disagreements, values shown as dollars per million after Vercel conversion, are:

| ID | Vercel input/output | models.dev input/output |
|---|---:|---:|
| `meta/llama-3.3-70b` | 0.72 / 0.72 | 0 / 0 |
| `meta/llama-4-scout` | 0.17 / 0.66 | 0 / 0 |
| `meta/llama-4-maverick` | 0.24 / 0.97 | 0 / 0 |
| `mistral/ministral-3b` | 0.10 / 0.10 | 0.04 / 0.04 |
| `mistral/ministral-8b` | 0.15 / 0.15 | 0.10 / 0.10 |

Zero is explicit in models.dev for the first three; it is not missing cost. Do not infer a cause from these files.

There are 108 joined Vercel IDs with no models.dev `cost`, by Vercel type: video 35, image 28, embedding 26, speech 6, reranking 5, transcription 5, realtime 3. This is not a claim that all such models lack every possible metadata field. There are no models.dev Vercel records with cost but no Vercel pricing counterpart in this join.

## Pricing fields

`pricing.image` occurs on 21 Vercel records. For example, `bfl/flux-kontext-max` has `{"image":"0.08"}`. The observed Vercel pricing key set is:

`audio_input_token_cost`, `audio_output_token_cost`, `fast`, `image`, `image_dimension_quality_pricing`, `input`, `input_cache_read`, `input_cache_read_tiers`, `input_cache_write`, `input_cache_write_tiers`, `input_tiers`, `maps_search`, `output`, `output_tiers`, `peak_pricing`, `realtime_client_message_cost`, `realtime_session_duration_cost_per_second`, `regional`, `service_tiers`, `speech_input_character_cost`, `transcription_duration_cost_per_second`, `varies_by_provider`, `video_duration_pricing`, `video_token_pricing`, `web_search`.

These files do not contain the previously claimed GPT-4o example; use exact IDs and fields from the current bodies when consuming them.

## Collection feasibility

The captured Vercel response has no ETag or Last-Modified, so conditional revalidation was not evidenced for that response. Models.dev supplies an ETag and `max-age=0, must-revalidate`, so conditional revalidation is possible subject to server behavior. Both responses were public successful GETs in the captured evidence. No auth, rate-limit, or licensing conclusion is established beyond those observations.
