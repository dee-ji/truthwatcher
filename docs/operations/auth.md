# Authentication and RBAC Operations Guide

This document describes the initial authentication and authorization scaffold for Truthwatcher.

## Goals of the current implementation

- Provide explicit authentication context injection for every HTTP request.
- Enforce RBAC checks on mutating endpoints by default.
- Keep read-only endpoints available with less strict checks while the platform matures.
- Support local development without requiring a live OIDC provider.

## Current auth modes

Auth behavior is driven by `internal/authn.Config`:

- `mode: disabled` disables bearer token parsing.
- `mode: jwt` enables bearer token parsing from `Authorization: Bearer <jwt>`.
- `local_dev_bypass: true` injects a synthetic identity and roles for local development.

> ⚠️ Local bypass logs an explicit warning on every request and must never be used in production.

## JWT parsing scope

Current JWT parsing intentionally validates structure and claim extraction only.

- Claims parsed: `sub`, `email`, `roles`, `permissions`, `iss`, `aud`.
- Signature verification is **not yet implemented**.
- This is a scaffold to keep integration with a real OIDC/JWKS verifier clean in a follow-up.

## RBAC model

The API enforces permissions through a simple evaluator interface:

- identity roles are resolved to permissions via a role catalog
- optional direct permissions in claims are merged in
- endpoint handlers call explicit permission checks (`requirePermission`)

### Seeded roles

- `admin`: full control
- `operator`: mutate intent/deploy/reconcile, read topology
- `viewer`: read-only visibility

### Permission examples

- `intent:write` required for `POST /api/v1/intents`
- `deployment:write` required for `POST /api/v1/deployments`
- `topology:write` required for `POST /api/v1/topology/import`
- `reconcile:write` required for `POST /api/v1/reconcile/runs`

## Local development

`cmd/tw-server` defaults to local bypass enabled for safety of developer workflow.

Environment knobs:

- `AUTH_LOCAL_DEV_BYPASS` (`true`/`false`)
- `AUTH_BYPASS_SUBJECT` (default `local-dev`)
- `AUTH_BYPASS_ROLE` (default `admin`)

When bypass is disabled, send a bearer JWT-like token with JSON claims payload.

## OIDC integration path (next step)

Planned clean extension points:

1. Replace `authn.Parser` with a verifier-backed parser (JWKS + signature verification).
2. Add issuer/audience/expiry enforcement.
3. Optionally bind claims to database `users` and `role_bindings`.
4. Incrementally tighten read-only endpoint checks.
