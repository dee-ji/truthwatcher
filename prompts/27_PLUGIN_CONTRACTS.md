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


Task: define extensibility contracts without implementing full plugin runtime.

Deliverables:
1. Define interfaces for:
   - Collector
   - Parser
   - Importer
   - Exporter
   - Enricher
   - Planner
2. Document plugin boundaries in `EXTENSIBILITY_MODEL.md`.
3. Add examples of future connectors:
   - IPAM
   - monitoring
   - EMS
   - cloud API
   - config archive
4. Keep kernel independent of any external system.

Constraints:
- Do not implement dynamic plugin loading yet.
- Do not add HashiCorp go-plugin yet unless explicitly requested.
