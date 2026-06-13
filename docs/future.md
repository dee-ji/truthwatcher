# Future Phases

This document records future candidates without making them part of v0.1.

Truthwatcher v0.1 remains focused on the evidence kernel:

- raw evidence storage
- assets, facts, and relationships
- confidence and uncertainty
- read-only safety policy
- fake fixture discovery
- explicit discovery runs
- PostgreSQL-backed graph projection
- embedded UI and local API

Future work must preserve the single-binary plus PostgreSQL model unless a later explicit architecture decision changes that boundary.

## Read-Only Protocol Collectors

### NETCONF Collector

Why it matters:

- NETCONF is common in service-provider networks and can expose structured operational data.
- Structured replies can reduce parser fragility compared with CLI scraping.

What must exist before it:

- Stable collector interface behavior.
- Policy checks for RPC/action allowlisting.
- Evidence records that preserve raw XML replies.
- Parser path for turning NETCONF replies into normalized candidates.
- Integration tests that do not require live devices by default.

Why it is not v0.1:

- The current kernel still needs stronger evidence-to-fact persistence workflows.
- NETCONF adds protocol and vendor surface area before the core evidence chain is fully proven.

### SNMP Collector

Why it matters:

- SNMP is widely available and useful for inventory, interfaces, LLDP/CDP, and basic device facts.
- It can provide low-impact read-only evidence in environments where SSH is limited.

What must exist before it:

- Credential reference handling for communities or SNMPv3 credentials.
- Explicit OID allowlists.
- Evidence storage for raw varbinds and walk metadata.
- Rate and scope controls to avoid accidental broad polling.

Why it is not v0.1:

- SNMP can easily drift toward monitoring/observability.
- Truthwatcher must first keep discovery evidence distinct from polling and alerting behavior.

### gNMI Collector

Why it matters:

- gNMI can provide structured operational state from modern network devices.
- It can eventually support vendor-neutral models where available.

What must exist before it:

- Read-only path allowlists.
- TLS and credential reference strategy.
- Evidence format for raw notifications/responses.
- Parser mapping from paths to generic assets, facts, and relationships.

Why it is not v0.1:

- It adds transport, authentication, and data-model complexity.
- The first milestone does not need streaming telemetry or subscription behavior.

### RESTCONF Collector

Why it matters:

- RESTCONF can expose structured YANG-backed operational data over HTTP.
- It may be easier to integrate with some platforms than NETCONF.

What must exist before it:

- HTTP client boundary with read-only method enforcement.
- Endpoint/path allowlists.
- Evidence storage for raw JSON or XML responses.
- Parser mappings for supported models.

Why it is not v0.1:

- It broadens protocol support before the initial evidence kernel is complete.
- It must not become a generic HTTP automation client.

### Terminal Server And Jump Host Collector

Why it matters:

- Many provider environments use jump hosts, bastions, or terminal servers for device access.
- Modeling access paths is useful for understanding how discovery should safely happen.

What must exist before it:

- Access path data model.
- Credential reference model.
- Strict approval and audit semantics.
- Clear separation between discovering an access path and using it.
- Read-only command policy enforced at the final device boundary.

Why it is not v0.1:

- Multi-hop access increases operational risk.
- The first kernel should prove direct, controlled, read-only evidence collection before adding access traversal.

## External System Connectors

### EMS Connectors

Why it matters:

- EMS and controllers often know managed devices, regions, platforms, and topology hints.
- They can help Truthwatcher discover how to discover without starting from device-by-device access.

What must exist before it:

- Importer/enricher contracts stable enough for external inventory.
- Evidence storage for API responses.
- Confidence model for EMS-provided facts.
- Adapter boundary that avoids coupling the kernel to one EMS.

Why it is not v0.1:

- EMS behavior is vendor-specific and organization-specific.
- The core model should remain stable before adding EMS assumptions.

### IPAM Connector

Why it matters:

- IPAM can provide prefixes, addresses, sites, tenants, VRFs, and ownership context.
- Comparing observed network evidence with IPAM data is central to source-of-truth bootstrapping.

What must exist before it:

- Generic IPAM adapter contract.
- Seeded/imported fact confidence semantics.
- Conflict handling between observed network evidence and imported IPAM context.
- Clear data mapping into assets, facts, relationships, and evidence.

Why it is not v0.1:

- Truthwatcher should not depend on one IPAM implementation.
- IPAM is context and evidence, not the kernel itself.

### Monitoring Connector

Why it matters:

- Monitoring systems can provide last-seen state, alarm context, management reachability, and device lists.
- This can enrich asset confidence and discovery planning.

What must exist before it:

- Enricher contract.
- Rules that prevent monitoring data from becoming the source of truth by default.
- Evidence storage for imported monitoring records.
- Clear distinction between enrichment and observability.

Why it is not v0.1:

- Monitoring can pull the project into alerting and observability.
- Observability is explicitly deferred until the evidence kernel is mature.

### Nautobot And NetBox Export

Why it matters:

- Many teams already use Nautobot or NetBox as a source of truth.
- Export can help bootstrap or reconcile those systems from observed evidence.

What must exist before it:

- Exporter contract.
- Stable mapping from Truthwatcher assets/facts/relationships to destination objects.
- Conflict and confidence rules for what is safe to export.
- Dry-run output and human review.

Why it is not v0.1:

- Export can be mistaken for authoritative synchronization.
- Truthwatcher must first prove its own evidence-backed model before pushing data outward.

## Access Control And Users

### OIDC Auth

Why it matters:

- OIDC would let teams integrate Truthwatcher with existing identity providers.
- It is a likely prerequisite for shared server deployments.

What must exist before it:

- Clear API boundary for authenticated identity.
- Session or token handling design.
- Local development mode that remains simple.
- Audit fields that can record a real user identity.

Why it is not v0.1:

- v0.1 is focused on local, single-binary workflows.
- Authentication adds operational and security complexity before the core model is finished.

### Multi-User RBAC

Why it matters:

- Shared deployments need role boundaries for viewing evidence, approving discovery, exporting data, and administering settings.

What must exist before it:

- Authenticated user identity.
- Permission model tied to current API actions.
- Audit storage that records decisions and actor context.
- Clear product decisions about roles.

Why it is not v0.1:

- RBAC without stable workflows usually becomes premature complexity.
- The project first needs stable evidence, discovery, and graph behavior.

## Service And Change Workflows

### Service Planning Workflows

Why it matters:

- Provider engineers reason in terms of services such as Internet access, L3VPN, E-Line, EVLAN, DIA, and managed CPE.
- Service planning can turn graph knowledge into operational planning.

What must exist before it:

- Service asset modeling.
- Device, site, port, and relationship confidence.
- Enough graph completeness to reason about paths and dependencies.
- Explicit separation between planning and device changes.

Why it is not v0.1:

- Service planning depends on trustworthy asset and relationship data.
- Building it too early would skip the evidence kernel.

### MOP Generation

Why it matters:

- Method-of-procedure generation could help engineers turn known network state into reviewable operational plans.

What must exist before it:

- Service planning model.
- Evidence-backed graph context.
- Human review workflow.
- Strong disclaimers and no automatic execution.

Why it is not v0.1:

- MOP generation is downstream of discovery, modeling, and service planning.
- It risks becoming automation theatre if the underlying graph is not trustworthy.

### Config Candidate Generation

Why it matters:

- Config candidates could eventually help translate intended service changes into reviewable vendor syntax.

What must exist before it:

- Intent/service model.
- Vendor-specific rendering boundaries.
- Human review and approval.
- Hard separation from device execution.
- Safety rules preventing write-capable network automation.

Why it is not v0.1:

- Configuration generation is not evidence collection.
- It is explicitly later than inventory, relationships, and service intent modeling.

## Optional Enrichment And Scale

### Observability Integration

Why it matters:

- Runtime state can enrich the graph with last-seen, alarm, maintenance, and availability context.

What must exist before it:

- Stable asset identity.
- Enricher contract.
- Rules that keep observability separate from source-of-truth evidence.
- Clear UI labeling for operational state versus modeled truth.

Why it is not v0.1:

- Observability is a non-goal for the initial kernel.
- Truthwatcher must not become an NMS or alert platform.

### Distributed Workers

Why it matters:

- Larger environments may eventually need distributed collection capacity or geographic execution points.

What must exist before it:

- Proven single-binary workflows.
- Durable audit and execution state.
- Clear job model.
- Authentication and authorization.
- Operational need that justifies extra moving parts.

Why it is not v0.1:

- Distributed workers imply coordination, deployment, and failure-mode complexity.
- The project constitution requires a single Go binary plus PostgreSQL unless explicitly changed.

### Graph Database Backend

Why it matters:

- A graph database may eventually help with deep path queries, dependency analysis, or service topology traversal.

What must exist before it:

- Evidence that PostgreSQL cannot satisfy required graph queries.
- Stable graph schema and query patterns.
- Migration or synchronization strategy.
- Clear operational justification.

Why it is not v0.1:

- PostgreSQL is the accepted primary database.
- The first problem is evidence correctness, not graph database selection.

### External LLM Integration

Why it matters:

- LLMs may help summarize evidence, explain graph context, or propose discovery plans in natural language.

What must exist before it:

- Grounded query layer with evidence references.
- Strict tool boundaries.
- No arbitrary command execution.
- Explicit configuration and opt-in.
- Redaction and data handling rules.

Why it is not v0.1:

- The existing agent workspace is deterministic by design.
- External LLMs could hallucinate or overreach unless the evidence kernel and policy boundaries are mature.

## Rule For Promoting Future Work

A future phase should become implementation work only when:

- the evidence-first chain remains intact
- the feature can be tested without real production devices by default
- safety policy applies before any collection action
- user approval is explicit for risky actions
- data is stored as evidence or clearly marked as seeded/inferred
- the single-binary plus PostgreSQL model still holds, or a new ADR explicitly changes it
