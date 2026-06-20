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


Task: make uncertainty first-class.

Deliverables:
1. Define confidence scoring model:
   - observed
   - inferred
   - user_seeded
   - conflicting
   - unknown
2. Add fields if missing:
   - confidence
   - confidence_reason
   - state
3. Add logic to mark conflicting facts instead of overwriting silently.
4. Add API output showing confidence and evidence references.
5. Document in `steering-docs/DATA_MODEL.md`.

Constraints:
- Do not attempt complex ML scoring.
- Use simple deterministic scoring.
