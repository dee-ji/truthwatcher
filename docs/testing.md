# Truthwatcher Testing Strategy

Truthwatcher tests must support the project safety model: normal test runs must not require real network devices, Docker, Kubernetes, cloud services, message brokers, or write-capable automation.

## Standard Command

Run the normal test suite with:

```bash
make test
```

The `Makefile` sets a repository-local Go build cache by default so tests do not depend on user-specific cache paths.

For a full local check before committing, run:

```bash
make test
make lint
make build
```

## Test Boundaries

Normal tests must be deterministic and local.

Allowed in normal tests:

- Go unit tests.
- `httptest` API handler tests.
- Fixture-backed fake discovery.
- Fixture-backed parser tests.
- In-memory fake repositories.
- Embedded migration ordering tests.
- Local temporary files via `t.TempDir()`.

Not allowed in normal tests:

- Real network devices.
- SSH to live routers.
- Docker or containerized databases.
- Cloud services.
- External IPAM, EMS, monitoring, or source-of-truth systems.
- Tests that require raw credentials.
- Tests that perform write-capable network actions.

## Fake Collector Tests

The fake collector is the supported end-to-end discovery test boundary.

Rules:

- Targets must use `fixture://`.
- Fixtures live under `examples/fixtures`.
- Fixture output must be stable text files committed to the repository.
- The collector must preserve task order and return deterministic outputs for repeated calls.
- Unsafe commands in profiles must still be rejected by policy even when using fixtures.

Useful packages:

- `internal/discovery`: fake collector, profile validation, workflow tests.
- `internal/evidence`: hashing and evidence validation tests.

## Parser Tests

Parser tests should use committed fixture files and should not call collectors or devices.

Rules:

- Read fixture text from `examples/fixtures`.
- Build `evidence.Evidence` values in memory.
- Assert normalized parser outputs: identities, inventory, neighbors, BGP peers, facts, and relationships.
- Parser failures must preserve the evidence reference and return warnings/errors without deleting raw evidence.

Useful package:

- `internal/parser`

## API Handler Tests

API tests should use `httptest` and in-memory fake services/repositories.

Rules:

- Do not start a real listener.
- Do not require PostgreSQL.
- Verify response envelopes: `data`, `error`, and `metadata`.
- Verify pagination/filtering response shape where relevant.
- Verify explicit discovery execution metadata and audit records.

Useful package:

- `internal/api`

## Repository Tests

Repository tests that require PostgreSQL are not part of the default suite yet because there is no live DB test harness and Docker is intentionally not required.

Current persistence coverage uses:

- Service-level tests with fake repositories.
- Migration ordering/embedding tests.
- SQL kept explicit and isolated in `internal/db`.

When a DB harness is added later, it should be opt-in and documented separately. It must not make `make test` depend on Docker or a real database unless the project explicitly changes that policy.

## Security Tests

Security-sensitive tests should cover:

- Denied command patterns.
- Shell metacharacter rejection.
- Script runner disabled-by-default behavior.
- Script allowlisting.
- Audit metadata presence.
- Credential/reference redaction hooks.

Useful packages:

- `internal/policy`
- `internal/audit`
- `internal/extensibility`
