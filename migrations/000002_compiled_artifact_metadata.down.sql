ALTER TABLE compiled_artifacts
  DROP COLUMN IF EXISTS artifact_metadata,
  DROP COLUMN IF EXISTS artifact_format;
