.DEFAULT_GOAL := build
TOOL := bin/agent-routing

.PHONY: build test check lint install redeploy research-rebuild research-coverage research-backlog

build:
	go build -o $(TOOL) ./cmd/agent-routing

test: build
	AR_REQUIRE_POSTGRES=1 go test ./...

# Full artifact gate: input pins, review decisions, rebuilt manifest
# byte-compare, integrity/FK/coverage/status gates on the published DB,
# backlog drift gate.
check: test lint
	$(TOOL) research check
	$(TOOL) research backlog --check docs/research/backlog.md

lint:
	@test -z "$$(gofmt -l .)" || { echo 'Run gofmt -w on:'; gofmt -l .; exit 1; }
	go vet ./...

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
