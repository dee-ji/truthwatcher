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


Task: design and scaffold a read-only NETCONF collector boundary.

Deliverables:
1. Define NETCONF collector config and evidence output shape.
2. Add allowlisted abstract NETCONF tasks or RPC descriptors.
3. Store raw XML replies as evidence candidates.
4. Add unit tests for policy validation and evidence mapping.
5. Document why NETCONF remains read-only and bounded.

Constraints:
- Do not require live devices for normal tests.
- Do not add broad NETCONF RPC execution.
- Do not implement write/edit-config, commit, lock/unlock, or config mutation.
- Do not make NETCONF the default discovery workflow.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
