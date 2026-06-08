# TruthWatcher Project Charter

## One-Sentence Mission

TruthWatcher is a single-binary, Go-based, read-only network cartography and source-of-truth bootstrap platform that transforms discovered evidence into a dynamic model of network assets, relationships, access paths, services, and intent.

## Core Problem

Large service-provider networks often do not have a trustworthy, complete, dynamic understanding of what exists, how it is connected, how it is accessed, what manages it, and which services depend on it.

The hardest problem is not only discovering devices. The hardest problem is:

> Discovering how to discover.

TruthWatcher exists to help engineers initialize and continuously improve a source-of-truth model for complex networks using safe, read-only evidence collection.

## What TruthWatcher Is

TruthWatcher is:

- A network evidence engine.
- A network cartography system.
- A source-of-truth bootstrap platform.
- A graph-based relationship model for network infrastructure.
- A service-provider inventory and modeling foundation.
- A read-only discovery framework.
- A future agentic network engineering workbench.
- A single Go binary with embedded frontend assets.
- A CLI-first server application inspired by tools like HashiCorp Vault.

## What TruthWatcher Is Not in v0.1

TruthWatcher is not:

- An observability platform.
- An alarm system.
- A monitoring system.
- A config deployment system.
- A remediation platform.
- A full NMS.
- A replacement for Nautobot or NetBox.
- A Kubernetes-first application.
- A Docker-first application.
- A chat-first AI app in the first milestone.

Observability, chat, service activation, and remediation may become later phases, but they must not distract from the v0.1 kernel.

## Product Philosophy

TruthWatcher is evidence-first.

The system must store raw evidence before it creates facts. It must track where every fact came from, when it was collected, how it was parsed, and how confident the system is.

TruthWatcher should never pretend to know something without evidence.

TruthWatcher treats uncertainty as a first-class concept:

- Known.
- Unknown.
- Partially known.
- Conflicting evidence.
- Inferred.
- Human-confirmed.

## Primary User

The primary user is a network engineer or network automation engineer working in a large service-provider, carrier, MSP, enterprise, or hybrid infrastructure environment where inventory and access methods are incomplete or fragmented.

## Long-Term Vision

TruthWatcher should allow an engineer to eventually ask:

- What devices exist in this market?
- How do I log into this device?
- Is this device behind an EMS?
- What route reflectors know about this PE?
- What services depend on this asset?
- What chassis, cards, ports, and optics are installed?
- Which sites are connected by this service?
- What is unknown or contradictory about this network?
- Help me plan an E-Line or EVLAN between these two locations.
- Generate a method of procedure using discovered network knowledge.

## Initial Constraint

The first version must do one thing well:

> Given a seed network device and read-only access, collect evidence, parse basic identity and topology facts, store assets and relationships, and expose them through a simple API and UI.

## Build Principle

Do not build the dream first. Build the kernel.

The kernel must prove:

- I collected this evidence.
- I parsed this fact.
- I believe this asset exists.
- I believe this relationship exists.
- Here is why.
