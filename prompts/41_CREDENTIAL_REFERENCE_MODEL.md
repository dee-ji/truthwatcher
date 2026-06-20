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


Task: formalize credential references without storing raw secrets.

Deliverables:
1. Add a credential reference model for local environment and future external providers.
2. Ensure collectors accept references, not stored plaintext secrets.
3. Validate that local-dev password use is explicit and environment-based.
4. Add redaction helpers for logs, audit records, and API responses.
5. Add tests for redaction and validation.
6. Update `steering-docs/SAFETY_MODEL.md` and install docs if behavior changes.

Constraints:
- Do not build a secrets vault.
- Do not store raw credentials in PostgreSQL.
- Do not add cloud secret-manager dependencies.

Final response requirement:
- Include a git commit message for the update.
- Render it as raw text in a single fenced `text` code block.
- Do not use bullets or numbered items inside the commit message.
