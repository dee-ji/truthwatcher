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


Task: build graph query service over relational tables.

Deliverables:
1. Implement query service:
   - get asset with facts
   - get neighbors for asset
   - get relationship path candidates if simple
2. Add API:
   - `GET /api/v1/assets/<built-in function id>/graph`
   - `GET /api/v1/graph/neighbors?asset_id=...`
3. Return nodes and edges JSON suitable for frontend graph rendering.
4. Add tests for graph response shape.

Constraints:
- Do not add Neo4j yet.
- Postgres remains the only database.
