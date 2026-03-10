CREATE TABLE IF NOT EXISTS reconcile_runs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  intent_id UUID REFERENCES intent_sets(id),
  status TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  findings_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

ALTER TABLE config_snapshots
  ADD COLUMN IF NOT EXISTS captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS artifact_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'offline-fixture';

ALTER TABLE operational_snapshots
  ADD COLUMN IF NOT EXISTS captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'offline-fixture';

ALTER TABLE drift_findings
  ADD COLUMN IF NOT EXISTS reconcile_run_id UUID REFERENCES reconcile_runs(id),
  ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'config_mismatch',
  ADD COLUMN IF NOT EXISTS summary TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS intended_artifact TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS actual_snapshot_id UUID REFERENCES config_snapshots(id),
  ADD COLUMN IF NOT EXISTS remediation_plan JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_config_snapshots_device_captured_at ON config_snapshots(device_id, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_drift_findings_reconcile_run ON drift_findings(reconcile_run_id);
