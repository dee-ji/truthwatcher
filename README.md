# Truthwatcher

Truthwatcher is an open-source, intent-driven network management platform for hyperscaler-style control planes.

## Current maturity
Initial production-grade scaffold: buildable binaries, API/CLI skeletons, migrations, drivers, and docs.

## Quickstart
1. `cp .env.example .env`
2. `make build`
3. `make test`
4. `make compose-up`
5. `curl localhost:8080/healthz`

## Main services
- `tw-server`: API + control-plane entrypoint.
- `tw-worker`: queue worker scaffold.
- `twctl`: operator CLI.
- `tw-migrate`: migration command scaffold.

## Local stack
Docker Compose provides PostgreSQL, Redis, API, and worker.

## Roadmap summary
- Persist domain services with PostgreSQL repositories.
- Introduce Redis-backed queue/stream workflows.
- Expand compiler and vendor drivers.
- Add OIDC/JWT and RBAC enforcement.
