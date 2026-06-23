# Evidence First

Truthwatcher stores raw evidence before it creates facts, assets, relationships, graph views, or answers.

This is the central product rule.

## Why Evidence Comes First

Network inventory is often incomplete, stale, or contradictory. Parsers can be wrong. Hostnames and IP addresses can move. EMS, IPAM, monitoring, and vendor records may disagree.

Raw evidence gives Truthwatcher a durable audit trail:

- what was collected
- where it came from
- when it was collected
- which method produced it
- which command or API returned it
- which parser interpreted it
- which facts or relationships refer back to it

Without raw evidence, a network model becomes another unsupported claim.

## Evidence In The Current Kernel

Evidence records include:

- discovery run ID
- target
- method
- command or API
- raw output
- raw output hash
- parser name, when known
- collection timestamp
- metadata

The raw output hash allows integrity checks and deduplication workflows later without changing the evidence contract.

## Evidence To Knowledge

Truthwatcher follows this chain:

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

The system should not skip steps. If identity, an asset, a fact, a relationship, a graph edge, or an answer exists, it should be explainable through evidence or explicitly marked as seeded, inferred, conflicting, or unknown. The skill names are conceptual responsibilities, not a requirement that v0.1 implement a runtime plugin system.

## Seeded Context Is Not Evidence

Users can seed architecture hints such as known ASNs, vendors, regions, EMS systems, services, or route reflectors. These hints are useful context for planning, but they are not observed network proof.

Seeded facts use a distinct source and confidence state so they are not confused with device evidence.

## Parser Failures

Parser failure must not destroy evidence. Raw evidence remains valuable even when current parsers cannot understand it. Future parsers should be able to reprocess stored evidence.

## User-Facing Rule

If Truthwatcher says it knows something, it should be able to show why.
