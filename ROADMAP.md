# Truthwatcher Roadmap

Truthwatcher is being built in phases.

Each phase should produce a working, testable outcome.

The goal is not to maximize features.

The goal is to maximize understanding.

Truthwatcher follows this chain:

```text
Evidence
    ↓
Identity
    ↓
Assets
    ↓
Facts
    ↓
Relationships
    ↓
Graph
    ↓
Understanding
```

⸻

## Phase 0 — OSS Foundation

### Goal

Create a maintainable open source project foundation.

### Success Criteria

A new contributor can clone, build, test, and understand the project.

### Tasks

#### Project Governance

* [x] Create CONTEXT.md
* [x] Create OSS_GOVERNANCE.md
* [x] Create RELEASE_STRATEGY.md
* [x] Create CONTRIBUTING.md
* [x] Create SECURITY.md
* [x] Create CODE_OF_CONDUCT.md

#### CI/CD

* [x] GitHub Actions test workflow
* [x] GitHub Actions lint workflow
* [x] GitHub Actions build workflow
* [x] GitHub Actions release workflow
* [x] Automatic changelog generation
* [x] Release artifact generation
* [x] Checksum generation

#### Repository Hygiene

* [x] Verify project license
* [x] Remove local IDE artifacts
* [x] Standardize issue templates
* [x] Standardize PR templates

### Completed

Phase 0 foundation is complete: governance documents exist, GPLv3 licensing is present, CI workflows validate test/lint/build, tagged releases generate changelogs, artifacts, and checksums, issue and PR templates are standardized, and local IDE artifacts have been removed from version control.

⸻

## Phase 1 — Network Knowledge Kernel

### Release

v0.1.0-alpha.1

### Goal

Prove that Truthwatcher can transform discovered evidence into an explainable network knowledge graph.

### Success Criteria

A fixture-based workflow can demonstrate:

```text
Fixture
    ↓
Discovery
    ↓
Raw Evidence
    ↓
Identity Candidate
    ↓
Asset
    ↓
Fact
    ↓
Relationship
    ↓
Graph
    ↓
UI
```

### Discovery

* Fixture-based discovery tasks
* Discovery task registry
* Discovery run tracking
* Discovery audit trail

### Evidence

* Raw evidence storage
* Evidence metadata
* Evidence provenance tracking
* Evidence retrieval API

### Identity

* Identity candidates
* Candidate confidence scoring
* Candidate review queue
* Candidate approval workflow
* Candidate rejection workflow

### Assets

* Asset creation
* Asset retrieval
* Asset search
* Asset history tracking

### Facts

* Fact persistence
* Fact provenance
* Fact confidence scoring
* Fact history

### Relationships

* Relationship persistence
* Relationship provenance
* Relationship visualization

### Graph

* Graph projection engine
* Graph query API
* Graph UI visualization

### User Interface

* Asset explorer
* Relationship explorer
* Evidence viewer
* Discovery viewer
* Graph visualization

### Testing

* End-to-end fixture workflow
* Documentation walkthrough
* Alpha acceptance test

⸻

## Phase 2 — Discovery Kernel

### Release

v0.1.0

### Goal

Introduce safe, real-world network discovery.

### Success Criteria

Truthwatcher can safely discover real devices using approved discovery tasks.

### Connectivity

* SSH transport
* Connection profiles
* Credential references
* Transport abstraction layer

### Discovery Tasks

* Inventory task
* Interface task
* ARP task
* NDP task
* BGP summary task
* LLDP task
* CDP task

### Safety

* Command allowlists
* Discovery policy engine
* Discovery scope controls
* Discovery approval workflows
* Discovery audit logging

### Vendors

#### Cisco

* IOS-XE support
* IOS-XR support

#### Juniper

* Junos support

### Candidate Expansion

* ARP candidates
* NDP candidates
* LLDP candidates
* CDP candidates
* BGP candidates

⸻

## Phase 3 — Topology Kernel

### Goal

Transform discovered relationships into network topology understanding.

### Success Criteria

Truthwatcher can explain how devices are connected.

### Layer 2

* LLDP topology
* CDP topology
* Port relationships

### Layer 3

* ARP relationships
* NDP relationships
* Routing relationships

### BGP

* BGP peer relationships
* ASN modeling
* Route reflector modeling
* Provider ASN modeling
* Customer ASN modeling
* Partner ASN modeling

### Ownership Boundaries

* Internal infrastructure classification
* Customer classification
* Partner classification
* External classification

### Visualization

* Layer 2 topology view
* Layer 3 topology view
* BGP topology view

⸻

## Phase 4 — Provider Domain Modeling

### Goal

Model the operational reality of service providers.

### Locations

* Location assets
* Hierarchical locations
* Markets
* Data centers
* POPs
* Customer premises
* Racks

### Customers

* Customer assets
* Customer relationships
* Customer ownership tracking

### Services

* Service assets
* Service history
* Service dependencies
* Service hierarchy

### Circuits

* Circuit assets
* A locations
* Z locations
* Carrier references
* Vendor references

### Impact Analysis

* Device to service mapping
* Service to customer mapping
* Impact visualization

⸻

## Phase 5 — Network Understanding

### Goal

Allow operators to understand network consequences and dependencies.

### Dependency Analysis

* Asset dependency graph
* Service dependency graph
* Customer dependency graph

### Questions Truthwatcher Should Answer

* What exists?
* Why do we believe it exists?
* What is unknown?
* What changed?
* What services depend on this asset?
* What customers depend on this asset?

### Change Planning

* Impact estimation
* Dependency tracing
* Failure domain analysis

⸻

## Phase 6 — Agentic Intelligence

### Goal

Provide safe reasoning and guidance based on discovered network knowledge.

### Discovery Guidance

* Discovery recommendations
* Discovery prioritization
* Discovery confidence analysis

### Reasoning

* Explain evidence chains
* Explain asset relationships
* Explain service dependencies

### Planning Assistance

* Change planning guidance
* MOP drafting assistance
* Risk assessment

### Safety

* Read-only reasoning guarantees
* Evidence-backed responses
* Explainable recommendations

⸻

## Future Considerations

The following are intentionally out of scope until earlier phases are complete.

* Observability
* Monitoring
* Alerting
* Configuration deployment
* Remediation
* Closed-loop automation
* Full AI orchestration
* Multi-cluster deployments
* Kubernetes-first deployments

Truthwatcher should earn complexity through successful milestones rather than assume it from the beginning.
