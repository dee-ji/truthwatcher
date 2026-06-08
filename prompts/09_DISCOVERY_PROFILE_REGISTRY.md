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


Task: create discovery profiles that map abstract tasks to vendor-aware read-only commands.

Deliverables:
1. Add discovery profile structs:
   - platform
   - vendor
   - tasks
   - commands
   - parser hints
2. Add built-in profiles for:
   - juniper_junos
   - cisco_iosxr
3. Include only safe commands.
4. Profiles should be loaded from embedded YAML or JSON, or Go structs. Choose the simplest approach and document it.
5. Add tests:
   - profile exists
   - task maps to allowed commands
   - deny patterns still apply

Constraints:
- No runtime plugin loading yet.
- No SSH yet.
