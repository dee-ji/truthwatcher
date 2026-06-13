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


Task: implement the first real parsers for fixture-driven discovery.

Deliverables:
1. Implement simple Junos parser for:
   - show version
   - show chassis hardware
   - show lldp neighbors
2. Implement simple IOS-XR parser for:
   - show version
   - show inventory
   - show lldp neighbors
3. Parsers should create normalized outputs only, not directly write DB.
4. Add tests using fixture files.
5. Ensure parser failure does not lose raw evidence.

Constraints:
- Do not chase perfect parsing.
- Parse only enough to create one device asset and neighbor relationships.
