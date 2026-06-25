# Truthwatcher

Truthwatcher is a vendor-neutral source-of-truth, discovery, reasoning, and intent platform for complex service-provider networks.

It continuously turns raw network evidence into identity, assets, facts, relationships, graph views, and operational understanding so engineering teams can keep up with multi-vendor networks without depending on static inventory, tribal knowledge, or disconnected tools.

Truthwatcher is early-stage proof-of-concept software. The current kernel focuses on safe evidence collection, evidence storage, basic modeling APIs, graph projection, deterministic planning, local fixture-backed workflows, and clear adapter boundaries that keep vendor-specific behavior outside the core model.

## Why Truthwatcher Exists

Network operators often have monitoring, alarms, SNMP polling, EMS platforms, spreadsheets, CMDBs, config archives, and tribal knowledge, but still cannot answer basic operational questions with confidence:

- What exactly exists?
- How do we access it?
- What is it connected to?
- Which service does it support?
- Which facts are verified by evidence?
- Which facts are stale, inferred, conflicting, or unknown?

Truthwatcher exists to make the problem of dynamically keeping up with multi-vendor networks a thing of the past. It collects safe read-only evidence, validates and models what the evidence says, relates the results into a graph, and makes every conclusion explainable.

## What Truthwatcher Is

- A vendor-neutral network evidence and source-of-truth platform.
- A discovery and reasoning kernel for complex service-provider networks.
- A relationship model that links evidence, identity, assets, facts, and intent.
- A local proof-of-concept packaged as a single Go binary with embedded migrations and UI assets.
- A CLI and HTTP server for inspectable, incremental workflows.
- A foundation for safe discovery planning, human review, and replaceable integrations.

## What Truthwatcher Is Not

- Not an observability, monitoring, or alerting platform.
- Not a configuration deployment or remediation tool.
- Not a Kubernetes, Docker, Helm, or microservice-first application.
- Not tied to any one source-of-truth product, vendor ecosystem, protocol, cloud, IPAM, EMS, or monitoring stack.
- Not a chat-first AI application.
- Not a system that runs arbitrary commands on network devices.

## Core Principle

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

Truthwatcher should not pretend something is true unless it can show supporting evidence or explicitly label the knowledge as seeded, inferred, conflicting, or unknown. Skills in this pipeline are conceptual responsibilities; implementations may be built into the kernel or provided through adapters as the proof of concept matures.

## Quickstart

Build the binary:

```sh
make build
```

Create a local PostgreSQL database:

```sh
createdb truthwatcher
export TRUTHWATCHER_DATABASE_URL='postgres://localhost/truthwatcher?sslmode=disable'
```

Run embedded migrations:

```sh
./bin/truthwatcher migrate up
```

Start the server:

```sh
./bin/truthwatcher server
```

Open the embedded UI:

```text
http://127.0.0.1:8080
```

Run fake fixture-backed discovery without touching a network:

```sh
./bin/truthwatcher discover fake --target fixture://junos-mx
```

Register a local device seed without running discovery. Vendor fields are optional metadata and do not bind Truthwatcher to a vendor-specific model:

```sh
./bin/truthwatcher devices add --hostname edge-01 --management-ip 192.0.2.10 --vendor example-vendor --role edge --site lab
./bin/truthwatcher devices list
```

See [docs/v0.1.0-quickstart.md](docs/v0.1.0-quickstart.md) for the copy-paste v0.1.0 local workflow, [docs/v0.1.0-known-limitations.md](docs/v0.1.0-known-limitations.md) for release boundaries, and [docs/install.md](docs/install.md) for the full local install flow.

## Example Workflow

1. Start PostgreSQL, export `TRUTHWATCHER_DATABASE_URL`, and run migrations.
2. Run fake discovery from fixtures:

```sh
./bin/truthwatcher discover fake --target fixture://junos-mx
```

3. Inspect discovery runs:

```sh
curl http://127.0.0.1:8080/api/v1/discovery-runs
```

4. Inspect evidence for a run:

```sh
curl http://127.0.0.1:8080/api/v1/discovery-runs/<discovery-run-id>/evidence
curl http://127.0.0.1:8080/api/v1/evidence/<evidence-id>
```

5. Parse stored evidence into assets, facts, and relationships:

```sh
./bin/truthwatcher parse discovery-run --id <discovery-run-id> --platform junos
```

Or through the API:

```sh
curl -X POST http://127.0.0.1:8080/api/v1/discovery-runs/<discovery-run-id>/parse \
  -H 'Content-Type: application/json' \
  -d '{"platform":"junos"}'
```

6. Inspect graph data when assets and relationships exist:

```sh
curl http://127.0.0.1:8080/api/v1/assets
curl http://127.0.0.1:8080/api/v1/assets/<asset-id>/graph
curl 'http://127.0.0.1:8080/api/v1/graph/neighbors?asset_id=<asset-id>'
```

You can also browse assets, facts, relationships, and linked read-only evidence in the embedded UI:

```text
http://127.0.0.1:8080/#/assets
```

Review safe discovery plan suggestions without executing them:

```text
http://127.0.0.1:8080/#/discovery-plans
```

Seed architecture context without treating it as observed proof:

```text
http://127.0.0.1:8080/#/architecture-seeds
```

Current limitation: fake discovery stores raw evidence first. Parser persistence is an explicit second step so raw evidence is preserved even when parsing produces warnings or skips unsupported commands.

## Proof-Of-Concept Scope

The proof of concept is intentionally narrow but strategically complete. It must demonstrate the full evidence-to-understanding loop without pretending to be an enterprise platform on day one:

1. Start from one seed network device or fixture target.
2. Use an approved read-only discovery profile.
3. Store raw evidence before facts are created.
4. Parse basic identity and topology outputs.
5. Create assets and relationships with confidence and evidence references.
6. Display evidence, assets, and graph relationships in the UI.

Some pieces of this target exist today; others are intentionally tracked as separate roadmap work. The POC is successful when it proves that raw evidence can become explainable network understanding through a repeatable, vendor-neutral pipeline. See [ROADMAP.md](ROADMAP.md) for current completion status.

## Safety Model

Truthwatcher is read-only by design.

- Discovery commands must come from built-in safe profiles.
- Dangerous command patterns such as `configure`, `commit`, `delete`, `reload`, `clear`, `write memory`, `copy`, and reboot requests are denied.
- Fake discovery uses local fixture files and does not touch a network.
- Optional SSH collection is behind the collector boundary and must pass policy checks before execution.
- Chat or agent-style features do not execute discovery or network actions.
- Seeded architecture hints are context, not observed proof.

See [docs/concepts/evidence-first.md](docs/concepts/evidence-first.md) and [steering-docs/SAFETY_MODEL.md](steering-docs/SAFETY_MODEL.md).

## Architecture

Truthwatcher keeps the early architecture deliberately boring so the hard problem remains the model, not the deployment topology:

```text
single Go binary
  -> CLI commands
  -> HTTP API
  -> embedded UI
  -> embedded migrations
  -> PostgreSQL
```

Core packages model:

- discovery runs
- raw evidence
- assets
- facts
- relationships
- graph projections
- safety policy
- deterministic planning
- compile-time extensibility contracts

PostgreSQL remains the only database in the POC. Graph relationships are modeled in relational tables first; a separate graph database is not part of the current kernel unless a later architecture decision proves it is necessary.

## Long-Term Direction

Truthwatcher evolves from POC to enterprise-ready platform by expanding adapters, review workflows, policy controls, scale characteristics, and integration surfaces while preserving the evidence-first kernel. Later phases may support:

- architecture seeding questionnaires
- agentic discovery planning
- service-aware modeling
- MOP generation
- config candidate generation
- source-of-truth export adapters
- EMS/controller integrations
- optional observability overlays

These are not all current capabilities. The v0.1 priority remains the evidence kernel.

## Concepts

- [v0.1.0 Quickstart](docs/v0.1.0-quickstart.md)
- [v0.1.0 Known Limitations](docs/v0.1.0-known-limitations.md)
- [POC Walkthrough](docs/poc-walkthrough.md)
- [Evidence First](docs/concepts/evidence-first.md)
- [Discover How To Discover](docs/concepts/discover-how-to-discover.md)
- [Assets, Facts, Relationships](docs/concepts/assets-facts-relationships.md)
- [Extensibility](docs/concepts/extensibility.md)

## Useful Commands

```sh
./bin/truthwatcher --help
./bin/truthwatcher version
./bin/truthwatcher server --help
./bin/truthwatcher migrate --help
./bin/truthwatcher discover fake --help
./bin/truthwatcher devices add --help
./bin/truthwatcher devices list --help
./bin/truthwatcher dev check-knowledge
```

For optional local agent context from the sibling Mistspren repository, see [docs/local-knowledge.md](docs/local-knowledge.md). Mistspren is development-time context only and is not required for production runtime workflows.

## License

Truthwatcher is licensed under the GNU General Public License v3.0 or later.

Copyright (C) 2026 dee-ji

See [LICENSE](LICENSE) for the full license text.

## Additional Docs

- [v0.1.0 Quickstart](docs/v0.1.0-quickstart.md)
- [v0.1.0 Known Limitations](docs/v0.1.0-known-limitations.md)
- [Install](docs/install.md)
- [API](docs/api.md)
- [Testing](docs/testing.md)
- [Local Knowledge Providers](docs/local-knowledge.md)
- [Future Phases](docs/future.md)
- [Roadmap](ROADMAP.md)
- [Agent Collaboration Contract](steering-docs/AGENT_COLLABORATION_CONTRACT.md)


## Audit inspection

Truthwatcher includes read-only audit inspection for discovery execution through `GET /api/v1/audit-records` and the embedded `#/audit` UI page. See [docs/audit.md](docs/audit.md).
