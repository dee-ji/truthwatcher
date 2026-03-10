# migrations/

SQL migrations for Archive-backed persistence scaffolding.

## Current status
Migration files define conceptual schema evolution, but `cmd/tw-migrate` is still a command scaffold and does not yet apply SQL to a live database.

## Naming
Files use golang-migrate style naming: `NNNNNN_description.(up|down).sql`.

## TODO
- TODO(truthwatcher): integrate real migration execution, locking, and status tracking.
