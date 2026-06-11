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
