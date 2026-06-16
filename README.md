# Truthwatcher

Truthwatcher is a Go-based, single-binary, evidence-first network cartography platform for service-provider-style environments.

It stores raw read-only discovery evidence, derives facts from that evidence, models assets and relationships, and exposes the resulting knowledge through a local API and embedded UI.

Truthwatcher is early-stage. The current kernel focuses on safe collection, evidence storage, basic modeling APIs, graph projection, deterministic planning, and local fixture-backed workflows.

## Why Truthwatcher Exists

Network operators often have monitoring, alarms, SNMP polling, EMS platforms, spreadsheets, CMDBs, config archives, and tribal knowledge, but still cannot answer basic operational questions with confidence:

- What exactly exists?
- How do we access it?
- What is it connected to?
- Which service does it support?
- Which facts are verified by evidence?
- Which facts are stale, inferred, conflicting, or unknown?

Truthwatcher exists to collect safe read-only evidence and turn it into assets, facts, relationships, and graph views that explain what the system believes and why.

## What Truthwatcher Is

- A read-only network evidence engine.
- A source-of-truth bootstrap tool.
- A relational network graph model backed by PostgreSQL.
- A single Go binary with embedded migrations and embedded UI assets.
- A CLI and HTTP server for local, inspectable workflows.
- A foundation for safe discovery planning and future adapters.

## What Truthwatcher Is Not

- Not an observability, monitoring, or alerting platform.
- Not a configuration deployment or remediation tool.
- Not a Kubernetes, Docker, Helm, or microservice-first application.
- Not a replacement for NetBox, Nautobot, or an existing source of truth.
- Not a chat-first AI application.
- Not a system that runs arbitrary commands on network devices.

## Core Principle

```text
Evidence first.
Inventory second.
Relationships third.
Intent later.
Automation last.
```

Truthwatcher should not pretend something is true unless it can show supporting evidence or explicitly label the knowledge as seeded, inferred, conflicting, or unknown.

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

See [docs/install.md](docs/install.md) for the full local install flow.

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

## Target Milestone

The evidence kernel is moving toward this first complete workflow:

1. Start from one seed network device or fixture target.
2. Use an approved read-only discovery profile.
3. Store raw evidence before facts are created.
4. Parse basic identity and topology outputs.
5. Create assets and relationships with confidence and evidence references.
6. Display evidence, assets, and graph relationships in the UI.

Some pieces of this target exist today; others are intentionally tracked as separate roadmap work. See [steering-docs/ROADMAP.md](steering-docs/ROADMAP.md) for current completion status.

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

Truthwatcher keeps the early architecture deliberately boring:

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

PostgreSQL remains the only database. Graph relationships are modeled in relational tables first; a separate graph database is not part of the current kernel.

## Long-Term Direction

Truthwatcher may eventually support:

- architecture seeding questionnaires
- agentic discovery planning
- service-aware modeling
- MOP generation
- config candidate generation
- Nautobot or NetBox export
- EMS/controller integrations
- optional observability overlays

These are not all current capabilities. The v0.1 priority remains the evidence kernel.

## Concepts

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
./bin/truthwatcher dev check-knowledge
```

For optional local agent context from the sibling Mistspren repository, see [docs/local-knowledge.md](docs/local-knowledge.md). Mistspren is development-time context only and is not required for production runtime workflows.

## License

Truthwatcher is licensed under the GNU General Public License v3.0 or later.

Copyright (C) 2026 dee-ji

See [LICENSE](LICENSE) for the full license text.

## Additional Docs

- [Install](docs/install.md)
- [API](docs/api.md)
- [Testing](docs/testing.md)
- [Local Knowledge Providers](docs/local-knowledge.md)
- [Future Phases](docs/future.md)
- [Roadmap](steering-docs/ROADMAP.md)
- [Agent Collaboration Contract](steering-docs/AGENT_COLLABORATION_CONTRACT.md)
