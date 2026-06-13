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

## Kernel Extension Contracts

Truthwatcher defines extension contracts as Go interfaces first. These are compile-time contracts, not a dynamic plugin runtime.

The contract package must not make the kernel depend on any specific external system. A connector may depend on an external API client in its own package later, but the core kernel must continue to depend only on Truthwatcher primitives.

Initial contract types:

- `Collector`: gathers read-only raw outputs for approved discovery tasks.
- `Parser`: converts stored evidence into normalized model candidates without writing directly to the database.
- `Importer`: reads data from an external system and returns evidence-backed model candidates.
- `Exporter`: sends a kernel snapshot to another system when explicitly invoked.
- `Enricher`: proposes additional evidence-backed facts or relationships from existing kernel data plus optional external context.
- `Planner`: proposes safe next discovery steps but does not execute them.

All contracts must preserve these boundaries:

- Connectors return evidence, candidates, snapshots, or plans.
- The kernel owns persistence decisions.
- The kernel owns confidence scoring rules.
- The kernel owns conflict handling.
- The kernel owns policy enforcement for discovery execution.
- Planners must not execute.
- Importers and enrichers must not silently overwrite facts.
- Collectors must be read-only and policy-gated.
- Parsers must not require network access.
- Exporters must be optional and explicitly invoked.

Dynamic loading is intentionally deferred.

Do not add HashiCorp `go-plugin`, Go `plugin`, WASM loading, subprocess plugin execution, or marketplace-style runtime registration until the stable kernel contracts have been exercised by at least one boring in-repo adapter.

## Future Connector Examples

These examples describe expected boundaries only. They are not implementation commitments for the current milestone.

### IPAM Importer

Purpose:

- Import prefixes, IP addresses, VRFs, sites, tenants, reservations, and utilization metadata.

Expected behavior:

- Read from an IPAM system using explicit credentials or credential references.
- Store imported records as evidence.
- Produce assets and facts such as site, prefix, VRF, assigned IP, owner, and reservation state.
- Mark imported facts as sourced from the IPAM adapter, not as observed network truth.

Kernel boundary:

- Truthwatcher must not assume one IPAM product.
- IPAM data should be compared with observed routing/interface evidence, not treated as automatically correct.

### Monitoring Enricher

Purpose:

- Add operational context such as last seen time, maintenance state, or known alarm references.

Expected behavior:

- Read operational metadata only.
- Return enrichment candidates linked to external evidence.
- Avoid becoming alerting, polling, or observability inside Truthwatcher.

Kernel boundary:

- Monitoring data enriches the graph but does not turn Truthwatcher into an NMS.
- Availability and alarm workflows remain outside the kernel.

### EMS Importer Or Collector

Purpose:

- Import inventory and controller-managed relationships from an element management system.

Expected behavior:

- Use read-only EMS/API access.
- Store EMS responses as evidence.
- Produce assets for managed devices/controllers and relationships such as `managed_by`.
- Keep EMS-specific object details in JSONB metadata unless they become stable kernel nouns.

Kernel boundary:

- EMS access must require explicit configuration and approval.
- EMS hints do not authorize device access or scope expansion.

### Cloud API Importer

Purpose:

- Import cloud network inventory such as VPCs/VNets, subnets, gateways, route tables, interfaces, and tags.

Expected behavior:

- Use read-only cloud API permissions.
- Store API responses as evidence.
- Produce generic assets and relationships without coupling the kernel to one cloud provider.

Kernel boundary:

- Cloud provider SDKs belong in connector packages, not the kernel.
- Cloud resources should map to stable Truthwatcher primitives.

### Config Archive Importer Or Parser

Purpose:

- Import historical or current network configuration files from a read-only archive.

Expected behavior:

- Store each config file as evidence.
- Parse only safe facts and relationships.
- Preserve parser warnings when syntax is unsupported or ambiguous.

Kernel boundary:

- Config archives are evidence, not live device state.
- Imported config must not trigger network actions.

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

### File JSON Adapter

The first boring adapter is a local JSON file importer/exporter.

Purpose:

- Prove import/export contracts without depending on NetBox, Nautobot, IPAM, cloud APIs, or another external platform.
- Allow engineers to move a bounded Truthwatcher graph snapshot between environments for testing and review.

Export behavior:

- Exports assets.
- Exports facts.
- Exports relationships.
- Exports evidence metadata only.
- Does not export raw evidence output.

Import behavior:

- Imports the same JSON snapshot format.
- Preserves fact source, confidence, state, evidence references, and metadata where present.
- Preserves evidence metadata such as target, method, command/API, hash, parser name, collection time, and metadata.
- Creates evidence candidates only when raw output is explicitly present in the imported file.

Kernel boundary:

- The adapter returns candidates; the kernel still owns validation, conflict handling, and persistence.
- Imported file data is evidence-backed context, not automatically observed truth.
- The adapter does not implement NetBox, Nautobot, IPAM, monitoring, EMS, or cloud integration.

### Script Adapter

A user brings their own script or executable.

The script receives structured input and returns structured output.

The script must return evidence, facts, and relationships in a documented schema.

BYO script execution is high risk if treated casually. Truthwatcher must keep it local, explicit, and disabled by default.

Script runner rules:

- Server mode must not allow arbitrary script execution by default.
- A local caller must explicitly enable the script runner.
- A local caller must allowlist the exact script path.
- The runner must execute the script directly, not through a shell.
- The runner must pass input JSON on stdin.
- The script must return output JSON on stdout.
- The runner must enforce a timeout.
- The runner must validate requested tasks through the policy engine before execution.
- The runner must validate returned evidence `command_or_api` values through the policy engine.
- Scripts must not mutate network state.
- Scripts must not guess credentials.
- Scripts must not perform brute force, scans, config changes, clears, reloads, commits, copies, deletes, or writes.
- Scripts must return evidence or normalized candidates; arbitrary logs are not valid output.

Input JSON:

```json
{
  "target": "fixture://junos-mx",
  "method": "script",
  "profile": "juniper_junos",
  "tasks": ["identify_device"],
  "credential_ref": "optional-local-reference",
  "dry_run": true,
  "context": {
    "note": "adapter-specific context"
  }
}
```

Output JSON:

```json
{
  "evidence": [
    {
      "target": "fixture://junos-mx",
      "method": "script",
      "command_or_api": "show version",
      "raw_output": "Hostname: fixture-junos-mx",
      "metadata": {
        "script": "emit_static_version.sh"
      }
    }
  ],
  "candidates": {
    "facts": [
      {
        "asset_id": "asset-placeholder",
        "name": "hostname",
        "value": "fixture-junos-mx",
        "source": "byo_script",
        "confidence": 0.4,
        "confidence_reason": "returned by local BYO script",
        "state": "user_seeded"
      }
    ]
  },
  "warnings": [
    "script output is untrusted until persisted by the kernel"
  ]
}
```

Exit codes:

- `0`: script completed and stdout contains valid output JSON.
- Non-zero: script failed; stderr may contain a short diagnostic.
- Timeout: runner kills the process and treats the run as failed.

Evidence-first rule:

- If the script learned something from an external source, it should return raw evidence.
- Normalized facts or relationships should be linked to evidence when possible.
- Seeded or imported facts must not pretend to be observed device evidence.

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
