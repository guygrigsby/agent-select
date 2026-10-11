# agent-select

Small, pure Go library that ranks candidate models. Keep it that way.

- `route/` is the library; `cmd/agent-select` is a thin CLI over it with no logic of its own.
- No dependencies beyond the standard library. No network, clock or hidden state in `Rank`.
- Unknown is never zero or perfect. Hosts supply candidates from their live registry; there is no model list here.
- Quality evidence is a `Scorer` plug-in, not a prerequisite. New evidence sources are new Scorers.
- Decisions go in `docs/adr/` (append-only). `make check` before done. Use `bd` for follow-up work.
