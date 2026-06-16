DROP TABLE IF EXISTS identity_candidate_reviews;

ALTER TABLE identity_candidates
    DROP CONSTRAINT IF EXISTS identity_candidates_review_state_check,
    ADD CONSTRAINT identity_candidates_review_state_check
        CHECK (review_state IN ('pending', 'auto_accepted', 'accepted', 'rejected', 'superseded'));
