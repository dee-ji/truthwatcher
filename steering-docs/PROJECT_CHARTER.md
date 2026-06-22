# TruthWatcher Project Charter

## One-Sentence Mission

TruthWatcher is a vendor-neutral source-of-truth, discovery, reasoning, and intent platform for complex service-provider networks. It turns raw evidence into identity, assets, facts, relationships, knowledge graphs, and operational understanding so teams can keep up with multi-vendor networks as they change.

## Core Problem

Large service-provider networks change faster than static inventory, tribal knowledge, manual discovery, and disconnected tooling can keep up. Operators may have monitoring, IP address records, EMS exports, spreadsheets, CMDBs, config archives, and vendor portals, yet still lack one trustworthy explanation of what exists, how it is connected, how it is accessed, which services depend on it, and which claims are stale or unproven.

The result is operational risk: failed automation, slow troubleshooting, low change confidence, duplicated engineering work, and planning decisions based on partial truth.

The hardest problem is not only discovering devices. The hardest problem is:

> Dynamically discovering how to discover, validate, relate, and reason over network reality.

TruthWatcher exists to make dynamically keeping up with multi-vendor networks a solved workflow rather than a recurring fire drill.

## What TruthWatcher Is

TruthWatcher is:

- A vendor-neutral evidence and source-of-truth platform.
- A read-only discovery and modeling kernel for complex networks.
- A graph-based relationship model for infrastructure, access paths, services, and intent.
- A reasoning foundation that can explain what the system believes and why.
- A proof-of-concept packaged as a single Go binary with embedded frontend assets and PostgreSQL storage.
- A CLI-first server application with explicit integration boundaries.
- A future network engineering workbench for safer automation, planning, and review.

## What TruthWatcher Is Not in the POC

TruthWatcher is not:

- A monitoring, alerting, or observability replacement.
- A configuration deployment, remediation, or service activation platform.
- A full NMS, CMDB, IPAM, EMS, or cloud inventory product.
- A clone of any existing source-of-truth system.
- A Kubernetes-first, Docker-first, or microservice-first application.
- A chat-first AI application.
- A system that runs arbitrary commands on network devices.
- A platform tied to one vendor, protocol, data source, or operational stack.

Those capabilities may become integrations or later phases, but they must not distract from the POC kernel: evidence-backed understanding.

## Product Philosophy

TruthWatcher is evidence-first and uncertainty-aware.

The system must preserve raw evidence before it creates identity, facts, assets, relationships, graph views, or answers. It must track where every conclusion came from, when the evidence was collected, how it was interpreted, and how confident the system is.

TruthWatcher should never pretend to know something without evidence.

TruthWatcher treats uncertainty as a first-class concept:

- Known.
- Unknown.
- Partially known.
- Conflicting evidence.
- Inferred.
- Seeded by a human or external system.
- Human-confirmed.

## Central Conceptual Pipeline

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

The named skills are conceptual responsibilities, not mandatory runtime plugins. The POC may implement them directly in the kernel; later versions may expose some of them through adapter or plugin contracts.

## Primary User

The primary user is a network engineer, network automation engineer, or infrastructure architect working in a large service-provider, carrier, MSP, enterprise, or hybrid network where inventory, access methods, and service relationships are incomplete, stale, or fragmented across teams and tools.

## Why Traditional Tools Have Not Solved This Completely

Traditional tools usually own one slice of truth:

- Monitoring knows symptoms but not full intent.
- IP address and inventory systems know assigned records but may not know observed state.
- EMS and controller systems know their managed domain but not the whole network.
- Config archives know historical text but not always identity, relationships, or confidence.
- Spreadsheets and tribal knowledge move quickly but do not provide durable provenance.
- Automation frameworks need truth before they can act safely.

TruthWatcher does not replace all of these systems. It relates their evidence and observed network data through a neutral model so engineers can understand where sources agree, disagree, or remain unknown.

## Initial Constraint

The proof of concept must do one thing well:

> Starting from a seed target or fixture and approved read-only collection, preserve evidence, derive identity, assets, facts, and relationships, project a graph, and expose the result through a simple API and UI with confidence and provenance.

## Build Principle

Do not build the dream first. Build the kernel.

The kernel must prove:

- I collected or imported this evidence.
- I derived this identity.
- I believe this asset exists.
- I extracted this fact.
- I believe this relationship exists.
- I can show why, including uncertainty and conflicts.

## Adaptability Principle

TruthWatcher must be designed as a stable kernel with adaptable edges.

The platform should never assume that one organization uses the same vendors, IPAM, CMDB, EMS, monitoring platform, credential vault, cloud provider, automation stack, or service model as another organization. Large service-provider environments are heterogeneous by default.

The core product owns:

- evidence storage
- identity lifecycle
- asset modeling
- fact modeling
- relationship modeling
- confidence and state
- discovery runs
- safety policy
- graph construction
- reasoning boundaries
- API contracts

External systems may provide or consume context through adapters, including DNS, DHCP, IP address management, monitoring, EMS/controller inventory, credential references, cloud inventory, vendor support records, config archives, service orders, circuit data, customer metadata, and existing source-of-truth platforms.

The rule is:

> Stable core, replaceable integrations.

If a feature only makes sense for one vendor, one company, one protocol, one source-of-truth product, one NMS, or one EMS, it belongs behind an adapter boundary.

## Path From POC To Enterprise Platform

TruthWatcher can become enterprise-ready by expanding from the kernel outward:

1. Prove evidence preservation, identity, fact extraction, relationships, and graph projection with fixtures and safe read-only collection.
2. Add human review for identity merges, conflicts, and seeded context.
3. Expand adapters without changing the core model.
4. Add policy, audit, credential-reference, and permission controls around discovery and imports.
5. Improve scale, HA, packaging, and operational hardening only after the model proves value.
6. Introduce planning, intent, service modeling, and automation guardrails on top of explainable understanding.
