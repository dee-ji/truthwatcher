CREATE TABLE IF NOT EXISTS assets (
    id uuid PRIMARY KEY,
    asset_type text NOT NULL,
    identity_key text NOT NULL,
    vendor text,
    model text,
    serial text,
    system_mac text,
    metadata jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS assets_identity_key_unique_idx
    ON assets (identity_key)
    WHERE identity_key <> '';

CREATE INDEX IF NOT EXISTS assets_asset_type_idx ON assets (asset_type);
CREATE INDEX IF NOT EXISTS assets_vendor_idx ON assets (vendor);
CREATE INDEX IF NOT EXISTS assets_serial_idx ON assets (serial) WHERE serial IS NOT NULL AND serial <> '';
CREATE INDEX IF NOT EXISTS assets_system_mac_idx ON assets (system_mac) WHERE system_mac IS NOT NULL AND system_mac <> '';

CREATE TABLE IF NOT EXISTS facts (
    id uuid PRIMARY KEY,
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    name text NOT NULL,
    value jsonb NOT NULL,
    source text NOT NULL,
    confidence numeric NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    evidence_id uuid REFERENCES evidence(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS facts_asset_id_idx ON facts (asset_id);
CREATE INDEX IF NOT EXISTS facts_name_idx ON facts (name);
CREATE INDEX IF NOT EXISTS facts_evidence_id_idx ON facts (evidence_id);

CREATE TABLE IF NOT EXISTS relationships (
    id uuid PRIMARY KEY,
    source_asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    target_asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    relationship_type text NOT NULL,
    confidence numeric NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    evidence_id uuid REFERENCES evidence(id),
    metadata jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS relationships_source_asset_id_idx ON relationships (source_asset_id);
CREATE INDEX IF NOT EXISTS relationships_target_asset_id_idx ON relationships (target_asset_id);
CREATE INDEX IF NOT EXISTS relationships_relationship_type_idx ON relationships (relationship_type);
CREATE INDEX IF NOT EXISTS relationships_evidence_id_idx ON relationships (evidence_id);
