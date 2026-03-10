# Offline reconciliation example

This walkthrough demonstrates **safe offline reconciliation** in Truthwatcher.

The reconcile run compares intended compiled artifacts against stored snapshots in `config_snapshots`; it does not poll devices directly.

## Prepare snapshots

Use fixtures under `examples/state/` as the source of truth for actual state snapshots.

## Start a reconcile run

```bash
curl -X POST http://localhost:8080/api/v1/reconcile/runs \
  -H 'content-type: application/json' \
  -d '{"intent_id":"<intent-id>","actor":"operator"}'
```

## Fetch run details

```bash
curl http://localhost:8080/api/v1/reconcile/runs/<run-id>
```

## List drift findings

```bash
curl http://localhost:8080/api/v1/drift/findings
```

## CLI compare

```bash
twctl state compare <intent-id>
```

## Notes

- Drift detection currently uses deterministic text comparison of intended artifact content vs stored snapshot content.
- TODO: integrate live snapshot collectors for controlled online capture.
- TODO: implement auto-remediation planning and approval workflows.
