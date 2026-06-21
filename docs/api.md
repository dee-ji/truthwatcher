# Truthwatcher API

Truthwatcher exposes a small JSON HTTP API from the single `truthwatcher server` binary.

## Response Envelope

All API responses use the same envelope:

```json
{
  "data": {},
  "error": null,
  "metadata": {}
}
```

Error responses use the same shape:

```json
{
  "data": null,
  "error": {
    "message": "target is required"
  },
  "metadata": {}
}
```

## Health

### `GET /healthz`

Returns process health.

### `GET /readyz`

Returns process readiness.

## Version

### `GET /api/v1/version`

Returns binary name and version.

## Architecture Seeds

Architecture seeds let a user provide planning context without claiming observed truth.

Seeded hints are stored as facts with:

- `source`: `user_seeded`
- `state`: `user_seeded`
- low deterministic confidence

Seeded hints are not proof. They may guide discovery planning, but they do not authorize discovery execution and they do not replace observed evidence.

The embedded UI exposes the same seed workflow at:

```text
/#/architecture-seeds
```

That page stores context only. It does not trigger discovery, approve execution, or convert seeded hints into observed facts.

### `POST /api/v1/architecture-seeds`

Stores user-provided architecture hints.

Request:

```json
{
  "organization_network_type": "service_provider",
  "known_asns": ["65000"],
  "known_route_reflectors": ["rr1.example.net"],
  "known_vendors": ["juniper"],
  "known_ems_systems": ["ems-a"],
  "known_services": ["l3vpn"],
  "known_regions_markets": ["nyc"]
}
```

Response `data`:

```json
{
  "architecture_seed": {
    "asset": {
      "type": "architecture_context",
      "identity_key": "architecture:seed:default",
      "state": "user_seeded"
    },
    "facts": [],
    "warning": "architecture hints are context only and must not be treated as observed network facts"
  }
}
```

## Local JSON Import and Export

The CLI exposes local file connector commands for graph snapshots:

```text
truthwatcher export json --output <path>
truthwatcher import json --input <path>
```

Export writes assets, facts, relationships, and evidence metadata. It does not export raw evidence output by default.

Import validates candidates and prints a summary. It does not persist records or treat imported data as observed proof.

See [Local JSON Import and Export](import-export.md).

## Discovery Plans

Discovery plans suggest safe read-only next steps from current graph data and user seed input.

Plans are not executable approvals:

- `approval_required` is always `true`.
- `execution_allowed` is always `false`.
- Scope-expanding targets such as CIDRs, wildcards, and comma-separated lists are rejected.
- Seeded architecture hints may inform suggested tasks or profile defaults, but they remain unobserved context.

The embedded UI exposes the same planning workflow at:

```text
/#/discovery-plans
```

That page submits seed input to this endpoint and renders suggested steps for review. It does not execute plans or add an approval execution path.

### `POST /api/v1/discovery-plans`

Request:

```json
{
  "target": "router-a",
  "method": "ssh",
  "profile": "juniper_junos"
}
```

Response `data`:

```json
{
  "discovery_plan": {
    "approval_required": true,
    "execution_allowed": false,
    "steps": [
      {
        "target": "router-a",
        "method": "ssh",
        "profile": "juniper_junos",
        "task": "get_neighbors",
        "reason": "stored graph has no relationships for this asset",
        "expected_evidence": "raw output from show lldp neighbors for get_neighbors",
        "risk_level": "low_read_only"
      }
    ]
  }
}
```

## Discovery Runs

Discovery APIs are explicit and evidence-first. Discovery execution creates a discovery run, validates the requested profile and tasks against policy, collects fixture-backed outputs, stores raw evidence, and only then returns the completed run.

No authentication exists yet. Until auth is added, execution audit metadata identifies the initiator as `api`.

### `POST /api/v1/discovery-runs`

Creates a pending discovery run record without executing collection.

Request:

```json
{
  "seed_input": {
    "target": "fixture://junos-mx"
  }
}
```

Response `data`:

```json
{
  "discovery_run": {
    "id": "11111111-1111-4111-8111-111111111111",
    "status": "pending",
    "seed_input": {
      "target": "fixture://junos-mx"
    }
  }
}
```

### `GET /api/v1/discovery-runs`

Lists discovery runs.

Response `data`:

```json
{
  "discovery_runs": []
}
```

### `GET /api/v1/discovery-runs/{id}`

Returns one discovery run.

Response `data`:

```json
{
  "discovery_run": {
    "id": "11111111-1111-4111-8111-111111111111",
    "status": "completed"
  }
}
```

### `POST /api/v1/discovery-runs/execute`

Executes a discovery run using the local fake collector. This endpoint is intentionally explicit:

- `collector` is required and must be `fake`.
- `target` is required and must use `fixture://`.
- `profile` is required and must be a built-in profile.
- `tasks` is required and each task must be allowed by policy and mapped by the selected profile.

Request:

```json
{
  "collector": "fake",
  "target": "fixture://junos-mx",
  "profile": "juniper_junos",
  "tasks": ["identify_device", "get_inventory"],
  "fixture_root": "examples/fixtures"
}
```

Response `data`:

```json
{
  "discovery_run": {
    "id": "11111111-1111-4111-8111-111111111111",
    "status": "completed"
  },
  "evidence": []
}
```

Response `metadata.audit`:

```json
{
  "initiator": "api",
  "collector": "fake",
  "target": "fixture://junos-mx",
  "profile": "juniper_junos",
  "tasks": ["identify_device", "get_inventory"],
  "discovery_run": "11111111-1111-4111-8111-111111111111",
  "run_status": "completed",
  "evidence_count": 2
}
```

### `POST /api/v1/discovery-runs/{id}/parse`

Parses already-stored evidence for one discovery run and persists derived assets, facts, and relationships. Parser persistence may also record identity candidates for review. This endpoint does not run discovery, does not touch a network, records parser warnings without deleting raw evidence, and does not merge assets or rewrite canonical asset identity.

Built-in fixture parser coverage currently includes Junos and IOS-XR version, inventory, LLDP neighbor, and BGP summary outputs. BGP summary parsing stores routing-context and peer placeholders, evidence-linked BGP peer facts, and `bgp_peer_of` relationships when the fixture output includes enough peer data.

Request:

```json
{
  "platform": "junos"
}
```

Response `data`:

```json
{
  "parse_result": {
    "discovery_run_id": "11111111-1111-4111-8111-111111111111",
    "evidence_count": 4,
    "parse_results": [
      {
        "evidence_id": "22222222-2222-4222-8222-222222222222",
        "parser_name": "junos_show_version",
        "status": "parsed",
        "warnings": []
      }
    ],
    "identity_candidates": [],
    "assets": [],
    "facts": [],
    "relationships": []
  }
}
```

## Identity Candidates

Identity candidates are read-only review records derived from parser evidence. They preserve parser-derived identity clues separately from canonical assets so hostname, neighbor-name, serial, system MAC, and similar clues can be inspected without silently merging or rewriting assets.

Strong vendor+serial or system-MAC candidates may be marked `auto_accepted` only when deterministic parser persistence finds no plausible conflict with existing canonical asset identifiers. Hostname, name, weak, provisional, ambiguous, or conflicting candidates remain `pending`. Candidate `metadata` includes operator-visible `identity_review_rule` and `identity_review_explanation` fields describing why a candidate was auto-accepted or queued.

### `GET /api/v1/identity-candidates`

Optional query filters:

- `discovery_run_id`
- `evidence_id`
- `review_state`: `pending`, `auto_accepted`, `accepted`, `rejected`, `superseded`, `deferred`, or `more_evidence_requested`
- `strength`: `strong`, `provisional`, or `weak`
- `candidate_identity_key`

Response `data`:

```json
{
  "identity_candidates": [
    {
      "id": "33333333-3333-4333-8333-333333333333",
      "discovery_run_id": "11111111-1111-4111-8111-111111111111",
      "evidence_id": "22222222-2222-4222-8222-222222222222",
      "parser_name": "junos_show_version",
      "asset_type": "device",
      "candidate_identity_key": "device:hostname:mx-edge-01",
      "strength": "provisional",
      "confidence": 0.55,
      "reason": "hostname is not globally unique and may change",
      "hostname": "mx-edge-01",
      "review_state": "pending",
      "metadata": {
        "identity_review_rule": "queue_non_strong_candidate",
        "identity_review_explanation": "queued for review because hostname, name, weak, or provisional identity evidence is not silently authoritative"
      }
    }
  ]
}
```

### `GET /api/v1/identity-candidates/review-queue`

Lists pending identity candidates for review. This is a read-only queue view; it does not execute discovery, merge assets, or rewrite canonical identity.

Supports the same optional filters as `GET /api/v1/identity-candidates`, except `review_state` is fixed to `pending`.

### `GET /api/v1/identity-candidates/handoff-report`

Returns a concise read-only identity-review handoff summary for Mistspren intake/workbench review. The report is derived review output, not raw evidence, not an accepted ADR, and not an authoritative Mistspren decision.

Optional query filters:

- `discovery_run_id`
- `evidence_id`

Response `data`:

```json
{
  "identity_review_handoff": {
    "report_type": "identity_review_handoff",
    "boundary": "Truthwatcher derived review output for Mistspren intake/workbench review; not an accepted ADR or authoritative Mistspren decision",
    "derived_output": true,
    "entries": [
      {
        "handoff_status": "ready_for_mistspren_review",
        "output_label": "derived_identity_review_output_not_raw_evidence",
        "candidate": {
          "id": "33333333-3333-4333-8333-333333333333",
          "evidence_id": "22222222-2222-4222-8222-222222222222",
          "parser_name": "junos_show_chassis_hardware",
          "candidate_identity_key": "chassis:vendor_serial:juniper:jn1234abcdef",
          "review_state": "auto_accepted"
        },
        "latest_review": {
          "id": "44444444-4444-4444-8444-444444444444",
          "action": "auto_accept",
          "rationale": "auto-accepted because durable identity has no plausible conflict",
          "effect": "deterministically auto-accepted evidence-backed strong identity candidate; no canonical asset merge or identity rewrite performed"
        },
        "evidence_reference": {
          "evidence_id": "22222222-2222-4222-8222-222222222222",
          "discovery_run_id": "11111111-1111-4111-8111-111111111111",
          "present": true
        },
        "parser_source": {
          "parser_name": "junos_show_chassis_hardware"
        },
        "review_summary": "auto-accepted because durable identity has no plausible conflict",
        "identity_effect": "deterministically auto-accepted evidence-backed strong identity candidate; no canonical asset merge or identity rewrite performed",
        "mistspren_intake_note": "derived Truthwatcher review output for intake review only"
      }
    ],
    "integrity": {
      "missing_evidence_references": 0,
      "orphaned_review_records": 0,
      "unresolved_pending_entries": 0
    },
    "non_destructive_guarantee": "report generation is read-only and does not merge canonical assets, rewrite assets.identity_key, or write to Mistspren"
  }
}
```

Pending candidates are included as `unresolved_pending_review` entries so conflicts and incomplete decisions are not hidden.

### `POST /api/v1/identity-candidates/{id}/review`

Records a non-destructive review decision for one identity candidate and writes an audit row tied to the candidate, discovery run, and evidence record.

Allowed `action` values:

- `accept`
- `reject`
- `defer`
- `request_more_evidence`

`auto_accept` is reserved for deterministic parser decisions and is not accepted as a manual review action.

Request:

```json
{
  "reviewer": "netops",
  "action": "request_more_evidence",
  "rationale": "neighbor name needs corroborating inventory evidence",
  "metadata": {
    "ticket": "TW-123"
  }
}
```

Response `data`:

```json
{
  "identity_candidate_review": {
    "id": "44444444-4444-4444-8444-444444444444",
    "identity_candidate_id": "33333333-3333-4333-8333-333333333333",
    "discovery_run_id": "11111111-1111-4111-8111-111111111111",
    "evidence_id": "22222222-2222-4222-8222-222222222222",
    "reviewer": "netops",
    "action": "request_more_evidence",
    "previous_review_state": "pending",
    "resulting_review_state": "more_evidence_requested",
    "rationale": "neighbor name needs corroborating inventory evidence",
    "effect": "review requested more evidence for candidate; no discovery execution, canonical asset merge, or identity rewrite performed",
    "metadata": {
      "ticket": "TW-123"
    }
  }
}
```

Review actions update the candidate review state and append audit metadata. Accepting a candidate requires a `proposed_asset_id`; when an operator accepts that candidate, Truthwatcher records an explicit `identity_aliases` row linking the candidate identity key to that asset. This does not rewrite `assets.identity_key`, does not collapse asset records, and does not expose any new collector execution path.

## Evidence

### `GET /api/v1/discovery-runs/{id}/evidence`

Lists raw evidence for one discovery run.

Response `data`:

```json
{
  "evidence": []
}
```

### `GET /api/v1/evidence/{id}`

Returns one raw evidence record.

Response `data`:

```json
{
  "evidence": {
    "id": "22222222-2222-4222-8222-222222222222",
    "raw_output_hash": "sha256:..."
  }
}
```

## Assets, Facts, and Relationships

Assets, facts, and relationships expose uncertainty fields:

- `confidence`: deterministic score from `0` to `1`.
- `confidence_reason`: why the score/state was assigned.
- `state`: one of `observed`, `inferred`, `user_seeded`, `conflicting`, or `unknown`.
- `evidence_id`: present on facts and relationships when linked to raw evidence.

List endpoints support offset pagination:

- `limit`: optional, default `100`, maximum `500`.
- `offset`: optional, default `0`.

Pagination metadata is returned in `metadata.pagination`:

```json
{
  "limit": 100,
  "offset": 0,
  "count": 1,
  "total": 1,
  "has_next": false
}
```

### `GET /api/v1/assets/provisional-identities`

Lists assets whose identity is weak or provisional. This is a review endpoint only; it does not merge or rewrite assets.

Response `data`:

```json
{
  "assets": [
    {
      "id": "33333333-3333-4333-8333-333333333333",
      "identity_key": "device:hostname:mx-edge-01",
      "metadata": {
        "identity_strength": "provisional",
        "identity_reason": "hostname is not globally unique and may change",
        "identity_provisional": true
      }
    }
  ]
}
```

### `GET /api/v1/assets`

Lists assets.

Supported exact-match filters:

- `type`
- `vendor`
- `serial`
- `identity_key`

Supported search filters:

- `search` searches asset ID, asset type, identity key, vendor, model, serial, and system MAC with a case-insensitive substring match.

Examples:

```text
GET /api/v1/assets?type=device&vendor=juniper&limit=50&offset=0
GET /api/v1/assets?search=mx-edge
```

Response `data`:

```json
{
  "assets": []
}
```

### `GET /api/v1/assets/{id}`

Returns one asset.

Response `data`:

```json
{
  "asset": {
    "id": "asset-a",
    "type": "device",
    "identity_key": "device:serial:aaa"
  }
}
```

### `GET /api/v1/assets/{id}/history`

Returns a compact asset history timeline assembled from the asset creation record, observed facts, and relationships. Fact and relationship events include `evidence_id` when provenance is available. This is a read-only projection; it does not mutate asset identity or evidence.

Response `data`:

```json
{
  "asset": {
    "id": "asset-a"
  },
  "history": [
    {
      "event_type": "asset_created",
      "record_id": "asset-a"
    },
    {
      "event_type": "fact_observed",
      "record_id": "fact-a",
      "evidence_id": "evidence-a"
    },
    {
      "event_type": "relationship_observed",
      "record_id": "relationship-a",
      "relationship_to": "asset-b",
      "evidence_id": "evidence-b"
    }
  ]
}
```

### `GET /api/v1/assets/{id}/facts`

Lists facts for one asset.

Response `data`:

```json
{
  "facts": []
}
```

### `GET /api/v1/facts/conflicts`

Lists facts marked `conflicting` because they disagree with an existing fact for the same asset. Existing facts are not overwritten.

Response `data`:

```json
{
  "facts": [
    {
      "id": "44444444-4444-4444-8444-444444444444",
      "asset_id": "33333333-3333-4333-8333-333333333333",
      "name": "software_version",
      "state": "conflicting",
      "confidence_reason": "conflicts with existing fact 55555555-5555-4555-8555-555555555555"
    }
  ]
}
```

### `GET /api/v1/assets/{id}/relationships`

Lists relationships where the asset is either the source or target.

Response `data`:

```json
{
  "relationships": []
}
```

### `GET /api/v1/assets/{id}/evidence`

Lists evidence linked to facts or relationships for one asset.

This endpoint does not infer evidence from target strings or hostnames. Evidence must be linked through `evidence_id` on facts or relationships.

Response `data`:

```json
{
  "evidence": []
}
```

### `GET /api/v1/facts/{id}/evidence`

Lists evidence linked to one fact. A fact currently links to at most one evidence record.

Response `data`:

```json
{
  "evidence": []
}
```

## Graph

### `GET /api/v1/assets/{id}/graph`

Returns a graph projection for one asset and its direct relationships.

### `GET /api/v1/graph/neighbors?asset_id={id}`

Returns the direct neighbor graph for one asset.

Graph responses are shaped for frontend rendering:

```json
{
  "graph": {
    "nodes": [],
    "edges": []
  }
}
```
