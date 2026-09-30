# Catalog source evidence

Collected 2026-09-30. Request URLs recorded by the orchestrator:

- `https://ai-gateway.vercel.sh/v1/models`, retrieved about 01:33 UTC. Body: `data/raw/vercel-models.json`; headers: `data/raw/vercel.headers`.
- `https://models.dev/api.json`, retrieved about 01:33 UTC. Body: `data/raw/models-dev.json`; headers: `data/raw/models-dev.headers`.

The saved Vercel headers show HTTP/2 200, `server: Vercel`, `x-matched-path: /v1/models`, `content-type: application/json`, and no ETag or Last-Modified. They do not contain the request URL. The saved models.dev headers show HTTP/2 200, `cache-control: public, max-age=0, must-revalidate`, and `etag: "a174062643509bd51dc6338a3e8bb26a"`.

Minimal verification GETs on 2026-09-30 about 01:50 UTC with User-Agent `agent-routing-research/0.1`:

- `https://ai-sdk.dev/providers/ai-sdk-providers/ai-gateway`: 200, Vercel-served documentation.
- `https://github.com/vercel/ai/tree/5eb85cc45a259553501f535b8ac79a77d0e79223/packages/gateway`: 200, GitHub page for the pinned Vercel AI Gateway package path.
- `https://raw.githubusercontent.com/vercel-labs/models.dev/main/LICENSE`: 404. No license conclusion is drawn from that failed lookup or from the JSON structures.
