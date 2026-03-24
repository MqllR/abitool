BINARY := abitool
GO     := go

.DEFAULT_GOAL := help

.PHONY: help build test lint

help: ## List available tasks
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GO) build -o $(BINARY) .

test: ## Run tests
	$(GO) test ./...

lint: ## Run linter (golangci-lint)
	golangci-lint run ./...
