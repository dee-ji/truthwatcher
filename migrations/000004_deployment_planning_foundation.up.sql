ALTER TABLE deployment_plans
  ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'dry-run',
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'planned',
  ADD COLUMN IF NOT EXISTS rollout_plan JSONB NOT NULL DEFAULT '{"waves":[]}'::jsonb,
  ADD COLUMN IF NOT EXISTS stop_conditions JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS rollback_plan JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS artifact_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS target_refs JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE deployment_runs
  ADD COLUMN IF NOT EXISTS simulation BOOLEAN NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS summary JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE deployment_targets
  ADD COLUMN IF NOT EXISTS artifact_ref TEXT,
  ADD COLUMN IF NOT EXISTS validation JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE UNIQUE INDEX IF NOT EXISTS idx_deployment_plans_idempotency_key
  ON deployment_plans(idempotency_key)
  WHERE idempotency_key IS NOT NULL;
