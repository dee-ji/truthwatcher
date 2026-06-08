# Truthwatcher Extensibility Model

Truthwatcher must be adaptable enough for different organizations, vendors, service-provider architectures, and operational toolchains.

The project should be designed as a stable kernel with replaceable integration edges.

## Core Design Rule

Do not hard-code one organization's environment into the platform.

Truthwatcher should support service-provider patterns, but it must not assume a specific company's:

- IPAM
- CMDB
- monitoring system
- EMS
- credential vault
- vendor mix
- naming standard
- ASN structure
- service catalog
- topology design
- change-management system

## Stable Kernel

The kernel is the part of the system that should stay stable for years.

The kernel owns:

- Assets
- Facts
- Relationships
- Evidence
- Discovery runs
- Discovery profiles
- Access paths
- Credential references
- Confidence scoring
- Safety policy
- Audit records
- Graph construction
- API contracts

The kernel should not care whether a fact came from SSH, NETCONF, SNMP, IPAM, DNS, a monitoring system, an EMS, a spreadsheet, or a human-approved import.

Everything becomes evidence.

## Replaceable Edges

External systems should connect through adapters.

Examples:

- IPAM adapter
- DNS adapter
- DHCP adapter
- Monitoring adapter
- EMS adapter
- Cloud adapter
- Credential vault adapter
- Config archive adapter
- CMDB adapter
- Ticketing/change adapter
- Nautobot/NetBox exporter

Adapters should translate external data into Truthwatcher primitives:

- Evidence
- Assets
- Facts
- Relationships

Adapters should not directly mutate core records without going through the same evidence and confidence workflow as native discovery.

## Adapter Contract

Every adapter should answer these questions:

1. What external system does it connect to?
2. What read-only capabilities does it use?
3. What facts can it produce?
4. What assets can it identify?
5. What relationships can it infer?
6. What evidence does it store?
7. How confident are its findings?
8. What permissions does it require?
9. What data should never be imported?
10. How can the adapter be disabled safely?

## Integration Types

Truthwatcher should support multiple integration patterns.

### Pull Adapter

Truthwatcher connects to an external system and imports read-only data.

Examples:

- pull prefixes from IPAM
- pull device list from EMS
- pull alerts from monitoring
- pull DNS records

### Push Adapter

An external system sends data into Truthwatcher through an API.

Examples:

- webhook from a monitoring system
- external discovery script posts evidence
- asset import job posts vendor contract data

### Export Adapter

Truthwatcher sends modeled data to another system.

Examples:

- export devices to Nautobot
- export graph data to a data lake
- export service inventory to a reporting platform

### Script Adapter

A user brings their own script or executable.

The script receives structured input and returns structured output.

The script must return evidence, facts, and relationships in a documented schema.

### Plugin Adapter

A compiled or loaded Go plugin may add support for a new vendor, protocol, parser, or data source.

Plugins must obey the same safety policy and audit requirements as built-in collectors.

## Data Flow for Integrations

External system
  -> adapter
  -> raw evidence
  -> normalized facts
  -> assets and relationships
  -> graph
  -> API/UI/exports

The platform should never skip raw evidence storage.

## Bring-Your-Own-System Philosophy

Truthwatcher should allow users to bring:

- their own IPAM
- their own monitoring platform
- their own EMS
- their own scripts
- their own parsers
- their own credential provider
- their own service catalog
- their own source-of-truth system

Truthwatcher's job is to unify these inputs into an evidence-backed network knowledge graph.

## Monitoring Is Not Core v0.1

Monitoring and observability are valuable but are not part of the v0.1 kernel.

Monitoring integrations may eventually enrich the graph with operational context, such as:

- last seen timestamp
- alarm history
- maintenance state
- availability history
- known risk indicators

However, Truthwatcher must not become an NMS in early versions.

Observability is an enrichment layer, not the foundation.

## IPAM Is an Important Adapter, Not the Kernel

IPAM data is extremely useful, but Truthwatcher should not depend on a specific IPAM.

An IPAM adapter may provide:

- prefixes
- IP addresses
- VRFs
- sites
- tenants/customers
- reservations
- utilization

Truthwatcher should treat IPAM data as evidence and compare it against observed network evidence.

Example questions:

- IPAM says this prefix exists; is it routed?
- The network shows this connected subnet; does IPAM know it?
- This management IP exists on a device; who owns it?

## Extensibility Guardrails

Agents and contributors must follow these rules:

- Do not add a dependency on a specific external platform to the kernel.
- Do not couple core database tables to one vendor or one integration.
- Do not create special-case tables unless the concept is broadly reusable.
- Prefer stable primitives plus JSONB facts for unknown vendor-specific surface area.
- Treat integration data as evidence, not absolute truth.
- Make every integration optional.
- Make every integration auditable.
- Make every integration disableable.
- Keep the single-binary deployment model intact.

## First Adapter Targets

After the SSH discovery MVP, good adapter candidates are:

1. Static file import: YAML/JSON/CSV seed data
2. IPAM import: generic prefix/IP/site model
3. Nautobot/NetBox export
4. SNMP read-only collector
5. Config archive import
6. Monitoring enrichment adapter
7. EMS inventory adapter

Do not build all of these at once.

Build the adapter framework first, then add one simple adapter.
