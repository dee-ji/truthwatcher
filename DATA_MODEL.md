# TruthWatcher Data Model

## Data Model Philosophy

TruthWatcher should not attempt to perfectly model every vendor object upfront.

The model should use stable nouns and flexible facts.

Stable nouns:

- Asset.
- Fact.
- Relationship.
- Evidence.
- DiscoveryRun.
- AccessPath.
- CredentialRef.
- DiscoveryProfile.

Dynamic details belong in JSONB facts and metadata.

## Identity Philosophy

Hostnames and IP addresses are not reliable long-term identities.

Better physical identity anchors include:

- Chassis serial number.
- Asset tag.
- System MAC address.
- Vendor-assigned hardware ID.
- Contract/vendor records.

However, identity is complicated:

- Chassis have serials.
- Cards have serials.
- Optics have serials.
- Virtual devices may not have physical serials.
- Stacks/clusters may have multiple serials.
- Hardware replacements change serials.

TruthWatcher should allow multiple identity keys and track confidence.

Implementation rule:

- `identity_key` is a system-generated stable key, not a raw hostname or IP address.
- Prefer keys derived from stronger evidence such as vendor plus serial, system MAC, asset tag, or another durable external identifier.
- Hostnames and IP addresses may be stored as facts, but they must not be treated as globally unique asset identity.
- When only weak identifiers are known, generate a provisional identity key and replace or merge it later when stronger evidence exists.

## Core Tables

### assets

Represents anything that can exist in the network model.

Examples:

- Device.
- Chassis.
- Card.
- Port.
- Optic.
- Site.
- Rack.
- EMS.
- Controller.
- Service.
- Customer premise.
- Route reflector.
- Provider edge.
- Core router.
- Backbone node.

Suggested fields:

```sql
id uuid primary key,
asset_type text not null,
name text,
identity_key text,
vendor text,
model text,
serial_number text,
system_mac text,
status text,
confidence numeric,
metadata jsonb not null default '{}',
created_at timestamptz not null default now(),
updated_at timestamptz not null default now()
```

Identity constraints should be cautious. Use partial unique indexes where confidence is high.

Examples:

- Unique physical serial per vendor when serial is known and non-empty.
- Unique system MAC when known and non-empty.
- Generated UUID for logical assets.

### facts

Represents observed, inferred, imported, or human-confirmed properties of an asset.

Suggested fields:

```sql
id uuid primary key,
asset_id uuid not null references assets(id),
fact_name text not null,
fact_value jsonb not null,
fact_source text not null,
confidence numeric not null,
evidence_id uuid references evidence(id),
valid_from timestamptz,
valid_to timestamptz,
created_at timestamptz not null default now()
```

Examples:

- software_version.
- management_ip.
- loopback_ip.
- bgp_asn.
- platform.
- os_family.
- interface_admin_state.
- interface_oper_state.
- optic_wavelength.
- route_reflector_client_count.

### relationships

Represents graph edges between assets.

Suggested fields:

```sql
id uuid primary key,
source_asset_id uuid not null references assets(id),
target_asset_id uuid not null references assets(id),
relationship_type text not null,
confidence numeric not null,
evidence_id uuid references evidence(id),
metadata jsonb not null default '{}',
created_at timestamptz not null default now(),
updated_at timestamptz not null default now()
```

Relationship examples:

- contains.
- installed_in.
- has_port.
- has_optic.
- connected_to.
- lldp_neighbor_of.
- cdp_neighbor_of.
- bgp_peer_of.
- ospf_neighbor_of.
- isis_neighbor_of.
- route_reflector_for.
- managed_by.
- accessed_through.
- belongs_to_site.
- provides_service.
- consumes_service.
- customer_edge_for.
- provider_edge_for.
- part_of_market.
- part_of_backbone.

### evidence

Stores raw discovery output or external source data.

Suggested fields:

```sql
id uuid primary key,
discovery_run_id uuid references discovery_runs(id),
target text not null,
method text not null,
command_or_request text,
raw_output text,
raw_output_hash text,
parser_name text,
parser_version text,
metadata jsonb not null default '{}',
collected_at timestamptz not null default now()
```

Evidence examples:

- CLI command output.
- NETCONF RPC response.
- REST API response.
- SNMP walk result.
- EMS export.
- Config archive.
- User-provided seed file.

### discovery_runs

Represents a bounded discovery operation.

Suggested fields:

```sql
id uuid primary key,
status text not null,
seed_input jsonb not null default '{}',
started_at timestamptz not null default now(),
completed_at timestamptz,
error text,
metadata jsonb not null default '{}'
```

Statuses:

- pending.
- running.
- completed.
- failed.
- partially_completed.
- canceled.

### access_paths

Represents how to access something.

Suggested fields:

```sql
id uuid primary key,
asset_id uuid references assets(id),
access_type text not null,
target text not null,
credential_ref_id uuid references credential_refs(id),
path jsonb not null default '[]',
confidence numeric,
metadata jsonb not null default '{}',
created_at timestamptz not null default now()
```

Access types:

- ssh.
- netconf.
- restconf.
- gnmi.
- snmp.
- ems_api.
- cloud_api.
- terminal_server.
- jump_host.

Example path:

```json
[
  {"type": "jump_host", "target": "jump01"},
  {"type": "terminal_server", "target": "ts01", "port": 2007},
  {"type": "device_console", "target": "router01"}
]
```

### credential_refs

Represents references to credentials, not raw secrets.

Suggested fields:

```sql
id uuid primary key,
name text not null,
provider text not null,
reference text not null,
metadata jsonb not null default '{}',
created_at timestamptz not null default now()
```

Providers:

- env.
- file.
- vault.
- delinea.
- cyberark.
- onepassword.
- static_dev_only.

### discovery_profiles

Represents vendor/platform-specific read-only discovery instructions.

Suggested fields:

```sql
id uuid primary key,
name text not null,
vendor text,
platform text,
os_family text,
transport text not null,
tasks jsonb not null,
created_at timestamptz not null default now()
```

Example tasks:

```json
{
  "identify_device": ["show version", "show inventory"],
  "get_neighbors": ["show lldp neighbors"],
  "get_bgp": ["show bgp summary"]
}
```

## Indexing Strategy

Start with:

- `assets(asset_type)`.
- `assets(vendor)`.
- `assets(serial_number)` partial index where serial is not null.
- `assets(system_mac)` partial index where system_mac is not null.
- `relationships(source_asset_id)`.
- `relationships(target_asset_id)`.
- `relationships(relationship_type)`.
- `facts(asset_id, fact_name)`.
- `evidence(discovery_run_id)`.
- `evidence(raw_output_hash)`.
- GIN indexes on important JSONB metadata later.

## Avoid Premature Dynamic Tables

Do not generate tables per vendor or per device type in early versions.

Use:

- stable assets table.
- facts table.
- relationships table.
- JSONB metadata.

This provides flexibility without sacrificing consistency.
