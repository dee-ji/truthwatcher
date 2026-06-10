CREATE TABLE IF NOT EXISTS discovery_runs (
    id uuid PRIMARY KEY,
    status text NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'canceled')),
    seed_input jsonb NOT NULL DEFAULT '{}',
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS discovery_runs_status_idx ON discovery_runs (status);
CREATE INDEX IF NOT EXISTS discovery_runs_created_at_idx ON discovery_runs (created_at DESC);
