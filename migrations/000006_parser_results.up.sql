CREATE TABLE IF NOT EXISTS parser_results (
    id uuid PRIMARY KEY,
    discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
    evidence_id uuid NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    parser_name text NOT NULL,
    status text NOT NULL CHECK (status IN ('parsed', 'skipped', 'failed')),
    warnings jsonb NOT NULL DEFAULT '[]',
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS parser_results_discovery_run_id_idx
    ON parser_results (discovery_run_id);

CREATE INDEX IF NOT EXISTS parser_results_evidence_id_idx
    ON parser_results (evidence_id);
