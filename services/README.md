# services/

Service-level architecture notes and coordination guidance.

Most implementation currently lives under `internal/*` packages, while `cmd/*` provides runnable entrypoints.

## Service map
- `radiant`: control-plane orchestration
- `spanreed`: API/interface boundary
- `highstorm`: deployment orchestration
- `stormlight`: drift detection/reconcile trigger
- `seekers`: topology discovery/ingestion
- `squire`: distributed execution worker

## TODO
- TODO(truthwatcher): add per-service runbooks and failure-mode notes as behavior moves beyond scaffold.
