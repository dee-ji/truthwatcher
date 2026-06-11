package planner

import (
	"context"
	"encoding/json"
	"strings"
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

func TestCreatePlanUsesArchitectureHints(t *testing.T) {
	service := NewService(Options{Assets: graphWithArchitectureHints(), Policy: policy.NewEngine()})

	plan, err := service.CreatePlan(context.Background(), Request{
		Target: "rr1.example.net",
		Method: "ssh",
	})
	if err != nil {
		t.Fatalf("CreatePlan returned error: %v", err)
	}

	if len(plan.Steps) == 0 {
		t.Fatal("plan has no steps")
	}
	if !containsTask(plan.Steps, policy.TaskGetBGPSummary) {
		t.Fatalf("steps = %#v, want BGP summary for seeded route reflector", plan.Steps)
	}
	for _, step := range plan.Steps {
		if step.Profile != discovery.ProfileJuniperJunos {
			t.Fatalf("profile = %q, want inferred Junos profile", step.Profile)
		}
	}
	if !containsWarning(plan.Warnings, "architecture seed hints are user-provided context") {
		t.Fatalf("warnings = %#v, want seeded context warning", plan.Warnings)
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

func containsWarning(warnings []string, text string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, text) {
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

func graphWithArchitectureHints() fakeAssets {
	graph := testGraph()
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	graph.assets = append(graph.assets, assets.Asset{
		ID:               "asset-seed",
		Type:             "architecture_context",
		IdentityKey:      "architecture:seed:default",
		Confidence:       0.45,
		ConfidenceReason: "provided by user seed input; useful planning context but not proof",
		State:            assets.StateUserSeeded,
		Metadata:         json.RawMessage(`{}`),
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	graph.facts = append(graph.facts,
		assets.Fact{
			ID:               "fact-vendors",
			AssetID:          "asset-seed",
			Name:             "known_vendors",
			Value:            json.RawMessage(`["juniper"]`),
			Source:           "user_seeded",
			Confidence:       0.45,
			ConfidenceReason: "provided by user seed input; useful planning context but not proof",
			State:            assets.StateUserSeeded,
			CreatedAt:        now,
		},
		assets.Fact{
			ID:               "fact-rr",
			AssetID:          "asset-seed",
			Name:             "known_route_reflectors",
			Value:            json.RawMessage(`["rr1.example.net"]`),
			Source:           "user_seeded",
			Confidence:       0.45,
			ConfidenceReason: "provided by user seed input; useful planning context but not proof",
			State:            assets.StateUserSeeded,
			CreatedAt:        now,
		},
	)
	return graph
}

func stringPtr(value string) *string {
	return &value
}
