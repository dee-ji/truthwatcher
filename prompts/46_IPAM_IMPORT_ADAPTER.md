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


Task: implement a generic file-backed IPAM import adapter.

Deliverables:
1. Define a generic IPAM import schema for prefixes, IP addresses, VRFs, sites, tenants, and reservations.
2. Import IPAM records as evidence-backed or imported context with appropriate confidence/state.
3. Map records to assets, facts, and relationships through kernel services.
4. Detect conflicts with observed facts instead of overwriting silently.
5. Add tests with sample JSON fixtures.
6. Document the import format.

Constraints:
- Do not integrate with a specific IPAM product.
- Do not treat IPAM data as observed device evidence.
- Do not add cloud dependencies.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
