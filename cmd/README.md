# cmd/

Service entrypoints for Truthwatcher binaries.

## Primary service binaries
- `cmd/radiant`: orchestration control plane.
- `cmd/spanreed`: API and external interface layer.
- `cmd/highstorm`: deployment engine.
- `cmd/stormlight`: drift detection and reconcile trigger engine.
- `cmd/seekers`: topology discovery and inventory ingestion.
- `cmd/squire`: distributed execution worker.
- `cmd/twctl`: operator CLI.

## Utility / compatibility binaries
- `cmd/tw-migrate`: migration helper scaffold.
- `cmd/tw-render`: render preview helper scaffold.
- `cmd/tw-server`: deprecated wrapper retained for compatibility.
- `cmd/tw-worker`: deprecated wrapper retained for compatibility.
