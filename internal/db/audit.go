package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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
