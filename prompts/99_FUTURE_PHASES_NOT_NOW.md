Before coding, read these files:
- steering-docs/PROJECT_TRUTHWATCHER.md
- steering-docs/AGENT_COLLABORATION_CONTRACT.md
- steering-docs/ARCHITECTURE_DECISIONS.md
- steering-docs/DATA_MODEL.md
- steering-docs/SAFETY_MODEL.md
- steering-docs/MVP_SPEC.md
- steering-docs/EXTENSIBILITY_MODEL.md
- ROADMAP.md

Act as an informed engineering sidekick. Challenge scope creep. Do not simply agree. If a requested implementation conflicts with the project constitution, stop and explain the conflict.

Rules:
- Keep Truthwatcher as a single Go binary plus PostgreSQL unless explicitly asked otherwise.
- Do not introduce Docker, Kubernetes, microservices, message brokers, or cloud dependencies.
- Do not introduce write-capable network automation.
- Do not build observability features.
- Do not build a full chat platform before the evidence kernel exists.
- Prefer boring, explicit Go code.
- Add tests where practical.
- Update ROADMAP.md only with completed work and next steps.
- Do not rewrite vision documents unless directly asked.


Task: document future phases but do not implement them.

Future candidates:
- NETCONF collector
- SNMP collector
- gNMI collector
- RESTCONF collector
- terminal server/jump host collector
- EMS connectors
- IPAM connector
- monitoring connector
- Nautobot/NetBox export
- OIDC auth
- multi-user RBAC
- service planning workflows
- MOP generation
- config candidate generation
- observability integration
- distributed workers
- graph database backend
- external LLM integration

Deliverables:
1. Add or update `docs/future.md`.
2. For each future phase, describe:
   - why it matters
   - what must exist before it
   - why it is not part of v0.1

Constraints:
- Do not implement any future phase.
