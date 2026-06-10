package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"truthwatcher/migrations"
)

type MigrationStatus struct {
	Migration migrations.Migration
	Applied   bool
}

type Migrator struct {
	db         *sql.DB
	migrations []migrations.Migration
}

func NewMigrator(conn *sql.DB, source []migrations.Migration) Migrator {
	return Migrator{
		db:         conn,
		migrations: source,
	}
}

func (m Migrator) Up(ctx context.Context) ([]migrations.Migration, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	var ran []migrations.Migration
	for _, migration := range m.migrations {
		if applied[migration.Version] {
			continue
		}
		if strings.TrimSpace(migration.UpSQL) == "" {
			return nil, fmt.Errorf("migration %s has no up sql", migration.ID)
		}

		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("begin migration %s: %w", migration.ID, err)
		}

		if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("run migration %s: %w", migration.ID, err)
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO schema_migrations (version, name)
VALUES ($1, $2)
ON CONFLICT (version) DO NOTHING
`, migration.Version, migration.Name); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("record migration %s: %w", migration.ID, err)
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit migration %s: %w", migration.ID, err)
		}

		applied[migration.Version] = true
		ran = append(ran, migration)
	}

	return ran, nil
}

func (m Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database is required")
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	status := make([]MigrationStatus, 0, len(m.migrations))
	for _, migration := range m.migrations {
		status = append(status, MigrationStatus{
			Migration: migration,
			Applied:   applied[migration.Version],
		})
	}

	return status, nil
}

func (m Migrator) appliedVersions(ctx context.Context) (map[int]bool, error) {
	rows, err := m.db.QueryContext(ctx, `
SELECT version
FROM schema_migrations
ORDER BY version
`)
	if err != nil {
		if isUndefinedTable(err) {
			return map[int]bool{}, nil
		}
		return nil, fmt.Errorf("query schema migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan schema migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read schema migrations: %w", err)
	}

	return applied, nil
}

func isUndefinedTable(err error) bool {
	return strings.Contains(err.Error(), `relation "schema_migrations" does not exist`) ||
		strings.Contains(err.Error(), "SQLSTATE 42P01")
}
