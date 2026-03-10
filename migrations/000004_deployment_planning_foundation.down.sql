DROP INDEX IF EXISTS idx_deployment_plans_idempotency_key;

ALTER TABLE deployment_targets
  DROP COLUMN IF EXISTS validation,
  DROP COLUMN IF EXISTS artifact_ref;

ALTER TABLE deployment_runs
  DROP COLUMN IF EXISTS summary,
  DROP COLUMN IF EXISTS simulation;

ALTER TABLE deployment_plans
  DROP COLUMN IF EXISTS target_refs,
  DROP COLUMN IF EXISTS artifact_refs,
  DROP COLUMN IF EXISTS rollback_plan,
  DROP COLUMN IF EXISTS stop_conditions,
  DROP COLUMN IF EXISTS rollout_plan,
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS mode;
