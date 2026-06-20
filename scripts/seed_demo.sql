-- Seed Truthwatcher with a deterministic v0.1.0-alpha.1 demo topology.
-- Requires the embedded migrations to be applied first.
-- Make the seed idempotent by removing any previous demo data first.
\i scripts/unseed_demo.sql

BEGIN;

INSERT INTO discovery_runs (id, status, seed_input, started_at, completed_at, created_at, updated_at)
VALUES
  ('10000000-0000-0000-0000-000000000001', 'completed', '{"demo":"v0.1.0-alpha.1","seed":"truthwatcher-demo","scope":"service-provider-metro"}', now() - interval '3 hours', now() - interval '2 hours 45 minutes', now() - interval '3 hours', now() - interval '2 hours 45 minutes'),
  ('10000000-0000-0000-0000-000000000002', 'completed', '{"demo":"v0.1.0-alpha.1","seed":"truthwatcher-demo","scope":"identity-review"}', now() - interval '90 minutes', now() - interval '75 minutes', now() - interval '90 minutes', now() - interval '75 minutes');

INSERT INTO devices (id, hostname, management_ip, vendor, model, serial_number, role, site, created_at, updated_at)
VALUES
  ('20000000-0000-0000-0000-000000000001','nyc1-core-01.corp.example.net','10.10.0.11','Juniper','MX480','JN-demo-nyc1-core-01','core','nyc1',now(),now()),
  ('20000000-0000-0000-0000-000000000002','nyc1-core-02.corp.example.net','10.10.0.12','Juniper','MX480','JN-demo-nyc1-core-02','core','nyc1',now(),now()),
  ('20000000-0000-0000-0000-000000000003','nyc1-pe-01.corp.example.net','10.10.1.21','Cisco','ASR 9001','FOC-demo-nyc1-pe-01','provider_edge','nyc1',now(),now()),
  ('20000000-0000-0000-0000-000000000004','nyc1-pe-02.corp.example.net','10.10.1.22','Cisco','ASR 9006','FOC-demo-nyc1-pe-02','provider_edge','nyc1',now(),now()),
  ('20000000-0000-0000-0000-000000000005','nyc1-agg-01.corp.example.net','10.10.2.31','Arista','7280SR3','JPE-demo-nyc1-agg-01','aggregation','nyc1',now(),now()),
  ('20000000-0000-0000-0000-000000000006','nyc1-agg-02.corp.example.net','10.10.2.32','Arista','7280SR3','JPE-demo-nyc1-agg-02','aggregation','nyc1',now(),now()),
  ('20000000-0000-0000-0000-000000000007','chi1-core-01.corp.example.net','10.20.0.11','Juniper','MX960','JN-demo-chi1-core-01','core','chi1',now(),now()),
  ('20000000-0000-0000-0000-000000000008','chi1-pe-01.corp.example.net','10.20.1.21','Cisco','NCS 5501','FOC-demo-chi1-pe-01','provider_edge','chi1',now(),now()),
  ('20000000-0000-0000-0000-000000000009','dal1-core-01.corp.example.net','10.30.0.11','Nokia','7750 SR-7','NS-demo-dal1-core-01','core','dal1',now(),now()),
  ('20000000-0000-0000-0000-000000000010','dal1-pe-01.corp.example.net','10.30.1.21','Cisco','ASR 9901','FOC-demo-dal1-pe-01','provider_edge','dal1',now(),now()),
  ('20000000-0000-0000-0000-000000000011','sea1-edge-01.corp.example.net','10.40.1.21','Juniper','ACX7100','JN-demo-sea1-edge-01','edge','sea1',now(),now()),
  ('20000000-0000-0000-0000-000000000012','atl1-edge-01.corp.example.net','10.50.1.21','Cisco','Catalyst 8500','FOC-demo-atl1-edge-01','edge','atl1',now(),now()),
  ('20000000-0000-0000-0000-000000000013','nyc1-oob-sw01.corp.example.net','10.10.99.10','Cisco','Catalyst 9300','FCW-demo-nyc1-oob-sw01','oob_management','nyc1',now(),now()),
  ('20000000-0000-0000-0000-000000000014','route-reflector-01.corp.example.net','10.60.0.10','Juniper','vRR','VRR-demo-rr-01','route_reflector','iad1',now(),now()),
  ('20000000-0000-0000-0000-000000000015','ipam01.corp.example.net','10.70.0.15','NetBox Labs','NetBox','NB-demo-01','source_of_truth','iad1',now(),now()),
  ('20000000-0000-0000-0000-000000000016','ems01.corp.example.net','10.70.0.20','Juniper','Apstra','AP-demo-01','ems','iad1',now(),now());

INSERT INTO evidence (id, discovery_run_id, target, method, command_or_api, raw_output, raw_output_hash, parser_name, collected_at, metadata)
VALUES
  ('30000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','nyc1-core-01.corp.example.net','ssh','show version','Hostname: nyc1-core-01.corp.example.net\nModel: MX480\nJunos: 22.4R3-S2\nSerial: JN-demo-nyc1-core-01','demo-hash-001','junos_show_version',now() - interval '170 minutes','{"demo_seed":true,"source":"ssh"}'),
  ('30000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000001','nyc1-core-01.corp.example.net','ssh','show lldp neighbors','nyc1-core-02.corp.example.net ge-0/0/0\nnyc1-pe-01.corp.example.net xe-1/0/0\nroute-reflector-01.corp.example.net ae10','demo-hash-002','junos_lldp_neighbors',now() - interval '169 minutes','{"demo_seed":true,"source":"ssh"}'),
  ('30000000-0000-0000-0000-000000000003','10000000-0000-0000-0000-000000000001','nyc1-pe-01.corp.example.net','ssh','show bgp summary','Local AS 64512 router-id 198.51.100.21\nNeighbor 203.0.113.10 AS 64512 Established\nNeighbor 203.0.113.20 AS 64513 Active','demo-hash-003','iosxr_bgp_summary',now() - interval '166 minutes','{"demo_seed":true,"source":"ssh"}'),
  ('30000000-0000-0000-0000-000000000004','10000000-0000-0000-0000-000000000002','ipam01.corp.example.net','api','GET /api/dcim/devices/','NetBox demo export: 16 devices, 6 sites, 4 vendors','demo-hash-004','netbox_inventory',now() - interval '88 minutes','{"demo_seed":true,"source":"api"}'),
  ('30000000-0000-0000-0000-000000000005','10000000-0000-0000-0000-000000000002','ems01.corp.example.net','api','GET /api/blueprints/service-provider/fabric','Apstra demo export: expected lldp adjacencies and intent tags','demo-hash-005','apstra_fabric',now() - interval '85 minutes','{"demo_seed":true,"source":"api"}');

INSERT INTO assets (id, asset_type, identity_key, vendor, model, serial, system_mac, confidence, confidence_reason, state, metadata, created_at, updated_at)
SELECT id, 'network_device', lower(hostname), vendor, model, serial_number, NULL, 0.93, 'seeded demo asset based on mock registry and evidence', 'observed', jsonb_build_object('demo_seed', true, 'site', site, 'role', role, 'management_ip', management_ip), now(), now()
FROM devices
WHERE role NOT IN ('source_of_truth','ems');

INSERT INTO assets (id, asset_type, identity_key, vendor, model, serial, confidence, confidence_reason, state, metadata, created_at, updated_at)
VALUES
 ('20000000-0000-0000-0000-000000000015','inventory_system','ipam01.corp.example.net','NetBox Labs','NetBox','NB-demo-01',0.80,'seeded as external inventory source for demo','user_seeded','{"demo_seed":true,"integration":"ipam"}',now(),now()),
 ('20000000-0000-0000-0000-000000000016','ems_system','ems01.corp.example.net','Juniper','Apstra','AP-demo-01',0.80,'seeded as external EMS source for demo','user_seeded','{"demo_seed":true,"integration":"ems"}',now(),now());

INSERT INTO facts (id, asset_id, name, value, source, confidence, confidence_reason, state, evidence_id, created_at)
SELECT ('40000000-0000-0000-0000-' || lpad(row_number() over ()::text, 12, '0'))::uuid, a.id, 'management_ip', to_jsonb(d.management_ip), 'demo_seed_registry', 0.90, 'seeded from mock registry', 'observed', '30000000-0000-0000-0000-000000000004', now()
FROM assets a JOIN devices d ON a.identity_key = lower(d.hostname);

INSERT INTO facts (id, asset_id, name, value, source, confidence, confidence_reason, state, evidence_id, created_at)
VALUES
 ('41000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000001','software_version','"Junos 22.4R3-S2"','demo_seed_evidence',0.95,'parsed from mock show version evidence','observed','30000000-0000-0000-0000-000000000001',now()),
 ('41000000-0000-0000-0000-000000000002','20000000-0000-0000-0000-000000000003','bgp_local_as','64512','demo_seed_evidence',0.92,'parsed from mock BGP summary','observed','30000000-0000-0000-0000-000000000003',now()),
 ('41000000-0000-0000-0000-000000000003','20000000-0000-0000-0000-000000000003','bgp_neighbor_state','{"neighbor":"203.0.113.20","remote_as":64513,"state":"Active"}','demo_seed_evidence',0.78,'mock evidence intentionally shows one non-established peer','observed','30000000-0000-0000-0000-000000000003',now()),
 ('41000000-0000-0000-0000-000000000004','20000000-0000-0000-0000-000000000009','vendor','"Nokia"','demo_seed_registry',0.70,'registry says Nokia while later candidate review can exercise identity handling','user_seeded','30000000-0000-0000-0000-000000000004',now()),
 ('41000000-0000-0000-0000-000000000005','20000000-0000-0000-0000-000000000014','route_reflector_cluster','"rr-cluster-east"','demo_seed_registry',0.85,'seeded route reflector context','user_seeded','30000000-0000-0000-0000-000000000004',now()),
 ('41000000-0000-0000-0000-000000000006','20000000-0000-0000-0000-000000000015','inventory_scope','["devices","sites","vendors","management_ips"]','demo_seed_registry',0.80,'seeded integration capability','user_seeded','30000000-0000-0000-0000-000000000004',now()),
 ('41000000-0000-0000-0000-000000000007','20000000-0000-0000-0000-000000000016','intent_source','"fabric adjacency expectations"','demo_seed_ems',0.80,'seeded EMS capability','user_seeded','30000000-0000-0000-0000-000000000005',now());

INSERT INTO relationships (id, source_asset_id, target_asset_id, relationship_type, confidence, confidence_reason, state, evidence_id, metadata, created_at, updated_at)
VALUES
 ('50000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000002','lldp_neighbor',0.96,'mock LLDP evidence','observed','30000000-0000-0000-0000-000000000002','{"demo_seed":true,"interface":"ge-0/0/0"}',now(),now()),
 ('50000000-0000-0000-0000-000000000002','20000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000003','lldp_neighbor',0.94,'mock LLDP evidence','observed','30000000-0000-0000-0000-000000000002','{"demo_seed":true,"interface":"xe-1/0/0"}',now(),now()),
 ('50000000-0000-0000-0000-000000000003','20000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000014','bgp_session',0.88,'mock BGP and route-reflector seed evidence','observed','30000000-0000-0000-0000-000000000003','{"demo_seed":true,"afi_safi":["inet-vpn","evpn"]}',now(),now()),
 ('50000000-0000-0000-0000-000000000004','20000000-0000-0000-0000-000000000003','20000000-0000-0000-0000-000000000004','evpn_peer',0.76,'seeded EMS intent; lower confidence until direct evidence arrives','inferred','30000000-0000-0000-0000-000000000005','{"demo_seed":true,"service":"business-internet"}',now(),now()),
 ('50000000-0000-0000-0000-000000000005','20000000-0000-0000-0000-000000000015','20000000-0000-0000-0000-000000000001','inventory_claims',0.80,'mock NetBox source-of-truth claim','user_seeded','30000000-0000-0000-0000-000000000004','{"demo_seed":true}',now(),now()),
 ('50000000-0000-0000-0000-000000000006','20000000-0000-0000-0000-000000000016','20000000-0000-0000-0000-000000000005','ems_manages',0.80,'mock Apstra managed fabric','user_seeded','30000000-0000-0000-0000-000000000005','{"demo_seed":true,"blueprint":"service-provider"}',now(),now());

INSERT INTO parser_results (id, discovery_run_id, evidence_id, parser_name, status, warnings, created_at)
VALUES
 ('60000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000001','junos_show_version','parsed','[]',now()),
 ('60000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000003','iosxr_bgp_summary','parsed','["neighbor 203.0.113.20 is not established"]',now());

INSERT INTO identity_candidates (id, discovery_run_id, evidence_id, parser_name, asset_type, candidate_identity_key, strength, confidence, reason, vendor, model, serial, hostname, proposed_asset_id, review_state, metadata, created_at)
VALUES
 ('70000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000001','junos_show_version','network_device','nyc1-core-01.corp.example.net','strong',0.96,'hostname plus serial matched mock evidence','Juniper','MX480','JN-demo-nyc1-core-01','nyc1-core-01.corp.example.net','20000000-0000-0000-0000-000000000001','auto_accepted','{"demo_seed":true}',now()),
 ('70000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000002','30000000-0000-0000-0000-000000000004','netbox_inventory','network_device','dal1-core-01.corp.example.net','provisional',0.66,'inventory record lacks direct device evidence in this demo','Nokia','7750 SR-7','NS-demo-dal1-core-01','dal1-core-01.corp.example.net','20000000-0000-0000-0000-000000000009','deferred','{"demo_seed":true,"needs":"direct discovery"}',now());

INSERT INTO identity_candidate_reviews (id, identity_candidate_id, discovery_run_id, evidence_id, reviewer, action, previous_review_state, resulting_review_state, rationale, effect, metadata, created_at)
VALUES
 ('80000000-0000-0000-0000-000000000001','70000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000001','truthwatcher-demo','auto_accept','pending','auto_accepted','strong identity evidence in demo seed','asset linked and alias recorded','{"demo_seed":true}',now()),
 ('80000000-0000-0000-0000-000000000002','70000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000002','30000000-0000-0000-0000-000000000004','truthwatcher-demo','defer','pending','deferred','show UI state for identity candidates needing more evidence','no asset mutation','{"demo_seed":true}',now());

INSERT INTO identity_aliases (id, asset_id, identity_candidate_id, alias_identity_key, alias_strength, evidence_id, discovery_run_id, reviewer, rationale, metadata, created_at)
VALUES
 ('90000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000001','70000000-0000-0000-0000-000000000001','serial:JN-demo-nyc1-core-01','strong','30000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','truthwatcher-demo','serial number is a strong alias for demo identity workflow','{"demo_seed":true}',now());

INSERT INTO audit_records (id, action, initiator, request_id, discovery_run_id, target, method, profile, task, command_or_api, status, evidence_id, started_at, completed_at, context, created_at)
VALUES
 ('a0000000-0000-0000-0000-000000000001','collect_evidence','truthwatcher-demo','demo-seed-001','10000000-0000-0000-0000-000000000001','nyc1-core-01.corp.example.net','ssh','juniper_junos','get_version','show version','succeeded','30000000-0000-0000-0000-000000000001',now() - interval '171 minutes',now() - interval '170 minutes','{"demo_seed":true}',now()),
 ('a0000000-0000-0000-0000-000000000002','collect_evidence','truthwatcher-demo','demo-seed-002','10000000-0000-0000-0000-000000000002','ipam01.corp.example.net','api','netbox','inventory_import','GET /api/dcim/devices/','succeeded','30000000-0000-0000-0000-000000000004',now() - interval '89 minutes',now() - interval '88 minutes','{"demo_seed":true}',now());

COMMIT;

SELECT 'truthwatcher demo seed complete' AS message,
       (SELECT count(*) FROM devices) AS devices,
       (SELECT count(*) FROM assets) AS assets,
       (SELECT count(*) FROM evidence) AS evidence,
       (SELECT count(*) FROM relationships) AS relationships;
