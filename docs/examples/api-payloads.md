# API Payload Examples

These examples mirror currently implemented Spanreed endpoints.

## Create intent
`POST /api/v1/intents`

```json
{
  "name": "leaf-fabric",
  "description": "baseline underlay policy",
  "revision": 1,
  "spec": {
    "metadata": {"name": "campus-leaf-1"},
    "routing_intent": {"bgp": {"asn": 65000}},
    "target_scope": {"sites": ["dc1"]},
    "desired_services": [{"type": "l3-underlay"}]
  }
}
```

## Compile intent
`POST /api/v1/intents/{id}/compile`

```json
{"vendor": "junos"}
```

## Import topology snapshot
`POST /api/v1/topology/import`

```json
{
  "vendors": [{"id": "v-junos", "name": "junos"}],
  "platforms": [{"id": "p-qfx", "vendor_id": "v-junos", "name": "qfx"}],
  "sites": [{"id": "dc1", "name": "dc1"}],
  "devices": [{"id": "leaf-1", "hostname": "leaf-1", "vendor": "junos", "platform": "qfx", "site": "dc1"}],
  "interfaces": [{"id": "leaf-1:xe-0/0/0", "device_id": "leaf-1", "name": "xe-0/0/0"}],
  "links": []
}
```

## Create deployment
`POST /api/v1/deployments`

```json
{
  "intent_id": "leaf-fabric",
  "idempotency_key": "deploy-001",
  "mode": "dry-run",
  "targets": ["leaf-1", "leaf-2"],
  "batch_size": 1,
  "canary_targets": 1,
  "require_manual_approval": true
}
```

## Create reconcile run
`POST /api/v1/reconcile/runs`

```json
{
  "intent_id": "leaf-fabric",
  "actor": "twctl"
}
```
