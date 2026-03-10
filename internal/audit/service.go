package audit

import (
	"context"
	"database/sql"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

type Service interface {
	List(context.Context) ([]domain.AuditEvent, error)
	Emit(context.Context, string, string, map[string]any) error
}

type StubService struct{}

func NewStubService() *StubService { return &StubService{} }
func (s *StubService) List(context.Context) ([]domain.AuditEvent, error) {
	return []domain.AuditEvent{}, nil
}
func (s *StubService) Emit(context.Context, string, string, map[string]any) error { return nil }

type PostgresService struct {
	db *sql.DB
}

func NewPostgresService(db *sql.DB) *PostgresService { return &PostgresService{db: db} }

func (s *PostgresService) Emit(ctx context.Context, actor, action string, payload map[string]any) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit_events (actor, action, payload) VALUES ($1, $2, $3::jsonb)`, actor, action, mustJSON(payload))
	return err
}

func (s *PostgresService) List(ctx context.Context) ([]domain.AuditEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, actor, action, payload, created_at FROM audit_events ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AuditEvent{}
	for rows.Next() {
		var ev domain.AuditEvent
		var payload []byte
		if err := rows.Scan(&ev.ID, &ev.Actor, &ev.Action, &payload, &ev.CreatedAt); err != nil {
			return nil, err
		}
		ev.Payload = unmarshalJSON(payload)
		out = append(out, ev)
	}
	return out, rows.Err()
}
