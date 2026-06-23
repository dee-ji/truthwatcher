# Documentation Truth Map

This file records the documentation alignment pass and identifies which documents are authoritative for each major topic. It is intentionally high level; implementation details remain in the API, testing, install, and code-level documents.

## Documentation Inventory

| File or folder | Current role |
| --- | --- |
| `README.md` | Public project overview, POC story, quickstart, safety model, architecture summary, and concept links. |
| `CONTEXT.md` | Product context and narrative background for contributors and AI-assisted coding sessions. |
| `ROADMAP.md` | Phase plan and current completion tracking. |
| `CONTRIBUTING.md` | Contributor workflow, scope guardrails, and review expectations. |
| `SECURITY.md` | Security reporting and read-only discovery expectations. |
| `CODE_OF_CONDUCT.md` | Contributor conduct. |
| `OSS_GOVERNANCE.md` | Open-source governance and merge policy. |
| `RELEASE_STRATEGY.md` and `docs/RELEASES.md` | Release process and tag semantics. |
| `steering-docs/PROJECT_TRUTHWATCHER.md` | Authoritative mission, product philosophy, scope, vendor-neutrality, and POC-to-enterprise path. |
| `steering-docs/PROJECT_CHARTER.md` | Legacy/alternate charter kept synchronized with `PROJECT_TRUTHWATCHER.md`. |
| `steering-docs/MVP_SPEC.md` | Authoritative proof-of-concept specification and success criteria. |
| `steering-docs/ARCHITECTURE_DECISIONS.md` | ADR-style architecture decisions and constraints. |
| `steering-docs/DATA_MODEL.md` | Data modeling philosophy and table-level concepts. |
| `steering-docs/SAFETY_MODEL.md` | Read-only safety policy and command boundaries. |
| `steering-docs/EXTENSIBILITY_MODEL.md` | Adapter, plugin, import/export, and replaceable-edge model. |
| `steering-docs/AGENT_*`, `START_HERE_CODEX_PROMPT.md`, `CODING_SESSION_TEMPLATE.md`, and `README_PROMPT_PACK.md` | AI-agent collaboration, prompt-pack usage, and coding-session guardrails. |
| `steering-docs/PROMPT_INDEX.md` and `prompts/*.md` | Sequenced implementation prompts. These may mention concrete adapters or fixtures as examples, but the steering docs define the vendor-neutral product direction. |
| `docs/poc-walkthrough.md` | Fixture-backed proof-of-concept walkthrough from evidence collection to understanding. |
| `docs/concepts/*.md` | User-facing concept documentation for evidence-first modeling, discovery planning, assets/facts/relationships, and extensibility. |
| `docs/diagrams/*.md` | Mermaid diagrams for evidence flow, safe discovery, identity review, request traceability, and extensibility. |
| `docs/api.md` | Implementation-level API reference. Examples may use fixtures or sample platforms only as examples. |
| `docs/install.md` | Local installation and runnable POC workflow. |
| `docs/import-export.md` | Local JSON import/export contract and trust model. |
| `docs/testing.md` | Test strategy and boundaries. |
| `docs/traceability.md` | Request, evidence, and model provenance expectations. |
| `docs/future.md` | Future-phase backlog intentionally outside the POC. |
| `docs/local-knowledge.md` | Optional local development knowledge-provider configuration. |
| `docs/planning/*.md` | Planning analyses for specific implementation slices. |

## Conflicts Found And Resolved

| Conflict | Resolution |
| --- | --- |
| Some documents described Truthwatcher primarily as network cartography or source-of-truth bootstrap tooling, while newer direction requires a broader discovery, reasoning, and intent platform. | Updated the README, charter, POC spec, context, contributing, and security language to frame Truthwatcher as a vendor-neutral evidence-to-understanding platform. |
| The previous MVP spec centered on one seed device, SSH, and named sample platforms. | Reframed it as a proof-of-concept specification with seed targets, fixtures, imports, read-only profiles, evidence preservation, identity, assets, facts, relationships, graph views, and reasoning boundaries. Vendor examples are now explicitly examples. |
| Several files implied the chain was `Evidence -> Facts -> Assets -> Relationships -> Graph -> Understanding`. | Aligned the core conceptual model to `Evidence -> Identity -> Assets -> Facts -> Relationships -> KnowledgeGraph -> Understanding` with conceptual skill responsibilities. |
| Mentions of specific products such as source-of-truth systems, IPAM, EMS, monitoring, and cloud tools could read as foundations. | Reworded the high-level docs so those systems are examples of adapters or sources of context, not required foundations. |
| `PROJECT_TRUTHWATCHER.md` and `PROJECT_CHARTER.md` were duplicate charter-style documents that could drift. | Synchronized both files with the same current charter and documented `PROJECT_TRUTHWATCHER.md` as authoritative. |
| The long-term direction mixed future features with current capabilities. | Clarified the POC scope and path to enterprise readiness while keeping future capabilities as later phases. |

## Files That Should Be Deleted, Merged, Or Renamed Later

No major file was deleted during this pass because several documents are referenced by prompt-pack and contributor workflows. Recommended later cleanup:

- Merge or retire `steering-docs/PROJECT_CHARTER.md` after all references point to `steering-docs/PROJECT_TRUTHWATCHER.md`.
- Rename `steering-docs/MVP_SPEC.md` to `steering-docs/POC_SPEC.md`; keep a compatibility stub or update all prompt references first.
- Consider merging `RELEASE_STRATEGY.md` and `docs/RELEASES.md` if release instructions start to diverge.
- Consider splitting `CONTEXT.md` into a shorter product narrative plus separate archival notes if it continues to grow.
- Audit older `prompts/*.md` before future implementation waves so vendor examples remain clearly bounded as fixtures, profiles, or adapters.

## Authoritative Topic Map

| Topic | Authoritative file |
| --- | --- |
| Mission, problem, product philosophy, and vendor-neutrality | `steering-docs/PROJECT_TRUTHWATCHER.md` |
| Public overview and quickstart | `README.md` |
| POC scope, success criteria, and non-goals | `steering-docs/MVP_SPEC.md` |
| Roadmap status | `ROADMAP.md` |
| Architecture decisions | `steering-docs/ARCHITECTURE_DECISIONS.md` |
| Data modeling | `steering-docs/DATA_MODEL.md` |
| Safety and read-only boundaries | `steering-docs/SAFETY_MODEL.md` |
| Adapter and extensibility principles | `steering-docs/EXTENSIBILITY_MODEL.md` and `docs/concepts/extensibility.md` |
| POC walkthrough | `docs/poc-walkthrough.md` |
| Evidence-first model | `docs/concepts/evidence-first.md` |
| Assets, facts, and relationships | `docs/concepts/assets-facts-relationships.md` |
| Discovery planning | `docs/concepts/discover-how-to-discover.md` |
| API reference | `docs/api.md` |
| Install and local execution | `docs/install.md` |
| Testing strategy | `docs/testing.md` |
| Traceability | `docs/traceability.md` |
| Future phases and out-of-scope ideas | `docs/future.md` |
| Prompt execution order | `steering-docs/PROMPT_INDEX.md` and `prompts/*.md` |

## Next-Step Recommendation

The next documentation improvement should add a small expected-output appendix to `docs/poc-walkthrough.md` after the fixture outputs and API responses stabilize. That appendix should show representative redacted JSON snippets for evidence, identity, assets, facts, relationships, graph context, and reasoning summaries.
