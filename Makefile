.DEFAULT_GOAL := build
TOOL := bin/agent-routing

.PHONY: build test check lint install redeploy research-rebuild research-coverage

build:
	go build -o $(TOOL) ./cmd/agent-routing

test: build
	go test ./...

# Full artifact gate: input pins, review decisions, rebuilt manifest
# byte-compare, integrity/FK/coverage/status gates on the published DB.
check: test lint
	$(TOOL) research check

lint:
	@test -z "$$(gofmt -l internal cmd)" || { echo 'Run gofmt -w on:'; gofmt -l internal cmd; exit 1; }
	go vet ./...

# Rebuild the research artifact set from durable inputs (staging, gates,
# backup, journaled publish). Run after source review decisions change.
research-rebuild: build
	$(TOOL) research rebuild

research-coverage: build
	$(TOOL) research coverage

install redeploy:
	@echo 'Research phase only. Router and collector are not implemented.'
	@exit 1
