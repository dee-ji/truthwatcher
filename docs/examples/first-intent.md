# First Intent Workflow

1. Author an intent YAML in `examples/intents/leaf-fabric.yaml`.
2. Validate and create the intent with `twctl intent validate examples/intents/leaf-fabric.yaml`.
3. List intents via `GET /api/v1/intents`.
4. Retrieve one intent via `GET /api/v1/intents/{id}`.
5. Trigger control-plane validation via `POST /api/v1/intents/{id}/validate`.
6. Compile via `POST /api/v1/intents/{id}/compile` with optional body `{ "vendor": "junos" }`.
7. Inspect compiled Junos artifact by calling `GET /api/v1/intents/{id}` and reading `artifacts`.
8. Use `twctl render preview {id} --vendor=junos` for CLI preview.
9. Query audit events via `GET /api/v1/audit/events`.
