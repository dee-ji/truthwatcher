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


Task: add a simple UI for architecture seeding.

Deliverables:
1. Add a form for organization/network type, known ASNs, route reflectors, vendors, EMS systems, services, and regions/markets.
2. Submit hints to the existing architecture seeding API.
3. Show a clear warning that seeded hints are context, not observed proof.
4. Display seeded facts with `user_seeded` state and low confidence.
5. Add tests for request shape or UI static behavior where practical.
6. Update docs with the seed workflow.

Constraints:
- Do not treat seeded facts as observed evidence.
- Do not trigger discovery from the seeding UI.
- Do not add auth or multi-user workflows yet.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
