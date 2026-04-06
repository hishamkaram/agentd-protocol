# agentd-protocol Makefile — shared wire protocol types

GO := go

.PHONY: build test vet lint clean help

## build: compile all packages
build:
	$(GO) build ./...

## test: run all tests with race detector
test:
	$(GO) test -race -count=1 -p 4 ./...

## vet: static analysis
vet:
	$(GO) vet ./...

## lint: run golangci-lint
lint:
	@if command -v golangci-lint > /dev/null; then golangci-lint run ./...; else echo "golangci-lint not installed, skipping"; fi

## clean: remove caches
clean:
	go clean -cache -testcache

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
