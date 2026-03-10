# Data Model

PostgreSQL migrations define foundational entities for:
- topology inventory (vendors, platforms, sites, devices, interfaces, links)
- intent sets and immutable revisions
- compiled artifacts
- deployment plans/runs/targets
- observed state snapshots
- drift findings and reconcile runs
- audit events
- authn/authz role mappings

Primary definitions live in `migrations/000001_init.up.sql` with focused follow-up migrations for deployment planning, state/reconcile, and RBAC.

## Naming consistency
The schema intentionally tracks platform vocabulary directly: `intent_revisions`, `compiled_artifacts`, `deployment_*`, `drift_findings`, and `reconcile_runs`.

## Explicit TODOs
- TODO(truthwatcher): add foreign-key relationships from audit events to principal identities where available.
- TODO(truthwatcher): add history/version tables for topology and state snapshots once retention policy is defined.
