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

Parses already-stored evidence for one discovery run and persists derived assets, facts, and relationships. This endpoint does not run discovery, does not touch a network, and records parser warnings without deleting raw evidence.

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
    "assets": [],
    "facts": [],
    "relationships": []
  }
}
```

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

Example:

```text
GET /api/v1/assets?type=device&vendor=juniper&limit=50&offset=0
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
