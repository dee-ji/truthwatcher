GO ?= go
BINS := radiant spanreed highstorm stormlight seekers squire twctl

.PHONY: all build test lint run clean

all: build

build:
	@mkdir -p bin
	@for bin in $(BINS); do \
		$(GO) build -o bin/$$bin ./cmd/$$bin; \
	done

test:
	$(GO) test ./...

lint:
	$(GO) fmt ./...
	$(GO) vet ./...

run:
	$(GO) run ./cmd/twctl

clean:
	rm -rf bin
