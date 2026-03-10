.PHONY: help build test lint fmt run-spanreed run-squire run-api run-worker migrate-up migrate-down compose-up compose-down openapi fixtures

help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'

build: ## Build all binaries and packages
	go build ./...

test: ## Run unit/integration tests in-repo
	go test ./...

lint: ## Run lightweight static checks
	go vet ./...

fmt: ## Format Go sources
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

run-spanreed: ## Run Spanreed API service (default :8080)
	go run ./cmd/spanreed

run-squire: ## Run Squire worker service
	go run ./cmd/squire

# TODO(truthwatcher): remove once all scripts use run-spanreed.
run-api: run-spanreed ## Compatibility alias for run-spanreed

# TODO(truthwatcher): remove once all scripts use run-squire.
run-worker: run-squire ## Compatibility alias for run-squire

migrate-up: ## Run migration scaffold command (conceptual placeholder)
	go run ./cmd/tw-migrate up

migrate-down: ## Run migration rollback scaffold command (conceptual placeholder)
	go run ./cmd/tw-migrate down

compose-up: ## Start local Postgres/Redis + Spanreed/Squire containers
	docker compose up -d --build

compose-down: ## Stop local compose stack and remove volumes
	docker compose down -v

openapi: ## Validate Spanreed OpenAPI document exists
	@test -f openapi/truthwatcher.yaml

fixtures: ## List example fixtures
	@find examples -type f
