package intent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/elsecall"
)

var ErrNotFound = errors.New("intent not found")

type Service interface {
	Create(context.Context, domain.Intent) (domain.Intent, error)
	List(context.Context) ([]domain.Intent, error)
	Get(context.Context, string) (domain.Intent, error)
	Validate(context.Context, string) error
	Compile(context.Context, string, string) (string, error)
}

type InMemoryService struct {
	mu      sync.RWMutex
	intents map[string]domain.Intent
	seq     int
}

func NewInMemoryService() *InMemoryService {
	return &InMemoryService{intents: map[string]domain.Intent{}}
}
func (s *InMemoryService) nextID() string { s.seq++; return fmt.Sprintf("intent-%d", s.seq) }
func (s *InMemoryService) Create(_ context.Context, in domain.Intent) (domain.Intent, error) {
	if in.Spec == nil {
		in.Spec = map[string]any{}
	}
	if err := validateSpec(in.Spec); err != nil {
		return domain.Intent{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	in.ID = s.nextID()
	in.CreatedAt = time.Now().UTC()
	if in.Revision == 0 {
		in.Revision = 1
	}
	s.intents[in.ID] = in
	return in, nil
}
func (s *InMemoryService) List(_ context.Context) ([]domain.Intent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Intent, 0, len(s.intents))
	for _, v := range s.intents {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (s *InMemoryService) Get(_ context.Context, id string) (domain.Intent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.intents[id]
	if !ok {
		return domain.Intent{}, ErrNotFound
	}
	return v, nil
}
func (s *InMemoryService) Validate(ctx context.Context, id string) error {
	in, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	return validateSpec(in.Spec)
}
func (s *InMemoryService) Compile(ctx context.Context, id, vendor string) (string, error) {
	in, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if vendor == "" {
		vendor = "junos"
	}
	compiler := elsecall.NewCompilerService()
	artifact, err := compiler.RenderForVendor(ctx, vendor, compiler.BuildIntermediate(in.Spec))
	if err != nil {
		return "", err
	}
	return "compiled " + artifact.Vendor, nil
}

type PostgresService struct {
	db       *sql.DB
	audit    audit.Service
	compiler *elsecall.CompilerService
}

func NewPostgresService(db *sql.DB, a audit.Service, c *elsecall.CompilerService) *PostgresService {
	return &PostgresService{db: db, audit: a, compiler: c}
}

func (s *PostgresService) Create(ctx context.Context, in domain.Intent) (domain.Intent, error) {
	if in.Spec == nil {
		in.Spec = map[string]any{}
	}
	if err := validateSpec(in.Spec); err != nil {
		return domain.Intent{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Intent{}, err
	}
	defer tx.Rollback()
	var id string
	if err := tx.QueryRowContext(ctx, `INSERT INTO intent_sets (name) VALUES ($1) RETURNING id::text`, in.Name).Scan(&id); err != nil {
		return domain.Intent{}, err
	}
	payload, _ := json.Marshal(in.Spec)
	if _, err := tx.ExecContext(ctx, `INSERT INTO intent_revisions (intent_set_id, revision, payload) VALUES ($1::uuid, 1, $2::jsonb)`, id, string(payload)); err != nil {
		return domain.Intent{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Intent{}, err
	}
	_ = s.audit.Emit(ctx, "spanreed", "intent.created", map[string]any{"intent_id": id, "name": in.Name})
	return s.Get(ctx, id)
}

func (s *PostgresService) List(ctx context.Context) ([]domain.Intent, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT s.id::text, s.name, COALESCE(r.revision, 1), COALESCE(r.payload, '{}'::jsonb), s.created_at
FROM intent_sets s
LEFT JOIN LATERAL (
    SELECT revision, payload FROM intent_revisions ir
    WHERE ir.intent_set_id = s.id
    ORDER BY revision DESC LIMIT 1
) r ON true
ORDER BY s.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Intent{}
	for rows.Next() {
		var in domain.Intent
		var payload []byte
		if err := rows.Scan(&in.ID, &in.Name, &in.Revision, &payload, &in.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(payload, &in.Spec)
		out = append(out, in)
	}
	return out, rows.Err()
}

func (s *PostgresService) Get(ctx context.Context, id string) (domain.Intent, error) {
	var in domain.Intent
	var payload []byte
	err := s.db.QueryRowContext(ctx, `
SELECT s.id::text, s.name, COALESCE(r.revision, 1), COALESCE(r.payload, '{}'::jsonb), s.created_at
FROM intent_sets s
LEFT JOIN LATERAL (
    SELECT id, revision, payload FROM intent_revisions ir
    WHERE ir.intent_set_id = s.id
    ORDER BY revision DESC LIMIT 1
) r ON true
WHERE s.id = $1::uuid`, id).Scan(&in.ID, &in.Name, &in.Revision, &payload, &in.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Intent{}, ErrNotFound
		}
		return domain.Intent{}, err
	}
	_ = json.Unmarshal(payload, &in.Spec)

	artRows, err := s.db.QueryContext(ctx, `
SELECT target_vendor, artifact_format, artifact, artifact_metadata, created_at
FROM compiled_artifacts
WHERE intent_revision_id = (
  SELECT id FROM intent_revisions WHERE intent_set_id = $1::uuid ORDER BY revision DESC LIMIT 1
)
ORDER BY created_at DESC`, id)
	if err == nil {
		defer artRows.Close()
		for artRows.Next() {
			var a domain.CompiledArtifactView
			var metadata []byte
			_ = artRows.Scan(&a.Vendor, &a.Format, &a.Artifact, &metadata, &a.CreatedAt)
			_ = json.Unmarshal(metadata, &a.Metadata)
			in.Artifacts = append(in.Artifacts, a)
		}
	}
	return in, nil
}

func (s *PostgresService) Validate(ctx context.Context, id string) error {
	in, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := validateSpec(in.Spec); err != nil {
		return err
	}
	_ = s.audit.Emit(ctx, "spanreed", "intent.validated", map[string]any{"intent_id": id})
	return nil
}

func (s *PostgresService) Compile(ctx context.Context, id, vendor string) (string, error) {
	in, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if err := validateSpec(in.Spec); err != nil {
		return "", err
	}
	if vendor == "" {
		vendor = "junos"
	}
	artifact, err := s.compiler.RenderForVendor(ctx, vendor, s.compiler.BuildIntermediate(in.Spec))
	if err != nil {
		return "", err
	}
	metadata, _ := json.Marshal(artifact.Metadata)
	_, err = s.db.ExecContext(ctx, `
INSERT INTO compiled_artifacts (intent_revision_id, target_vendor, artifact_format, artifact_metadata, artifact)
VALUES ((SELECT id FROM intent_revisions WHERE intent_set_id = $1::uuid ORDER BY revision DESC LIMIT 1), $2, $3, $4::jsonb, $5)
`, id, artifact.Vendor, artifact.Format, string(metadata), artifact.Contents)
	if err != nil {
		return "", err
	}
	_ = s.audit.Emit(ctx, "spanreed", "intent.compiled", map[string]any{"intent_id": id, "vendor": artifact.Vendor, "format": artifact.Format})
	return "compiled", nil
}

func validateSpec(spec map[string]any) error {
	metadata, ok := spec["metadata"].(map[string]any)
	if !ok {
		return errors.New("spec.metadata is required")
	}
	if _, ok := metadata["name"].(string); !ok {
		return errors.New("spec.metadata.name is required")
	}
	return nil
}
