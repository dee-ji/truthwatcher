package deploy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
)

var (
	ErrDeploymentNotFound = errors.New("deployment not found")
	ErrInvalidRequest     = errors.New("invalid deployment request")
)

type Service interface {
	Create(context.Context, domain.DeploymentPlanRequest) (domain.Deployment, error)
	Get(context.Context, string) (domain.Deployment, error)
}

type StubService struct {
	mu                sync.RWMutex
	deployments       map[string]domain.Deployment
	idempotencyToPlan map[string]string
	seq               int
	audit             audit.Service
	intent            intent.Service
}

func NewStubService() *StubService {
	return &StubService{
		deployments:       map[string]domain.Deployment{},
		idempotencyToPlan: map[string]string{},
		audit:             audit.NewStubService(),
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
	defer s.mu.Unlock()
	if existingID, ok := s.idempotencyToPlan[req.IdempotencyKey]; ok {
		return s.deployments[existingID], nil
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
		StopConditions: []domain.StopCondition{
			{Type: "error_rate", Threshold: ">2%", Reason: "placeholder policy while Oathgate checks are integrated"},
		},
		RollbackPlan: domain.RollbackPlan{
			Strategy: "previous-known-good",
			Steps: []string{
				"capture pre-deployment state snapshot",
				"re-apply last-known-good artifact",
				"TODO: enforce rollback orchestration through Highstorm",
			},
		},
		CreatedAt: created,
	}

	// TODO: Add strict approval workflow enforcement once Oathgate/Highstorm policy service is available.
	// TODO: Add stop-condition runtime enforcement during deployment execution.
	s.deployments[id] = dep
	s.idempotencyToPlan[req.IdempotencyKey] = id
	_ = s.audit.Emit(ctx, "spanreed", "deployment.plan.created", map[string]any{
		"deployment_id":      dep.ID,
		"intent_id":          dep.IntentID,
		"idempotency_key":    dep.IdempotencyKey,
		"mode":               dep.Mode,
		"target_count":       len(dep.Targets),
		"artifact_ref_count": len(dep.ArtifactRefs),
	})
	return dep, nil
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
