# Contributing to Truthwatcher

Truthwatcher is a single-binary, Go-based, evidence-first network cartography platform. Contributions should keep the project focused on the read-only evidence kernel before expanding into later roadmap phases.

## Before Coding

- Read `steering-docs/PROJECT_TRUTHWATCHER.md`, `steering-docs/AGENT_COLLABORATION_CONTRACT.md`, `steering-docs/ARCHITECTURE_DECISIONS.md`, `steering-docs/DATA_MODEL.md`, `steering-docs/SAFETY_MODEL.md`, `steering-docs/MVP_SPEC.md`, `steering-docs/EXTENSIBILITY_MODEL.md`, and `steering-docs/ROADMAP.md`.
- Read the active prompt from `prompts/`.
- Confirm the task fits the current roadmap phase.
- Challenge requests that add scope before the evidence kernel exists.

## Working Rules

- Execute one prompt per task.
- Keep commits small and reviewable.
- Modify only files required for the prompt.
- Prefer boring, explicit Go code.
- Add tests where practical.
- Do not introduce external dependencies unless the prompt requires them.

## Scope Guardrails

- Do not introduce Docker, Kubernetes, microservices, message brokers, or cloud dependencies.
- Do not build observability features in the early kernel phases.
- Do not build a full chat platform before evidence, facts, assets, relationships, and graph querying are working.
- Do not create application logic, database schema, or integrations outside the active prompt.

## Network Safety

- Truthwatcher discovery must remain read-only.
- Do not add device configuration changes, reloads, clears, deletes, copies, file transfers, or write-capable network automation.
- Do not allow arbitrary network commands.
- Commands and API calls must go through allowlisted discovery profiles and policy checks when those components exist.
- Raw evidence must be stored before parsed facts are created.
- Derived facts and relationships must link back to evidence when practical.

## Review Checklist

- The change fits the active prompt.
- The change preserves the single Go binary plus PostgreSQL architecture.
- The change preserves read-only safety.
- The change keeps evidence first.
- Tests were added or the reason for no tests is clear.
- `steering-docs/ROADMAP.md` was updated only for completed work and immediate next steps.
