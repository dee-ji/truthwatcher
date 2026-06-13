Before coding, read these files:
- steering-docs/PROJECT_TRUTHWATCHER.md
- steering-docs/AGENT_COLLABORATION_CONTRACT.md
- steering-docs/ARCHITECTURE_DECISIONS.md
- steering-docs/DATA_MODEL.md
- steering-docs/SAFETY_MODEL.md
- steering-docs/MVP_SPEC.md
- steering-docs/EXTENSIBILITY_MODEL.md
- steering-docs/ROADMAP.md

Act as an informed engineering sidekick. Challenge scope creep. Do not simply agree. If a requested implementation conflicts with the project constitution, stop and explain the conflict.

Rules:
- Keep Truthwatcher as a single Go binary plus PostgreSQL unless explicitly asked otherwise.
- Do not introduce Docker, Kubernetes, microservices, message brokers, or cloud dependencies.
- Do not introduce write-capable network automation.
- Do not build observability features.
- Do not build a full chat platform before the evidence kernel exists.
- Prefer boring, explicit Go code.
- Add tests where practical.
- Update steering-docs/ROADMAP.md only with completed work and next steps.
- Do not rewrite vision documents unless directly asked.


Task: create chat/agent workspace shell without advanced reasoning.

Deliverables:
1. Add UI panel for “Ask Truthwatcher”.
2. Add backend endpoint:
   - `POST /api/v1/agent/messages`
3. For now, implement deterministic canned tool responses:
   - list known assets
   - explain asset evidence
   - summarize discovery run
4. Store conversation history locally or in DB if simple.

Constraints:
- Do not connect external LLM yet unless explicitly configured.
- Do not let agent execute discovery yet.
- No network actions from chat.
