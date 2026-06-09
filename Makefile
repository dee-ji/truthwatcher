.PHONY: fmt test lint build run

BINARY := truthwatcher
GO_FILES := $(shell find . -type f -name '*.go' -not -path './.git/*')
GO_PACKAGES := $(shell go list ./... 2>/dev/null)

fmt:
	@if [ -n "$(GO_FILES)" ]; then gofmt -w $(GO_FILES); else echo "no Go files to format"; fi

test:
	@if [ -n "$(GO_PACKAGES)" ]; then go test $(GO_PACKAGES); else echo "no Go packages to test"; fi

lint:
	@if [ -n "$(GO_PACKAGES)" ]; then go vet $(GO_PACKAGES); else echo "no Go packages to vet"; fi

build:
	@echo "build target placeholder: cmd/truthwatcher will be added by the project skeleton prompt"

run:
	@echo "run target placeholder: cmd/truthwatcher will be added by the project skeleton prompt"
