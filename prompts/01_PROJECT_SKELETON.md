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


Task: implement the initial Go project skeleton.

Deliverables:
1. Initialize or validate `go.mod`.
2. Create a single binary entrypoint:
   - `cmd/truthwatcher/main.go`
3. Implement CLI command structure with at least:
   - `truthwatcher version`
   - `truthwatcher server`
4. Use standard library where reasonable.
5. Create internal package layout:
   - `internal/app`
   - `internal/config`
   - `internal/logging`
   - `internal/api`
   - `internal/db`
   - `internal/discovery`
   - `internal/evidence`
   - `internal/assets`
   - `internal/policy`
   - `internal/audit`
6. `truthwatcher server` should start and print a clear startup message, but it does not need DB yet.
7. Add unit tests for version/config basics if practical.

Constraints:
- No frontend yet.
- No database migrations yet.
- No network collectors yet.
- No agents.
