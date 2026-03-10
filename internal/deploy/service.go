package deploy

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/queue"
)

var (
	ErrDeploymentNotFound = errors.New("deployment not found")
	ErrRunNotFound        = errors.New("deployment run not found")
	ErrInvalidRequest     = errors.New("invalid deployment request")
)

type Service interface {
	Create(context.Context, domain.DeploymentPlanRequest) (domain.Deployment, error)
	Get(context.Context, string) (domain.Deployment, error)
	GetRun(context.Context, string) (domain.DeploymentRun, error)
	ExecuteRun(context.Context, string) (domain.DeploymentRun, error)
	StopRun(context.Context, string, string) error
}

type Metrics struct {
	mu               sync.Mutex
	queuedJobs       int64
	runDurations     []time.Duration
	lastCompletedRun time.Time
	lastCompletedID  string
}

func (m *Metrics) recordQueued() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queuedJobs++
}

func (m *Metrics) recordDuration(runID string, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runDurations = append(m.runDurations, d)
	m.lastCompletedID = runID
	m.lastCompletedRun = time.Now().UTC()
}

type StopConditionEvaluator interface {
	ShouldStop(domain.StopCondition, domain.DeploymentRun) (bool, string)
}

type DefaultStopConditionEvaluator struct{}

func (e DefaultStopConditionEvaluator) ShouldStop(cond domain.StopCondition, run domain.DeploymentRun) (bool, string) {
	if cond.Type != "error_rate" || cond.Threshold == "" {
		return false, ""
	}
	if len(run.Targets) == 0 {
		return false, ""
	}
	failed := 0
	for _, t := range run.Targets {
		if t.Status == "failed" {
			failed++
		}
	}
	rate := float64(failed) / float64(len(run.Targets))
	if strings.HasPrefix(cond.Threshold, ">") {
		n := strings.TrimPrefix(cond.Threshold, ">")
		n = strings.TrimSuffix(n, "%")
		v, _ := strconv.ParseFloat(n, 64)
		if rate*100 > v {
			return true, fmt.Sprintf("error rate %.2f%% exceeded %s", rate*100, cond.Threshold)
		}
	}
	return false, ""
}

type StubService struct {
	mu                sync.RWMutex
	deployments       map[string]domain.Deployment
	runs              map[string]domain.DeploymentRun
	idempotencyToPlan map[string]string
	seq               int
	runSeq            int
	audit             audit.Service
	intent            intent.Service
	queue             queue.Queue
	metrics           *Metrics
	stopEval          StopConditionEvaluator
}

func NewStubService() *StubService {
	return &StubService{
		deployments:       map[string]domain.Deployment{},
		runs:              map[string]domain.DeploymentRun{},
		idempotencyToPlan: map[string]string{},
		audit:             audit.NewStubService(),
		metrics:           &Metrics{},
		stopEval:          DefaultStopConditionEvaluator{},
	}
}

func NewStubServiceWithDependencies(a audit.Service, i intent.Service) *StubService {
	s := NewStubService()
	if a != nil {
		s.audit = a
	}
	if i != nil {
		s.intent = i
	}
	return s
}

func (s *StubService) WithQueue(q queue.Queue) *StubService { s.queue = q; return s }

func (s *StubService) Create(ctx context.Context, req domain.DeploymentPlanRequest) (domain.Deployment, error) {
	if err := validateRequest(req); err != nil {
		return domain.Deployment{}, err
	}
	if s.intent != nil {
		in, err := s.intent.Get(ctx, req.IntentID)
		if err != nil {
			return domain.Deployment{}, fmt.Errorf("resolve intent: %w", err)
		}
		if len(in.Artifacts) == 0 {
			return domain.Deployment{}, fmt.Errorf("%w: no compiled artifacts for intent", ErrInvalidRequest)
		}
	}

	s.mu.Lock()
	if existingID, ok := s.idempotencyToPlan[req.IdempotencyKey]; ok {
		dep := s.deployments[existingID]
		s.mu.Unlock()
		return dep, nil
	}

	s.seq++
	id := fmt.Sprintf("dep-%d", s.seq)
	created := time.Now().UTC()
	mode := req.Mode
	if mode == "" {
		mode = "dry-run"
	}
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}
	canary := req.CanaryTargets
	if canary <= 0 {
		canary = 1
	}

	artifactRefs := []string{"latest-compiled-artifacts"}
	if s.intent != nil {
		if in, err := s.intent.Get(ctx, req.IntentID); err == nil {
			artifactRefs = make([]string, 0, len(in.Artifacts))
			for _, art := range in.Artifacts {
				artifactRefs = append(artifactRefs, fmt.Sprintf("%s/%s", art.Vendor, art.Format))
			}
		}
	}

	targets := req.Targets
	if len(targets) == 0 {
		targets = []string{"all-intent-targets"}
	}
	dep := domain.Deployment{
		ID:             id,
		IntentID:       req.IntentID,
		Status:         "planned",
		IdempotencyKey: req.IdempotencyKey,
		Mode:           mode,
		ArtifactRefs:   artifactRefs,
		Targets:        append([]string{}, targets...),
		Rollout: domain.Rollout{Waves: []domain.RolloutWave{
			{Name: "canary", Order: 1, MaxTargets: canary, CanaryTargets: canary, RequiresApproval: req.RequireManualApproval, PlannedTargetCount: min(canary, len(targets))},
			{Name: "batch", Order: 2, MaxTargets: batchSize, RequiresApproval: req.RequireManualApproval, PlannedTargetCount: max(0, len(targets)-canary)},
		}},
		StopConditions: []domain.StopCondition{{Type: "error_rate", Threshold: ">50%", Reason: "simulation stop hook"}},
		RollbackPlan:   domain.RollbackPlan{Strategy: "previous-known-good", Steps: []string{"capture pre-deployment state snapshot", "re-apply last-known-good artifact"}},
		CreatedAt:      created,
	}
	s.runSeq++
	runID := fmt.Sprintf("run-%d", s.runSeq)
	run := domain.DeploymentRun{ID: runID, DeploymentPlanID: id, Status: "queued", Simulation: true, CreatedAt: created}
	run.Targets = make([]domain.DeploymentTarget, 0, len(targets))
	for idx, target := range targets {
		run.Targets = append(run.Targets, domain.DeploymentTarget{
			ID:              fmt.Sprintf("%s-target-%d", runID, idx+1),
			DeploymentRunID: runID,
			DeviceID:        target,
			ArtifactRef:     artifactRefs[idx%len(artifactRefs)],
			Wave:            assignWave(dep.Rollout.Waves, idx),
			Status:          "queued",
			CreatedAt:       created,
		})
	}

	s.deployments[id] = dep
	s.runs[runID] = run
	s.idempotencyToPlan[req.IdempotencyKey] = id
	s.mu.Unlock()

	if s.queue != nil {
		_ = s.queue.Enqueue(ctx, queue.Job{ID: runID, Type: queue.JobTypeDeploy, CreatedAt: created, Payload: map[string]any{"run_id": runID}})
		s.metrics.recordQueued()
	}
	_ = s.audit.Emit(ctx, "tw-server", "deployment.run.queued", map[string]any{"deployment_id": dep.ID, "run_id": runID})
	return dep, nil
}

func assignWave(waves []domain.RolloutWave, idx int) int {
	if len(waves) == 0 {
		return 1
	}
	if idx < waves[0].PlannedTargetCount {
		return waves[0].Order
	}
	return waves[len(waves)-1].Order
}

func (s *StubService) Get(_ context.Context, id string) (domain.Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.deployments[id]
	if !ok {
		return domain.Deployment{}, ErrDeploymentNotFound
	}
	return d, nil
}

func (s *StubService) GetRun(_ context.Context, id string) (domain.DeploymentRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.runs[id]
	if !ok {
		return domain.DeploymentRun{}, ErrRunNotFound
	}
	return r, nil
}

func (s *StubService) StopRun(ctx context.Context, id, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok {
		return ErrRunNotFound
	}
	if run.Status == "succeeded" || run.Status == "failed" {
		return nil
	}
	run.Status = "stopped"
	run.StoppedReason = reason
	s.runs[id] = run
	_ = s.audit.Emit(ctx, "tw-worker", "deployment.run.stopped", map[string]any{"run_id": id, "reason": reason})
	return nil
}

func (s *StubService) ExecuteRun(ctx context.Context, id string) (domain.DeploymentRun, error) {
	start := time.Now()
	s.mu.Lock()
	run, ok := s.runs[id]
	if !ok {
		s.mu.Unlock()
		return domain.DeploymentRun{}, ErrRunNotFound
	}
	run.Status = "running"
	run.StartedAt = time.Now().UTC()
	s.runs[id] = run
	dep := s.deployments[run.DeploymentPlanID]
	s.mu.Unlock()
	_ = s.audit.Emit(ctx, "tw-worker", "deployment.run.started", map[string]any{"run_id": id})

	for i, target := range run.Targets {
		run.Targets[i].Status = deterministicTargetStatus(id, target.DeviceID)
		if run.Targets[i].Status == "failed" {
			run.Targets[i].Result = "simulated failure"
		} else {
			run.Targets[i].Result = "simulated success"
		}
		for _, cond := range dep.StopConditions {
			tmp := run
			if stop, reason := s.stopEval.ShouldStop(cond, tmp); stop {
				run.Status = "stopped"
				run.StoppedReason = reason
				goto done
			}
		}
	}

	if hasFailed(run) {
		run.Status = "failed"
	} else {
		run.Status = "succeeded"
	}

done:
	run.FinishedAt = time.Now().UTC()
	run.DurationSeconds = run.FinishedAt.Sub(run.StartedAt).Seconds()
	s.mu.Lock()
	s.runs[id] = run
	s.mu.Unlock()
	s.metrics.recordDuration(id, time.Since(start))
	_ = s.audit.Emit(ctx, "tw-worker", "deployment.run.finished", map[string]any{"run_id": id, "status": run.Status, "duration_seconds": run.DurationSeconds})
	return run, nil
}

func hasFailed(run domain.DeploymentRun) bool {
	for _, t := range run.Targets {
		if t.Status == "failed" {
			return true
		}
	}
	return false
}

func deterministicTargetStatus(runID, target string) string {
	h := sha1.Sum([]byte(runID + ":" + target))
	if hex.EncodeToString(h[:])[0]%5 == 0 {
		return "failed"
	}
	return "succeeded"
}

func validateRequest(req domain.DeploymentPlanRequest) error {
	if strings.TrimSpace(req.IntentID) == "" {
		return fmt.Errorf("%w: intent_id is required", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidRequest)
	}
	if req.Mode != "" && req.Mode != "dry-run" && req.Mode != "execute" {
		return fmt.Errorf("%w: mode must be dry-run or execute", ErrInvalidRequest)
	}
	if req.BatchSize < 0 || req.CanaryTargets < 0 {
		return fmt.Errorf("%w: batch_size and canary_targets must be >= 0", ErrInvalidRequest)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
