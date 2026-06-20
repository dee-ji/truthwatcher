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


Task: add local authentication foundation without enterprise RBAC.

Deliverables:
1. Add optional local API token authentication controlled by environment config.
2. Keep unauthenticated local-dev mode explicit and documented.
3. Protect mutating API endpoints when token auth is enabled.
4. Add request actor context for audit records.
5. Add tests for protected and unprotected routes.
6. Document the auth boundary and future OIDC/RBAC path.

Constraints:
- Do not implement OIDC yet.
- Do not implement multi-user RBAC.
- Do not store user passwords.
- Do not add external identity provider dependencies.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
