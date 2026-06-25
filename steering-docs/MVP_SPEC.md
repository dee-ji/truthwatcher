# TruthWatcher POC Specification

## POC Name

TruthWatcher v0.1: Vendor-Neutral Evidence-To-Understanding Kernel

## POC Goal

Given a seed target, fixture, or imported context and an approved read-only collection path, TruthWatcher should preserve raw evidence, derive identity, create assets, extract facts, build relationships, project a graph, and display the result with confidence and provenance.

## POC User Story

As a network engineer in a complex multi-vendor environment, I want to start with limited trustworthy input and let TruthWatcher build an explainable model of what exists, how it is related, and what remains uncertain so that planning, troubleshooting, and future automation are based on evidence rather than stale assumptions.

## Conceptual Pipeline

```text
Evidence
  | IdentitySkill
  v
Identity
  | AssetDiscoverySkill
  v
Assets
  | FactExtractionSkill
  v
Facts
  | RelationshipSkill
  v
Relationships
  | GraphBuilderSkill
  v
KnowledgeGraph
  | ReasoningSkill
  v
Understanding
```

The skill names describe product responsibilities. They do not require a plugin architecture in v0.1.

## POC Inputs

- Seed target, fixture target, or imported local context.
- Optional platform, role, site, service, region, or vendor hints.
- Approved read-only discovery profile or file import contract.
- Credential reference when real collection is enabled.
- Human-seeded architecture context that is clearly marked as context, not observed proof.

## POC Outputs

- Discovery run or import record.
- Raw evidence records with provenance.
- Identity candidates and review state.
- Assets such as devices, chassis, cards, ports, optics, interfaces, services, or other modeled entities as parser support matures.
- Facts about assets.
- Relationships between assets and facts.
- Graph views and API responses.
- Basic web display with evidence links.

## POC CLI Commands

```bash
truthwatcher version
truthwatcher migrate
truthwatcher server --config ./truthwatcher.yaml
truthwatcher discover fake --target fixture://junos-mx
truthwatcher discover ssh --target 10.0.0.1 --profile <approved-profile> --credential-ref <readonly-ref>
truthwatcher import json --input ./truthwatcher-snapshot.json
truthwatcher export json --output ./truthwatcher-snapshot.json
```

Example fixture and profile names may reference a vendor sample, but they are test fixtures and adapter examples rather than product foundations.

## POC HTTP Surface

The API should expose the kernel concepts rather than a vendor-specific workflow:

```text
GET  /healthz
GET  /readyz
GET  /api/v1/version
GET  /api/v1/discovery-runs
POST /api/v1/discovery-runs
GET  /api/v1/discovery-runs/{id}
GET  /api/v1/discovery-runs/{id}/evidence
POST /api/v1/discovery-runs/{id}/parse
GET  /api/v1/evidence/{id}
GET  /api/v1/assets
GET  /api/v1/assets/{id}
GET  /api/v1/assets/{id}/facts
GET  /api/v1/assets/{id}/relationships
GET  /api/v1/graph/neighbors
```

See `docs/api.md` for the implementation-level endpoint reference.

## POC Frontend

Embedded frontend served by the Go binary.

Initial pages:

- Dashboard.
- Discovery runs.
- Evidence detail and evidence drawer.
- Assets.
- Asset detail.
- Relationship and graph views.
- Discovery plan review.
- Architecture seed review.

## POC Discovery Tasks

Discovery tasks should be expressed as vendor-neutral intents and mapped to safe implementation details by profiles or adapters:

- identify device or endpoint
- collect inventory evidence
- collect interface evidence
- collect neighbor evidence
- collect routing or service-context evidence when safe and supported

## Supported Platform Policy

The POC may include sample-driven parsers and fixtures for specific network operating systems only when sample output and tests exist. A fixture or parser is an integration example, not a statement that TruthWatcher is tied to that vendor.

Unsupported vendors, controllers, APIs, and data sources belong behind adapter boundaries until their evidence contracts and tests are explicit.

## POC Success Criteria

The POC is successful when:

- The binary starts a web server.
- The binary connects to PostgreSQL.
- Migrations create the schema.
- A discovery run or import can be created.
- Raw evidence can be stored and retrieved.
- At least one parser or import workflow creates evidence-backed identity and assets from sample data.
- At least one workflow creates evidence-backed relationships.
- The API and UI display assets, relationships, graph context, evidence, confidence, and unknown/conflict state.
- Vendor-specific behavior is isolated to profiles, parsers, fixtures, or adapters.

## POC Non-Goals

Do not build:

- Chat-first interface.
- Autonomous remediation.
- Write-capable network automation.
- Observability, alerting, or polling replacement.
- Config generation or service provisioning as a core feature.
- Full dynamic schema generator.
- Multi-tenant SaaS.
- Kubernetes or Docker requirement.
- Hard dependency on any vendor, IPAM, source-of-truth, EMS, cloud, or monitoring product.
