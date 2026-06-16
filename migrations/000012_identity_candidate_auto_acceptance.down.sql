DELETE FROM identity_candidate_reviews
WHERE action = 'auto_accept' OR resulting_review_state = 'auto_accepted';

ALTER TABLE identity_candidate_reviews
    DROP CONSTRAINT IF EXISTS identity_candidate_reviews_action_check,
    ADD CONSTRAINT identity_candidate_reviews_action_check
        CHECK (action IN ('accept', 'reject', 'defer', 'request_more_evidence'));

ALTER TABLE identity_candidate_reviews
    DROP CONSTRAINT IF EXISTS identity_candidate_reviews_resulting_review_state_check,
    ADD CONSTRAINT identity_candidate_reviews_resulting_review_state_check
        CHECK (resulting_review_state IN ('accepted', 'rejected', 'deferred', 'more_evidence_requested'));
