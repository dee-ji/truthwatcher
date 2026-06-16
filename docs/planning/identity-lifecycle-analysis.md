# ADR-0002 Identity Lifecycle Implementation Analysis

Status: planning analysis only. No application code, migrations, or ADRs were changed.

This document analyzes the current Truthwatcher implementation for ADR-0002 identity lifecycle work. It distinguishes observed code facts from implementation proposals. The controlling project rule remains: preserve raw evidence before derived truth.

## Observed Code Facts

### Current Evidence-To-Parser-To-Asset Flow

1. Discovery execution can be requested through `POST /api/v1/discovery-runs/execute`, but that API path is intentionally limited to the `fake` collector and `fixture://` targets. It validates the built-in discovery profile and tasks through the policy engine before collecting.
2. Discovery execution stores raw evidence first. Evidence rows contain the discovery run, target, method, command/API, unchanged raw output, raw output hash, optional parser name, collection time, and metadata.
3. Parser persistence is a separate explicit step through `POST /api/v1/discovery-runs/{id}/parse` or the CLI parse command. It reads already-stored evidence for a discovery run and selects a built-in parser by platform and command.
4. Parser output is in-memory normalized candidate data: device identities, inventory components, interfaces, neighbors, BGP peers, parsed facts, relationships, warnings, parser name, and evidence ID.
5. Parser persistence builds an in-memory index of existing assets by normalized `identity_key`. It creates missing assets, facts, relationships, and parser result rows. Existing assets with the same identity key are reused.
6. Facts and relationships link to evidence when parser output supplies an evidence ID. Parser result rows preserve parse status and warnings, but not the full structured parser output.

### Current Tables, Types, And API Routes Involved

Raw evidence:

- Table: `evidence`.
- Type: `evidence.Evidence`.
- Fields include `discovery_run_id`, `target`, `method`, `command_or_api`, `raw_output`, `raw_output_hash`, `parser_name`, `collected_at`, and `metadata`.
- Routes: `GET /api/v1/discovery-runs/{id}/evidence`, `GET /api/v1/evidence/{id}`.

Parser output and parse metadata:

- Table: `parser_results`.
- Types: `parser.Result`, `DeviceIdentity`, `InventoryComponent`, `Interface`, `Neighbor`, `BGPPeer`, `ParsedFact`, `ParsedRelationship`, and `ParseRecord`.
- Persisted parse result fields include discovery run ID, evidence ID, parser name, status `parsed|skipped|failed`, warnings, error message, and created time.
- Route: `POST /api/v1/discovery-runs/{id}/parse`.

Assets, facts, and relationships:

- Tables: `assets`, `facts`, `relationships`.
- Types: `assets.Asset`, `assets.Fact`, `assets.Relationship`.
- `assets` stores `asset_type`, `identity_key`, optional `vendor`, `model`, `serial`, `system_mac`, confidence fields, state, metadata, and timestamps.
- `facts` stores asset ID, fact name, JSON value, source, confidence fields, state, optional evidence ID, and created time.
- `relationships` stores source asset ID, target asset ID, relationship type, confidence fields, state, optional evidence ID, metadata, and timestamps.
- Routes: `GET /api/v1/assets`, `GET /api/v1/assets/{id}`, `GET /api/v1/assets/{id}/facts`, `GET /api/v1/assets/{id}/relationships`, `GET /api/v1/assets/{id}/evidence`, `GET /api/v1/facts/{id}/evidence`.

Uncertainty, conflicts, and provisional/review-like state:

- Current state vocabulary is `observed`, `inferred`, `user_seeded`, `conflicting`, and `unknown`.
- Identity strength vocabulary is `strong`, `provisional`, and `weak`.
- Fact conflicts are detected during fact creation. A new differing fact for the same asset and fact name is stored as `conflicting`; the existing fact is not overwritten.
- Provisional identity review is read-only. `GET /api/v1/assets/provisional-identities` lists assets whose stored identity is not strong.
- Conflict review is read-only. `GET /api/v1/facts/conflicts` lists conflicting facts.

Graph projection:

- Graph is relational projection, not a separate graph database.
- Types: `graph.Node`, `graph.Edge`, `graph.Graph`.
- Routes: `GET /api/v1/assets/{id}/graph`, `GET /api/v1/graph/neighbors?asset_id=...`.

Audit metadata:

- Table: `audit_records`.
- Type: `audit.Record`.
- Discovery execution can persist action, initiator, request ID, discovery run ID, target, method, profile, task, command/API, status, evidence ID, error message, timestamps, and context.
- Parser persistence and identity review do not currently create audit records.

Adjacent seed/registry state:

- Architecture seeds store user-provided context as `user_seeded` facts under an `architecture_context` asset. They are not observed proof.
- The `devices` table and `internal/devices` package represent a local device registry/seed path. Current parser persistence does not use that table as canonical asset identity, and no discovery/parser API route exposes it.

### Where Parser-Derived Identity Clues Become Persistent State

- `parseJunosShowVersion` and `parseIOSXRShowVersion` derive device identity from hostname when version output does not include durable serial or system MAC. That yields identity keys such as `device:hostname:mx-edge-01`, with provisional identity metadata.
- Inventory parsers derive component identities from vendor plus serial, yielding strong identity keys such as `chassis:vendor_serial:juniper:jn1234abcdef`.
- LLDP parsers derive remote device identity from remote system name and local interface identity from interface name. These are provisional identity references.
- `persistResult` turns parser candidates into persistent assets, facts, and relationships.
- `ensurePlaceholderAsset` creates an asset when a fact or relationship references an identity key that does not yet exist.
- `assets.Service.CreateAsset` normalizes the identity key and annotates metadata with `identity_strength`, `identity_reason`, and `identity_provisional`.

### Risk Of Silently Mutating Canonical Asset Identity

Observed current behavior does not silently rewrite existing canonical asset identity. The parser persistence path creates assets when identity keys are missing and reuses assets when identity keys already exist. The database repository uses `INSERT` for assets, facts, relationships, evidence, parser results, and audit records; there is no parser-driven `UPDATE assets` path.

Current risks are still material:

- A provisional hostname or neighbor-name identity becomes an `assets` row immediately. It is labeled provisional, but there is no separate persisted identity candidate or decision record.
- A stronger identity discovered later becomes a separate asset unless it has the same identity key. There is no first-class proposed merge or candidate match path.
- `assets.identity_key` has a global partial unique index, but the schema does not model multiple identifiers for one real-world asset.
- Parser result rows store warnings and status only. Structured identity clues are not persisted except indirectly through created assets, facts, relationships, and metadata.
- Fact conflict handling is non-destructive, but asset identity conflict handling does not yet exist.

## Implementation Proposals

### Minimal Identity Lifecycle Vocabulary

- **Identity clue**: raw parser-observed identity material, such as hostname, serial, system MAC, vendor, model, neighbor system name, or interface name. It must reference evidence.
- **Identity candidate**: normalized identity proposal with asset type, candidate identity key, strength, confidence, reason, source parser, attributes, metadata, and evidence reference.
- **Canonical asset**: an `assets` row used by facts, relationships, APIs, and graph projection.
- **Proposed match**: a non-destructive assertion that a candidate may refer to an existing asset or another candidate.
- **Identity decision**: review state for a candidate or proposed match: `pending`, `auto_accepted`, `accepted`, `rejected`, or `superseded`.
- **Material identity change**: anything that would rewrite `assets.identity_key`, attach a stronger identity to an existing asset, or merge assets.
- **Non-destructive auto-acceptance**: recording an evidence-backed candidate, or creating a canonical asset only when no plausible existing asset/candidate conflicts.

### Proposed First Implementation Slice: Persisted Identity Candidates

Smallest testable milestone:

1. Add only an `identity_candidates` table and repository/service support. Do not implement asset merges or identity rewrites.
2. Persist parser-derived identity candidates alongside parser results with discovery run ID, evidence ID, parser name, asset type, candidate identity key, identity attributes, strength, confidence, reason, review state, and metadata.
3. Set hostname/IP/name/unknown candidates to `pending`. Strong vendor+serial or system-MAC candidates may be `auto_accepted` only when no existing canonical asset or candidate plausibly conflicts.
4. Preserve current raw evidence, asset, fact, and relationship behavior for the first slice. The candidate table is an added review surface, not a replacement for all persistence yet.
5. Add a read-only API endpoint to list identity candidates by discovery run, evidence ID, review state, strength, and identity key. This endpoint must not expose non-fake collectors or execution flows.
6. Defer audit records for acceptance/rejection transitions until those transitions exist. Candidate creation metadata plus evidence references are sufficient for the first slice.

Suggested initial table:

```sql
CREATE TABLE identity_candidates (
    id uuid PRIMARY KEY,
    discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
    evidence_id uuid NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    parser_name text NOT NULL,
    asset_type text NOT NULL,
    candidate_identity_key text NOT NULL,
    strength text NOT NULL CHECK (strength IN ('strong', 'provisional', 'weak')),
    confidence numeric NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    reason text NOT NULL,
    vendor text,
    model text,
    serial text,
    system_mac text,
    hostname text,
    proposed_asset_id uuid REFERENCES assets(id) ON DELETE SET NULL,
    review_state text NOT NULL DEFAULT 'pending'
        CHECK (review_state IN ('pending', 'auto_accepted', 'accepted', 'rejected', 'superseded')),
    metadata jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);
```

Recommended first indexes:

- Unique dedupe index on `(evidence_id, parser_name, candidate_identity_key)`.
- Lookup index on `(review_state, created_at DESC)`.
- Lookup index on `candidate_identity_key`.
- Lookup index on `discovery_run_id`.

### Candidate Tests For Fixture-Backed Parser Output

- Junos `show version` fixture persists one hostname identity candidate with strength `provisional`, parser name, evidence ID, confidence, and reason.
- IOS-XR `show version` fixture does the same for IOS-XR hostname identity.
- Junos `show chassis hardware` fixture persists vendor+serial identity candidates with strength `strong`.
- IOS-XR `show inventory` fixture persists vendor+serial identity candidates with strength `strong`.
- Re-running parse for the same evidence/parser/candidate key does not duplicate identity candidate rows.
- LLDP fixture output records neighbor system-name candidates as pending/provisional and does not merge them with existing assets.
- A parser output that implies one provisional hostname asset and one strong serial asset may describe the same device stores candidates/proposed matches only; it does not rewrite either asset.
- Unsupported commands still create skipped parser results and no identity candidates.
- Fact conflicts continue to be stored as `conflicting` facts and do not mutate asset identity metadata.

### Explicit Non-Goals

- No graph database.
- No microservices.
- No source-of-truth export, synchronization, or broad external system writeback.
- No broad observability, monitoring, alerting, or telemetry system.
- No chat UX for identity review.
- No arbitrary command execution.
- No new API/UI flow exposing SSH or any other non-fake collector.
- No automatic destructive asset merge.
- No identity rewrite approval workflow in the first slice.
- No ADR edits as part of this analysis.

### Open Questions And Conflicts

- Should future parser persistence keep creating canonical assets from provisional hostname identities, or should it persist identity candidates first and create canonical assets only for strong no-conflict cases?
- Should identity candidates reference `parser_results.id` directly, or are `evidence_id` plus `parser_name` enough for the first slice?
- What exact rule makes a strong candidate safe for `auto_accepted`? Vendor+serial and system MAC are strong today, but virtual devices, chassis replacement, and reused serial formats complicate this.
- Should `assets.vendor`, `assets.serial`, and `assets.system_mac` eventually become evidence-backed facts instead of canonical mutable columns?
- Should proposed matches live in `identity_candidates.proposed_asset_id` initially, or should the first slice include a separate `identity_matches` table?
- How should placeholder assets created only from relationships be presented once identity candidates exist?
- Should parser warnings include pending-identity-review warnings, or should warnings remain parser-quality-only?
- How should the device registry table relate to canonical assets once both local seeds and parser-derived identity candidates exist?
