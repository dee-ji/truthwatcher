DROP INDEX IF EXISTS idx_drift_findings_reconcile_run;
DROP INDEX IF EXISTS idx_config_snapshots_device_captured_at;

ALTER TABLE drift_findings
  DROP COLUMN IF EXISTS remediation_plan,
  DROP COLUMN IF EXISTS actual_snapshot_id,
  DROP COLUMN IF EXISTS intended_artifact,
  DROP COLUMN IF EXISTS summary,
  DROP COLUMN IF EXISTS kind,
  DROP COLUMN IF EXISTS reconcile_run_id;

ALTER TABLE operational_snapshots
  DROP COLUMN IF EXISTS source,
  DROP COLUMN IF EXISTS captured_at;

ALTER TABLE config_snapshots
  DROP COLUMN IF EXISTS source,
  DROP COLUMN IF EXISTS artifact_ref,
  DROP COLUMN IF EXISTS captured_at;

DROP TABLE IF EXISTS reconcile_runs;
