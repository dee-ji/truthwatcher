package deploy

import (
	"context"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/queue"
)

func TestRunStateTransitions(t *testing.T) {
	ctx := context.Background()
	intentSvc := intent.NewInMemoryService()
	in, _ := intentSvc.Create(ctx, domain.Intent{Name: "x", Spec: map[string]any{"metadata": map[string]any{"name": "x"}}, Artifacts: []domain.CompiledArtifactView{{Vendor: "junos", Format: "set"}}})
	q := queue.NewInMemoryQueue()
	svc := NewStubServiceWithDependencies(nil, intentSvc).WithQueue(q)
	_, err := svc.Create(ctx, domain.DeploymentPlanRequest{IntentID: in.ID, IdempotencyKey: "k1", Targets: []string{"leaf-1", "leaf-2"}})
	if err != nil {
		t.Fatalf("create deployment: %v", err)
	}
	job, err := q.Dequeue(ctx, []queue.JobType{queue.JobTypeDeploy}, 0)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	run, err := svc.GetRun(ctx, job.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.Status != "queued" {
		t.Fatalf("expected queued, got %s", run.Status)
	}
	run, err = svc.ExecuteRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("execute run: %v", err)
	}
	if run.Status != "succeeded" && run.Status != "failed" && run.Status != "stopped" {
		t.Fatalf("unexpected terminal status %s", run.Status)
	}
	if run.StartedAt.IsZero() || run.FinishedAt.IsZero() {
		t.Fatalf("expected timestamps set")
	}
}
