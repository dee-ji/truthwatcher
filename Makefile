.PHONY: help build test lint fmt run-spanreed run-squire run-api run-worker migrate-up migrate-down compose-up compose-down openapi fixtures

help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'

build: ## Build all binaries
	go build ./...

test: ## Run tests
	go test ./...

lint: ## Run lightweight lint
	go vet ./...

fmt: ## Format code
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

run-spanreed: ## Run Spanreed API service
	go run ./cmd/spanreed

run-squire: ## Run Squire worker service
	go run ./cmd/squire

# TODO: Remove compatibility aliases once local scripts migrate.
run-api: run-spanreed ## Backward-compatible alias for run-spanreed

# TODO: Remove compatibility aliases once local scripts migrate.
run-worker: run-squire ## Backward-compatible alias for run-squire

migrate-up: ## Apply migrations (placeholder)
	go run ./cmd/tw-migrate up

migrate-down: ## Rollback migrations (placeholder)
	go run ./cmd/tw-migrate down

compose-up: ## Start local stack
	docker compose up -d --build

compose-down: ## Stop local stack
	docker compose down -v

openapi: ## Validate OpenAPI exists
	@test -f openapi/truthwatcher.yaml

fixtures: ## show fixtures
	@find examples -type f
