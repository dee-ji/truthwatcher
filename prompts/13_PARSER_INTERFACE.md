Before coding, read these files:
- PROJECT_TRUTHWATCHER.md
- AGENT_COLLABORATION_CONTRACT.md
- ARCHITECTURE_DECISIONS.md
- DATA_MODEL.md
- SAFETY_MODEL.md
- MVP_SPEC.md
- EXTENSIBILITY_MODEL.md
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


Task: create parser framework.

Deliverables:
1. Define Parser interface:
   - Supports(platform, command)
   - Parse(evidence) returns normalized facts and relationships
2. Define normalized parse outputs:
   - device identity
   - inventory components
   - interfaces
   - neighbors
   - bgp peers
3. Create parser registry.
4. Add NoopParser for unsupported evidence.
5. Add tests for parser selection.

Constraints:
- Do not attempt to parse every command.
- Keep parser outputs generic and tied to Asset/Fact/Relationship concepts.
