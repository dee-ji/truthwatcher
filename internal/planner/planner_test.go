package planner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/policy"
)

type fakeAssets struct {
	assets        []assets.Asset
	facts         []assets.Fact
	relationships []assets.Relationship
}

func (f fakeAssets) GetAsset(ctx context.Context, id string) (assets.Asset, error) {
	for _, item := range f.assets {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Asset{}, assets.ErrNotFound
}

func (f fakeAssets) ListAssets(ctx context.Context) ([]assets.Asset, error) {
	return f.assets, nil
}

func (f fakeAssets) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	var result []assets.Fact
	for _, item := range f.facts {
		if item.AssetID == assetID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (f fakeAssets) ListRelationships(ctx context.Context) ([]assets.Relationship, error) {
	return f.relationships, nil
}

func TestCreatePlanForUnknownExplicitTarget(t *testing.T) {
	service := NewService(Options{Assets: testGraph(), Policy: policy.NewEngine()})

	plan, err := service.CreatePlan(context.Background(), Request{
		Target:  "fixture://junos-mx",
		Method:  discovery.FakeMethod,
		Profile: discovery.ProfileJuniperJunos,
	})
	if err != nil {
		t.Fatalf("CreatePlan returned error: %v", err)
	}

	if !plan.ApprovalRequired {
		t.Fatal("plan does not require approval")
	}
	if plan.ExecutionAllowed {
		t.Fatal("plan allows execution")
	}
	if got, want := len(plan.Steps), 3; got != want {
		t.Fatalf("step count = %d, want %d", got, want)
	}
	for _, step := range plan.Steps {
		if step.Target != "fixture://junos-mx" {
			t.Fatalf("target = %q, want fixture://junos-mx", step.Target)
		}
		if step.RiskLevel != "low_read_only" {
			t.Fatalf("risk = %q, want low_read_only", step.RiskLevel)
		}
		if step.ExpectedEvidence == "" {
			t.Fatal("expected evidence is empty")
		}
	}
}

func TestCreatePlanUsesGraphGaps(t *testing.T) {
	service := NewService(Options{Assets: testGraph(), Policy: policy.NewEngine()})

	plan, err := service.CreatePlan(context.Background(), Request{
		SeedInput: json.RawMessage(`{"target":"router-a","method":"ssh","profile":"juniper_junos"}`),
	})
	if err != nil {
		t.Fatalf("CreatePlan returned error: %v", err)
	}

	if len(plan.Steps) == 0 {
		t.Fatal("plan has no steps")
	}
	if !containsTask(plan.Steps, policy.TaskGetInventory) {
		t.Fatalf("steps = %#v, want inventory task for missing model/serial", plan.Steps)
	}
	if !containsTask(plan.Steps, policy.TaskGetNeighbors) {
		t.Fatalf("steps = %#v, want neighbors task for missing relationships", plan.Steps)
	}
}

func TestCreatePlanRejectsScopeExpansionTarget(t *testing.T) {
	service := NewService(Options{Assets: testGraph(), Policy: policy.NewEngine()})

	_, err := service.CreatePlan(context.Background(), Request{
		Target:  "10.0.0.0/24",
		Method:  "ssh",
		Profile: discovery.ProfileJuniperJunos,
	})
	if err == nil {
		t.Fatal("CreatePlan returned nil error for CIDR target")
	}
}

func TestCreatePlanRejectsDeniedTask(t *testing.T) {
	service := NewService(Options{Assets: testGraph(), Policy: policy.NewEngine()})

	_, err := service.CreatePlan(context.Background(), Request{
		Target:  "router-a",
		Method:  "ssh",
		Profile: discovery.ProfileJuniperJunos,
		Tasks:   []policy.Task{"configure"},
	})
	if err == nil {
		t.Fatal("CreatePlan returned nil error for denied task")
	}
}

func containsTask(steps []Step, task policy.Task) bool {
	for _, step := range steps {
		if step.Task == task {
			return true
		}
	}
	return false
}

func testGraph() fakeAssets {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	evidenceID := "evidence-a"
	return fakeAssets{
		assets: []assets.Asset{{
			ID:               "asset-a",
			Type:             "device",
			IdentityKey:      "device:serial:aaa",
			Vendor:           stringPtr("juniper"),
			Confidence:       0.9,
			ConfidenceReason: "directly observed",
			State:            assets.StateObserved,
			Metadata:         json.RawMessage(`{}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}},
		facts: []assets.Fact{{
			ID:               "fact-a",
			AssetID:          "asset-a",
			Name:             "hostname",
			Value:            json.RawMessage(`"router-a"`),
			Source:           "parser",
			Confidence:       0.9,
			ConfidenceReason: "directly observed",
			State:            assets.StateObserved,
			EvidenceID:       &evidenceID,
			CreatedAt:        now,
		}},
	}
}

func stringPtr(value string) *string {
	return &value
}
