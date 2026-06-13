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


Task: create repository guardrails only.

Deliverables:
1. Create or update `.gitignore` for Go, GoLand, local binaries, environment files, coverage files, temporary data, and frontend build artifacts.
2. Create `CONTRIBUTING.md` explaining:
   - small commits
   - one prompt per task
   - read project docs before coding
   - no scope drift
   - read-only network safety requirement
3. Create `CODEOWNERS` if useful, but keep it simple.
4. Create `.editorconfig`.
5. Create `Makefile` with placeholder targets:
   - `fmt`
   - `test`
   - `lint`
   - `build`
   - `run`

Constraints:
- Do not create application logic yet.
- Do not create database schema yet.
- Do not add external dependencies unless necessary.
