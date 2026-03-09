# First Intent Workflow
1. Edit `examples/intents/leaf-fabric.yaml`.
2. Validate with `twctl intent validate`.
3. Compile with `POST /api/v1/intents/{id}/compile`.
4. Inspect rendered examples in `examples/rendered-configs`.
5. Create deployment via `POST /api/v1/deployments`.
6. Simulate rollout (placeholder in deploy service).
7. Query audit events via `GET /api/v1/audit/events`.
8. Trigger reconciliation via `POST /api/v1/reconcile/runs`.
