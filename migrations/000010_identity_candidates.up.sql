CREATE TABLE IF NOT EXISTS identity_candidates (
    id uuid PRIMARY KEY,
    discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
    evidence_id uuid NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    parser_name text NOT NULL,
    asset_type text NOT NULL,
    candidate_identity_key text NOT NULL,
    strength text NOT NULL CHECK (strength IN ('strong', 'provisional', 'weak')),
    confidence numeric NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    reason text NOT NULL,
    vendor text,
    model text,
    serial text,
    system_mac text,
    hostname text,
    proposed_asset_id uuid REFERENCES assets(id) ON DELETE SET NULL,
    review_state text NOT NULL DEFAULT 'pending'
        CHECK (review_state IN ('pending', 'auto_accepted', 'accepted', 'rejected', 'superseded')),
    metadata jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS identity_candidates_evidence_parser_key_unique_idx
    ON identity_candidates (evidence_id, parser_name, candidate_identity_key);

CREATE INDEX IF NOT EXISTS identity_candidates_review_state_created_at_idx
    ON identity_candidates (review_state, created_at DESC);

CREATE INDEX IF NOT EXISTS identity_candidates_candidate_identity_key_idx
    ON identity_candidates (candidate_identity_key);

CREATE INDEX IF NOT EXISTS identity_candidates_discovery_run_id_idx
    ON identity_candidates (discovery_run_id);
