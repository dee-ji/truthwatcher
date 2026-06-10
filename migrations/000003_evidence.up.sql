CREATE TABLE IF NOT EXISTS evidence (
    id uuid PRIMARY KEY,
    discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
    target text NOT NULL,
    method text NOT NULL,
    command_or_api text NOT NULL,
    raw_output text NOT NULL,
    raw_output_hash text NOT NULL,
    parser_name text,
    collected_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS evidence_discovery_run_id_idx ON evidence (discovery_run_id);
CREATE INDEX IF NOT EXISTS evidence_raw_output_hash_idx ON evidence (raw_output_hash);
CREATE INDEX IF NOT EXISTS evidence_collected_at_idx ON evidence (collected_at DESC);
