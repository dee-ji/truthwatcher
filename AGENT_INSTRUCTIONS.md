# Agent Instructions for Codex/AI Builders

## Required Reading

Before modifying code, read these files:

1. PROJECT_TRUTHWATCHER.md
2. ROADMAP.md
3. ARCHITECTURE_DECISIONS.md
4. DATA_MODEL.md
5. SAFETY_MODEL.md
6. MVP_SPEC.md

## Prime Directive

Do not turn this project into a generic AI app, chat app, NMS, observability system, or config automation platform.

The first product is a read-only network evidence and cartography kernel.

## Implementation Style

Use Go.

Prefer:

- Simple standard-library-first design.
- Clear interfaces.
- Small packages.
- Explicit errors.
- Structured logging.
- PostgreSQL.
- Embedded frontend assets.
- CLI-first server UX.

Avoid:

- Premature Kubernetes.
- Premature Docker requirement.
- Premature microservices.
- Dynamic table generation.
- Arbitrary command execution.
- Storing raw secrets.
- Large unreviewable changes.

## Recommended Package Layout

```text
cmd/truthwatcher/
internal/api/
internal/app/
internal/config/
internal/db/
internal/migrations/
internal/discovery/
internal/collectors/ssh/
internal/policy/
internal/parsers/
internal/graph/
internal/assets/
internal/evidence/
internal/logging/
internal/ui/
web/
docs/
examples/
```

## Build One Small Thing at a Time

Good task prompt:

> Implement the DiscoveryRun database model, migration, and repository. Do not modify unrelated packages. Follow DATA_MODEL.md and MVP_SPEC.md.

Bad task prompt:

> Build the whole agentic network discovery platform.

## Commit Discipline

Each change should be small and should answer:

- What did this add?
- Why does it belong in v0.1?
- Does it preserve read-only safety?
- Does it maintain evidence-first design?

## Guardrail Checklist

Before adding a feature, verify:

- Does it support evidence collection, asset modeling, relationships, or safe discovery?
- Is it needed for the MVP?
- Does it avoid write operations to network devices?
- Does it avoid storing raw secrets?
- Does it link facts back to evidence?

If not, defer it.

## Agent Workflow Rule

Agents may eventually propose tasks, but deterministic code must execute them.

Agent:

```text
I recommend running identify_device against target X using profile Y.
```

Policy engine:

```text
Approved or denied.
```

Collector:

```text
Runs only approved read-only commands.
```

## Do Not Hallucinate Vendor Coverage

Only claim support for a vendor/platform if sample outputs and parser tests exist.

Add unsupported vendors as discovery profile placeholders only.

## Extensibility Instructions

Truthwatcher must remain adaptable to many environments.

When implementing features, agents must preserve the stable-kernel / replaceable-edge architecture.

Before adding support for any external system, ask:

- Is this part of the kernel, or is it an adapter?
- Can this be disabled?
- Does this force one company's workflow into the core product?
- Does this store raw evidence before normalized facts?
- Does this preserve the single Go binary deployment goal?

Examples:

- IPAM support must be an adapter.
- Monitoring support must be an adapter/enrichment layer.
- EMS support must be an adapter.
- Vendor command support must live in discovery profiles/parsers.
- User scripts must plug into a documented adapter contract.

Do not turn Truthwatcher into a tightly coupled platform that only works with one ecosystem.
