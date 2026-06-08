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


Task: implement DiscoveryRun as the first product table and API object.

Deliverables:
1. Migration for `discovery_runs`:
   - id UUID or ULID
   - status: pending | running | completed | failed | canceled
   - seed_input JSONB
   - started_at
   - completed_at
   - error_message nullable
   - created_at
   - updated_at
2. Repository:
   - CreateDiscoveryRun
   - GetDiscoveryRun
   - ListDiscoveryRuns
   - UpdateDiscoveryRunStatus
3. API:
   - `POST /api/v1/discovery-runs`
   - `GET /api/v1/discovery-runs`
   - `GET /api/v1/discovery-runs/<built-in function id>`
4. Tests for repository behavior if DB test setup exists, otherwise service-level unit tests.

Constraints:
- Do not run network discovery yet.
- Do not create evidence records yet.
