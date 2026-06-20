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


Task: implement HTTP API foundation.

Deliverables:
1. `truthwatcher server` starts an HTTP server.
2. Add endpoints:
   - `GET /healthz`
   - `GET /readyz`
   - `GET /api/v1/version`
3. Use graceful shutdown on SIGINT/SIGTERM.
4. Add middleware:
   - request ID
   - structured request logging
   - panic recovery
5. Keep routing simple with standard library or a small router if already chosen.
6. Add endpoint tests using `httptest`.

Constraints:
- No auth yet.
- No UI yet.
- No discovery endpoints yet.
