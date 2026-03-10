package deploy

import (
	"context"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
)

func TestCreateDeploymentPlan(t *testing.T) {
	ctx := context.Background()
	intentSvc := intent.NewInMemoryService()
	created, err := intentSvc.Create(ctx, domain.Intent{
		Name:      "fabric-update",
		Spec:      map[string]any{"metadata": map[string]any{"name": "fabric-update"}},
		Artifacts: []domain.CompiledArtifactView{{Vendor: "junos", Format: "set", Artifact: "set system host-name leaf1"}},
	})
	if err != nil {
		t.Fatalf("create intent: %v", err)
	}

	svc := NewStubServiceWithDependencies(nil, intentSvc)
	plan, err := svc.Create(ctx, domain.DeploymentPlanRequest{
		IntentID:              created.ID,
		IdempotencyKey:        "idem-1",
		Mode:                  "dry-run",
		Targets:               []string{"leaf-1", "leaf-2", "leaf-3"},
		BatchSize:             2,
		CanaryTargets:         1,
		RequireManualApproval: true,
	})
	if err != nil {
		t.Fatalf("create deployment: %v", err)
	}
	if plan.Mode != "dry-run" {
		t.Fatalf("expected dry-run mode, got %q", plan.Mode)
	}
	if len(plan.Rollout.Waves) != 2 || plan.Rollout.Waves[0].Name != "canary" {
		t.Fatalf("expected canary and batch waves, got %+v", plan.Rollout.Waves)
	}
	if len(plan.ArtifactRefs) == 0 {
		t.Fatalf("expected artifact refs")
	}
}

func TestCreateDeploymentIdempotent(t *testing.T) {
	ctx := context.Background()
	svc := NewStubService()
	req := domain.DeploymentPlanRequest{IntentID: "intent-1", IdempotencyKey: "idem-abc", Targets: []string{"leaf-1"}}
	first, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	second, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("second create: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected idempotent create to return same deployment id, got %s and %s", first.ID, second.ID)
	}
}

func TestCreateDeploymentValidation(t *testing.T) {
	svc := NewStubService()
	_, err := svc.Create(context.Background(), domain.DeploymentPlanRequest{IntentID: "intent-1"})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
