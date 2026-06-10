package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"truthwatcher/internal/evidence"
	"truthwatcher/internal/policy"
)

func TestStartDiscoveryRunStoresEvidenceAndCompletes(t *testing.T) {
	repo := newWorkflowRunRepository()
	evidenceStore := &workflowEvidenceStore{}
	profile, ok := BuiltInProfile(ProfileJuniperJunos)
	if !ok {
		t.Fatal("expected Junos profile")
	}

	result, err := NewService(repo).StartDiscoveryRun(context.Background(), StartDiscoveryRunParams{
		Seed: DiscoverySeed{
			Target: "fixture://junos-mx",
			Method: FakeMethod,
		},
		Profile:   profile,
		Tasks:     []policy.Task{policy.TaskIdentifyDevice},
		Collector: NewFakeCollector("../../examples/fixtures", policy.NewEngine()),
		Evidence:  evidenceStore,
		Policy:    policy.NewEngine(),
	})
	if err != nil {
		t.Fatalf("StartDiscoveryRun returned error: %v", err)
	}
	if result.DiscoveryRun.Status != StatusCompleted {
		t.Fatalf("status = %q, want completed", result.DiscoveryRun.Status)
	}
	if got, want := len(result.Evidence), 1; got != want {
		t.Fatalf("evidence count = %d, want %d", got, want)
	}
	if result.Evidence[0].CommandOrAPI != "show version" {
		t.Fatalf("command = %q, want show version", result.Evidence[0].CommandOrAPI)
	}
	if evidenceStore.items[0].DiscoveryRunID != result.DiscoveryRun.ID {
		t.Fatalf("evidence discovery_run_id = %q, want %q", evidenceStore.items[0].DiscoveryRunID, result.DiscoveryRun.ID)
	}
}

func TestStartDiscoveryRunMarksFailedWhenCollectorFails(t *testing.T) {
	repo := newWorkflowRunRepository()
	evidenceStore := &workflowEvidenceStore{}
	profile, ok := BuiltInProfile(ProfileJuniperJunos)
	if !ok {
		t.Fatal("expected Junos profile")
	}

	result, err := NewService(repo).StartDiscoveryRun(context.Background(), StartDiscoveryRunParams{
		Seed: DiscoverySeed{
			Target: "fixture://missing",
			Method: FakeMethod,
		},
		Profile:   profile,
		Tasks:     []policy.Task{policy.TaskIdentifyDevice},
		Collector: NewFakeCollector("../../examples/fixtures", policy.NewEngine()),
		Evidence:  evidenceStore,
		Policy:    policy.NewEngine(),
	})
	if err == nil {
		t.Fatal("StartDiscoveryRun returned nil error")
	}
	if result.DiscoveryRun.Status != StatusFailed {
		t.Fatalf("status = %q, want failed", result.DiscoveryRun.Status)
	}
	if result.DiscoveryRun.ErrorMessage == nil || *result.DiscoveryRun.ErrorMessage == "" {
		t.Fatal("expected failed run error message")
	}
	if len(result.Evidence) != 0 {
		t.Fatalf("evidence count = %d, want 0", len(result.Evidence))
	}
}

func newWorkflowRunRepository() *workflowRunRepository {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	return &workflowRunRepository{now: now}
}

type workflowRunRepository struct {
	now  time.Time
	runs []DiscoveryRun
}

func (f *workflowRunRepository) CreateDiscoveryRun(ctx context.Context, params CreateDiscoveryRunParams) (DiscoveryRun, error) {
	run := DiscoveryRun{
		ID:        "11111111-1111-4111-8111-111111111111",
		Status:    StatusPending,
		SeedInput: append(json.RawMessage(nil), params.SeedInput...),
		StartedAt: f.now,
		CreatedAt: f.now,
		UpdatedAt: f.now,
	}
	f.runs = append(f.runs, run)
	return run, nil
}

func (f *workflowRunRepository) GetDiscoveryRun(ctx context.Context, id string) (DiscoveryRun, error) {
	for _, run := range f.runs {
		if run.ID == id {
			return run, nil
		}
	}
	return DiscoveryRun{}, ErrNotFound
}

func (f *workflowRunRepository) ListDiscoveryRuns(ctx context.Context) ([]DiscoveryRun, error) {
	return append([]DiscoveryRun(nil), f.runs...), nil
}

func (f *workflowRunRepository) UpdateDiscoveryRunStatus(ctx context.Context, params UpdateDiscoveryRunStatusParams) (DiscoveryRun, error) {
	for i := range f.runs {
		if f.runs[i].ID == params.ID {
			f.runs[i].Status = params.Status
			f.runs[i].CompletedAt = params.CompletedAt
			f.runs[i].ErrorMessage = params.ErrorMessage
			f.runs[i].UpdatedAt = f.now.Add(time.Second)
			return f.runs[i], nil
		}
	}
	return DiscoveryRun{}, ErrNotFound
}

type workflowEvidenceStore struct {
	items []evidence.Evidence
}

func (f *workflowEvidenceStore) CreateEvidence(ctx context.Context, params evidence.CreateEvidenceParams) (evidence.Evidence, error) {
	if params.CommandOrAPI == "fail" {
		return evidence.Evidence{}, errors.New("store evidence")
	}
	item := evidence.Evidence{
		ID:             "22222222-2222-4222-8222-222222222222",
		DiscoveryRunID: params.DiscoveryRunID,
		Target:         params.Target,
		Method:         params.Method,
		CommandOrAPI:   params.CommandOrAPI,
		RawOutput:      params.RawOutput,
		RawOutputHash:  evidence.HashRawOutput(params.RawOutput),
		CollectedAt:    time.Date(2026, 6, 10, 12, 1, 0, 0, time.UTC),
		Metadata:       append(json.RawMessage(nil), params.Metadata...),
	}
	f.items = append(f.items, item)
	return item, nil
}
