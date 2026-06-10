package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
)

type EvidenceRepository struct {
	db *sql.DB
}

func NewEvidenceRepository(conn *sql.DB) EvidenceRepository {
	return EvidenceRepository{db: conn}
}

func (r EvidenceRepository) CreateEvidence(ctx context.Context, params evidence.CreateEvidenceParams) (evidence.Evidence, error) {
	if r.db == nil {
		return evidence.Evidence{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return evidence.Evidence{}, err
	}
	if len(params.Metadata) == 0 {
		params.Metadata = json.RawMessage(`{}`)
	}

	rawOutputHash := evidence.HashRawOutput(params.RawOutput)

	result, err := scanEvidence(r.db.QueryRowContext(ctx, `
INSERT INTO evidence (
    id,
    discovery_run_id,
    target,
    method,
    command_or_api,
    raw_output,
    raw_output_hash,
    parser_name,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, discovery_run_id, target, method, command_or_api, raw_output, raw_output_hash, parser_name, collected_at, metadata
`,
		id,
		params.DiscoveryRunID,
		params.Target,
		params.Method,
		params.CommandOrAPI,
		params.RawOutput,
		rawOutputHash,
		params.ParserName,
		params.Metadata,
	))
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("create evidence: %w", err)
	}

	return result, nil
}

func (r EvidenceRepository) GetEvidence(ctx context.Context, id string) (evidence.Evidence, error) {
	if r.db == nil {
		return evidence.Evidence{}, fmt.Errorf("database is required")
	}

	result, err := scanEvidence(r.db.QueryRowContext(ctx, `
SELECT id, discovery_run_id, target, method, command_or_api, raw_output, raw_output_hash, parser_name, collected_at, metadata
FROM evidence
WHERE id = $1
`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return evidence.Evidence{}, evidence.ErrNotFound
	}
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("get evidence: %w", err)
	}

	return result, nil
}

func (r EvidenceRepository) ListEvidenceByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]evidence.Evidence, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, discovery_run_id, target, method, command_or_api, raw_output, raw_output_hash, parser_name, collected_at, metadata
FROM evidence
WHERE discovery_run_id = $1
ORDER BY collected_at ASC, id ASC
`, discoveryRunID)
	if err != nil {
		return nil, fmt.Errorf("list evidence: %w", err)
	}
	defer rows.Close()

	var results []evidence.Evidence
	for rows.Next() {
		item, err := scanEvidence(rows)
		if err != nil {
			return nil, fmt.Errorf("scan evidence: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read evidence: %w", err)
	}

	return results, nil
}

func scanEvidence(s scanner) (evidence.Evidence, error) {
	var item evidence.Evidence
	var parserName sql.NullString

	if err := s.Scan(
		&item.ID,
		&item.DiscoveryRunID,
		&item.Target,
		&item.Method,
		&item.CommandOrAPI,
		&item.RawOutput,
		&item.RawOutputHash,
		&parserName,
		&item.CollectedAt,
		&item.Metadata,
	); err != nil {
		return evidence.Evidence{}, err
	}

	if parserName.Valid {
		item.ParserName = &parserName.String
	}

	return item, nil
}
