# System Traceability

Traceability is one of Truthwatcher's core differentiators: the system is designed to explain not only *what* it believes, but *why* it believes it and *which request or discovery action produced it*.

## Traceability Goals

Truthwatcher should make it possible to answer these questions from API responses, stored records, and UI views:

- Which HTTP request initiated this operation?
- Which discovery run collected the raw evidence?
- Which target, method, profile, task, command, or API produced the evidence?
- Which parser interpreted the evidence?
- Which facts, assets, and relationships reference that evidence?
- Which records are observed, inferred, user-seeded, conflicting, weak, provisional, or unknown?
- Which audit records describe the action lifecycle?

## Request IDs

All HTTP requests flow through request ID middleware.

- If a caller sends `X-Request-ID`, Truthwatcher preserves that value.
- If the header is absent or blank, Truthwatcher generates a UUIDv7 value.
- The generated or preserved value is written to the request header before handlers run.
- The same value is returned in the response `X-Request-ID` header.

UUIDv7 request IDs give each request a globally unique identifier with a sortable timestamp component. That is useful before authentication exists because operators can still correlate logs, API responses, discovery metadata, and audit rows by request.

## Evidence-To-Model Chain

Truthwatcher does not treat inventory as a flat table. The traceability chain is:

```text
HTTP request -> discovery run -> raw evidence -> parser output -> fact/asset/relationship -> graph/API/UI answer
```

Each layer intentionally keeps enough context to point backward:

1. **HTTP request**: `X-Request-ID` is assigned at the edge and included in request logs.
2. **Discovery run**: execution receives the request ID and records run-level context.
3. **Evidence**: raw output stores target, method, command/API, collection time, metadata, and raw output hash.
4. **Audit**: discovery actions store initiator, request ID, run ID, target, profile, task, command/API, status, error, and timing.
5. **Parser persistence**: parser output creates model records that retain source, confidence, state, and evidence references.
6. **Graph and agent answers**: higher-level views are projections over assets, facts, relationships, and evidence rather than unsupported claims.

## Why This Separates Truthwatcher

Many inventory tools eventually become another place where stale facts live. Truthwatcher is different because every important assertion should either:

- point to observed evidence,
- identify itself as user-seeded context,
- identify itself as inferred deterministic knowledge,
- remain visible as provisional or weak identity, or
- surface as a conflict instead of overwriting history.

This makes the model reviewable. Operators can inspect raw evidence, compare it with parsed facts, and decide whether a proposed identity merge or conflicting fact should be accepted.

## Traceability States

Truthwatcher uses explicit state and confidence fields so uncertainty is visible:

| State | Meaning | Traceability behavior |
| --- | --- | --- |
| `observed` | Directly supported by collected evidence. | Should link back to evidence. |
| `inferred` | Derived deterministically from known data. | Should explain the derivation path. |
| `user_seeded` | Supplied by a human or import as context. | Must not masquerade as device proof. |
| `conflicting` | Contradicts another known fact. | Preserves both values for review. |
| `unknown` | Not known yet. | Keeps gaps visible instead of guessing. |

Identity also carries review-oriented traceability:

- `strong`: stable evidence-backed identity such as vendor plus serial or system MAC.
- `provisional`: useful but not globally reliable, such as hostname or IP address.
- `weak`: insufficient identity anchor and should remain reviewable.

## Operational Correlation Example

When an operator executes fixture-backed discovery through the API:

1. The API assigns or preserves `X-Request-ID`.
2. The discovery workflow stores that request ID in discovery context and action audit records.
3. Raw command output is stored as evidence with metadata.
4. Parsing later derives facts and relationships with evidence references.
5. API and UI graph views can show the model while evidence endpoints explain the source.
6. Logs can be searched by the same request ID returned to the caller.

## Related Diagrams

Traceability is also shown as individual Mermaid diagrams so each feature path can evolve independently:

- [Request traceability](diagrams/request-traceability.md)
- [Evidence-first knowledge pipeline](diagrams/evidence-first-pipeline.md)
- [Safe discovery execution](diagrams/safe-discovery-execution.md)
- [Parsing and identity review](diagrams/parsing-identity-review.md)
- [Planning and seeded context](diagrams/planning-seeded-context.md)
- [Import, export, and extensibility boundaries](diagrams/import-export-extensibility.md)

## Current Boundary

Traceability does not require authentication. Authentication will eventually add user identity, authorization, and stronger actor attribution. Until then, UUIDv7 request IDs provide a stable correlation key with embedded request time while the audit records still capture the initiator string supplied by current workflows.
