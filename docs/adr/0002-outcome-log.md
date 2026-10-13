# 2. Append-only outcome log for CLI hosts

Status: Accepted
Date: 2026-10-13

## Context

`History` lives in memory. A Go host holds one and persists outcomes its own way (autophage uses Postgres). A host that shells out to `agent-select` (the pi extension) runs one process per call, so the CLI must load outcomes from somewhere and accept new ones. `History` already marshals its counts as JSON, but recording into that snapshot is read, modify, write: two concurrent pi sessions lose an update unless they lock.

## Decision

The CLI keeps outcomes in an append-only JSON Lines log, one `route.Outcome` per line. `record` appends one line in a single write to a file opened `O_APPEND`, so concurrent appenders do not interleave and need no lock. `rank --history FILE` replays the log into a `History` (`--min-runs`, default 5) and chains it ahead of the request's `scores` table: a mechanical verdict on enough runs beats an operator's opinion. A missing log is no history; malformed lines are skipped with a warning, so a torn final write costs one outcome, not the log. The format lives in `route` (`Outcome`, `AppendOutcome`, `History.ReadLog`) so the CLI holds no logic of its own.

## Consequences

- Replay is linear in the log. Fine at hundreds of outcomes a day; compaction (rewrite as counts) is the answer if it ever is not.
- The log is auditable and editable by hand: delete a line to retract a verdict.
- History and table scores are on different scales; first-hit `Chain` means a model with history is compared on its pass rate against others' declared scores. Operators should declare table scores as expected pass rates.
