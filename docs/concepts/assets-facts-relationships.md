# Assets, Facts, Relationships

Truthwatcher models network knowledge using three stable concepts: assets, facts, and relationships.

## Assets

An asset is something that can exist in the network model.

Examples:

- device
- chassis
- card
- port
- optic
- site
- EMS
- controller
- service
- route reflector
- provider edge

Assets have a system-generated `identity_key`. Hostnames and IP addresses are useful facts, but they are not reliable global identities.

Stronger identity anchors include:

- vendor plus serial number
- system MAC address
- vendor hardware ID
- external durable identifier

When only weak identifiers are available, Truthwatcher uses provisional identity and leaves it visible for review. It does not assume hostnames, interface names, or IP addresses are globally unique.

Current deterministic identity behavior:

- `vendor` plus `serial` creates a strong identity such as `device:vendor_serial:juniper:jn1234`.
- `system_mac` creates a strong identity such as `device:system_mac:00:11:22:33:44:55`.
- `serial` alone is treated as strong but less preferred than vendor plus serial.
- `hostname`, `ip`, and `name` identities are provisional.
- Unknown identity keys are weak.

Assets include identity metadata:

- `identity_strength`: `strong`, `provisional`, or `weak`
- `identity_reason`: why that strength was assigned
- `identity_provisional`: `true` unless the identity is strong

Review provisional or weak identities through:

```text
GET /api/v1/assets/provisional-identities
```

## Facts

A fact describes something about an asset.

Examples:

- hostname
- software version
- management address
- platform model
- interface name
- BGP ASN
- seeded region

Facts carry:

- value
- source
- confidence
- confidence reason
- state
- evidence reference, when observed

## Relationships

A relationship links two assets.

Examples:

- device has chassis
- chassis has card
- card has port
- device neighbors device
- service depends on device
- device managed by EMS

Relationships also carry confidence and evidence references.

## Confidence And State

Truthwatcher makes uncertainty explicit. Current states are:

- `observed`: directly supported by evidence
- `inferred`: derived deterministically but not directly observed
- `user_seeded`: provided by a user as context
- `conflicting`: contradicts another known fact
- `unknown`: not yet known

The model should mark conflicts instead of silently overwriting facts.

When a new fact disagrees with an existing non-conflicting fact of the same name for the same asset, Truthwatcher records the new fact with `state=conflicting` and keeps the existing fact. This is intentionally non-destructive; humans or later explicit merge workflows can review the disagreement.

Review conflicting facts through:

```text
GET /api/v1/facts/conflicts
```

## Graph Views

Graph APIs project assets and relationships into frontend-friendly nodes and edges:

```text
GET /api/v1/assets/<asset-id>/graph
GET /api/v1/graph/neighbors?asset_id=<asset-id>
```

Graph views are a projection of the relational model, not a separate graph database.

## Current Boundary

Truthwatcher can explicitly parse stored discovery-run evidence into persisted assets, facts, and relationships. It still does not perform automatic destructive merges when stronger identity evidence appears later.
