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


Task: formalize asset, fact, relationship, and evidence APIs.

Deliverables:
1. Add endpoints:
   - list assets
   - get asset
   - list facts for asset
   - list relationships for asset
   - list evidence for asset/fact if supported
2. Add filters:
   - type
   - vendor
   - serial
   - identity_key
3. Add pagination.
4. Add tests for response shape and pagination.

Constraints:
- Do not build search engine yet.
- Do not build full query language yet.
