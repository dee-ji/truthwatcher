package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"truthwatcher/internal/audit"
	"truthwatcher/internal/discovery"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(conn *sql.DB) AuditRepository {
	return AuditRepository{db: conn}
}

func (r AuditRepository) CreateAuditRecord(ctx context.Context, params audit.CreateRecordParams) (audit.Record, error) {
	if r.db == nil {
		return audit.Record{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return audit.Record{}, err
	}
	if len(params.Context) == 0 {
		params.Context = json.RawMessage(`{}`)
	}

	result, err := scanAuditRecord(r.db.QueryRowContext(ctx, `
INSERT INTO audit_records (
    id,
    action,
    initiator,
    request_id,
    discovery_run_id,
    target,
    method,
    profile,
    task,
    command_or_api,
    status,
    evidence_id,
    error_message,
    started_at,
    completed_at,
    context
)
VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, '')::uuid, $6, $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11, NULLIF($12, '')::uuid, NULLIF($13, ''), $14, $15, $16)
RETURNING id, action, initiator, COALESCE(request_id, ''), COALESCE(discovery_run_id::text, ''), target, method, COALESCE(profile, ''), COALESCE(task, ''), COALESCE(command_or_api, ''), status, COALESCE(evidence_id::text, ''), COALESCE(error_message, ''), started_at, completed_at, context
`,
		id,
		params.Action,
		params.Initiator,
		params.RequestID,
		params.DiscoveryRunID,
		params.Target,
		params.Method,
		params.Profile,
		params.Task,
		params.CommandOrAPI,
		params.Status,
		params.EvidenceID,
		params.ErrorMessage,
		params.StartedAt,
		params.CompletedAt,
		params.Context,
	))
	if err != nil {
		return audit.Record{}, fmt.Errorf("create audit record: %w", err)
	}

	return result, nil
}

func (r AuditRepository) ListAuditRecords(ctx context.Context, filters audit.ListRecordsFilters) ([]audit.Record, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}
	clauses := []string{}
	args := []any{}
	add := func(format string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(format, len(args)))
	}
	if strings.TrimSpace(filters.DiscoveryRunID) != "" {
		add("discovery_run_id = $%d::uuid", filters.DiscoveryRunID)
	}
	if strings.TrimSpace(filters.EvidenceID) != "" {
		add("evidence_id = $%d::uuid", filters.EvidenceID)
	}
	if strings.TrimSpace(filters.RequestID) != "" {
		add("request_id = $%d", filters.RequestID)
	}
	if strings.TrimSpace(filters.Action) != "" {
		add("action = $%d", filters.Action)
	}
	if strings.TrimSpace(filters.Status) != "" {
		add("status = $%d", filters.Status)
	}
	if strings.TrimSpace(filters.Target) != "" {
		add("target ILIKE '%' || $%d || '%'", filters.Target)
	}
	if strings.TrimSpace(filters.Method) != "" {
		add("method = $%d", filters.Method)
	}
	if strings.TrimSpace(filters.Profile) != "" {
		add("profile = $%d", filters.Profile)
	}
	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	query := `
SELECT id, action, initiator, COALESCE(request_id, ''), COALESCE(discovery_run_id::text, ''), target, method, COALESCE(profile, ''), COALESCE(task, ''), COALESCE(command_or_api, ''), status, COALESCE(evidence_id::text, ''), COALESCE(error_message, ''), started_at, completed_at, context
FROM audit_records`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit)
	query += fmt.Sprintf(" ORDER BY started_at DESC, created_at DESC LIMIT $%d", len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit records: %w", err)
	}
	defer rows.Close()
	var results []audit.Record
	for rows.Next() {
		item, err := scanAuditRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan audit record: %w", err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list audit records: %w", err)
	}
	return results, nil
}

func scanAuditRecord(s scanner) (audit.Record, error) {
	var item audit.Record
	if err := s.Scan(
		&item.ID,
		&item.Action,
		&item.Initiator,
		&item.RequestID,
		&item.DiscoveryRunID,
		&item.Target,
		&item.Method,
		&item.Profile,
		&item.Task,
		&item.CommandOrAPI,
		&item.Status,
		&item.EvidenceID,
		&item.ErrorMessage,
		&item.StartedAt,
		&item.CompletedAt,
		&item.Context,
	); err != nil {
		return audit.Record{}, err
	}
	return item, nil
}
