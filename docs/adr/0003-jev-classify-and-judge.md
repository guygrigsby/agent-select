# 3. Jev classifies tasks and judges outcomes, outside Rank

Status: Accepted
Date: 2026-10-13

## Context

`Rank` is only as good as its inputs. Hosts hand-write the task name, needed capabilities and token estimates, and `History` learns only from outcomes a host records, which few do. Both gaps need a reading of natural language: what a prompt asks for, and whether a result did it. Jev (TypeSafe) answers those as typed judgments with probabilities. The first design reserved Jev as a v2 judge (`ar-fs6`); that went to the `archive` branch with the evidence pipeline.

## Decision

A `jev` package beside `route`, stdlib only, with two calls:

- `Classify(prompt, Taxonomy)` fills a `route.Task`'s name, needs and token sizes in one request: a Choice over the host's task names plus `none`, a Noul per capability tag, a Score over four answer sizes. A task choice below `MinConfidence` (default 0.5) or `none` leaves the name unknown. A need is required at even odds, since missing one routes to a model that cannot do the work. Output tokens are the size Jev is 90% sure the answer fits; input tokens are estimated in code.
- `Judge(Run)` asks one Noul: did the result do what was asked. At 0.8 or above it passes, at 0.2 or below it fails, between it records nothing.

The CLI exposes both (`classify`, `judge`) for hosts that shell out. `route.Outcome` gains an optional `judge` naming the model behind a verdict; empty means mechanical. `Rank` stays pure: hosts call Jev before ranking and after running.

## Consequences

- Jev verdicts and mechanical verdicts count alike in `History`. A host that trusts only mechanical checks keeps judged outcomes in a separate log, or does not call `judge`.
- The judge reads the run's result text, so a result that argues for its own success can sway it. Hosts should pass the observable result (diff, test output), not the agent's summary.
- Thresholds are starting points, not calibrated. Calibrate the judge against mechanical verdicts where a host has both.
- Every classify and judge is a network call with cost and latency; failures exit 1 and the host routes as before.
