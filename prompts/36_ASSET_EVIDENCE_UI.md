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


Task: build asset and evidence browsing UI.

Deliverables:
1. Add UI views for listing assets with filters already supported by the API.
2. Add asset detail view showing facts, relationships, confidence, state, and evidence references.
3. Allow opening raw evidence from linked facts and relationships.
4. Keep evidence read-only and clearly labeled.
5. Add basic UI tests or handler/static asset tests where practical.
6. Update user docs if workflow changes.

Constraints:
- Do not add a frontend framework unless already introduced.
- Do not add websockets or real-time updates.
- Do not hide uncertainty or conflicting state.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
