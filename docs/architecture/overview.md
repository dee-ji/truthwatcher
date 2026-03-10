# Architecture Overview

Truthwatcher uses a monorepo with domain-focused packages and service entrypoints.

Intent flows through this lifecycle:
1. intent authoring
2. revision creation
3. artifact compilation
4. deployment planning
5. execution
6. state observation
7. drift detection
8. reconcile workflows
9. audit recording

The current repository state emphasizes coherent interfaces and composition-safe scaffolding over full production behavior.

## Extension points (important)
- **Archive repositories**: persistence contracts under `internal/*` should be backed by PostgreSQL without leaking SQL details into handlers.
- **Driver registry**: Elsecall should support additional vendor drivers through a stable `Render` contract.
- **Execution adapters**: Squire/Highstorm should gain pluggable adapters for device communication and staged rollout control.
- **Policy gates**: Oathgate should evolve from placeholder checks to explicit approval and simulation policy gates.
- **Identity providers**: authn middleware should move from local/dev behavior to OIDC-backed claim verification.

## Explicit TODOs
- TODO(truthwatcher): wire real migration and repository bootstrapping into startup paths.
- TODO(truthwatcher): persist reconcile runs and findings beyond in-memory/service-local behavior where applicable.
- TODO(truthwatcher): expand eventing and queue contracts for multi-service orchestration.
