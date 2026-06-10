# TruthWatcher Roadmap

## Phase 0: Project Constitution

Goal: prevent scope drift and agent hallucination.

Deliverables:

- PROJECT_TRUTHWATCHER.md
- ROADMAP.md
- ARCHITECTURE_DECISIONS.md
- DATA_MODEL.md
- AGENT_INSTRUCTIONS.md
- SAFETY_MODEL.md
- MVP_SPEC.md

Rules:

- Every coding session must start by reading these files.
- Every major design decision must update ARCHITECTURE_DECISIONS.md.
- Every Codex/agent task must be small and bounded.

Completed:

- Repository guardrails added: `.gitignore`, `Makefile`, and `CONTRIBUTING.md`; local `.editorconfig` files are ignored.

Next steps:
- Continue with Phase 1 single-binary kernel tasks.

## Phase 1: Single-Binary Kernel

Goal: create a Go application that runs as a single CLI/server binary.

Deliverables:

- `truthwatcher server` command.
- `truthwatcher migrate` command.
- `truthwatcher version` command.
- Embedded static frontend placeholder.
- Local config file support.
- Structured logging.
- PostgreSQL connection.
- Health endpoint.

Non-goals:

- No Kubernetes.
- No Docker requirement.
- No agent workflow yet.
- No chat UI yet.

Completed:

- Initial Go project skeleton added with `cmd/truthwatcher`.
- Standard-library CLI supports `truthwatcher version` and `truthwatcher server`.
- Internal package layout created for app, config, logging, API, DB, discovery, evidence, assets, policy, and audit.
- Server starts without database, migrations, frontend, collectors, or agents.
- Environment-based configuration and standard-library structured logging added for the server.
- PostgreSQL database foundation added with `database/sql`, embedded migrations, and `truthwatcher migrate up/status`.

Next steps:

- Execute `prompts/04_HTTP_API_FOUNDATION.md` only when explicitly requested.

## Phase 2: Evidence Store

Goal: store raw discovery evidence before parsing.

Deliverables:

- `discovery_runs` table.
- `evidence` table.
- API to create/list discovery runs.
- API to list evidence for a run.
- Raw output hashing.
- Timestamps.
- Target metadata.
- Method metadata.

Acceptance criteria:

- Raw command output can be stored unchanged.
- Evidence can be traced to a discovery run.
- Evidence can be hashed for deduplication and integrity.

## Phase 3: Asset, Fact, Relationship Model

Goal: transform evidence into network knowledge.

Deliverables:

- `assets` table.
- `facts` table.
- `relationships` table.
- Confidence scoring field.
- Evidence linkage from facts and relationships.
- Stable core schema plus JSONB for dynamic vendor-specific data.

Acceptance criteria:

- A device asset can be created from evidence.
- A chassis/card/port/optic can be represented as an asset.
- Parent-child relationships can be created.
- Device-to-device neighbor relationships can be created.

## Phase 4: Read-Only SSH Discovery

Goal: safely collect basic identity and topology data from one seed device.

Deliverables:

- SSH collector interface.
- Credential reference abstraction.
- Read-only command allowlist.
- Vendor discovery profiles.
- Cisco IOS-XR profile.
- Juniper Junos profile.
- Raw evidence storage for every command.

Initial safe commands:

- show version
- show inventory
- show chassis hardware
- show interfaces terse / brief
- show lldp neighbors
- show cdp neighbors
- show arp
- show ipv6 neighbors
- show bgp summary
- show route summary

Acceptance criteria:

- Given a seed target, TruthWatcher can run approved read-only commands and store raw output.
- Arbitrary commands are rejected.

## Phase 5: Parser Framework

Goal: normalize raw evidence into facts.

Deliverables:

- Parser interface.
- Parser registry.
- Device identity parser.
- Interface parser.
- Neighbor parser.
- BGP neighbor parser.
- Parser confidence output.

Acceptance criteria:

- Junos and IOS-XR sample outputs can produce assets/facts/relationships.
- Parser failures are stored as errors without failing the entire discovery run.

## Phase 6: Basic Web UI

Goal: make the discovered knowledge visible.

Deliverables:

- Embedded frontend.
- Dashboard page.
- Assets table.
- Asset detail page.
- Evidence drawer.
- Relationship list.
- Basic graph visualization.

UX principle:

The UI should help engineers understand evidence and relationships, not force them through endless tables.

## Phase 7: Discovery Planner

Goal: begin solving “discover how to discover.”

Deliverables:

- Seed questionnaire.
- Discovery plan model.
- Architecture hints.
- Known vendors.
- Known ASNs.
- Known route reflectors.
- Known EMS/controllers.
- Known access methods.
- Known subnets/domains.

Acceptance criteria:

- A user can seed architectural knowledge.
- The system can recommend next safe read-only discovery actions.

## Phase 8: Agentic Workflow

Goal: add a controlled agent that plans but does not directly execute unsafe actions.

Rules:

- Agent never receives raw credentials.
- Agent never runs arbitrary commands.
- Agent proposes read-only tasks.
- Policy engine approves/rejects tasks.
- Every action is audited.

Deliverables:

- Chat/agent interface.
- Agent task proposal model.
- Policy approval model.
- Discovery task execution through existing collectors only.

## Phase 9: Service-Aware Modeling

Goal: model how providers sell and deliver services.

Service examples:

- Internet access.
- L3VPN.
- E-Line.
- EVLAN.
- DIA.
- Managed CPE.
- Cloud connectivity.

Deliverables:

- Service assets.
- Service-to-device relationships.
- Customer-premise relationships.
- Provider-edge relationships.
- Route reflector/core/backbone relationships.
- MOP generation later.

## Phase 10: Protocol Expansion

Add collectors for:

- NETCONF.
- RESTCONF.
- gNMI/gRPC.
- SNMP.
- EMS APIs.
- Cloud APIs.
- Terminal server paths.
- Jump hosts.
- Config archives.
- IPAM/DNS/DHCP.

## Phase 11: Integrations

Export/sync with:

- Nautobot.
- NetBox.
- JSON.
- GraphQL.
- CSV.
- Git-backed reports.

## Phase 12: Observability Optional

Observability is intentionally deferred.

If added, it should attach runtime state to the existing graph rather than becoming the core product.

## Extensibility Roadmap

Extensibility is a core design requirement, but it must be introduced carefully.

### Phase E0: Kernel Boundary

Define clear interfaces for:

- collector plugins
- parser plugins
- import adapters
- export adapters
- enrichment adapters

### Phase E1: File Import Adapter

Add a simple YAML/JSON/CSV import adapter to seed known facts.

This proves the adapter architecture without depending on a third-party system.

### Phase E2: IPAM Adapter

Add a generic IPAM adapter interface.

Do not hard-code one IPAM as the only supported model.

### Phase E3: Source-of-Truth Export

Add exporter support for systems such as Nautobot or NetBox.

Truthwatcher should be able to bootstrap and enrich a source-of-truth system.

### Phase E4: Monitoring Enrichment

Add optional monitoring/NMS enrichment.

This should not become observability or alerting. It should only enrich asset and relationship context.
