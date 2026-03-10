package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

type Repository interface {
	CreateConfigSnapshot(context.Context, domain.ConfigSnapshot) (domain.ConfigSnapshot, error)
	LatestConfigSnapshots(context.Context) ([]domain.ConfigSnapshot, error)
	CreateOperationalSnapshot(context.Context, domain.OperationalSnapshot) (domain.OperationalSnapshot, error)
	CreateDriftFindings(context.Context, []domain.DriftFinding) error
	ListDriftFindings(context.Context) ([]domain.DriftFinding, error)
}

type Service interface {
	CreateConfigSnapshot(context.Context, domain.ConfigSnapshot) (domain.ConfigSnapshot, error)
	LatestConfigSnapshots(context.Context) ([]domain.ConfigSnapshot, error)
	CreateOperationalSnapshot(context.Context, domain.OperationalSnapshot) (domain.OperationalSnapshot, error)
	StoreDriftFindings(context.Context, []domain.DriftFinding) error
	ListDriftFindings(context.Context) ([]domain.DriftFinding, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) CreateConfigSnapshot(ctx context.Context, in domain.ConfigSnapshot) (domain.ConfigSnapshot, error) {
	if in.CapturedAt.IsZero() {
		in.CapturedAt = time.Now().UTC()
	}
	if in.Source == "" {
		in.Source = "offline-fixture"
	}
	return s.repo.CreateConfigSnapshot(ctx, in)
}
func (s *service) LatestConfigSnapshots(ctx context.Context) ([]domain.ConfigSnapshot, error) {
	return s.repo.LatestConfigSnapshots(ctx)
}
func (s *service) CreateOperationalSnapshot(ctx context.Context, in domain.OperationalSnapshot) (domain.OperationalSnapshot, error) {
	if in.CapturedAt.IsZero() {
		in.CapturedAt = time.Now().UTC()
	}
	if in.Source == "" {
		in.Source = "offline-fixture"
	}
	return s.repo.CreateOperationalSnapshot(ctx, in)
}
func (s *service) StoreDriftFindings(ctx context.Context, findings []domain.DriftFinding) error {
	return s.repo.CreateDriftFindings(ctx, findings)
}
func (s *service) ListDriftFindings(ctx context.Context) ([]domain.DriftFinding, error) {
	return s.repo.ListDriftFindings(ctx)
}

type inMemoryRepository struct {
	mu       sync.RWMutex
	seq      int
	cfg      []domain.ConfigSnapshot
	op       []domain.OperationalSnapshot
	findings []domain.DriftFinding
}

func NewInMemoryRepository() Repository { return &inMemoryRepository{} }
func (r *inMemoryRepository) nextID(prefix string) string {
	r.seq++
	return fmt.Sprintf("%s-%d", prefix, r.seq)
}
func (r *inMemoryRepository) CreateConfigSnapshot(_ context.Context, in domain.ConfigSnapshot) (domain.ConfigSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	in.ID = r.nextID("cfgsnap")
	in.CreatedAt = time.Now().UTC()
	r.cfg = append(r.cfg, in)
	return in, nil
}
func (r *inMemoryRepository) LatestConfigSnapshots(_ context.Context) ([]domain.ConfigSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	latest := map[string]domain.ConfigSnapshot{}
	for _, s := range r.cfg {
		if cur, ok := latest[s.DeviceID]; !ok || cur.CapturedAt.Before(s.CapturedAt) {
			latest[s.DeviceID] = s
		}
	}
	out := make([]domain.ConfigSnapshot, 0, len(latest))
	for _, s := range latest {
		out = append(out, s)
	}
	return out, nil
}
func (r *inMemoryRepository) CreateOperationalSnapshot(_ context.Context, in domain.OperationalSnapshot) (domain.OperationalSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	in.ID = r.nextID("opsnap")
	in.CreatedAt = time.Now().UTC()
	r.op = append(r.op, in)
	return in, nil
}
func (r *inMemoryRepository) CreateDriftFindings(_ context.Context, in []domain.DriftFinding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range in {
		in[i].ID = r.nextID("drift")
		if in[i].CreatedAt.IsZero() {
			in[i].CreatedAt = time.Now().UTC()
		}
		r.findings = append(r.findings, in[i])
	}
	return nil
}
func (r *inMemoryRepository) ListDriftFindings(_ context.Context) ([]domain.DriftFinding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := append([]domain.DriftFinding(nil), r.findings...)
	return out, nil
}

type postgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) Repository { return &postgresRepository{db: db} }

func (r *postgresRepository) CreateConfigSnapshot(ctx context.Context, in domain.ConfigSnapshot) (domain.ConfigSnapshot, error) {
	err := r.db.QueryRowContext(ctx, `INSERT INTO config_snapshots (device_id, captured_at, artifact_ref, content, source) VALUES ($1::uuid, $2, $3, $4, $5) RETURNING id::text, created_at`, in.DeviceID, in.CapturedAt, in.ArtifactRef, in.Content, in.Source).Scan(&in.ID, &in.CreatedAt)
	return in, err
}
func (r *postgresRepository) LatestConfigSnapshots(ctx context.Context) ([]domain.ConfigSnapshot, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT ON (device_id) id::text, device_id::text, captured_at, COALESCE(artifact_ref,''), content, COALESCE(source,''), created_at FROM config_snapshots ORDER BY device_id, captured_at DESC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ConfigSnapshot
	for rows.Next() {
		var s domain.ConfigSnapshot
		if err := rows.Scan(&s.ID, &s.DeviceID, &s.CapturedAt, &s.ArtifactRef, &s.Content, &s.Source, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
func (r *postgresRepository) CreateOperationalSnapshot(ctx context.Context, in domain.OperationalSnapshot) (domain.OperationalSnapshot, error) {
	content, _ := json.Marshal(in.Content)
	err := r.db.QueryRowContext(ctx, `INSERT INTO operational_snapshots (device_id, captured_at, content, source) VALUES ($1::uuid, $2, $3::jsonb, $4) RETURNING id::text, created_at`, in.DeviceID, in.CapturedAt, string(content), in.Source).Scan(&in.ID, &in.CreatedAt)
	return in, err
}
func (r *postgresRepository) CreateDriftFindings(ctx context.Context, findings []domain.DriftFinding) error {
	for _, f := range findings {
		findingJSON, _ := json.Marshal(f.Finding)
		remediationJSON, _ := json.Marshal(f.Remediation)
		if _, err := r.db.ExecContext(ctx, `INSERT INTO drift_findings (reconcile_run_id, device_id, severity, kind, summary, intended_artifact, actual_snapshot_id, finding, remediation_plan) VALUES (NULLIF($1,'')::uuid, $2::uuid, $3, $4, $5, NULLIF($6,''), NULLIF($7,'')::uuid, $8::jsonb, $9::jsonb)`, f.ReconcileRunID, f.DeviceID, f.Severity, f.Kind, f.Summary, f.IntendedArtifact, f.ActualSnapshotID, string(findingJSON), string(remediationJSON)); err != nil {
			return err
		}
	}
	return nil
}
func (r *postgresRepository) ListDriftFindings(ctx context.Context) ([]domain.DriftFinding, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id::text, COALESCE(reconcile_run_id::text,''), device_id::text, severity, kind, summary, COALESCE(intended_artifact,''), COALESCE(actual_snapshot_id::text,''), finding, remediation_plan, created_at FROM drift_findings ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DriftFinding{}
	for rows.Next() {
		var f domain.DriftFinding
		var findingJSON, remediationJSON []byte
		if err := rows.Scan(&f.ID, &f.ReconcileRunID, &f.DeviceID, &f.Severity, &f.Kind, &f.Summary, &f.IntendedArtifact, &f.ActualSnapshotID, &findingJSON, &remediationJSON, &f.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(findingJSON, &f.Finding)
		_ = json.Unmarshal(remediationJSON, &f.Remediation)
		out = append(out, f)
	}
	return out, rows.Err()
}
