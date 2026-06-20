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


Task: add service-aware modeling foundation without generating configurations.

Deliverables:
1. Add generic service asset types for Internet access, L3VPN, E-Line, EVLAN, DIA, and managed CPE.
2. Add relationships for service depends_on asset, service terminates_on asset, and customer/site context where modeled.
3. Add APIs to create/list/get service assets and service relationships.
4. Include confidence, state, and evidence/source references.
5. Add tests for service modeling behavior.
6. Document service modeling boundaries.

Constraints:
- Do not generate configs.
- Do not generate MOPs.
- Do not activate services.
- Do not assume one provider service catalog.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
