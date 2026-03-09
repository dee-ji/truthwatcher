# Truthwatcher

Truthwatcher is an intent-driven network management system that lets engineers:

- Express high-level network intent.
- Translate intent into vendor-specific configurations.
- Deploy changes safely in phases.
- Monitor drift from desired state.

## Repository Layout

- `cmd/`: command binaries.
- `internal/`: private domain modules.
- `pkg/`: reusable public Go packages.
- `docs/`: architecture and design documentation.
- `deployments/`: deployment artifacts (including Dockerfiles).
- `configs/`: sample configuration.
- `examples/`: future usage examples.
- `test/`: cross-package/integration tests.
- `build/`: build-related assets.

## Quick Start

```bash
make build
make test
make run
```
