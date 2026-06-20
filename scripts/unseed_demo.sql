-- Clear Truthwatcher application data while preserving schema_migrations.
BEGIN;
TRUNCATE TABLE
    identity_aliases,
    identity_candidate_reviews,
    identity_candidates,
    parser_results,
    audit_records,
    relationships,
    facts,
    assets,
    evidence,
    discovery_runs,
    devices
RESTART IDENTITY CASCADE;
COMMIT;

SELECT 'truthwatcher app data cleared' AS message;
