CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  subject TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL DEFAULT '',
  display_name TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS roles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS permissions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS role_bindings (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  scope TEXT NOT NULL DEFAULT 'global',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, role_id, scope)
);

INSERT INTO roles(name, description) VALUES
  ('admin', 'Full control over intent, deployment, topology and reconciliation workflows'),
  ('operator', 'Can mutate intent and deployment workflows, read topology and reconciliation state'),
  ('viewer', 'Read-only access to intent, deployment, topology and audit views')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions(name, description) VALUES
  ('intent:read', 'Read intent objects'),
  ('intent:write', 'Create or modify intent objects'),
  ('deployment:read', 'Read deployment plans and runs'),
  ('deployment:write', 'Create deployment plans and trigger rollout actions'),
  ('topology:read', 'Read topology inventory and links'),
  ('topology:write', 'Import or modify topology data'),
  ('reconcile:read', 'Read reconciliation runs and drift findings'),
  ('reconcile:write', 'Create reconciliation runs'),
  ('audit:read', 'Read audit events')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
  'intent:read', 'intent:write', 'deployment:read', 'deployment:write',
  'topology:read', 'topology:write', 'reconcile:read', 'reconcile:write', 'audit:read'
)
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
  'intent:read', 'intent:write', 'deployment:read', 'deployment:write',
  'topology:read', 'reconcile:read', 'reconcile:write'
)
WHERE r.name = 'operator'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
  'intent:read', 'deployment:read', 'topology:read', 'reconcile:read', 'audit:read'
)
WHERE r.name = 'viewer'
ON CONFLICT DO NOTHING;
