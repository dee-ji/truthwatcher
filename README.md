# Truthwatcher

Truthwatcher is an open-source, intent-driven network management platform.

It is organized as a control-plane scaffold that keeps architectural language and interfaces stable while subsystem internals are incrementally implemented.

## Current maturity
Truthwatcher is in an intentional foundation phase: binaries compile, HTTP endpoints and migrations exist, tests exercise core scaffolding, and major subsystems expose explicit TODO surfaces rather than pretending to be production-complete.

## Platform vocabulary
Use these terms consistently across code, docs, and APIs:

- **intent**: desired network behavior and policy definition.
- **revision**: immutable version of an intent set.
- **artifact**: compiled, vendor-targeted output from Elsecall.
- **topology**: devices, links, and adjacency graph state.
- **deployment**: rollout plan and execution lifecycle for intent revisions.
- **state**: observed snapshots from the live network.
- **reconcile**: convergence workflow from drift findings back to declared intent.
- **drift**: mismatch between intended artifacts and observed state.
- **audit**: append-only record of control-plane actions.
- **driver**: vendor-specific renderer/execution adapter contract implementation.

## Quickstart (new contributor)
### 1) Build and test the monorepo
1. `make build`
2. `make test`

### 2) Start local platform dependencies and core services
1. `make compose-up`
2. `curl -s localhost:8080/healthz`
3. `curl -s localhost:8080/version`

Compose currently starts PostgreSQL, Redis, Spanreed (`:8080`), and Squire. Radiant/Highstorm/Stormlight/Seekers are available as local binaries.

### 3) Run foundational checks
1. `make migrate-up` (migration CLI scaffold; currently no real DB apply)
2. `make openapi`

### 4) Inspect starter assets
- `examples/intents/leaf-fabric.yaml` for intent input.
- `examples/topology/fabric-small.yaml` for topology fixtures.
- `examples/rendered-configs/` for compiled artifacts.

## Core services
- `radiant`: control-plane orchestration service.
- `spanreed`: API and external interface service.
- `highstorm`: deployment orchestration engine.
- `stormlight`: drift detection and reconcile trigger service.
- `seekers`: topology discovery and ingestion service.
- `squire`: distributed execution worker service.
- `twctl`: operator CLI.

Legacy compatibility binaries (`tw-server`, `tw-worker`) remain as deprecation wrappers.

## Intentional TODO surfaces
Truthwatcher keeps these areas explicitly stubbed for safe iteration:

- Real migration execution and state tracking in `cmd/tw-migrate`.
- Persistent repositories for core domain services (Archive-backed).
- Live execution adapters and transactional deployment orchestration.
- OIDC-backed authn/authz integration beyond local/dev modes.
- Rich topology analytics and simulation depth in Oathgate/Shadesmar.

See `ROADMAP.md` for phased expansion.
