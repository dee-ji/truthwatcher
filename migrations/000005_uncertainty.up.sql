ALTER TABLE assets
    ADD COLUMN confidence numeric NOT NULL DEFAULT 0.5 CHECK (confidence >= 0 AND confidence <= 1),
    ADD COLUMN confidence_reason text NOT NULL DEFAULT 'deterministically inferred without direct evidence',
    ADD COLUMN state text NOT NULL DEFAULT 'inferred' CHECK (state IN ('observed', 'inferred', 'user_seeded', 'conflicting', 'unknown'));

ALTER TABLE facts
    ADD COLUMN confidence_reason text NOT NULL DEFAULT 'directly observed from evidence',
    ADD COLUMN state text NOT NULL DEFAULT 'observed' CHECK (state IN ('observed', 'inferred', 'user_seeded', 'conflicting', 'unknown'));

ALTER TABLE relationships
    ADD COLUMN confidence_reason text NOT NULL DEFAULT 'directly observed from evidence',
    ADD COLUMN state text NOT NULL DEFAULT 'observed' CHECK (state IN ('observed', 'inferred', 'user_seeded', 'conflicting', 'unknown'));

CREATE INDEX IF NOT EXISTS assets_state_idx ON assets (state);
CREATE INDEX IF NOT EXISTS facts_state_idx ON facts (state);
CREATE INDEX IF NOT EXISTS relationships_state_idx ON relationships (state);
