# ADR 0003: Driver abstraction

## Status
Accepted

## Context
Truthwatcher needs a clear path from declared intent to vendor configuration artifacts without tightly coupling compilation logic to any single vendor syntax.

## Decision
Adopt a two-stage compile model:
1. Elsecall compiles raw intent JSON/YAML into a vendor-neutral intermediate representation (`DeviceConfigIR`).
2. Vendor drivers implement a small interface (`Vendor()`, `Render(context.Context, DeviceConfigIR)`) and return deterministic rendered artifacts with metadata.

The first production slice supports Junos under `drivers/vendor/junos` and is intentionally narrow:
- host-name
- BGP ASN under `routing-options`
- explicit TODO markers for unsupported intent sections

## Consequences
- Vendor-neutral IR remains separate from renderers.
- Driver unit tests can validate deterministic output with golden-like assertions.
- Spanreed can request a specific vendor at compile time (`POST /api/v1/intents/{id}/compile` body `{ "vendor": "junos" }`).
- Stored artifacts now include metadata and format for downstream workflows.

## Non-goals (current phase)
- No plugin framework/dynamic module loading.
- No broad vendor parity.
- No full intent coverage in renderer yet; unsupported sections produce TODO markers.
