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


Task: implement the raw evidence store.

Deliverables:
1. Migration for `evidence`:
   - id
   - discovery_run_id
   - target
   - method
   - command_or_api
   - raw_output
   - raw_output_hash
   - parser_name nullable
   - collected_at
   - metadata JSONB
2. Repository:
   - CreateEvidence
   - GetEvidence
   - ListEvidenceByDiscoveryRun
3. Hash raw output before storage.
4. API:
   - `GET /api/v1/discovery-runs/<built-in function id>/evidence`
   - `GET /api/v1/evidence/<built-in function id>`
5. Add tests for hashing and repository logic.

Core rule:
- Raw evidence is always stored before facts are created.

Constraints:
- No parsing yet.
- No redaction beyond clearly documented TODOs unless easy.
