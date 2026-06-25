# Proof-Of-Concept Walkthrough

This walkthrough shows one complete fixture-backed Truthwatcher flow from evidence collection to identity, assets, facts, relationships, graph view, and reasoning output.

For a copy-paste v0.1.0 local setup, see [v0.1.0 Quickstart](v0.1.0-quickstart.md).


The fixture and platform names in this document are examples only. They prove the POC workflow without making Truthwatcher a Junos, Cisco, Arista, Nautobot, NetBox, IPAM, EMS, monitoring, cloud, or any other vendor-specific project. In production, the same pipeline should accept evidence from replaceable profiles, parsers, imports, and adapters.

## What This Walkthrough Proves

Truthwatcher exists to make dynamically keeping up with multi-vendor networks a repeatable workflow. This POC walkthrough proves the core loop:

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

The important claim is not that one fixture is special. The important claim is that Truthwatcher can preserve evidence, derive model objects, keep provenance attached, expose uncertainty, and explain why it believes something about the network.

## Preconditions

Build the binary, create a local PostgreSQL database, run migrations, and start the server as described in the local install guide:

```sh
make build
createdb truthwatcher
export TRUTHWATCHER_DATABASE_URL='postgres://localhost/truthwatcher?sslmode=disable'
./bin/truthwatcher migrate up
./bin/truthwatcher server
```

The server defaults to:

```text
http://127.0.0.1:8080
```

## Step 1: Collect Fixture Evidence

Run fixture-backed discovery:

```sh
./bin/truthwatcher discover fake --target fixture://junos-mx
```

This does not touch a network. It uses local fixture data to exercise the same evidence-first shape that a real read-only collector or import adapter must follow.

Expected result:

- A discovery run is created.
- Raw evidence is persisted before any model record is trusted.
- Evidence metadata records target, method, command or API context, timestamps, parser hints when present, and raw output hash.

Inspect discovery runs:

```sh
curl http://127.0.0.1:8080/api/v1/discovery-runs
```

Inspect evidence for a run:

```sh
curl http://127.0.0.1:8080/api/v1/discovery-runs/<discovery-run-id>/evidence
curl http://127.0.0.1:8080/api/v1/evidence/<evidence-id>
```

Pipeline state:

```text
Fixture-backed collector -> Evidence
```

## Step 2: Derive Identity From Evidence

Parse the stored discovery run:

```sh
./bin/truthwatcher parse discovery-run --id <discovery-run-id> --platform junos
```

Or use the API:

```sh
curl -X POST http://127.0.0.1:8080/api/v1/discovery-runs/<discovery-run-id>/parse \
  -H 'Content-Type: application/json' \
  -d '{"platform":"junos"}'
```

The `junos` value here selects a sample parser for sample fixture output. It is not a product dependency. Other parsers, imports, and adapters should map their source-specific evidence into the same Truthwatcher primitives.

Expected result:

- Parsers read stored evidence rather than re-running discovery.
- Identity clues such as hostnames, serials, system MACs, or other stable identifiers become identity candidates or asset identity material.
- Weak, conflicting, or provisional identity remains visible instead of being silently promoted to truth.

Pipeline state:

```text
Evidence -> Identity
```

## Step 3: Create Assets

After identity is derived, Truthwatcher can create or update assets with confidence, state, and evidence references.

Inspect assets:

```sh
curl http://127.0.0.1:8080/api/v1/assets
curl http://127.0.0.1:8080/api/v1/assets/<asset-id>
```

Expected result:

- Device or component assets appear when the fixture parser supports them.
- Assets retain enough provenance to answer why they exist.
- Vendor-specific details remain data or metadata, not separate vendor-specific core tables.

Pipeline state:

```text
Evidence -> Identity -> Assets
```

## Step 4: Extract Facts

Facts are evidence-backed statements about assets. Examples include observed software version, model, serial, interface attributes, inventory details, or other parser-supported claims.

Inspect facts for an asset:

```sh
curl http://127.0.0.1:8080/api/v1/assets/<asset-id>/facts
```

Expected result:

- Facts point back to supporting evidence where available.
- Facts carry confidence and state so stale, inferred, conflicting, or unknown information is not hidden.
- Parser limitations do not destroy evidence; future parser improvements can reprocess stored evidence.

Pipeline state:

```text
Evidence -> Identity -> Assets -> Facts
```

## Step 5: Build Relationships

Relationships connect assets, facts, and topology observations into a navigable model. Examples include device-to-component, interface-to-neighbor, asset-to-service, or other relationships as parser and service-model support expands.

Inspect relationships:

```sh
curl http://127.0.0.1:8080/api/v1/assets/<asset-id>/relationships
```

Expected result:

- Relationships are created only when evidence or clearly labeled seeded context supports them.
- Relationship records preserve confidence, state, and evidence references.
- The model remains neutral even when the source evidence came from a vendor-specific CLI, protocol, API, file, or external system.

Pipeline state:

```text
Evidence -> Identity -> Assets -> Facts -> Relationships
```

## Step 6: View The Knowledge Graph

Graph views are projections of evidence-backed assets, facts, and relationships. They are not a separate source of truth; they are a way to understand and navigate the modeled network.

Inspect graph data:

```sh
curl http://127.0.0.1:8080/api/v1/assets/<asset-id>/graph
curl 'http://127.0.0.1:8080/api/v1/graph/neighbors?asset_id=<asset-id>'
```

Open the embedded UI:

```text
http://127.0.0.1:8080/#/assets
```

Expected result:

- Operators can move from an asset to nearby graph context.
- Graph nodes and edges remain explainable through evidence, facts, relationships, and confidence state.
- The UI makes the modeled result easier to inspect without hiding the evidence chain.

Pipeline state:

```text
Evidence -> Identity -> Assets -> Facts -> Relationships -> KnowledgeGraph
```

## Step 7: Produce Reasoning Output

In the POC, reasoning means explainable understanding, not autonomous action. A reasoning output should summarize what Truthwatcher believes, why it believes it, what evidence supports it, and what remains unknown.

A useful POC reasoning answer for an asset should be able to say:

- This asset exists because these evidence records produced these identity signals.
- These facts were extracted from these parser-supported outputs.
- These relationships are supported by this evidence or are clearly marked as seeded/inferred.
- These claims are strong, weak, conflicting, unknown, or awaiting review.
- The next safe discovery step would be to collect additional evidence, not to assume missing truth.

Today, this reasoning can be assembled by reading API and UI views together. Future work may expose a dedicated reasoning endpoint or assistant surface, but it must still be grounded in evidence and must not bypass human review or safety policy.

Pipeline state:

```text
Evidence -> Identity -> Assets -> Facts -> Relationships -> KnowledgeGraph -> Understanding
```

## Why Fixture Names Do Not Define The Product

This walkthrough uses `fixture://junos-mx` and `--platform junos` because a concrete example makes the POC easy to run and test. Those names are intentionally narrow examples.

Truthwatcher's product boundary is the evidence-to-understanding pipeline, not a vendor parser. A different environment should be able to add adapters for other CLIs, APIs, controllers, files, source-of-truth systems, IP address tools, monitoring exports, cloud inventories, or human-reviewed imports without changing the core model.

The rule is:

```text
Stable core, replaceable integrations.
```

## POC Completion Checklist

Use this checklist to decide whether a fixture-backed walkthrough demonstrates the intended proof of concept:

- Raw evidence is stored before parsing.
- Identity is derived from evidence or explicitly marked as seeded, inferred, conflicting, weak, or unknown.
- Assets are created with confidence and provenance.
- Facts are linked to evidence where available.
- Relationships are linked to evidence or clearly labeled context.
- Graph views can be navigated from assets and relationships.
- Reasoning output explains what is known, why it is known, and what remains uncertain.
- Vendor-specific behavior remains isolated to fixtures, profiles, parsers, or adapters.
