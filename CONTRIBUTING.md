# CONTRIBUTING

Thanks for contributing to Truthwatcher.

## Development workflow
1. Create focused, small changes.
2. Keep architectural vocabulary consistent (intent, revision, artifact, topology, deployment, state, reconcile, drift, audit, driver).
3. Run:
   - `make build`
   - `make test`
   - optional: `make lint`
4. Update docs/examples/OpenAPI when behavior or API contracts change.

## Principles
- Prefer honest placeholders and explicit TODOs over speculative implementation.
- Keep package boundaries domain-oriented and avoid circular dependencies.
- Do not introduce hidden insecure defaults, especially in authn/authz or deployment safety paths.

## Pull request checklist
- [ ] Scope is clear and focused.
- [ ] Tests or checks updated for behavior changes.
- [ ] OpenAPI + example payloads match implemented endpoints.
- [ ] Docs/README updates included for contributor-facing changes.
- [ ] TODO markers added where functionality remains intentionally stubbed.

## Security-sensitive changes
For authn, authz, request middleware, identity mapping, and deployment safety:
- Add tests for allow and deny flows.
- Document permission/role changes and migration impact.
- Keep dev-only bypass behavior explicit in docs and logs.
