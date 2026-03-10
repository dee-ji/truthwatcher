# test/

Repository-level test organization notes.

## Test placement
- Unit tests: colocated as `*_test.go` beside package code.
- Integration fixtures: `test/fixtures/` and `examples/`.
- Cross-package behavior checks: keep in owning package unless they require external processes.

## Current fixture focus
- `test/fixtures/intents/`: seed intent examples for parser/validation flows.

## TODO
- TODO(truthwatcher): add e2e smoke tests that exercise compose services and API workflows.
