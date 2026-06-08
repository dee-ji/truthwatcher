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


Task: add PostgreSQL database foundation and migrations.

Deliverables:
1. Add a DB package that opens and pings PostgreSQL using `database/sql`.
2. Choose a minimal Postgres driver and document the decision in `ARCHITECTURE_DECISIONS.md`.
3. Add a simple migration runner embedded in the Go binary using `embed`.
4. Create `migrations/000001_init.up.sql` and `migrations/000001_init.down.sql`.
5. First migration should create only a schema version/migration bookkeeping table if needed.
6. Add CLI commands:
   - `truthwatcher migrate up`
   - `truthwatcher migrate status`
7. Add tests where possible without requiring a live DB; isolate SQL generation/ordering logic.

Constraints:
- Do not introduce an ORM.
- Do not introduce Docker.
- Do not create product tables yet.
