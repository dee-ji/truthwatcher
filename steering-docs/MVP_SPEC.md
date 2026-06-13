# TruthWatcher MVP Specification

## MVP Name

TruthWatcher v0.1: Read-Only Network Evidence Kernel

## MVP Goal

Given one seed device and read-only access, TruthWatcher should collect evidence, parse basic identity/topology facts, create assets and relationships, and display the result.

## MVP User Story

As a network engineer, I want to provide a seed router and a read-only credential reference so that TruthWatcher can discover what the device is, what hardware it contains, what interfaces it has, and what neighbors it sees.

## MVP Inputs

- Seed target: hostname or IP.
- Vendor/platform hint: optional.
- Access method: SSH only for v0.1.
- Credential reference.
- Discovery profile.

## MVP Outputs

- Discovery run record.
- Raw evidence records.
- Device asset.
- Chassis/card/port/optic assets when parser supports them.
- Facts about the device.
- Relationships between assets.
- Neighbor relationships when discovered.
- API responses.
- Basic web display.

## MVP CLI Commands

```bash
truthwatcher version
truthwatcher migrate
truthwatcher server --config ./truthwatcher.yaml
truthwatcher discover ssh --target 10.0.0.1 --profile junos --credential-ref lab-readonly
```

## MVP HTTP Endpoints

```text
GET  /health
GET  /api/v1/discovery-runs
POST /api/v1/discovery-runs
GET  /api/v1/discovery-runs/{id}
GET  /api/v1/discovery-runs/{id}/evidence
GET  /api/v1/assets
GET  /api/v1/assets/{id}
GET  /api/v1/assets/{id}/facts
GET  /api/v1/assets/{id}/relationships
GET  /api/v1/relationships
```

## MVP Frontend

Embedded frontend served by the Go binary.

Initial pages:

- Dashboard.
- Discovery runs.
- Assets.
- Asset detail.
- Evidence detail.
- Relationship view.

## MVP Discovery Tasks

- identify_device.
- get_inventory.
- get_interfaces.
- get_neighbors.
- get_bgp_summary.

## MVP Supported Platforms

Start with sample-driven parsers for:

- Juniper Junos.
- Cisco IOS-XR.

Only implement what is testable with sample command output.

## MVP Success Criteria

The MVP is successful when:

- The binary starts a web server.
- The binary connects to PostgreSQL.
- Migrations create the schema.
- A discovery run can be created.
- Raw evidence can be stored.
- At least one parser creates an asset from sample output.
- At least one parser creates a relationship from sample neighbor output.
- The UI displays assets and evidence.

## MVP Non-Goals

Do not build:

- Chat interface.
- LLM agent.
- Observability.
- Config generation.
- Service provisioning.
- Full dynamic schema generator.
- Multi-tenant SaaS.
- Kubernetes deployment.
- Docker requirement.
