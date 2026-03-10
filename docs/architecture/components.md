# Components
- API/Spanreed boundary: HTTP/OpenAPI endpoints.
- Ideals: intent validation schemas and logic.
- Elsecall: compiles normalized intent specs into vendor-neutral `DeviceConfigIR`, then dispatches to vendor drivers for deterministic artifacts.
- Highstorm: deployment orchestration models.
- Stormlight: drift detection and reconciliation runs.

## Compile pipeline (first real slice)
1. Spanreed receives `POST /api/v1/intents/{id}/compile` with optional `vendor`.
2. Elsecall normalizes intent data (`metadata`, `routing_intent`, target scope, services) into `DeviceConfigIR`.
3. Elsecall invokes a vendor driver interface (`Vendor()`, `Render(...)`) for the requested target vendor.
4. Renderer output is persisted in `compiled_artifacts` with format and metadata for retrieval in `GET /api/v1/intents/{id}`.

Current concrete renderer: Junos set-format output with TODO markers for unsupported sections (for example, `interface_intent`).
