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


Task: implement a fake collector for safe local development.

Deliverables:
1. Define a Collector interface:
   - Collect(ctx, target, profile, tasks) returns evidence-like outputs
2. Implement FakeCollector that reads sample command outputs from `examples/fixtures`.
3. Add fixtures for Junos and IOS-XR:
   - show version
   - show inventory / show chassis hardware
   - show lldp neighbors
   - show bgp summary
4. Add a CLI command:
   - `truthwatcher discover fake --target fixture://junos-mx`
5. Store fake collected outputs as evidence.
6. Add tests.

Constraints:
- Do not implement real SSH yet.
- This phase should allow end-to-end evidence creation without touching a network.
