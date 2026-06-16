package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/parser"
)

type IdentityCandidateRepository struct {
	db *sql.DB
}

func NewIdentityCandidateRepository(conn *sql.DB) IdentityCandidateRepository {
	return IdentityCandidateRepository{db: conn}
}

func (r IdentityCandidateRepository) CreateIdentityCandidate(ctx context.Context, params parser.CreateIdentityCandidateParams) (parser.IdentityCandidate, error) {
	if r.db == nil {
		return parser.IdentityCandidate{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return parser.IdentityCandidate{}, err
	}
	if len(params.Metadata) == 0 {
		params.Metadata = json.RawMessage(`{}`)
	}

	item, err := scanIdentityCandidate(r.db.QueryRowContext(ctx, `
INSERT INTO identity_candidates (
    id,
    discovery_run_id,
    evidence_id,
    parser_name,
    asset_type,
    candidate_identity_key,
    strength,
    confidence,
    reason,
    vendor,
    model,
    serial,
    system_mac,
    hostname,
    proposed_asset_id,
    review_state,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
ON CONFLICT (evidence_id, parser_name, candidate_identity_key) DO NOTHING
RETURNING id, discovery_run_id, evidence_id, parser_name, asset_type, candidate_identity_key, strength, confidence, reason, vendor, model, serial, system_mac, hostname, proposed_asset_id, review_state, metadata, created_at
`,
		id,
		params.DiscoveryRunID,
		params.EvidenceID,
		params.ParserName,
		params.AssetType,
		params.CandidateIdentityKey,
		params.Strength,
		params.Confidence,
		params.Reason,
		params.Vendor,
		params.Model,
		params.Serial,
		params.SystemMAC,
		params.Hostname,
		params.ProposedAssetID,
		params.ReviewState,
		params.Metadata,
	))
	if err == nil {
		return item, nil
	}
	if err != sql.ErrNoRows {
		return parser.IdentityCandidate{}, fmt.Errorf("create identity candidate: %w", err)
	}

	item, err = scanIdentityCandidate(r.db.QueryRowContext(ctx, `
SELECT id, discovery_run_id, evidence_id, parser_name, asset_type, candidate_identity_key, strength, confidence, reason, vendor, model, serial, system_mac, hostname, proposed_asset_id, review_state, metadata, created_at
FROM identity_candidates
WHERE evidence_id = $1 AND parser_name = $2 AND candidate_identity_key = $3
`,
		params.EvidenceID,
		params.ParserName,
		params.CandidateIdentityKey,
	))
	if err != nil {
		return parser.IdentityCandidate{}, fmt.Errorf("get existing identity candidate: %w", err)
	}
	return item, nil
}

func (r IdentityCandidateRepository) ListIdentityCandidates(ctx context.Context, filters parser.IdentityCandidateFilters) ([]parser.IdentityCandidate, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	query := `
SELECT id, discovery_run_id, evidence_id, parser_name, asset_type, candidate_identity_key, strength, confidence, reason, vendor, model, serial, system_mac, hostname, proposed_asset_id, review_state, metadata, created_at
FROM identity_candidates
`
	var conditions []string
	var args []any
	addFilter := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if strings.TrimSpace(filters.DiscoveryRunID) != "" {
		addFilter("discovery_run_id = $%d", filters.DiscoveryRunID)
	}
	if strings.TrimSpace(filters.EvidenceID) != "" {
		addFilter("evidence_id = $%d", filters.EvidenceID)
	}
	if filters.ReviewState != "" {
		addFilter("review_state = $%d", filters.ReviewState)
	}
	if filters.Strength != "" {
		addFilter("strength = $%d", filters.Strength)
	}
	if strings.TrimSpace(filters.CandidateIdentityKey) != "" {
		addFilter("candidate_identity_key = $%d", filters.CandidateIdentityKey)
	}
	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + "\n"
	}
	query += "ORDER BY created_at DESC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list identity candidates: %w", err)
	}
	defer rows.Close()

	var results []parser.IdentityCandidate
	for rows.Next() {
		item, err := scanIdentityCandidate(rows)
		if err != nil {
			return nil, fmt.Errorf("scan identity candidate: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read identity candidates: %w", err)
	}
	return results, nil
}

func (r IdentityCandidateRepository) ListIdentityReviewHandoffEntries(ctx context.Context, filters parser.IdentityReviewHandoffFilters) ([]parser.IdentityReviewHandoffEntry, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	query := `
SELECT
    c.id, c.discovery_run_id, c.evidence_id, c.parser_name, c.asset_type, c.candidate_identity_key, c.strength, c.confidence, c.reason, c.vendor, c.model, c.serial, c.system_mac, c.hostname, c.proposed_asset_id, c.review_state, c.metadata, c.created_at,
    r.id, r.identity_candidate_id, r.discovery_run_id, r.evidence_id, r.reviewer, r.action, r.previous_review_state, r.resulting_review_state, r.rationale, r.effect, r.metadata, r.created_at,
    e.id IS NOT NULL
FROM identity_candidates c
LEFT JOIN LATERAL (
    SELECT id, identity_candidate_id, discovery_run_id, evidence_id, reviewer, action, previous_review_state, resulting_review_state, rationale, effect, metadata, created_at
    FROM identity_candidate_reviews
    WHERE identity_candidate_id = c.id
    ORDER BY created_at DESC, id DESC
    LIMIT 1
) r ON true
LEFT JOIN evidence e ON e.id = c.evidence_id
`
	var conditions []string
	var args []any
	addFilter := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if strings.TrimSpace(filters.DiscoveryRunID) != "" {
		addFilter("c.discovery_run_id = $%d", filters.DiscoveryRunID)
	}
	if strings.TrimSpace(filters.EvidenceID) != "" {
		addFilter("c.evidence_id = $%d", filters.EvidenceID)
	}
	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + "\n"
	}
	query += "ORDER BY c.created_at DESC, c.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list identity review handoff entries: %w", err)
	}
	defer rows.Close()

	var results []parser.IdentityReviewHandoffEntry
	for rows.Next() {
		entry, err := scanIdentityReviewHandoffEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("scan identity review handoff entry: %w", err)
		}
		results = append(results, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read identity review handoff entries: %w", err)
	}
	return results, nil
}

func (r IdentityCandidateRepository) ListOrphanedIdentityCandidateReviews(ctx context.Context) ([]parser.IdentityCandidateReview, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT r.id, r.identity_candidate_id, r.discovery_run_id, r.evidence_id, r.reviewer, r.action, r.previous_review_state, r.resulting_review_state, r.rationale, r.effect, r.metadata, r.created_at
FROM identity_candidate_reviews r
LEFT JOIN identity_candidates c ON c.id = r.identity_candidate_id
WHERE c.id IS NULL
ORDER BY r.created_at DESC, r.id DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list orphaned identity candidate reviews: %w", err)
	}
	defer rows.Close()

	var results []parser.IdentityCandidateReview
	for rows.Next() {
		item, err := scanIdentityCandidateReview(rows)
		if err != nil {
			return nil, fmt.Errorf("scan orphaned identity candidate review: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read orphaned identity candidate reviews: %w", err)
	}
	return results, nil
}

func (r IdentityCandidateRepository) GetIdentityCandidate(ctx context.Context, id string) (parser.IdentityCandidate, error) {
	if r.db == nil {
		return parser.IdentityCandidate{}, fmt.Errorf("database is required")
	}
	item, err := scanIdentityCandidate(r.db.QueryRowContext(ctx, `
SELECT id, discovery_run_id, evidence_id, parser_name, asset_type, candidate_identity_key, strength, confidence, reason, vendor, model, serial, system_mac, hostname, proposed_asset_id, review_state, metadata, created_at
FROM identity_candidates
WHERE id = $1
`, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return parser.IdentityCandidate{}, assets.ErrNotFound
	}
	if err != nil {
		return parser.IdentityCandidate{}, fmt.Errorf("get identity candidate: %w", err)
	}
	return item, nil
}

func (r IdentityCandidateRepository) ReviewIdentityCandidate(ctx context.Context, params parser.ReviewIdentityCandidateParams) (parser.IdentityCandidateReview, error) {
	return r.recordIdentityCandidateReview(ctx, recordIdentityCandidateReviewParams{
		IdentityCandidateID: params.IdentityCandidateID,
		Reviewer:            params.Reviewer,
		Action:              params.Action,
		Rationale:           params.Rationale,
		Metadata:            params.Metadata,
		OnlyFromPending:     false,
	})
}

func (r IdentityCandidateRepository) AutoAcceptIdentityCandidate(ctx context.Context, params parser.AutoAcceptIdentityCandidateParams) error {
	_, err := r.recordIdentityCandidateReview(ctx, recordIdentityCandidateReviewParams{
		IdentityCandidateID: params.IdentityCandidateID,
		Reviewer:            "parser:auto_acceptance",
		Action:              parser.IdentityReviewActionAutoAccept,
		Rationale:           params.Rationale,
		Metadata:            params.Metadata,
		OnlyFromPending:     true,
	})
	return err
}

type recordIdentityCandidateReviewParams struct {
	IdentityCandidateID string
	Reviewer            string
	Action              parser.IdentityReviewAction
	Rationale           string
	Metadata            json.RawMessage
	OnlyFromPending     bool
}

func (r IdentityCandidateRepository) recordIdentityCandidateReview(ctx context.Context, params recordIdentityCandidateReviewParams) (parser.IdentityCandidateReview, error) {
	if r.db == nil {
		return parser.IdentityCandidateReview{}, fmt.Errorf("database is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return parser.IdentityCandidateReview{}, fmt.Errorf("begin identity candidate review: %w", err)
	}
	defer tx.Rollback()

	candidate, err := scanIdentityCandidate(tx.QueryRowContext(ctx, `
SELECT id, discovery_run_id, evidence_id, parser_name, asset_type, candidate_identity_key, strength, confidence, reason, vendor, model, serial, system_mac, hostname, proposed_asset_id, review_state, metadata, created_at
FROM identity_candidates
WHERE id = $1
FOR UPDATE
`, params.IdentityCandidateID))
	if errors.Is(err, sql.ErrNoRows) {
		return parser.IdentityCandidateReview{}, assets.ErrNotFound
	}
	if err != nil {
		return parser.IdentityCandidateReview{}, fmt.Errorf("get identity candidate for review: %w", err)
	}
	if params.OnlyFromPending && candidate.ReviewState != parser.IdentityReviewPending {
		return parser.IdentityCandidateReview{}, nil
	}

	resultingState := parser.ResultingReviewState(params.Action)
	effect := parser.IdentityReviewEffect(params.Action)
	id, err := discovery.NewID()
	if err != nil {
		return parser.IdentityCandidateReview{}, err
	}
	if len(params.Metadata) == 0 {
		params.Metadata = json.RawMessage(`{}`)
	}

	review, err := scanIdentityCandidateReview(tx.QueryRowContext(ctx, `
INSERT INTO identity_candidate_reviews (
    id,
    identity_candidate_id,
    discovery_run_id,
    evidence_id,
    reviewer,
    action,
    previous_review_state,
    resulting_review_state,
    rationale,
    effect,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, identity_candidate_id, discovery_run_id, evidence_id, reviewer, action, previous_review_state, resulting_review_state, rationale, effect, metadata, created_at
`,
		id,
		candidate.ID,
		candidate.DiscoveryRunID,
		candidate.EvidenceID,
		params.Reviewer,
		params.Action,
		candidate.ReviewState,
		resultingState,
		params.Rationale,
		effect,
		params.Metadata,
	))
	if err != nil {
		return parser.IdentityCandidateReview{}, fmt.Errorf("create identity candidate review: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE identity_candidates
SET review_state = $1
WHERE id = $2
`, resultingState, candidate.ID); err != nil {
		return parser.IdentityCandidateReview{}, fmt.Errorf("update identity candidate review state: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return parser.IdentityCandidateReview{}, fmt.Errorf("commit identity candidate review: %w", err)
	}
	return review, nil
}

func scanIdentityCandidate(s scanner) (parser.IdentityCandidate, error) {
	var item parser.IdentityCandidate
	var vendor sql.NullString
	var model sql.NullString
	var serial sql.NullString
	var systemMAC sql.NullString
	var hostname sql.NullString
	var proposedAssetID sql.NullString

	if err := s.Scan(
		&item.ID,
		&item.DiscoveryRunID,
		&item.EvidenceID,
		&item.ParserName,
		&item.AssetType,
		&item.CandidateIdentityKey,
		&item.Strength,
		&item.Confidence,
		&item.Reason,
		&vendor,
		&model,
		&serial,
		&systemMAC,
		&hostname,
		&proposedAssetID,
		&item.ReviewState,
		&item.Metadata,
		&item.CreatedAt,
	); err != nil {
		return parser.IdentityCandidate{}, err
	}
	item.Vendor = nullableStringPtr(vendor)
	item.Model = nullableStringPtr(model)
	item.Serial = nullableStringPtr(serial)
	item.SystemMAC = nullableStringPtr(systemMAC)
	item.Hostname = nullableStringPtr(hostname)
	item.ProposedAssetID = nullableStringPtr(proposedAssetID)
	return item, nil
}

func scanIdentityCandidateReview(s scanner) (parser.IdentityCandidateReview, error) {
	var item parser.IdentityCandidateReview
	if err := s.Scan(
		&item.ID,
		&item.IdentityCandidateID,
		&item.DiscoveryRunID,
		&item.EvidenceID,
		&item.Reviewer,
		&item.Action,
		&item.PreviousReviewState,
		&item.ResultingReviewState,
		&item.Rationale,
		&item.Effect,
		&item.Metadata,
		&item.CreatedAt,
	); err != nil {
		return parser.IdentityCandidateReview{}, err
	}
	return item, nil
}

func scanIdentityReviewHandoffEntry(s scanner) (parser.IdentityReviewHandoffEntry, error) {
	var candidate parser.IdentityCandidate
	var vendor sql.NullString
	var model sql.NullString
	var serial sql.NullString
	var systemMAC sql.NullString
	var hostname sql.NullString
	var proposedAssetID sql.NullString
	var reviewID sql.NullString
	var reviewCandidateID sql.NullString
	var reviewDiscoveryRunID sql.NullString
	var reviewEvidenceID sql.NullString
	var reviewer sql.NullString
	var action sql.NullString
	var previousState sql.NullString
	var resultingState sql.NullString
	var rationale sql.NullString
	var effect sql.NullString
	var reviewMetadata []byte
	var reviewCreatedAt sql.NullTime
	var evidencePresent bool

	if err := s.Scan(
		&candidate.ID,
		&candidate.DiscoveryRunID,
		&candidate.EvidenceID,
		&candidate.ParserName,
		&candidate.AssetType,
		&candidate.CandidateIdentityKey,
		&candidate.Strength,
		&candidate.Confidence,
		&candidate.Reason,
		&vendor,
		&model,
		&serial,
		&systemMAC,
		&hostname,
		&proposedAssetID,
		&candidate.ReviewState,
		&candidate.Metadata,
		&candidate.CreatedAt,
		&reviewID,
		&reviewCandidateID,
		&reviewDiscoveryRunID,
		&reviewEvidenceID,
		&reviewer,
		&action,
		&previousState,
		&resultingState,
		&rationale,
		&effect,
		&reviewMetadata,
		&reviewCreatedAt,
		&evidencePresent,
	); err != nil {
		return parser.IdentityReviewHandoffEntry{}, err
	}
	candidate.Vendor = nullableStringPtr(vendor)
	candidate.Model = nullableStringPtr(model)
	candidate.Serial = nullableStringPtr(serial)
	candidate.SystemMAC = nullableStringPtr(systemMAC)
	candidate.Hostname = nullableStringPtr(hostname)
	candidate.ProposedAssetID = nullableStringPtr(proposedAssetID)

	entry := parser.IdentityReviewHandoffEntry{
		Candidate: candidate,
		EvidenceReference: parser.IdentityEvidenceRef{
			EvidenceID:     candidate.EvidenceID,
			DiscoveryRunID: candidate.DiscoveryRunID,
			Present:        evidencePresent,
		},
		ParserSource: parser.IdentityParserSource{
			ParserName: candidate.ParserName,
			Metadata:   candidate.Metadata,
		},
	}
	if reviewID.Valid {
		review := parser.IdentityCandidateReview{
			ID:                   reviewID.String,
			IdentityCandidateID:  reviewCandidateID.String,
			DiscoveryRunID:       reviewDiscoveryRunID.String,
			EvidenceID:           reviewEvidenceID.String,
			Reviewer:             reviewer.String,
			Action:               parser.IdentityReviewAction(action.String),
			PreviousReviewState:  parser.IdentityReviewState(previousState.String),
			ResultingReviewState: parser.IdentityReviewState(resultingState.String),
			Rationale:            rationale.String,
			Effect:               effect.String,
			Metadata:             json.RawMessage(reviewMetadata),
			CreatedAt:            reviewCreatedAt.Time,
		}
		entry.LatestReview = &review
	}
	return entry, nil
}
