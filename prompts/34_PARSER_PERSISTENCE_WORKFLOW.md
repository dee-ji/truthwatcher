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


Task: wire parser outputs into persisted assets, facts, and relationships.

Deliverables:
1. Add a service layer that takes stored evidence and parser output, then creates or updates assets, facts, and relationships.
2. Preserve the evidence-first rule: raw evidence must already exist before derived records are persisted.
3. Persist parser warnings without losing the discovery run or evidence.
4. Link every created fact and relationship to `evidence_id` when possible.
5. Add an explicit CLI or API path to parse evidence for a discovery run.
6. Add tests using fake fixture evidence and existing Junos/IOS-XR parsers.

Constraints:
- Do not run network discovery.
- Do not create vendor-specific tables.
- Do not silently overwrite conflicting facts.
- Do not parse commands beyond the existing fixture parser scope unless needed for the test workflow.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
