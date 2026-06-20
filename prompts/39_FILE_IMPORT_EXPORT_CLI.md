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


Task: wire local JSON import/export connector foundation into CLI commands.

Deliverables:
1. Add `truthwatcher export json --output <path>` for local graph snapshots.
2. Add `truthwatcher import json --input <path>` for local snapshot import candidates.
3. Export assets, facts, relationships, and evidence metadata.
4. Preserve source, confidence, state, and evidence references where possible.
5. Import should validate data and avoid treating imported records as observed proof.
6. Add tests for CLI parsing and import/export service behavior without requiring a live DB where practical.
7. Document the workflow in `docs/api.md` or a new user doc.

Constraints:
- Do not implement NetBox, Nautobot, IPAM, or cloud connectors.
- Do not export raw evidence output by default.
- Do not bypass kernel validation.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
