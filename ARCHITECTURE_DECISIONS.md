# Architecture Decisions

This file records durable design decisions. Update it whenever the project makes a meaningful architecture choice.

## ADR-001: TruthWatcher Is Evidence-First

Decision: Raw evidence must be stored before parsed facts are created.

Reason: In large networks, parsers may be wrong, incomplete, or vendor-specific. Raw evidence preserves auditability and allows future reparsing.

Implication: Every fact and relationship should link back to evidence when possible.

## ADR-002: Stable Core Schema + Dynamic JSONB

Decision: Use stable relational tables for core concepts and JSONB for vendor-specific details.

Core concepts:

- Asset.
- Fact.
- Relationship.
- Evidence.
- DiscoveryRun.
- AccessPath.
- CredentialRef.
- DiscoveryProfile.

Reason: Fully dynamic schemas are fragile. Vendor surface area is too broad to model upfront. JSONB allows adaptation while stable tables preserve queryability.

## ADR-003: Single Go Binary

Decision: TruthWatcher should initially run as a single Go binary with embedded frontend assets.

Reason: The project should be easy to download, run, test, and understand. Avoid unnecessary platform complexity.

Target UX:

- `truthwatcher server`
- `truthwatcher migrate`
- `truthwatcher discover`
- `truthwatcher version`

## ADR-004: PostgreSQL as the Primary Database

Decision: Use PostgreSQL as the primary database.

Reason: PostgreSQL supports relational integrity, JSONB, indexing, constraints, recursive queries, and can serve as an effective first graph-like store.

Future: A graph database may be added later only if PostgreSQL is proven insufficient.

## ADR-005: Graph Is a Data Model, Not Initially a Separate Database

Decision: Model graph relationships in PostgreSQL first.

Reason: Starting with Neo4j or a graph database too early increases operational complexity. The first problem is evidence correctness, not graph database selection.

## ADR-006: Read-Only by Default

Decision: All discovery must be read-only by default.

Reason: The tool must be safe for production service-provider networks.

Blocked in v0.1:

- Configuration changes.
- Commit/write memory.
- Reload/reboot.
- Clear commands.
- Delete commands.
- Copy commands.
- Arbitrary shell commands.

## ADR-007: Agents Plan, Collectors Execute

Decision: Agentic workflows may propose discovery actions but must not directly execute arbitrary commands.

Reason: Safety and repeatability require deterministic collectors and policy enforcement.

## ADR-008: Credential References, Not Stored Secrets

Decision: TruthWatcher should avoid storing raw credentials in v0.1.

Reason: Secrets handling is a major security concern. Store references to external credential providers or local environment/config mechanisms.

## ADR-009: Observability Is Deferred

Decision: Monitoring and alarming are not part of the initial product.

Reason: Observability can overwhelm the core mission. Runtime health can later attach to the graph.

## ADR-010: UI Should Be Graph-First and Chat-Ready

Decision: The UI should eventually center on a network graph and an engineering chat workspace.

Reason: Engineers reason about paths, services, sites, risks, and relationships more naturally than isolated tables.

Initial UI still starts simple: assets, evidence, relationships.

## ADR: Stable Kernel with Replaceable Integration Edges

Status: Accepted

Truthwatcher will be designed as a stable core with optional adapters for external systems.

The kernel owns the universal concepts: assets, facts, relationships, evidence, discovery runs, safety policy, confidence, and graph construction.

External systems such as IPAM, monitoring, EMS, cloud platforms, CMDBs, credential vaults, and config archives must integrate through adapters.

Rationale:

- Service providers have heterogeneous environments.
- No single IPAM, NMS, EMS, or CMDB can be assumed.
- The project must remain useful to many organizations.
- Integrations should enrich the graph without bloating the core.

Decision:

- All integrations are optional.
- Integration data is treated as evidence.
- Adapters translate external data into assets, facts, relationships, and evidence.
- The single-binary deployment goal remains intact.

## ADR-011: PostgreSQL Driver Uses database/sql with lib/pq

Decision: Use Go's standard `database/sql` package with `github.com/lib/pq` as the PostgreSQL driver.

Reason: Truthwatcher needs a small PostgreSQL foundation without an ORM or query framework. `lib/pq` provides a focused `database/sql` driver with a small API surface, keeping early database code explicit and boring.

Implication: Database access should stay behind internal packages. If later requirements need PostgreSQL-specific features that are awkward through `database/sql`, the driver choice can be revisited without changing the project architecture.
