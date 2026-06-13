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

When only weak identifiers are available, Truthwatcher can use provisional identity and improve it later when stronger evidence exists.

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

## Graph Views

Graph APIs project assets and relationships into frontend-friendly nodes and edges:

```text
GET /api/v1/assets/<asset-id>/graph
GET /api/v1/graph/neighbors?asset_id=<asset-id>
```

Graph views are a projection of the relational model, not a separate graph database.

## Current Boundary

Truthwatcher has parser interfaces and first fixture parsers, but automatic persistence of parser outputs into assets, facts, and relationships is not treated as complete unless the relevant workflow explicitly writes those records.
