# Feature Diagrams

This folder contains one Mermaid diagram per feature path. Splitting the diagrams into individual files makes each flow easier to review, link from product docs, and evolve independently as the implementation changes.

## Diagrams

- [Evidence-first knowledge pipeline](evidence-first-pipeline.md)
- [Request traceability](request-traceability.md)
- [Safe discovery execution](safe-discovery-execution.md)
- [Parsing and identity review](parsing-identity-review.md)
- [Planning and seeded context](planning-seeded-context.md)
- [Import, export, and extensibility boundaries](import-export-extensibility.md)

## How to read these diagrams

Each diagram uses the same convention:

- **Solid process boxes** are API handlers, services, parsers, collectors, or operator review steps.
- **Database cylinders** are persisted system-of-record data.
- **Decision diamonds** are policy, confidence, or review gates.
- **Notes below each diagram** explain why the flow exists and reference implementation or concept docs that support the design choice.

The diagrams are intentionally Markdown-native Mermaid blocks so documentation tooling can render them without committing generated images.
