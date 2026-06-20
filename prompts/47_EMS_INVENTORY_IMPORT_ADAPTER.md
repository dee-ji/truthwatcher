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


Task: implement a generic file-backed EMS inventory import adapter.

Deliverables:
1. Define a generic EMS inventory import schema for managed devices, platform hints, regions, and controller ownership.
2. Store imported EMS data as evidence or imported facts with source metadata.
3. Map data into assets, facts, and relationships without vendor-specific tables.
4. Use EMS context in discovery planner suggestions.
5. Add tests with sample EMS fixture files.
6. Document EMS import boundaries.

Constraints:
- Do not integrate with a specific EMS API yet.
- Do not let EMS hints authorize network access.
- Do not treat EMS data as stronger than observed evidence by default.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
