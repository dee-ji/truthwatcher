# RBAC Policy Examples

These examples show intended policy shape for role assignment and endpoint permissions.

## Example 1: Global admin

```yaml
subject: user:alice@example.com
bindings:
  - role: admin
    scope: global
```

## Example 2: Operator

```yaml
subject: user:ops@example.com
bindings:
  - role: operator
    scope: global
```

## Example 3: Viewer

```yaml
subject: user:noc@example.com
bindings:
  - role: viewer
    scope: global
```

## Effective permission summary

| Role     | intent:write | deployment:write | topology:write | reconcile:write |
|----------|--------------|------------------|----------------|-----------------|
| admin    | ✅           | ✅               | ✅             | ✅              |
| operator | ✅           | ✅               | ❌             | ✅              |
| viewer   | ❌           | ❌               | ❌             | ❌              |
