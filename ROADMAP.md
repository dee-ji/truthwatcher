# ROADMAP

This roadmap tracks the next practical phases for Truthwatcher after foundation hardening.

## Phase 1: Live execution adapters
- Implement Squire execution adapters for device/session operations.
- Wire Highstorm rollout steps to adapter-backed operations.
- Add rollback hooks and richer deployment run status transitions.

## Phase 2: Richer topology analysis
- Expand Shadesmar graph query capabilities (blast radius, path impact, dependency sets).
- Add topology import validation and schema evolution controls.
- Introduce topology change diffing between revisions.

## Phase 3: Real approval gates
- Move Oathgate from scaffold checks to policy-backed approval gates.
- Enforce pre-deploy simulation and stop-conditions before apply.
- Record gate outcomes in audit events and deployment summaries.

## Phase 4: OIDC integration
- Replace local-dev auth shortcuts with OIDC discovery/JWKS verification.
- Formalize role binding bootstrap and least-privilege defaults.
- Add denied/allowed authorization coverage for critical endpoints.

## Phase 5: Multi-vendor expansion
- Promote additional vendor drivers (EOS, IOS-XE, IOS-XR, SR-OS) to first-class support.
- Normalize artifact metadata and rendering test fixtures across vendors.
- Add driver capability reporting in API surfaces.

## Phase 6: UI
- Introduce a contributor-focused web UI for intent, topology, deployment, and drift workflows.
- Back UI flows with existing Spanreed endpoints before adding new API primitives.
- Keep all UI actions audit-visible and permission-gated.
