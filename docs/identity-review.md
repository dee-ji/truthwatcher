# Identity Review

Identity Review makes parser-derived identity candidates visible without treating them as canonical asset merges. It supports the v0.1.0 promise that Truthwatcher can show why an asset identity may exist, preserve uncertainty, and require review before risky identity decisions.

## Non-destructive identity model

Parsers can derive identity clues from evidence: vendor and serial numbers, system MAC addresses, hostnames, neighbor names, and other command output. Truthwatcher stores those clues as identity candidates, separate from canonical assets.

Reviewing a candidate does not rewrite `assets.identity_key`, collapse asset rows, execute discovery, call external services, or write to another system. Accepted candidates are recorded as explicit evidence-backed aliases for a proposed asset when a proposed asset is available.

## UI route

Open the embedded UI and choose **Identity Review** from the top navigation, or go directly to:

```text
http://127.0.0.1:8080/#/identity-review
```

The page shows pending candidates by default. Operators can filter by review state, identity strength, asset type, discovery run ID, proposed asset ID, and search text. Asset type, proposed asset ID, and broad search are UI-side filters because the current API exposes server-side filters for discovery run ID, evidence ID, review state, strength, and candidate identity key.

## Candidate fields

The candidate list and detail panel show the following when available:

- candidate identity key
- asset type
- strength and confidence
- review state
- vendor, model, serial, system MAC, and hostname
- parser name
- discovery run ID
- evidence ID
- proposed asset ID
- created timestamp
- parser metadata rendered as JSON

Evidence links open the existing read-only evidence drawer. Discovery run links open the discovery run detail route. Proposed asset links open the asset detail route when a proposed asset ID exists.

## Strengths

- `strong`: durable identity evidence, such as vendor plus serial or system MAC, that can be reviewed as a high-confidence clue.
- `provisional`: useful but non-authoritative identity evidence, such as a hostname or neighbor name that may change or be duplicated.
- `weak`: incomplete or ambiguous evidence that should not drive risky identity decisions without more context.

## Review states

- `pending`: waiting for human review.
- `accepted`: accepted by an operator as an evidence-backed alias for the proposed asset.
- `rejected`: rejected as an identity clue.
- `deferred`: explicitly left undecided for later review.
- `more_evidence_requested`: more evidence is needed before deciding.
- `auto_accepted`: deterministic parser auto-acceptance for strong candidates with no plausible conflict. The UI displays this state but does not expose manual `auto_accept`.
- `superseded`: retained historical state for candidates replaced by later identity evidence.

## Review actions

The UI supports the manual actions already exposed by the API:

- `accept`
- `reject`
- `defer`
- `request_more_evidence`

Reject, defer, and request-more-evidence actions require a short review note. Accepting a candidate requires the candidate to already have a `proposed_asset_id`; if it does not, the UI asks the reviewer to defer or request more evidence instead.

## Handoff report

Truthwatcher includes a read-only API handoff report at:

```text
GET /api/v1/identity-candidates/handoff-report
```

The report is derived review output for downstream human intake. It is not raw evidence, not an accepted external decision, and not a write integration.
