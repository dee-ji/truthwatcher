# Truthwatcher

Truthwatcher is an open-source, intent-driven network management platform for hyperscaler-style control planes.

## Current maturity
Initial scaffold: buildable binaries, API/CLI skeletons, migrations, drivers, and docs with intentional TODO stubs for subsystem internals.

## Quickstart
1. `make build`
2. `make test`
3. `make compose-up`
4. `curl localhost:8080/healthz`

## Core services
- `radiant`: control-plane orchestration service.
- `spanreed`: API and external interface service.
- `highstorm`: deployment orchestration engine.
- `stormlight`: drift detection and reconcile trigger service.
- `seekers`: topology discovery and ingestion service.
- `squire`: execution worker service.
- `twctl`: operator CLI.

Legacy compatibility binaries (`tw-server`, `tw-worker`) remain as deprecation wrappers.

## Local stack
Docker Compose provides PostgreSQL, Redis, Spanreed API, and Squire worker scaffolding.

## Roadmap summary
- Persist domain services with Archive-backed PostgreSQL repositories.
- Introduce Redis-backed queue/stream workflows.
- Expand Elsecall artifact rendering and vendor drivers.
- Add OIDC/JWT authentication and policy-backed authorization.
