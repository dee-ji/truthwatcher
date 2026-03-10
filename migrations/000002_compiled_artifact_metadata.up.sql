ALTER TABLE compiled_artifacts
  ADD COLUMN artifact_format TEXT NOT NULL DEFAULT 'text',
  ADD COLUMN artifact_metadata JSONB NOT NULL DEFAULT '{}'::jsonb;
