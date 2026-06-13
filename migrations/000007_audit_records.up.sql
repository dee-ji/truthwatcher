CREATE TABLE audit_records (
    id UUID PRIMARY KEY,
    action TEXT NOT NULL,
    initiator TEXT NOT NULL,
    request_id TEXT,
    discovery_run_id UUID REFERENCES discovery_runs(id) ON DELETE SET NULL,
    target TEXT NOT NULL,
    method TEXT NOT NULL,
    profile TEXT,
    task TEXT,
    command_or_api TEXT,
    status TEXT NOT NULL,
    evidence_id UUID REFERENCES evidence(id) ON DELETE SET NULL,
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL,
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_records_discovery_run_id ON audit_records(discovery_run_id);
CREATE INDEX idx_audit_records_started_at ON audit_records(started_at);
CREATE INDEX idx_audit_records_action ON audit_records(action);
