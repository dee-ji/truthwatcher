package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"truthwatcher/internal/discovery"
)

type scanner interface {
	Scan(dest ...any) error
}

type DiscoveryRunRepository struct {
	db *sql.DB
}

func NewDiscoveryRunRepository(conn *sql.DB) DiscoveryRunRepository {
	return DiscoveryRunRepository{db: conn}
}

func (r DiscoveryRunRepository) CreateDiscoveryRun(ctx context.Context, params discovery.CreateDiscoveryRunParams) (discovery.DiscoveryRun, error) {
	if r.db == nil {
		return discovery.DiscoveryRun{}, fmt.Errorf("database is required")
	}

	id, err := discovery.NewID()
	if err != nil {
		return discovery.DiscoveryRun{}, err
	}
	if len(params.SeedInput) == 0 {
		params.SeedInput = json.RawMessage(`{}`)
	}

	run, err := scanDiscoveryRun(r.db.QueryRowContext(ctx, `
INSERT INTO discovery_runs (id, status, seed_input, started_at)
VALUES ($1, $2, $3, now())
RETURNING id, status, seed_input, started_at, completed_at, error_message, created_at, updated_at
`, id, discovery.StatusPending, params.SeedInput))
	if err != nil {
		return discovery.DiscoveryRun{}, fmt.Errorf("create discovery run: %w", err)
	}

	return run, nil
}

func (r DiscoveryRunRepository) GetDiscoveryRun(ctx context.Context, id string) (discovery.DiscoveryRun, error) {
	if r.db == nil {
		return discovery.DiscoveryRun{}, fmt.Errorf("database is required")
	}

	run, err := scanDiscoveryRun(r.db.QueryRowContext(ctx, `
SELECT id, status, seed_input, started_at, completed_at, error_message, created_at, updated_at
FROM discovery_runs
WHERE id = $1
`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return discovery.DiscoveryRun{}, discovery.ErrNotFound
	}
	if err != nil {
		return discovery.DiscoveryRun{}, fmt.Errorf("get discovery run: %w", err)
	}

	return run, nil
}

func (r DiscoveryRunRepository) ListDiscoveryRuns(ctx context.Context) ([]discovery.DiscoveryRun, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, status, seed_input, started_at, completed_at, error_message, created_at, updated_at
FROM discovery_runs
ORDER BY created_at DESC, id DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list discovery runs: %w", err)
	}
	defer rows.Close()

	var runs []discovery.DiscoveryRun
	for rows.Next() {
		run, err := scanDiscoveryRun(rows)
		if err != nil {
			return nil, fmt.Errorf("scan discovery run: %w", err)
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read discovery runs: %w", err)
	}

	return runs, nil
}

func (r DiscoveryRunRepository) UpdateDiscoveryRunStatus(ctx context.Context, params discovery.UpdateDiscoveryRunStatusParams) (discovery.DiscoveryRun, error) {
	if r.db == nil {
		return discovery.DiscoveryRun{}, fmt.Errorf("database is required")
	}

	run, err := scanDiscoveryRun(r.db.QueryRowContext(ctx, `
UPDATE discovery_runs
SET status = $2,
    completed_at = $3,
    error_message = $4,
    updated_at = now()
WHERE id = $1
RETURNING id, status, seed_input, started_at, completed_at, error_message, created_at, updated_at
`, params.ID, params.Status, params.CompletedAt, params.ErrorMessage))
	if errors.Is(err, sql.ErrNoRows) {
		return discovery.DiscoveryRun{}, discovery.ErrNotFound
	}
	if err != nil {
		return discovery.DiscoveryRun{}, fmt.Errorf("update discovery run status: %w", err)
	}

	return run, nil
}

func scanDiscoveryRun(s scanner) (discovery.DiscoveryRun, error) {
	var run discovery.DiscoveryRun
	var completedAt sql.NullTime
	var errorMessage sql.NullString

	if err := s.Scan(
		&run.ID,
		&run.Status,
		&run.SeedInput,
		&run.StartedAt,
		&completedAt,
		&errorMessage,
		&run.CreatedAt,
		&run.UpdatedAt,
	); err != nil {
		return discovery.DiscoveryRun{}, err
	}

	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}
	if errorMessage.Valid {
		run.ErrorMessage = &errorMessage.String
	}

	return run, nil
}
