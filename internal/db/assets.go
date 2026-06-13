package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
)

type AssetRepository struct {
	db *sql.DB
}

func NewAssetRepository(conn *sql.DB) AssetRepository {
	return AssetRepository{db: conn}
}

func (r AssetRepository) CreateAsset(ctx context.Context, params assets.CreateAssetParams) (assets.Asset, error) {
	if r.db == nil {
		return assets.Asset{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return assets.Asset{}, err
	}

	result, err := scanAsset(r.db.QueryRowContext(ctx, `
INSERT INTO assets (id, asset_type, identity_key, vendor, model, serial, system_mac, confidence, confidence_reason, state, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, asset_type, identity_key, vendor, model, serial, system_mac, confidence, confidence_reason, state, metadata, created_at, updated_at
`,
		id,
		params.Type,
		params.IdentityKey,
		params.Vendor,
		params.Model,
		params.Serial,
		params.SystemMAC,
		params.Confidence,
		params.ConfidenceReason,
		params.State,
		params.Metadata,
	))
	if err != nil {
		return assets.Asset{}, fmt.Errorf("create asset: %w", err)
	}

	return result, nil
}

func (r AssetRepository) GetAsset(ctx context.Context, id string) (assets.Asset, error) {
	if r.db == nil {
		return assets.Asset{}, fmt.Errorf("database is required")
	}

	result, err := scanAsset(r.db.QueryRowContext(ctx, `
SELECT id, asset_type, identity_key, vendor, model, serial, system_mac, confidence, confidence_reason, state, metadata, created_at, updated_at
FROM assets
WHERE id = $1
`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return assets.Asset{}, assets.ErrNotFound
	}
	if err != nil {
		return assets.Asset{}, fmt.Errorf("get asset: %w", err)
	}

	return result, nil
}

func (r AssetRepository) ListAssets(ctx context.Context) ([]assets.Asset, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, asset_type, identity_key, vendor, model, serial, system_mac, confidence, confidence_reason, state, metadata, created_at, updated_at
FROM assets
ORDER BY created_at DESC, id DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var results []assets.Asset
	for rows.Next() {
		item, err := scanAsset(rows)
		if err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read assets: %w", err)
	}

	return results, nil
}

func (r AssetRepository) CreateFact(ctx context.Context, params assets.CreateFactParams) (assets.Fact, error) {
	if r.db == nil {
		return assets.Fact{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return assets.Fact{}, err
	}

	result, err := scanFact(r.db.QueryRowContext(ctx, `
INSERT INTO facts (id, asset_id, name, value, source, confidence, confidence_reason, state, evidence_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, asset_id, name, value, source, confidence, confidence_reason, state, evidence_id, created_at
`, id, params.AssetID, params.Name, params.Value, params.Source, params.Confidence, params.ConfidenceReason, params.State, params.EvidenceID))
	if err != nil {
		return assets.Fact{}, fmt.Errorf("create fact: %w", err)
	}

	return result, nil
}

func (r AssetRepository) GetFact(ctx context.Context, id string) (assets.Fact, error) {
	if r.db == nil {
		return assets.Fact{}, fmt.Errorf("database is required")
	}

	result, err := scanFact(r.db.QueryRowContext(ctx, `
SELECT id, asset_id, name, value, source, confidence, confidence_reason, state, evidence_id, created_at
FROM facts
WHERE id = $1
`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return assets.Fact{}, assets.ErrNotFound
	}
	if err != nil {
		return assets.Fact{}, fmt.Errorf("get fact: %w", err)
	}

	return result, nil
}

func (r AssetRepository) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, asset_id, name, value, source, confidence, confidence_reason, state, evidence_id, created_at
FROM facts
WHERE asset_id = $1
ORDER BY created_at DESC, id DESC
`, assetID)
	if err != nil {
		return nil, fmt.Errorf("list facts: %w", err)
	}
	defer rows.Close()

	var results []assets.Fact
	for rows.Next() {
		item, err := scanFact(rows)
		if err != nil {
			return nil, fmt.Errorf("scan fact: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read facts: %w", err)
	}

	return results, nil
}

func (r AssetRepository) ListFacts(ctx context.Context) ([]assets.Fact, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, asset_id, name, value, source, confidence, confidence_reason, state, evidence_id, created_at
FROM facts
ORDER BY created_at DESC, id DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list facts: %w", err)
	}
	defer rows.Close()

	var results []assets.Fact
	for rows.Next() {
		item, err := scanFact(rows)
		if err != nil {
			return nil, fmt.Errorf("scan fact: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read facts: %w", err)
	}

	return results, nil
}

func (r AssetRepository) CreateRelationship(ctx context.Context, params assets.CreateRelationshipParams) (assets.Relationship, error) {
	if r.db == nil {
		return assets.Relationship{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return assets.Relationship{}, err
	}

	result, err := scanRelationship(r.db.QueryRowContext(ctx, `
INSERT INTO relationships (id, source_asset_id, target_asset_id, relationship_type, confidence, confidence_reason, state, evidence_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, source_asset_id, target_asset_id, relationship_type, confidence, confidence_reason, state, evidence_id, metadata, created_at, updated_at
`, id, params.SourceAssetID, params.TargetAssetID, params.RelationshipType, params.Confidence, params.ConfidenceReason, params.State, params.EvidenceID, params.Metadata))
	if err != nil {
		return assets.Relationship{}, fmt.Errorf("create relationship: %w", err)
	}

	return result, nil
}

func (r AssetRepository) GetRelationship(ctx context.Context, id string) (assets.Relationship, error) {
	if r.db == nil {
		return assets.Relationship{}, fmt.Errorf("database is required")
	}

	result, err := scanRelationship(r.db.QueryRowContext(ctx, `
SELECT id, source_asset_id, target_asset_id, relationship_type, confidence, confidence_reason, state, evidence_id, metadata, created_at, updated_at
FROM relationships
WHERE id = $1
`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return assets.Relationship{}, assets.ErrNotFound
	}
	if err != nil {
		return assets.Relationship{}, fmt.Errorf("get relationship: %w", err)
	}

	return result, nil
}

func (r AssetRepository) ListRelationships(ctx context.Context) ([]assets.Relationship, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, source_asset_id, target_asset_id, relationship_type, confidence, confidence_reason, state, evidence_id, metadata, created_at, updated_at
FROM relationships
ORDER BY created_at DESC, id DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list relationships: %w", err)
	}
	defer rows.Close()

	var results []assets.Relationship
	for rows.Next() {
		item, err := scanRelationship(rows)
		if err != nil {
			return nil, fmt.Errorf("scan relationship: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read relationships: %w", err)
	}

	return results, nil
}

func scanAsset(s scanner) (assets.Asset, error) {
	var item assets.Asset
	var vendor sql.NullString
	var model sql.NullString
	var serial sql.NullString
	var systemMAC sql.NullString

	if err := s.Scan(
		&item.ID,
		&item.Type,
		&item.IdentityKey,
		&vendor,
		&model,
		&serial,
		&systemMAC,
		&item.Confidence,
		&item.ConfidenceReason,
		&item.State,
		&item.Metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return assets.Asset{}, err
	}

	item.Vendor = nullableStringPtr(vendor)
	item.Model = nullableStringPtr(model)
	item.Serial = nullableStringPtr(serial)
	item.SystemMAC = nullableStringPtr(systemMAC)
	return item, nil
}

func scanFact(s scanner) (assets.Fact, error) {
	var item assets.Fact
	var evidenceID sql.NullString

	if err := s.Scan(
		&item.ID,
		&item.AssetID,
		&item.Name,
		&item.Value,
		&item.Source,
		&item.Confidence,
		&item.ConfidenceReason,
		&item.State,
		&evidenceID,
		&item.CreatedAt,
	); err != nil {
		return assets.Fact{}, err
	}

	item.EvidenceID = nullableStringPtr(evidenceID)
	return item, nil
}

func scanRelationship(s scanner) (assets.Relationship, error) {
	var item assets.Relationship
	var evidenceID sql.NullString

	if err := s.Scan(
		&item.ID,
		&item.SourceAssetID,
		&item.TargetAssetID,
		&item.RelationshipType,
		&item.Confidence,
		&item.ConfidenceReason,
		&item.State,
		&evidenceID,
		&item.Metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return assets.Relationship{}, err
	}

	item.EvidenceID = nullableStringPtr(evidenceID)
	return item, nil
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
