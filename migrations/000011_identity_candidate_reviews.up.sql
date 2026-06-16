ALTER TABLE identity_candidates
    DROP CONSTRAINT IF EXISTS identity_candidates_review_state_check,
    ADD CONSTRAINT identity_candidates_review_state_check
        CHECK (review_state IN ('pending', 'auto_accepted', 'accepted', 'rejected', 'superseded', 'deferred', 'more_evidence_requested'));

CREATE TABLE IF NOT EXISTS identity_candidate_reviews (
    id uuid PRIMARY KEY,
    identity_candidate_id uuid NOT NULL REFERENCES identity_candidates(id) ON DELETE CASCADE,
    discovery_run_id uuid NOT NULL REFERENCES discovery_runs(id) ON DELETE CASCADE,
    evidence_id uuid NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    reviewer text NOT NULL,
    action text NOT NULL CHECK (action IN ('accept', 'reject', 'defer', 'request_more_evidence')),
    previous_review_state text NOT NULL,
    resulting_review_state text NOT NULL CHECK (resulting_review_state IN ('accepted', 'rejected', 'deferred', 'more_evidence_requested')),
    rationale text NOT NULL,
    effect text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS identity_candidate_reviews_candidate_id_idx
    ON identity_candidate_reviews (identity_candidate_id, created_at DESC);

CREATE INDEX IF NOT EXISTS identity_candidate_reviews_discovery_run_id_idx
    ON identity_candidate_reviews (discovery_run_id);

CREATE INDEX IF NOT EXISTS identity_candidate_reviews_evidence_id_idx
    ON identity_candidate_reviews (evidence_id);
