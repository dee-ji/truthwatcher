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


Task: implement the first real discovery workflow.

Deliverables:
1. Add service layer function:
   - StartDiscoveryRun(seed, profile, tasks, collector)
2. Workflow:
   - create discovery run
   - resolve profile
   - validate tasks/commands against policy
   - collect outputs
   - store raw evidence
   - mark run completed or failed
3. API endpoint:
   - `POST /api/v1/discovery-runs/<built-in function id>/execute` or equivalent
4. CLI:
   - `truthwatcher discover fake ...`
   - optionally `truthwatcher discover ssh ...`
5. Tests using FakeCollector.

Constraints:
- No parsing/facts yet in this workflow unless already implemented cleanly.
- Evidence first.
