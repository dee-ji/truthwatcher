package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"truthwatcher/internal/discovery"
	"truthwatcher/internal/parser"
)

type ParseResultRepository struct {
	db *sql.DB
}

func NewParseResultRepository(conn *sql.DB) ParseResultRepository {
	return ParseResultRepository{db: conn}
}

func (r ParseResultRepository) CreateParseResult(ctx context.Context, params parser.CreateParseResultParams) (parser.ParseRecord, error) {
	if r.db == nil {
		return parser.ParseRecord{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return parser.ParseRecord{}, err
	}
	warnings, err := json.Marshal(params.Warnings)
	if err != nil {
		return parser.ParseRecord{}, fmt.Errorf("encode parser warnings: %w", err)
	}

	item, err := scanParseRecord(r.db.QueryRowContext(ctx, `
INSERT INTO parser_results (id, discovery_run_id, evidence_id, parser_name, status, warnings, error_message)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, discovery_run_id, evidence_id, parser_name, status, warnings, error_message, created_at
`,
		id,
		params.DiscoveryRunID,
		params.EvidenceID,
		params.ParserName,
		params.Status,
		warnings,
		params.ErrorMessage,
	))
	if err != nil {
		return parser.ParseRecord{}, fmt.Errorf("create parse result: %w", err)
	}

	return item, nil
}

func (r ParseResultRepository) ListParseResultsByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]parser.ParseRecord, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, discovery_run_id, evidence_id, parser_name, status, warnings, error_message, created_at
FROM parser_results
WHERE discovery_run_id = $1
ORDER BY created_at ASC, id ASC
`, discoveryRunID)
	if err != nil {
		return nil, fmt.Errorf("list parse results: %w", err)
	}
	defer rows.Close()

	var results []parser.ParseRecord
	for rows.Next() {
		item, err := scanParseRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan parse result: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read parse results: %w", err)
	}

	return results, nil
}

func scanParseRecord(s scanner) (parser.ParseRecord, error) {
	var item parser.ParseRecord
	var errorMessage sql.NullString
	if err := s.Scan(
		&item.ID,
		&item.DiscoveryRunID,
		&item.EvidenceID,
		&item.ParserName,
		&item.Status,
		&item.Warnings,
		&errorMessage,
		&item.CreatedAt,
	); err != nil {
		return parser.ParseRecord{}, err
	}
	if errorMessage.Valid {
		item.ErrorMessage = &errorMessage.String
	}
	return item, nil
}
