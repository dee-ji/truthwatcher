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


Task: implement the read-only command policy engine.

Deliverables:
1. Create policy package for checking whether a discovery action is allowed.
2. Define allowed abstract tasks:
   - identify_device
   - get_inventory
   - get_interfaces
   - get_neighbors
   - get_arp
   - get_ipv6_neighbors
   - get_bgp_summary
   - get_routes
3. Define denied command patterns:
   - configure
   - commit
   - delete
   - reload
   - clear
   - write memory
   - copy
   - request system reboot
4. Add vendor/platform command mappings later, but create initial policy structure now.
5. Add tests proving dangerous commands are denied.

Constraints:
- Agents and collectors must call policy before execution.
- Do not allow arbitrary shell access.
