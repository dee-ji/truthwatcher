# TruthWatcher Roadmap

## Phase 0: Project Constitution

Goal: prevent scope drift and agent hallucination.

Deliverables:

- steering-docs/PROJECT_TRUTHWATCHER.md
- steering-docs/ROADMAP.md
- steering-docs/ARCHITECTURE_DECISIONS.md
- steering-docs/DATA_MODEL.md
- steering-docs/AGENT_INSTRUCTIONS.md
- steering-docs/SAFETY_MODEL.md
- steering-docs/MVP_SPEC.md

Rules:

- Every coding session must start by reading these files.
- Every major design decision must update steering-docs/ARCHITECTURE_DECISIONS.md.
- Every Codex/agent task must be small and bounded.

Completed:

- Repository guardrails added: `.gitignore`, `Makefile`, and `CONTRIBUTING.md`; local `.editorconfig` files are ignored.
- Root Markdown cleanup completed by moving steering and prompt-pack documents into `steering-docs/` while keeping `README.md` and `CONTRIBUTING.md` at the repository root.

Next steps:
- Continue with Phase 1 single-binary kernel tasks.
- Keep new steering documents under `steering-docs/` unless they are public root entry points.

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
- HTTP API foundation added with health, readiness, version, request ID, request logging, and panic recovery middleware.
- Single-binary local packaging added with command help, embedded migrations/UI validation, install documentation, and `make release-local`.
- First user-facing README and concept documentation added for evidence-first modeling, discovery planning, assets/facts/relationships, and extensibility.
- Future phase candidates documented with prerequisites and v0.1 deferral rationale.
- Prompt pack expanded with post-documentation prompts for evidence-kernel completion, adapter boundaries, security foundations, and deferred future capabilities.

Next steps:

- Add installer checksums or signed release artifacts only when explicitly requested.
- Keep user-facing docs aligned with parser persistence and graph workflow changes as those features are completed.
- Promote future phases to implementation only through small, explicit prompts after prerequisite kernel work exists.
- Execute new prompts one at a time only when explicitly requested.

## Phase 2: Evidence Store

Goal: store raw discovery evidence before parsing.

Completed:

- DiscoveryRun product table, repository, service object, and API endpoints added.
- Raw evidence table, repository, hashing service, and read API endpoints added.

Next steps:

- Execute `prompts/07_ASSETS_FACTS_RELATIONSHIPS.md` only when explicitly requested.

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

Completed:

- Stable Asset, Fact, and Relationship tables, services, and repositories added.
- Asset, fact, relationship, and evidence read APIs added with filters, pagination, and response-shape tests.
- Uncertainty fields and deterministic confidence states added for assets, facts, and relationships.

Next steps:

- Wire parser outputs into asset, fact, and relationship persistence only when explicitly requested.

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

Completed:

- Read-only policy engine added for vendor-neutral discovery task allowlisting and dangerous command denial.
- Built-in Juniper Junos and Cisco IOS-XR discovery profiles added with read-only command mappings and parser hints.
- Fake local collector added for fixture-backed evidence creation without network access.
- Optional SSH collector boundary added behind the collector interface with read-only policy checks before command execution.
- First evidence-first discovery execution workflow added for fake collector runs through service, CLI, and API paths.
- Discovery API responses formalized with consistent envelopes, explicit execution validation, audit metadata, and endpoint documentation.

Next steps:

- Wire parser outputs into persistence only when explicitly requested.

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

Completed:

- Parser interface, normalized parse output model, registry, and no-op fallback parser added.
- First fixture-driven Junos and IOS-XR parsers added for version, inventory, and LLDP neighbor evidence.
- Explicit parser persistence workflow added for stored discovery-run evidence, including parse result records, CLI/API execution paths, and evidence-linked asset/fact/relationship creation.
- Deterministic identity strength handling added with strong vendor/serial and system MAC preference, provisional hostname/name/IP identities, conflict review, and weak/provisional identity review APIs.

Next steps:

- Design an explicit, human-reviewed identity merge workflow for cases where version, inventory, and neighbor evidence describe the same physical device through different keys.
- Add more fixture parser coverage only where it supports the evidence kernel workflow.

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

Completed:

- Relational graph query service and graph JSON API endpoints added for asset neighborhood rendering.
- Embedded static frontend foundation added with dashboard placeholder and API status indicator.
- Discovery run UI added with run list, detail view, evidence counts, and fake discovery form.
- Basic graph view added with API-backed node/edge rendering, asset click details, and edge confidence labels.
- Evidence drawer added for graph facts and relationships with read-only raw output display and copy support.
- Asset browsing UI added with API-backed filters, asset detail facts and relationships, confidence/state visibility, and linked read-only evidence access.

Next steps:

- Add focused evidence search or filtering views only when explicitly requested.

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

Completed:

- Safe discovery planning API added with read-only suggested steps, explicit human approval requirement, and scope-expansion rejection.
- Architecture seeding API added for user-seeded network type, ASN, route-reflector, vendor, EMS, service, and region/market hints; planner consumes hints without treating them as proof.
- Discovery plan review UI added for submitting seed targets and rendering suggested read-only steps without automatic execution.
- Architecture seed UI added for submitting user-seeded network type, ASN, route-reflector, vendor, EMS, service, and region/market hints without triggering discovery.

Next steps:

- Refine planner use of seeded context only when explicitly requested.

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

Completed:

- Deterministic Ask Truthwatcher shell added with read-only canned responses and browser-local conversation history.
- Agent workspace can answer graph-grounded asset, neighbor, existence, and unknown-state questions with evidence references.

Next steps:

- Add policy-gated task proposal models only when explicitly requested.

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

Completed:

- Compile-time Go extensibility contracts defined for collectors, parsers, importers, exporters, enrichers, and planners without dynamic plugin loading.
- Local JSON file import/export connector foundation added for assets, facts, relationships, and evidence metadata.
- BYO script contract documented and disabled-by-default local script runner foundation added with policy checks and static examples.
- Security and audit hardening pass added typed discovery audit records, expanded denied command tests, credential handling notes, and audit/log redaction hooks.
- Testing strategy documented with deterministic fake collector expectations, fixture parser/API test boundaries, and no-Docker/no-device normal test rules.

Next steps:

- Add CLI or API wiring for file import/export only when explicitly requested.
- Add CLI-only BYO script execution wiring only when explicitly requested.
- Add persistent audit storage only when explicitly requested.
- Add opt-in PostgreSQL repository integration tests only when a non-Docker DB harness is explicitly requested.

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
