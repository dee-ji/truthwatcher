package worker

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/queue"
)

func TestWorkerProcessesDeployJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := queue.NewInMemoryQueue()
	intentSvc := intent.NewInMemoryService()
	in, _ := intentSvc.Create(ctx, domain.Intent{Name: "x", Spec: map[string]any{"metadata": map[string]any{"name": "x"}}, Artifacts: []domain.CompiledArtifactView{{Vendor: "junos", Format: "set"}}})
	deploySvc := deploy.NewStubServiceWithDependencies(nil, intentSvc).WithQueue(q)
	_, _ = deploySvc.Create(ctx, domain.DeploymentPlanRequest{IntentID: in.ID, IdempotencyKey: "k1", Targets: []string{"leaf-1"}})

	w := New(slog.Default(), q, deploySvc)
	go func() { _ = w.Run(ctx) }()
	time.Sleep(50 * time.Millisecond)
	cancel()

	run, err := deploySvc.GetRun(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.Status == "queued" || run.Status == "running" {
		t.Fatalf("expected terminal status, got %s", run.Status)
	}
}
