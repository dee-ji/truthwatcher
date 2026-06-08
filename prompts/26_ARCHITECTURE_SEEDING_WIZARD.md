Before coding, read these files:
- PROJECT_TRUTHWATCHER.md
- AGENT_COLLABORATION_CONTRACT.md
- ARCHITECTURE_DECISIONS.md
- DATA_MODEL.md
- SAFETY_MODEL.md
- MVP_SPEC.md
- EXTENSIBILITY_MODEL.md
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


Task: implement architecture seeding.

Deliverables:
1. Add data model for architecture hints:
   - organization/network type
   - known ASNs
   - known route reflectors
   - known vendors
   - known EMS systems
   - known services
   - known regions/markets
2. Add UI wizard or simple API to submit hints.
3. Store hints as seeded facts/evidence with source `user_seeded`.
4. Use hints in discovery planner.
5. Document that seeding is not proof; it is context.

Constraints:
- Do not treat seeded facts as observed facts.
- Maintain confidence distinction.
