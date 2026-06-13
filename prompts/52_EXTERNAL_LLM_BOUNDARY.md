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


Task: design external LLM integration boundaries without connecting to any provider.

Deliverables:
1. Add an LLM boundary design doc covering opt-in configuration, grounding, redaction, and tool restrictions.
2. Define interfaces for evidence-grounded summarization and question answering without provider SDKs.
3. Add tests for prompt/context construction if any code is added.
4. Document that deterministic local agent behavior remains the default.
5. Update `docs/future.md` with the boundary decision.

Constraints:
- Do not call external LLM APIs.
- Do not add provider SDKs.
- Do not allow the LLM to execute commands or discovery.
- Do not send raw credentials or secrets to any model.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
