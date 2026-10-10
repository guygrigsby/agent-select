.DEFAULT_GOAL := build
TOOL := bin/agent-routing

.PHONY: build test check lint race fuzz perf install redeploy research-rebuild research-coverage research-backlog

build:
	go build -o $(TOOL) ./cmd/agent-routing

test: build
	AR_REQUIRE_POSTGRES=1 go test ./...

# Full artifact gate: input pins, review decisions, rebuilt manifest
# byte-compare, integrity/FK/coverage/status gates on the published DB,
# backlog drift gate, routing race check and bounded routing fuzz runs.
# The fuzz stage runs via $(MAKE) in the recipe, after the prerequisite
# targets, so a saturating fuzz session never overlaps the test run.
check: test lint race
	$(MAKE) fuzz
	$(TOOL) research check
	$(TOOL) research backlog --check docs/research/backlog.md

lint:
	@test -z "$$(gofmt -l .)" || { echo 'Run gofmt -w on:'; gofmt -l .; exit 1; }
	go vet ./...

# Minimum mandatory concurrency gate: the routing package under the race
# detector. Runs the fuzz seed corpus, the engine adopt/route race tests and
# the concurrent adoption test.
race:
	go test -race ./routing

# Both decoders run concurrently, with two fuzz workers each and a 30s bound.
# Wait for both and preserve either failure. Crashers land in testdata/fuzz.
fuzz:
	@status=0; \
	go test -run '^$$' -fuzz '^FuzzDecodeRequestWire$$' -parallel=2 -fuzztime=30s ./routing & first=$$!; \
	go test -run '^$$' -fuzz '^FuzzDecodeSnapshot$$' -parallel=2 -fuzztime=30s ./routing & second=$$!; \
	wait $$first || status=$$?; \
	wait $$second || status=$$?; \
	exit $$status

# Measured latency percentiles and cold construction costs for Engine.Route
# at 1000/10000 snapshot and runtime offers. Fixed sample counts
# (-benchtime=1x): observed results only, no thresholds asserted.
perf:
	go test ./routing -run '^$$' -bench 'Percentiles' -benchtime=1x

# Rebuild the research artifact set from durable inputs (staging, gates,
# backup, journaled publish). Run after source review decisions change.
research-rebuild: build
	$(TOOL) research rebuild

research-coverage: build
	$(TOOL) research coverage

# Regenerate docs/research/backlog.md from the same records that feed the
# export (phase 5). make check fails if the committed report drifts.
research-backlog: build
	$(TOOL) research backlog --report docs/research/backlog.md

install redeploy:
	@echo 'Research phase only. Router and collector are not implemented.'
	@exit 1
