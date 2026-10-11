.DEFAULT_GOAL := build

.PHONY: build test check lint install redeploy

build:
	go build -o bin/agent-select ./cmd/agent-select

test:
	go test -race ./...

lint:
	@test -z "$$(gofmt -l .)" || { echo 'Run gofmt -w on:'; gofmt -l .; exit 1; }
	go vet ./...

check: lint test

install:
	go install ./cmd/agent-select

# A library and a one-shot CLI: nothing runs to redeploy.
redeploy: install
