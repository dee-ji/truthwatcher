# Truthwatcher Codex Prompt Pack

This pack contains a sequenced set of prompts for building Truthwatcher without drifting into a generic CMDB, NMS, chat app, or automation tool.

## How to use

For every Codex session:

1. Open the project in GoLand.
2. Make sure these project steering files exist:
   - `steering-docs/PROJECT_TRUTHWATCHER.md`
   - `steering-docs/AGENT_COLLABORATION_CONTRACT.md`
   - `steering-docs/ARCHITECTURE_DECISIONS.md`
   - `steering-docs/DATA_MODEL.md`
   - `steering-docs/SAFETY_MODEL.md`
   - `steering-docs/MVP_SPEC.md`
   - `steering-docs/EXTENSIBILITY_MODEL.md`
   - `ROADMAP.md`
3. Copy one prompt at a time into Codex.
4. Review the diff.
5. Run tests.
6. Commit.
7. Do not run the next prompt until the current one is clean.

## Build philosophy

Truthwatcher must be built in small, boring, testable steps.

The system starts as:

> A single Go binary plus one PostgreSQL database that stores read-only discovery evidence and turns that evidence into assets, facts, and relationships.

It is not initially:

- Kubernetes
- Docker-first
- Observability
- Full AI agent platform
- Config deployment
- Remediation
- A better Nautobot
- A full dynamic schema engine

## Recommended branch strategy

Use one branch per phase:

```bash
git checkout -b phase-00-project-skeleton
```

Then merge after tests pass.
