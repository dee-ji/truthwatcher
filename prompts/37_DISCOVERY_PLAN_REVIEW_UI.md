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


Task: add UI for reviewing safe discovery plans without executing them automatically.

Deliverables:
1. Add a page or panel that submits seed input to `POST /api/v1/discovery-plans`.
2. Render suggested steps with target, method, profile, task, reason, expected evidence, and risk level.
3. Clearly show that human approval is required and execution is not automatic.
4. Add tests for UI route/static assets where practical.
5. Update docs with the planning workflow.

Constraints:
- Do not auto-execute plans.
- Do not add approval execution yet.
- Do not suggest scope expansion or credential guessing.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
