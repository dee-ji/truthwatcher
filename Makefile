.PHONY: fmt test lint build-ui build run

BINARY := truthwatcher
GOCACHE_DIR ?= $(CURDIR)/.gocache
GO := GOCACHE=$(GOCACHE_DIR) go
GO_FILES := $(shell find . -type f -name '*.go' -not -path './.git/*')
GO_PACKAGES := $(shell GOCACHE=$(GOCACHE_DIR) go list ./... 2>/dev/null)

fmt:
	@if [ -n "$(GO_FILES)" ]; then gofmt -w $(GO_FILES); else echo "no Go files to format"; fi

test:
	@if [ -n "$(GO_PACKAGES)" ]; then $(GO) test $(GO_PACKAGES); else echo "no Go packages to test"; fi

lint:
	@if [ -n "$(GO_PACKAGES)" ]; then $(GO) vet $(GO_PACKAGES); else echo "no Go packages to vet"; fi

build-ui:
	@test -f web/index.html
	@test -f web/assets/app.css
	@test -f web/assets/app.js

build: build-ui
	@mkdir -p bin
	@$(GO) build -o bin/$(BINARY) ./cmd/truthwatcher

run:
	@$(GO) run ./cmd/truthwatcher $(ARGS)
