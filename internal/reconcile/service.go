package reconcile

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/state"
)

var ErrRunNotFound = errors.New("reconcile run not found")

type Repository interface {
	CreateRun(context.Context, domain.ReconcileRun) (domain.ReconcileRun, error)
	GetRun(context.Context, string) (domain.ReconcileRun, error)
	UpdateRun(context.Context, domain.ReconcileRun) error
}

type Service interface {
	CreateRun(context.Context, domain.ReconcileRunRequest) (domain.ReconcileRun, error)
	GetRun(context.Context, string) (domain.ReconcileRun, error)
	ListFindings(context.Context) ([]domain.DriftFinding, error)
}

type service struct {
	runs    Repository
	intents intent.Service
	state   state.Service
	audit   audit.Service
}

func NewService(runs Repository, intents intent.Service, stateSvc state.Service, auditSvc audit.Service) Service {
	return &service{runs: runs, intents: intents, state: stateSvc, audit: auditSvc}
}

func NewStubService() Service {
	return NewService(NewInMemoryRepository(), intent.NewInMemoryService(), state.NewService(state.NewInMemoryRepository()), audit.NewStubService())
}

func (s *service) CreateRun(ctx context.Context, req domain.ReconcileRunRequest) (domain.ReconcileRun, error) {
	run, err := s.runs.CreateRun(ctx, domain.ReconcileRun{IntentID: req.IntentID, Status: "running", CreatedAt: time.Now().UTC()})
	if err != nil {
		return domain.ReconcileRun{}, err
	}
	if req.Actor == "" {
		req.Actor = "spanreed"
	}
	_ = s.audit.Emit(ctx, req.Actor, "reconcile.run.started", map[string]any{"reconcile_run_id": run.ID, "intent_id": req.IntentID})

	in, err := s.intents.Get(ctx, req.IntentID)
	if err != nil {
		run.Status = "failed"
		run.Summary = err.Error()
		_ = s.runs.UpdateRun(ctx, run)
		return run, nil
	}
	snapshots, err := s.state.LatestConfigSnapshots(ctx)
	if err != nil {
		return domain.ReconcileRun{}, err
	}
	findings := CompareArtifactsWithSnapshots(run.ID, in.Artifacts, snapshots)
	for i := range findings {
		findings[i].Remediation = map[string]any{
			"plan_status": "pending",
			"notes":       "TODO: attach auto-remediation planner once Oathgate/Highstorm workflow is integrated",
		}
	}
	if len(findings) > 0 {
		if err := s.state.StoreDriftFindings(ctx, findings); err != nil {
			return domain.ReconcileRun{}, err
		}
	}
	run.Status = "completed"
	run.CompletedAt = time.Now().UTC()
	run.FindingsCount = len(findings)
	if len(findings) == 0 {
		run.Summary = "no drift detected"
	} else {
		run.Summary = fmt.Sprintf("%d drift finding(s) detected", len(findings))
	}
	run.Findings = findings
	if err := s.runs.UpdateRun(ctx, run); err != nil {
		return domain.ReconcileRun{}, err
	}
	_ = s.audit.Emit(ctx, req.Actor, "reconcile.run.completed", map[string]any{"reconcile_run_id": run.ID, "intent_id": req.IntentID, "findings_count": len(findings)})
	if len(findings) > 0 {
		_ = s.audit.Emit(ctx, req.Actor, "drift.detection.completed", map[string]any{"reconcile_run_id": run.ID, "findings_count": len(findings)})
	}
	return run, nil
}

func (s *service) GetRun(ctx context.Context, id string) (domain.ReconcileRun, error) {
	return s.runs.GetRun(ctx, id)
}
func (s *service) ListFindings(ctx context.Context) ([]domain.DriftFinding, error) {
	return s.state.ListDriftFindings(ctx)
}

func CompareArtifactsWithSnapshots(runID string, intended []domain.CompiledArtifactView, actual []domain.ConfigSnapshot) []domain.DriftFinding {
	byHost := map[string]domain.CompiledArtifactView{}
	for _, art := range intended {
		host := strings.TrimSpace(art.Metadata["hostname"])
		if host == "" {
			continue
		}
		if _, exists := byHost[host]; !exists {
			byHost[host] = art
		}
	}
	actualByDevice := map[string]domain.ConfigSnapshot{}
	for _, snap := range actual {
		actualByDevice[snap.DeviceID] = snap
	}
	keys := make([]string, 0, len(byHost))
	for k := range byHost {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	findings := []domain.DriftFinding{}
	for _, deviceID := range keys {
		art := byHost[deviceID]
		snap, ok := actualByDevice[deviceID]
		if !ok {
			findings = append(findings, domain.DriftFinding{ReconcileRunID: runID, DeviceID: deviceID, Severity: "high", Kind: "missing_snapshot", Summary: "No stored configuration snapshot available for intended device", IntendedArtifact: art.Artifact, Finding: map[string]any{"reason": "snapshot_missing"}, CreatedAt: time.Now().UTC()})
			continue
		}
		if normalizeText(art.Artifact) != normalizeText(snap.Content) {
			findings = append(findings, domain.DriftFinding{ReconcileRunID: runID, DeviceID: deviceID, Severity: "medium", Kind: "config_mismatch", Summary: "Stored configuration snapshot differs from intended artifact", IntendedArtifact: art.Artifact, ActualSnapshotID: snap.ID, Finding: map[string]any{"reason": "content_mismatch", "snapshot_artifact_ref": snap.ArtifactRef}, CreatedAt: time.Now().UTC()})
		}
	}
	return findings
}

func normalizeText(in string) string {
	lines := strings.Split(strings.ReplaceAll(in, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return strings.Join(out, "\n")
}

type inMemoryRepository struct {
	mu   sync.RWMutex
	seq  int
	runs map[string]domain.ReconcileRun
}

func NewInMemoryRepository() Repository {
	return &inMemoryRepository{runs: map[string]domain.ReconcileRun{}}
}
func (r *inMemoryRepository) nextID() string { r.seq++; return fmt.Sprintf("recon-%d", r.seq) }
func (r *inMemoryRepository) CreateRun(_ context.Context, in domain.ReconcileRun) (domain.ReconcileRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	in.ID = r.nextID()
	r.runs[in.ID] = in
	return in, nil
}
func (r *inMemoryRepository) GetRun(_ context.Context, id string) (domain.ReconcileRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	in, ok := r.runs[id]
	if !ok {
		return domain.ReconcileRun{}, ErrRunNotFound
	}
	return in, nil
}
func (r *inMemoryRepository) UpdateRun(_ context.Context, in domain.ReconcileRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[in.ID] = in
	return nil
}

type postgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) Repository { return &postgresRepository{db: db} }
func (r *postgresRepository) CreateRun(ctx context.Context, in domain.ReconcileRun) (domain.ReconcileRun, error) {
	err := r.db.QueryRowContext(ctx, `INSERT INTO reconcile_runs (intent_id, status, summary, findings_count) VALUES ($1::uuid, $2, $3, $4) RETURNING id::text, created_at`, in.IntentID, in.Status, in.Summary, in.FindingsCount).Scan(&in.ID, &in.CreatedAt)
	return in, err
}
func (r *postgresRepository) GetRun(ctx context.Context, id string) (domain.ReconcileRun, error) {
	var out domain.ReconcileRun
	err := r.db.QueryRowContext(ctx, `SELECT id::text, intent_id::text, status, COALESCE(summary,''), findings_count, created_at, completed_at FROM reconcile_runs WHERE id = $1::uuid`, id).Scan(&out.ID, &out.IntentID, &out.Status, &out.Summary, &out.FindingsCount, &out.CreatedAt, &out.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ReconcileRun{}, ErrRunNotFound
	}
	return out, err
}
func (r *postgresRepository) UpdateRun(ctx context.Context, in domain.ReconcileRun) error {
	_, err := r.db.ExecContext(ctx, `UPDATE reconcile_runs SET status=$2, summary=$3, findings_count=$4, completed_at=$5 WHERE id=$1::uuid`, in.ID, in.Status, in.Summary, in.FindingsCount, in.CompletedAt)
	return err
}
