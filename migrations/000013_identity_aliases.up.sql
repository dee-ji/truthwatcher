CREATE TABLE IF NOT EXISTS identity_aliases (
    id uuid PRIMARY KEY,
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    identity_candidate_id uuid NOT NULL REFERENCES identity_candidates(id) ON DELETE CASCADE,
    alias_identity_key text NOT NULL,
    alias_strength text NOT NULL CHECK (alias_strength IN ('strong', 'provisional', 'weak')),
    evidence_id uuid NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
    reviewer text NOT NULL,
    rationale text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (asset_id, alias_identity_key),
    UNIQUE (identity_candidate_id)
);

CREATE INDEX IF NOT EXISTS identity_aliases_asset_id_idx
    ON identity_aliases (asset_id, created_at DESC);

CREATE INDEX IF NOT EXISTS identity_aliases_evidence_id_idx
    ON identity_aliases (evidence_id);
