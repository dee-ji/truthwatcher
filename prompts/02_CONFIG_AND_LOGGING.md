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


Task: implement configuration and structured logging foundation.

Deliverables:
1. Add environment-based config loading:
   - `TRUTHWATCHER_ADDR` default `127.0.0.1:8080`
   - `TRUTHWATCHER_DATABASE_URL`
   - `TRUTHWATCHER_LOG_LEVEL` default `info`
   - `TRUTHWATCHER_DEV_MODE` default `false`
2. Add config validation.
3. Add structured logging with the Go standard library `log/slog`.
4. Wire config and logger into `truthwatcher server`.
5. Add tests for config defaults and validation.

Constraints:
- Do not add Viper unless strongly justified.
- Do not add config files yet.
- Do not add secrets storage.
