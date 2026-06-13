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


Task: create embedded frontend foundation served by the Go binary.

Deliverables:
1. Create minimal frontend under `web/`.
2. Choose simple approach:
   - plain TypeScript/Vite, or
   - minimal static HTML/JS first
3. Use Go `embed` to serve built frontend from `truthwatcher server`.
4. Add basic layout:
   - top nav
   - dashboard placeholder
   - API status indicator
5. Add Makefile targets:
   - build-ui
   - build
6. Ensure binary can run and serve UI.

Constraints:
- Do not build chat yet.
- Do not over-design UI.
