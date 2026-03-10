# Local Development

## Fast path
1. `make build`
2. `make test`
3. `make compose-up`
4. `curl localhost:8080/healthz`

## Running services outside compose
- `make run-spanreed`
- `make run-squire`

Compatibility aliases remain:
- `make run-api`
- `make run-worker`

## Migration command status
`make migrate-up` and `make migrate-down` currently exercise scaffold CLI behavior. They are conceptually wired but not yet connected to a real migration engine.
