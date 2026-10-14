# agent-select

Rank the models you can run for the task at hand. A Go library plus a thin CLI.

```go
cands := []route.Candidate{
    {ID: "qwen-27b", Caps: []string{"tools"}, ContextTokens: 131072, InputUSDPerMTok: ptr(0.0), OutputUSDPerMTok: ptr(0.0)},
    {ID: "sonnet", Caps: []string{"tools"}, ContextTokens: 1_000_000, InputUSDPerMTok: ptr(2.0), OutputUSDPerMTok: ptr(10.0)},
}
res := route.Rank(cands, route.Task{Name: "code", Mode: route.Quality, Need: []string{"tools"},
    InputTokens: 100_000, OutputTokens: 8_000}, route.Table{"sonnet": {"code": 0.9}, "qwen-27b": {"code": 0.5}})
best, ok := res.Best()
```

- You supply candidates (what you can run right now), a task (needs and bounds) and optionally a `Scorer`.
- Hard constraints reject with a reason code: missing capability, context or output too small or unknown, over budget, too slow, unscored, below minimum score.
- Order is cost or quality first, then the other, then latency, then ID. Unknown values rank after known ones and never count as free or perfect.
- Scorers: `Table` (operator-declared, per model and task with `*` wildcards), `History` (smoothed pass rate from outcomes you record, persists as JSON) and `Chain` (first opinion wins). Implement `Scorer` to plug in benchmarks or anything else.

`agent-select rank < request.json` does the same from the shell; see `cmd/agent-select`. Hosts in other languages keep outcomes in an append-only log: `agent-select record --history FILE --task T --id ID --verdict pass|fail` appends one, and `rank --history FILE` scores from it ahead of the request's table ([ADR 0002](docs/adr/0002-outcome-log.md)).

Package `jev` fills the inputs a host cannot compute: `Classify` turns a prompt into a task (name, needs, token sizes) from the host's own task and capability vocabulary, and `Judge` grades a finished run so `History` has outcomes to learn from. Both call TypeSafe's Jev; `Rank` never does. From the shell: `agent-select classify --taxonomy FILE < prompt` and `agent-select judge --history FILE --task T --id ID < run.json`, with `TYPESAFE_URL` pointing at a gateway that serves `/v1/systemone` (Aperture: `http://ai/v1/systemone`, no key) or `TYPESAFE_API_KEY` for TypeSafe directly ([ADR 0003](docs/adr/0003-jev-classify-and-judge.md)).

The earlier evidence-pipeline design (snapshots, Postgres collector, research artifact) lives on the `archive` branch. See [ADR 0001](docs/adr/0001-lean-core.md).

MIT licensed.
