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


Task: improve asset identity handling and conflict review.

Deliverables:
1. Add deterministic identity candidate logic for assets from parser outputs.
2. Prefer strong identifiers such as vendor plus serial or system MAC over hostname or IP.
3. Add conflict records or conflict state when two facts disagree.
4. Add APIs to list conflicting facts and weak/provisional identities.
5. Add tests for strong identity, provisional identity, and conflict detection.
6. Document the behavior in `docs/concepts/assets-facts-relationships.md`.

Constraints:
- Do not implement automatic destructive merges.
- Do not assume hostname or IP address is globally unique.
- Do not introduce a graph database.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
