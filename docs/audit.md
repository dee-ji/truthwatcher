# Audit records

Truthwatcher audit records are read-only safety records created during discovery execution. They help a user inspect what Truthwatcher attempted to do and how each action relates back to raw evidence and a discovery run.

## What audit records prove

An audit record can show:

- who or what initiated the action (`initiator`, for example `cli`, `api`, or `unknown`);
- which request ID was attached when the action came through the API;
- the discovery run and evidence record associated with the action;
- the target, method, profile, task, and command/API that Truthwatcher mapped;
- whether the action started, completed, failed, or stored evidence;
- the started and completed timestamps; and
- a redacted error message when one is available.

For v0.1.0 this is intentionally local and minimal. It is meant to support fixture-backed discovery review and explain what Truthwatcher recorded before later parsing or graph inspection.

## What audit records do not prove

Audit records do not prove that a production network device was contacted, that modeled assets are fully correct, or that a user approved a future action outside Truthwatcher. They are not authentication logs, RBAC records, SIEM events, or compliance attestations. They should be interpreted together with raw evidence, parser results, identity review state, and the known v0.1.0 limitations.

## Read-only inspection

Audit inspection is read-only. Truthwatcher exposes `GET /api/v1/audit-records` and the embedded `#/audit` page for review. There are no v0.1.0 audit create, update, or delete API endpoints for users. Audit records are created by discovery execution paths.

## Secret handling

Audit records should not contain raw credentials or secrets. Truthwatcher redacts sensitive assignment-like text before writing audit fields such as target, command/API, and error. Raw evidence remains separate and evidence-first; do not use audit fields for credential storage or secret transport.

## API filters

`GET /api/v1/audit-records` supports small, read-only filtering with a safe default limit:

- `discovery_run_id`
- `evidence_id`
- `request_id`
- `action`
- `status`
- `target`
- `method`
- `profile`
- `limit` (default `50`, maximum `200`)

Use the discovery run and evidence IDs to move from audit intent to the related persisted evidence.
