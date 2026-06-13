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


Task: design and scaffold a read-only SNMP collector boundary.

Deliverables:
1. Define SNMP collector config for target, version, credential reference, timeout, and allowed OID groups.
2. Add explicit OID allowlist structure for inventory, interfaces, and neighbor evidence candidates.
3. Map raw varbinds into evidence-like outputs without parsing into facts yet.
4. Add tests for OID allowlist enforcement and evidence mapping.
5. Document that SNMP is discovery evidence, not monitoring.

Constraints:
- Do not add polling loops, alerting, metrics, or observability features.
- Do not require real devices for normal tests.
- Do not perform broad walks outside configured allowlists.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
