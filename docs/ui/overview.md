# UI Overview

The `ui/` application is the first production-facing UI foundation for Truthwatcher.

## Stack
- React + TypeScript
- Vite build tooling
- Lightweight local state (hooks + context)

## Design goals
- Keep the UI contributor-friendly and explicit.
- Avoid over-designed visuals while baseline intent/deployment/topology/drift visibility is built out.
- Wire pages to real Spanreed endpoints where they exist.
- Use clear placeholder states where endpoints are intentionally missing.

## Current page coverage
- Login placeholder / auth status: `/auth`
- Dashboard landing page: `/`
- Intents list: `/intents`
- Intent detail: `/intents/:intentId`
- Topology devices: `/topology/devices`
- Deployments list placeholder: `/deployments`
- Deployment detail: `/deployments/:deploymentId`
- Drift findings: `/drift/findings`
- System health: `/system/health`

## API integration notes
- API base URL is configured via `VITE_API_BASE_URL`.
- Typed API client is in `ui/src/api/client.ts` and `ui/src/api/types.ts`.
- Loading/empty/error handling is standardized by `AsyncState`.
- `TODO(truthwatcher):` markers are used for missing backend surfaces (for example deployment list).

## Future expansion points
- RBAC-aware rendering (permissions from auth claims).
- Intent create/edit flows.
- Deployment run and execution visibility.
- Reconcile run flows and audit timeline views.
