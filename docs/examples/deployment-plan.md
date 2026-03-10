# Deployment plan example

This example shows a **safe-first deployment plan** generated from compiled artifacts.

## API request

```http
POST /api/v1/deployments
Content-Type: application/json

{
  "intent_id": "intent-42",
  "idempotency_key": "change-2026-03-10-1",
  "mode": "dry-run",
  "targets": ["leaf-1", "leaf-2", "leaf-3", "leaf-4"],
  "batch_size": 2,
  "canary_targets": 1,
  "require_manual_approval": true
}
```

## API response

```json
{
  "id": "dep-1",
  "intent_id": "intent-42",
  "status": "planned",
  "idempotency_key": "change-2026-03-10-1",
  "mode": "dry-run",
  "artifact_refs": ["junos/set", "eos/cfg"],
  "targets": ["leaf-1", "leaf-2", "leaf-3", "leaf-4"],
  "rollout": {
    "waves": [
      {
        "name": "canary",
        "order": 1,
        "max_targets": 1,
        "canary_targets": 1,
        "requires_approval": true,
        "planned_target_count": 1
      },
      {
        "name": "batch",
        "order": 2,
        "max_targets": 2,
        "requires_approval": true,
        "planned_target_count": 3
      }
    ]
  },
  "stop_conditions": [
    {
      "type": "error_rate",
      "threshold": ">2%",
      "reason": "placeholder policy while Oathgate checks are integrated"
    }
  ],
  "rollback_plan": {
    "strategy": "previous-known-good",
    "steps": [
      "capture pre-deployment state snapshot",
      "re-apply last-known-good artifact",
      "TODO: enforce rollback orchestration through Highstorm"
    ]
  },
  "created_at": "2026-03-10T00:00:00Z"
}
```

## Safety notes

- `mode: dry-run` simulates rollout planning only; no device execution is performed.
- idempotency keys prevent duplicate plan creation for retried requests.
- approval and stop-condition enforcement are TODO placeholders until execution engines are wired.
