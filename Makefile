BINARY  := abitool
GO      := go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)

.DEFAULT_GOAL := help

.PHONY: help build test lint

help: ## List available tasks
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GO) build -ldflags="$(LDFLAGS)" -o ./bin/$(BINARY) ./cmd/

test: ## Run tests
	$(GO) test ./...

lint: ## Run linter (golangci-lint)
	golangci-lint run ./...
