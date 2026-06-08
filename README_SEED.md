# TruthWatcher Seed README

TruthWatcher is a read-only network evidence and cartography engine for bootstrapping source-of-truth knowledge in large, complex networks.

It is designed for service-provider-style environments where inventory, access paths, EMS ownership, routing relationships, and physical asset identity are incomplete or fragmented.

## Why TruthWatcher Exists

Network operators often have monitoring, alarms, SNMP polling, EMS platforms, spreadsheets, CMDBs, and tribal knowledge, but still cannot answer simple questions confidently:

- What exactly exists?
- How do we access it?
- What is it connected to?
- What service does it support?
- Which facts are verified by evidence?
- Which facts are stale, inferred, or unknown?

TruthWatcher solves this by collecting read-only evidence and turning it into assets, facts, and relationships.

## Initial Architecture

```text
Single Go Binary
  ├── CLI
  ├── HTTP API
  ├── Embedded Web UI
  ├── Discovery Engine
  ├── Parser Registry
  ├── Policy Guardrails
  └── PostgreSQL
```

## Core Idea

```text
Evidence first.
Inventory second.
Relationships third.
Intent later.
Automation last.
```

## First Milestone

Given one seed network device, TruthWatcher should:

1. Connect read-only over SSH.
2. Run approved discovery commands.
3. Store raw evidence.
4. Parse identity and topology facts.
5. Create assets and relationships.
6. Display the results in a simple UI.

## Long-Term Direction

TruthWatcher may eventually support:

- Architecture seeding questionnaires.
- Agentic discovery planning.
- Service-aware modeling.
- MOP generation.
- Config candidate generation.
- Nautobot/NetBox export.
- EMS/controller integrations.
- Optional observability overlays.

But v0.1 is only the evidence kernel.


## AI Agent Collaboration

Before using Codex or another AI coding agent, include `AGENT_COLLABORATION_CONTRACT.md` in the context. This file defines how the agent should challenge ideas, preserve scope, and behave like an informed technical sidekick rather than a passive code generator.
