# Local JSON Import and Export

Truthwatcher can move local graph snapshots through explicit CLI commands.

This is a file connector foundation only. It is not a NetBox, Nautobot, IPAM, EMS, monitoring, or cloud connector.

## Export

```sh
TRUTHWATCHER_DATABASE_URL='postgres://user:pass@localhost:5432/truthwatcher?sslmode=disable' \
  ./bin/truthwatcher export json --output ./truthwatcher-snapshot.json
```

The JSON snapshot includes:

- assets
- facts
- relationships
- evidence metadata

The export preserves source, confidence, state, and evidence references where they exist.

Raw evidence output is not exported by default. Evidence metadata includes target, method, command/API, hash, parser name, collection time, and metadata.

## Import

```sh
./bin/truthwatcher import json --input ./truthwatcher-snapshot.json
```

Import currently validates and summarizes candidates. It does not persist records to PostgreSQL.

This is intentional: imported file data must not bypass kernel validation, conflict review, or evidence-first rules. If a file contains records marked `observed`, the CLI treats them as candidates only and warns that the command did not persist them as observed proof.

## Trust Model

Local JSON import/export is for review, testing, and moving snapshots between environments.

Imported records are context until reconciled by the kernel. They must not replace observed device evidence, and they must not silently overwrite existing facts.
