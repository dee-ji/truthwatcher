DROP INDEX IF EXISTS relationships_state_idx;
DROP INDEX IF EXISTS facts_state_idx;
DROP INDEX IF EXISTS assets_state_idx;

ALTER TABLE relationships
    DROP COLUMN IF EXISTS state,
    DROP COLUMN IF EXISTS confidence_reason;

ALTER TABLE facts
    DROP COLUMN IF EXISTS state,
    DROP COLUMN IF EXISTS confidence_reason;

ALTER TABLE assets
    DROP COLUMN IF EXISTS state,
    DROP COLUMN IF EXISTS confidence_reason,
    DROP COLUMN IF EXISTS confidence;
