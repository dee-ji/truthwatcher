package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

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
