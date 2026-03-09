# Truthwatcher Architecture Overview

Truthwatcher is a Go-based, intent-driven network management platform.

## Control Plane Flow

1. **Ideals** captures high-level desired state.
2. **Elsecall** translates intent into vendor-specific plans.
3. **Oathgate** orchestrates phased rollout.
4. **Stormlight / Highstorm** execute and monitor changes.
5. **Seekers / Squire** watch for configuration drift and report health.

## Monorepo Boundaries

- `cmd/` contains executable entrypoints only.
- `internal/` contains private implementation packages by domain.
- `pkg/` contains reusable public packages.

## Next Steps

- Define canonical intent schema and validation.
- Add adapter interfaces for network vendors.
- Implement safe, rollback-aware phased deployment engine.
