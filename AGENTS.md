# AGENTS.md

## Purpose
This file defines contributor conventions for both humans and AI agents across the entire repository.

Truthwatcher is a scaffold-first, intent-driven control plane. Keep changes honest, explicit, and compilation-safe.

## Required vocabulary
Use these domain terms consistently in code, docs, API payloads, examples, and commit messages:

- intent
- revision
- artifact
- topology
- deployment
- state
- reconcile
- drift
- audit
- driver

Avoid introducing synonyms when one of the canonical terms already exists.

## Architecture naming guardrails
Core subsystem names are architectural vocabulary and should remain stable:

- Radiant (control plane orchestration)
- Spanreed (API/interface layer)
- Archive (source of truth)
- Ideals (intent modeling/validation)
- Elsecall (translation/rendering)
- Shadesmar (topology graph)
- Oathgate (safety simulation)
- Highstorm (deployment engine)
- Stormlight (drift detection/reconcile)
- Seekers (discovery/ingestion)
- Squire (distributed execution worker)

## Development conventions
- Keep packages domain-focused; avoid circular dependencies.
- Prefer explicit TODOs over fake completeness.
- Maintain buildable code even for partial implementations.
- Keep API contracts and example payloads aligned with implemented handlers.
- Update docs when changing user-facing behavior or architecture boundaries.

## Testing conventions
- Run `make build` and `make test` for code changes.
- Prefer deterministic fixtures over fragile environment-dependent behavior.
- Put integration/e2e-style fixtures under `test/` and `examples/` with clear README notes.

## Security-sensitive paths
Changes in authn/authz, middleware, permission mapping, and deployment safety checks must:
1. Use explicit allow/deny checks near handlers.
2. Keep dev bypass behavior clearly visible in docs/logs.
3. Include tests for both allowed and denied outcomes.
4. Document permission or role changes.

## Stub policy (important)
Truthwatcher intentionally contains unfinished areas. Do not paper over them.

When a feature is incomplete:
- leave a clear `TODO(truthwatcher): ...` note,
- document what is intentionally missing,
- avoid pretending external integrations are production-ready.
