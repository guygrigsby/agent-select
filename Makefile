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

# Release archives for every platform, named by tag. The release workflow
# runs this on a pushed v* tag and uploads dist/.
VERSION ?= $(shell git describe --tags --always)
PLATFORMS := darwin/arm64 darwin/amd64 linux/arm64 linux/amd64

.PHONY: dist
dist:
	rm -rf dist && mkdir -p dist
	for p in $(PLATFORMS); do os=$${p%/*}; arch=$${p#*/}; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/agent-select ./cmd/agent-select && \
		tar -C dist -czf dist/agent-select_$(VERSION)_$${os}_$${arch}.tar.gz agent-select && rm dist/agent-select || exit 1; \
	done
	cd dist && shasum -a 256 *.tar.gz > checksums.txt
