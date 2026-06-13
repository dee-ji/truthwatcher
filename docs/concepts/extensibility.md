# Extensibility

Truthwatcher is designed as a stable kernel with replaceable integration edges.

The kernel owns:

- evidence storage
- discovery runs
- assets
- facts
- relationships
- confidence
- safety policy
- graph projection
- API contracts

External systems should connect through adapters rather than becoming assumptions inside the kernel.

## Why This Matters

Service-provider environments are heterogeneous. One network may use NetBox, another may use an EMS, another may use spreadsheets, custom scripts, cloud APIs, config archives, or multiple IPAM systems.

Truthwatcher should not require one source-of-truth product, one vendor, one protocol, or one operational stack.

## Current Contracts

The project defines compile-time Go interfaces for:

- collectors
- parsers
- importers
- exporters
- enrichers
- planners

These are normal Go boundaries, not a dynamic plugin runtime.

## Adapters Treat Data As Evidence

Integrations should translate external data into Truthwatcher primitives:

```text
external system -> adapter -> evidence -> facts -> assets/relationships -> graph
```

Imported data should preserve source and confidence metadata. External records are useful, but they are not automatically more truthful than observed evidence.

## Current File Adapter

The first import/export foundation is local JSON. It is intentionally boring:

- export assets
- export facts
- export relationships
- export evidence metadata
- avoid exporting raw evidence output by default
- preserve source, confidence, state, and evidence references where possible

This proves the contract without depending on NetBox, Nautobot, IPAM, cloud APIs, or any other external system.

## BYO Script Boundary

Bring-your-own scripts are documented and disabled by default. If enabled locally, scripts must:

- be explicitly allowlisted
- run without a shell
- receive JSON input
- return JSON output
- obey timeouts
- pass policy checks
- return evidence or normalized candidates
- avoid network mutation, credential guessing, brute force, reloads, clears, commits, writes, deletes, and copies

Server mode must not allow arbitrary scripts by default.

## What Is Not Implemented Yet

Truthwatcher does not currently implement:

- dynamic plugin loading
- NetBox or Nautobot connectors
- IPAM connectors
- cloud inventory connectors
- EMS connectors
- monitoring enrichment connectors

Those are future adapter candidates. The current goal is to keep the kernel stable and make integration boundaries explicit.
