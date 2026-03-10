package reconcile

import (
	"context"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/state"
)

func TestCompareArtifactsWithSnapshots(t *testing.T) {
	intended := []domain.CompiledArtifactView{{Vendor: "junos", Artifact: "set system host-name leaf-1\n", Metadata: map[string]string{"hostname": "leaf-1"}}}
	actual := []domain.ConfigSnapshot{{ID: "cfg-1", DeviceID: "leaf-1", Content: "set system host-name leaf-2\n"}}
	findings := CompareArtifactsWithSnapshots("run-1", intended, actual)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding got %d", len(findings))
	}
	if findings[0].Kind != "config_mismatch" {
		t.Fatalf("expected config_mismatch got %s", findings[0].Kind)
	}
}

func TestCompareArtifactsWithSnapshotsNoDrift(t *testing.T) {
	intended := []domain.CompiledArtifactView{{Vendor: "junos", Artifact: "set system host-name leaf-1\n", Metadata: map[string]string{"hostname": "leaf-1"}}}
	actual := []domain.ConfigSnapshot{{ID: "cfg-1", DeviceID: "leaf-1", Content: " set system host-name leaf-1 \n\n"}}
	findings := CompareArtifactsWithSnapshots("run-1", intended, actual)
	if len(findings) != 0 {
		t.Fatalf("expected no findings got %d", len(findings))
	}
}

func TestReconcileRunIntegration(t *testing.T) {
	ctx := context.Background()
	intentSvc := intent.NewInMemoryService()
	stateSvc := state.NewService(state.NewInMemoryRepository())
	svc := NewService(NewInMemoryRepository(), intentSvc, stateSvc, audit.NewStubService())

	in, err := intentSvc.Create(ctx, domain.Intent{Name: "leaf", Spec: map[string]any{"metadata": map[string]any{"name": "leaf-1"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := intentSvc.Compile(ctx, in.ID, "junos"); err != nil {
		t.Fatal(err)
	}
	_, _ = stateSvc.CreateConfigSnapshot(ctx, domain.ConfigSnapshot{DeviceID: "leaf-1", Content: "set system host-name leaf-1\nset routing-options autonomous-system 65000\n"})

	run, err := svc.CreateRun(ctx, domain.ReconcileRunRequest{IntentID: in.ID, Actor: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != "completed" {
		t.Fatalf("expected completed got %s", run.Status)
	}
	if run.FindingsCount != 0 {
		t.Fatalf("expected 0 findings got %d", run.FindingsCount)
	}
}
