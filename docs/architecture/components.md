# Components

- **Spanreed**: HTTP API boundary with OpenAPI contract and authn/authz middleware.
- **Ideals**: intent validation schemas and policy prechecks.
- **Elsecall**: compiles intent revisions into vendor-targeted artifacts through driver interfaces.
- **Shadesmar**: topology graph service and adjacency/query helpers.
- **Oathgate**: deployment safety/simulation guardrails (stubbed foundation).
- **Highstorm**: deployment lifecycle orchestration model.
- **Stormlight**: drift detection and reconcile run lifecycle.
- **Archive**: long-term source-of-truth data persistence (migrations + future repositories).
- **Audit**: append-only event history for control-plane actions.

## Compile pipeline (first implemented slice)
1. Spanreed receives `POST /api/v1/intents/{id}/compile` with optional `vendor`.
2. Elsecall normalizes revision payload (`metadata`, `routing_intent`, target scope, services) into `DeviceConfigIR`.
3. Elsecall invokes a vendor driver (`Vendor()`, `Render(...)`).
4. Artifact metadata is returned through intent APIs and can be stored via Archive persistence layers.

Current concrete renderer: Junos set-format output with TODO markers for unsupported sections (for example, `interface_intent`).

## Explicit TODOs
- TODO(truthwatcher): register EOS/IOS-XE/IOS-XR drivers in default compiler service once fixture coverage is complete.
- TODO(truthwatcher): add artifact provenance fields (schema/revision hash, compiler version).
