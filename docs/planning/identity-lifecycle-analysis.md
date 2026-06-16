# Identity Lifecycle Implementation Analysis

Status: planning analysis only; no application code or ADRs changed.

This document applies the external Mistspren ADR-0002 intent to the current Truthwatcher codebase: parser-derived observations must not silently merge canonical assets or overwrite stronger canonical asset identity attributes. Ambiguous matches, conflicting identifiers, proposed merges, and material canonical identity changes should become reviewable identity-resolution decisions with evidence references and audit metadata. Narrow deterministic auto-acceptance is acceptable only for low-risk, evidence-backed, non-destructive cases.

## Observed code facts

### Current evidence-to-parser-to-asset flow

1. Discovery execution is exposed at `POST /api/v1/discovery-runs/execute` and is deliberately limited to the fake collector through request validation and fake collector construction. The route passes a discovery service, evidence service, optional audit service, policy engine, initiator, request ID, and request context into discovery execution.
2. Raw evidence is a first-class model with discovery run ID, target, method, command/API, raw output, raw output hash, optional parser name, collected timestamp, and JSON metadata. The evidence service validates required fields and explicitly preserves raw output unchanged.
3. Parser persistence is a second explicit step exposed at `POST /api/v1/discovery-runs/{id}/parse`. It accepts a platform, lists evidence for the discovery run, selects built-in parsers by platform and command, and records parser results as parsed, skipped, or failed.
4. Parser output is represented as in-memory candidates: device identities, inventory components, interfaces, neighbors, BGP peers, parsed facts, and parsed relationships. Asset references use parser-generated `identity_key` strings before database asset IDs exist.
5. Persistence builds an in-memory asset index from existing assets keyed by normalized `identity_key`, then creates missing assets, facts, relationships, and parser result records. Existing assets with the same identity key are reused; they are not updated in the current persistence path.
6. Facts and relationships store `evidence_id` when available. Parser result records preserve parser name, evidence ID, status, warnings, and error message, but they do not persist the full structured parser output.

### Current tables, types, and routes involved

#### Raw evidence and parser output

- `evidence` domain type: `Evidence` stores `DiscoveryRunID`, `Target`, `Method`, `CommandOrAPI`, `RawOutput`, `RawOutputHash`, `ParserName`, `CollectedAt`, and `Metadata`.
- Parser output types: `Result`, `DeviceIdentity`, `InventoryComponent`, `Interface`, `Neighbor`, `BGPPeer`, `ParsedFact`, and `ParsedRelationship`.
- `parser_results` table: persists one row per evidence parse attempt, with discovery run ID, evidence ID, parser name, status, warnings, optional error, and created timestamp.
- Routes: `GET /api/v1/discovery-runs/{id}/evidence`, `GET /api/v1/evidence/{id}`, and `POST /api/v1/discovery-runs/{id}/parse`.

#### Assets, facts, relationships, uncertainty, conflicts, and provisional state

- `Asset` has type, identity key, optional vendor/model/serial/system MAC, confidence, confidence reason, state, metadata, and timestamps.
- `Fact` has asset ID, name, JSON value, source, confidence, confidence reason, state, optional evidence ID, and created timestamp.
- `Relationship` has source and target asset IDs, relationship type, confidence, confidence reason, state, optional evidence ID, metadata, and timestamps.
- The uncertainty state vocabulary is `observed`, `inferred`, `user_seeded`, `conflicting`, and `unknown`.
- Fact conflicts are detected at fact creation time when a new fact has the same name but a different value from an existing non-conflicting fact for the same asset. The new fact is stored as `conflicting`; the existing fact is not overwritten.
- Provisional identity is currently metadata-derived. Identity keys based on hostname, IP, name, unknown source, or non-durable source are classified as provisional or weak, and `GET /api/v1/assets/provisional-identities` lists non-strong identity assets for review only.
- Routes: `GET /api/v1/assets`, `GET /api/v1/assets/provisional-identities`, `GET /api/v1/assets/{id}`, `GET /api/v1/assets/{id}/facts`, `GET /api/v1/facts/conflicts`, `GET /api/v1/assets/{id}/relationships`, `GET /api/v1/assets/{id}/evidence`, `GET /api/v1/facts/{id}/evidence`, `GET /api/v1/assets/{id}/graph`, and `GET /api/v1/graph/neighbors`.

#### Audit metadata

- Discovery execution can persist audit records with action, initiator, request ID, discovery run ID, target, method, profile, task, command/API, status, evidence ID, error message, timestamps, and JSON context.
- API discovery execution also returns audit metadata in the response envelope.
- Parser persistence and asset creation do not currently emit identity-resolution audit records.

### Where parser-derived identity clues currently become persistent state

- Built-in version parsers derive device identities from hostname when no durable serial or MAC is parsed from version output. These become `DeviceIdentity` candidates with a provisional hostname identity key and metadata explaining identity strength and reason.
- Inventory parsers derive component identities from vendor plus serial when present, creating strong identity keys for chassis, cards, optics, and related inventory components.
- `persistResult` turns `DeviceIdentities`, `InventoryComponents`, `Interfaces`, `ParsedFacts`, and `ParsedRelationships` into persisted assets, facts, and relationships. It uses identity keys to create or reuse assets.
- `ensurePlaceholderAsset` creates placeholder assets from identity references that appear only in facts or relationships. Placeholder type is inferred from the identity key prefix, and the reason is `placeholder created from parser output identity reference`.
- Asset creation annotates metadata with `identity_strength`, `identity_reason`, and `identity_provisional` based on the identity key source.

### Current risk of silently mutating canonical asset identity

Observed current behavior is comparatively safe because parser persistence only creates assets when an identity key is absent and returns an existing asset when the key is already present. There is no current code path in parser persistence that updates an existing asset's `identity_key`, `vendor`, `serial`, `system_mac`, or metadata.

However, the current behavior still creates identity lifecycle risks:

- A parser-derived provisional hostname identity becomes a durable row in `assets` immediately. It is labeled provisional in metadata, but there is no separate persisted candidate/review table.
- Strong identifiers discovered later do not have a first-class path to propose that a provisional hostname asset and a durable serial asset represent the same real-world device.
- The global unique index on non-empty `assets.identity_key` makes identity key uniqueness deterministic, but it does not model proposed matches, rejected matches, or conflicts between multiple identifiers for one asset.
- Fact conflict handling protects fact history, but asset identity conflicts are not equivalent to fact conflicts and are not currently represented as reviewable decisions.
- Parser results persist warnings and status, not structured identity evidence. Reconstructing why an identity should be merged or rejected requires reparsing raw evidence or inferring from already-created assets/facts.

## Implementation proposals

### Minimal identity lifecycle vocabulary

Use a small vocabulary that fits the existing evidence-first relational model:

- **Identity clue**: a raw parser observation about hostname, serial, system MAC, vendor, model, chassis serial, neighbor system name, or interface-local identity. It must reference evidence.
- **Identity candidate**: a normalized, evidence-backed proposed identity key or identity attribute for an asset-like thing, with strength (`strong`, `provisional`, `weak`), confidence, parser/source, and reason.
- **Canonical asset**: an existing `assets` row used by facts, relationships, and graph projection.
- **Proposed match**: a candidate-to-asset or candidate-to-candidate assertion that two identities may refer to the same real-world object.
- **Identity decision**: review state for a proposed identity effect: `pending`, `auto_accepted`, `accepted`, `rejected`, or `superseded`.
- **Material identity change**: any change that would rewrite an existing canonical asset identity key or stronger identity attributes, merge two assets, or attach a stronger identifier to a previously provisional asset.
- **Non-destructive auto-acceptance**: creation of a new candidate record, or creation of a new canonical asset only when no plausible existing match exists and the candidate is strong and evidence-backed.

### Proposed first implementation slice: persisted identity candidates

Smallest testable milestone:

1. Add an `identity_candidates` table only. Do not implement merges or canonical asset mutation in the first slice.
2. Persist parser-derived identity candidates alongside parser results, referencing discovery run, evidence, parser name, candidate asset type, candidate identity key, candidate attributes (`vendor`, `model`, `serial`, `system_mac`, `hostname`, optional metadata), strength, confidence, reason, and review state.
3. Set all hostname/IP/name/unknown candidates to `pending` or `provisional` review state. Strong vendor+serial/system-MAC candidates may be `auto_accepted` only when no existing canonical asset or candidate plausibly conflicts.
4. Keep current asset/fact/relationship persistence intact initially, but make the new candidate rows the source for later review UI/API work.
5. Add a read-only API endpoint to list identity candidates by discovery run, evidence ID, review state, and identity key. This should not expose non-fake collectors or any execution flow.
6. Add audit metadata only for later decision transitions. In the first slice, creation metadata on the candidate plus evidence reference is sufficient; do not add an approval/merge workflow yet.

Suggested initial columns:

- `id uuid primary key`
- `discovery_run_id uuid not null references discovery_runs(id)`
- `evidence_id uuid not null references evidence(id)`
- `parser_name text not null`
- `asset_type text not null`
- `candidate_identity_key text not null`
- `strength text not null check (strength in ('strong','provisional','weak'))`
- `confidence numeric not null check (confidence >= 0 and confidence <= 1)`
- `reason text not null`
- `vendor text`, `model text`, `serial text`, `system_mac text`, `hostname text`
- `proposed_asset_id uuid references assets(id)` nullable
- `review_state text not null default 'pending' check (review_state in ('pending','auto_accepted','accepted','rejected','superseded'))`
- `metadata jsonb not null default '{}'`
- `created_at timestamptz not null default now()`

### Candidate tests for fixture-backed parser output

- Parsing `show version` fixture output persists a hostname-based identity candidate with strength `provisional`, references the original evidence ID, and does not require asset mutation.
- Parsing inventory fixture output persists vendor+serial identity candidates with strength `strong`, references the original evidence ID, and records parser name and confidence.
- Re-running parse on the same evidence should not create duplicate identity candidates for the same evidence ID, parser name, and candidate identity key.
- A parser output that proposes a hostname candidate matching one asset and a serial candidate matching another asset should produce pending candidates/proposed match data rather than merging assets.
- A conflicting hostname fact should still be stored as a conflicting fact and should not rewrite asset identity metadata.
- Unsupported commands should still create skipped parser results without identity candidates.

### Explicit non-goals

- Do not introduce a graph database.
- Do not introduce microservices.
- Do not introduce source-of-truth export or broad synchronization behavior.
- Do not build broad observability, monitoring, alerting, or telemetry features.
- Do not build chat UX for identity review.
- Do not run arbitrary commands or expose arbitrary command execution.
- Do not expose SSH or other non-fake collectors through new API/UI flows.
- Do not implement asset merging, identity rewriting, or review approval workflows in the first slice.
- Do not modify ADRs for this analysis task.

### Open questions and conflicts

- Should current parser persistence continue creating canonical assets immediately from provisional hostname identities, or should future parser persistence create candidates first and canonical assets only for strong no-conflict cases?
- What exact threshold makes a strong candidate safe for `auto_accepted`? Vendor+serial and system MAC are strong in current code, but chassis replacement, virtual devices, and reused serial formats can still create edge cases.
- Should identity candidates be tied directly to `parser_results.id` in addition to `evidence_id` and `parser_name`? This would improve auditability but requires parser result creation before candidate persistence.
- How should candidate-to-asset and candidate-to-candidate proposed matches be modeled: in the first table as nullable `proposed_asset_id`, or in a later `identity_matches` table?
- Should asset identity attributes (`vendor`, `serial`, `system_mac`) eventually be treated like facts with evidence and conflict states instead of mutable columns on `assets`?
- How should provisional placeholder assets created from relationship references be represented once identity candidates exist?
- Should `parser_results.warnings` include identity-review warnings when candidates are pending, or should warnings remain parser-quality-only?
