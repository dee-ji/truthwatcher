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


Task: add dry-run export mapping for Nautobot/NetBox style systems.

Deliverables:
1. Define export candidate structs for devices, interfaces, IPs, sites, and relationships.
2. Map Truthwatcher assets/facts/relationships into generic source-of-truth export candidates.
3. Add `dry_run` output that writes JSON locally for review.
4. Include confidence and evidence references in export candidates.
5. Add tests for mapping behavior.
6. Document that actual API push is not implemented.

Constraints:
- Do not call Nautobot or NetBox APIs.
- Do not push changes to external systems.
- Do not export conflicting or unknown facts as authoritative without clear warnings.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
