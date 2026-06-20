.PHONY: fmt test lint build-ui build release-local checksums run

BINARY := truthwatcher
GOCACHE_DIR ?= $(CURDIR)/.gocache
GO := GOCACHE=$(GOCACHE_DIR) go
GO_FILES := $(shell find . -type f -name '*.go' -not -path './.git/*')
GO_PACKAGES := $(shell GOCACHE=$(GOCACHE_DIR) go list ./... 2>/dev/null)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
RELEASE_DIR := dist/$(BINARY)-$(GOOS)-$(GOARCH)

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

release-local: test build-ui
	@mkdir -p $(RELEASE_DIR)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 $(GO) build -trimpath -ldflags "-s -w" -o $(RELEASE_DIR)/$(BINARY) ./cmd/truthwatcher
	@cp docs/install.md $(RELEASE_DIR)/
	@printf "local release written to %s\n" "$(RELEASE_DIR)"

checksums:
	@if [ -d dist ]; then find dist -type f ! -name SHA256SUMS.txt -print0 | sort -z | xargs -0 sha256sum > dist/SHA256SUMS.txt; else echo "dist directory not found"; exit 1; fi

run:
	@$(GO) run ./cmd/truthwatcher $(ARGS)
