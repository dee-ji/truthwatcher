#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${TRUTHWATCHER_DATABASE_URL:-}" ]]; then
  echo "TRUTHWATCHER_DATABASE_URL is required" >&2
  exit 1
fi

BIN="${TRUTHWATCHER_BIN:-./bin/truthwatcher}"
if [[ ! -x "$BIN" ]]; then
  echo "truthwatcher binary not found or not executable at $BIN; run make build first" >&2
  exit 1
fi

TARGET="${TRUTHWATCHER_ACCEPTANCE_TARGET:-fixture://junos-mx}"
PLATFORM="${TRUTHWATCHER_ACCEPTANCE_PLATFORM:-junos}"
FIXTURE_ROOT="${TRUTHWATCHER_ACCEPTANCE_FIXTURES:-examples/fixtures}"

echo "==> running migrations"
"$BIN" migrate up

echo "==> running fake discovery for $TARGET"
discover_output="$($BIN discover fake --target "$TARGET" --fixtures "$FIXTURE_ROOT")"
printf '%s\n' "$discover_output"

run_id="$(printf '%s\n' "$discover_output" | awk '/completed discovery run/ {print $4}' | tail -n 1)"
if [[ -z "$run_id" ]]; then
  echo "could not parse discovery run id from discovery output" >&2
  exit 1
fi

echo "==> parsing discovery run $run_id as $PLATFORM"
"$BIN" parse discovery-run --id "$run_id" --platform "$PLATFORM"

echo "==> checking persisted evidence, assets, and relationships"
read -r evidence_count asset_count relationship_count <<<"$(psql "$TRUTHWATCHER_DATABASE_URL" -v ON_ERROR_STOP=1 -At -F ' ' -c "SELECT (SELECT count(*) FROM evidence WHERE discovery_run_id = '$run_id'), (SELECT count(*) FROM assets), (SELECT count(*) FROM relationships);")"

echo "evidence_count=$evidence_count"
echo "asset_count=$asset_count"
echo "relationship_count=$relationship_count"

if [[ "$evidence_count" -lt 1 ]]; then
  echo "expected at least one evidence record for discovery run $run_id" >&2
  exit 1
fi
if [[ "$asset_count" -lt 1 ]]; then
  echo "expected at least one asset after parsing discovery run $run_id" >&2
  exit 1
fi
if [[ "$relationship_count" -lt 1 ]]; then
  echo "expected at least one relationship after parsing discovery run $run_id" >&2
  exit 1
fi

echo "acceptance-v0.1.0 passed for discovery run $run_id"
