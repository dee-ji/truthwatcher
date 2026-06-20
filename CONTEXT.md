# CONTEXT.md

## Purpose

This document captures the worldview, assumptions, priorities, constraints, and decision-making philosophy behind Truthwatcher.

It exists because architecture documents describe what the system does, but they often fail to capture why certain decisions were made.

Future contributors, maintainers, and AI coding agents should read this document before making significant architectural changes.

---

## What Truthwatcher Is

Truthwatcher is an evidence-first network knowledge platform designed for Internet Service Providers and other organizations operating large-scale networks.

Truthwatcher is not primarily an inventory platform.

Truthwatcher is not primarily an automation platform.

Truthwatcher is not primarily an observability platform.

Inventory, automation, topology, and planning are all outputs of a deeper goal:

Understanding the network.

Truthwatcher transforms discovered evidence into an explainable graph of physical and logical network reality.

The platform exists to answer questions such as:

* What exists?
* Why do we believe it exists?
* How is it connected?
* What evidence supports that conclusion?
* What remains unknown?
* What should be discovered next?

---

## The Core Problem

Large service provider environments rarely have a complete and trustworthy source of truth.

Knowledge becomes fragmented across:

* Spreadsheets
* EMS platforms
* IPAM systems
* Monitoring tools
* CMDBs
* Configuration archives
* Engineer tribal knowledge

Truthwatcher exists to continuously build and refine a network understanding model from evidence.

The system should always be capable of explaining why it believes something is true.

---

## Product Philosophy

Truthwatcher follows this chain:

Evidence → Identity → Assets → Facts → Relationships → Graph → Understanding

The graph is the product.

Inventory is a byproduct.

The purpose of the graph is not visualization.

The purpose of the graph is understanding.

---

## Target Audience

Primary users:

* Internet Service Providers
* Backbone operators
* Network automation engineers
* Network engineering teams
* Managed service providers
* Large enterprises with complex networks

The platform is optimized for environments containing:

* Core networks
* Backbone networks
* MPLS networks
* BGP-based infrastructures
* Carrier Ethernet services
* Multi-vendor environments

---

## Safety Philosophy

Safety always wins.

If safety and convenience conflict, safety wins.

If functionality and safety conflict, safety wins.

Truthwatcher should lose functionality before risking production impact.

The platform must be safe to run in production environments.

---

## Discovery Philosophy

Truthwatcher is not a router shell.

Users do not execute arbitrary commands.

AI agents do not execute arbitrary commands.

Discovery is task-based.

Examples:

* get_inventory
* get_neighbors
* get_bgp_summary
* get_interfaces
* get_arp
* get_ndp

Discovery tasks may map to vendor-specific commands internally, but users interact with approved discovery actions rather than raw CLI commands.

Custom discovery profiles are allowed.

Custom profiles should eventually support approval, signing, policy validation, and confidence scoring.

---

## Discovery Expansion Philosophy

Truthwatcher should not blindly scan networks.

Discovery should be graph-driven.

Discovery should begin from a trusted seed.

Examples:

* Seed device
* Seed subnet
* Seed location
* Seed inventory record

After discovery begins, Truthwatcher proposes additional discovery candidates based on evidence.

Examples:

* ARP entries
* NDP entries
* LLDP neighbors
* CDP neighbors
* BGP peers
* Static neighbor definitions
* Controller references

Truthwatcher should distinguish between:

* Provider-owned infrastructure
* Customer infrastructure
* Partner infrastructure
* External networks

before automatically expanding discovery.

For v0.1-alpha, Truthwatcher proposes candidates but does not automatically execute expanded discovery.

---

## Identity Philosophy

Physical identity is the foundation of truth.

Truthwatcher should prefer physical identity over logical identity whenever possible.

Identity strength hierarchy:

### Tier 1

* Vendor + Serial Number

### Tier 2

* System MAC Address

### Tier 3

* Asset Tags
* External Inventory IDs
* EMS Identifiers

### Tier 4

* Hostname
* Management IPv4 Address
* Management IPv6 Address
* DNS Records

### Tier 5

* Neighbor Relationships
* Location Information
* Protocol Participation

Serial number is the strongest identity anchor.

Hostnames and IP addresses are evidence.

They are not authoritative identity.

---

## Identity Candidates

Truthwatcher should not immediately assume a discovered object is a canonical asset.

Discovery should first create identity candidates.

Identity candidates may be:

* Accepted
* Rejected
* Deferred
* Require additional evidence

The system should prefer caution over destructive merges.

Automatic merges may exist in future releases but must always be explainable and evidence-backed.

---

## Truth Hierarchy

Observed network evidence is preferred over imported assertions.

Truthwatcher should classify information according to source and confidence.

Potential classifications:

* Observed Evidence
* Imported Evidence
* External Assertion
* Human Confirmed
* Inferred
* Historical

Imported systems are evidence providers.

Imported systems are not automatically authoritative.

Examples:

* IPAM
* EMS
* CMDB
* Monitoring
* CSV Imports
* NetBox
* Nautobot

If imported information conflicts with observed network evidence, observed evidence wins.

---

## History And Conflict Philosophy

Truthwatcher should preserve history.

Truthwatcher should preserve conflicting facts.

Truthwatcher should not hide uncertainty.

Conflicts are valuable information.

Users should be able to understand:

* What changed
* Why it changed
* When it changed
* What evidence supports the change

---

## Route Reflector Philosophy

A route reflector is both:

* A device
* A role

Roles should be modeled as facts.

Roles are not separate assets.

Relationships are more important than individual protocol sessions.

Truthwatcher should understand:

* Provider ASNs
* Customer ASNs
* Partner ASNs

as first-class concepts.

---

## Service Philosophy

Services are assets.

Examples:

* DIA
* E-Line
* EVLAN
* L3VPN
* Managed Router Services

Services may:

* Depend on circuits
* Depend on devices
* Depend on other services

Service history should be preserved.

A disconnected service becomes historical rather than disappearing.

---

## Circuit Philosophy

Circuits are first-class assets.

Circuits may be:

* Physical
* Logical

Truthwatcher should model:

* A Locations
* Z Locations
* Carrier References
* Vendor References

Services and circuits are strongly related concepts.

---

## Customer Philosophy

Customers are first-class assets.

Customers may:

* Own multiple services
* Depend on multiple circuits

In certain provider-to-provider scenarios, multiple customers may share service constructs.

Truthwatcher should eventually be capable of answering:

Which customers depend on this device?

and

Which customers are impacted by this failure?

---

## Location Philosophy

Locations are first-class assets.

Examples:

* Data Centers
* POPs
* Headends
* Hubs
* Central Offices
* Customer Premises

Locations may contain other locations.

Examples:

Market → Data Center → Cage → Rack

Racks should be modeled as assets.

Ports, cards, and optics should initially remain inventory records rather than top-level graph assets.

---

## Network Knowledge Philosophy

Truthwatcher is not trying to discover devices.

Truthwatcher is trying to discover network reality.

Understanding is more important than collection.

Relationships are more important than records.

Evidence is more important than assumptions.

The graph is more important than the tables.

The goal is to build an explainable model of the network that can continuously improve itself through evidence-driven discovery.

---

## v0.1.0-alpha.1

Milestone Name:

Network Knowledge Kernel

Goal:

Prove that Truthwatcher can transform discovered evidence into an explainable network knowledge graph.

Required workflow:

Fixture
→ Discovery
→ Raw Evidence
→ Identity Candidate
→ Asset
→ Fact
→ Relationship
→ Graph
→ UI

Alpha intentionally excludes:

* SSH discovery
* SNMP discovery
* NETCONF
* RESTCONF
* EMS integrations
* IPAM integrations
* Customer modeling
* Service modeling
* Circuit modeling
* AI reasoning
* Configuration generation
* MOP generation
* Credential vaults

These are future milestones.

---

## Long-Term Vision

Truthwatcher should eventually become capable of answering:

* What exists?
* How is it connected?
* What services depend on it?
* What customers depend on it?
* What should be discovered next?
* What is unknown?
* What changed?
* What would be impacted by a proposed change?
* How should this network evolve?

The long-term goal is a continuously improving network knowledge system built on evidence, relationships, and explainable reasoning.
