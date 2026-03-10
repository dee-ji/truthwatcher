# openapi/

OpenAPI specification for Spanreed HTTP API contracts.

- Primary document: `openapi/truthwatcher.yaml`
- Scope: implemented endpoints from `internal/apihttp/server.go`

## Contributor rule
When adding/changing endpoints:
1. update handler behavior,
2. update OpenAPI paths/request examples,
3. update example payload docs in `docs/examples/api-payloads.md`.

## TODO
- TODO(truthwatcher): add reusable component schemas and response envelopes once API stabilization starts.
