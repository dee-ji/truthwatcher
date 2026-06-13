package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"truthwatcher/internal/audit"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/policy"
)

func TestStartDiscoveryRunStoresEvidenceAndCompletes(t *testing.T) {
	repo := newWorkflowRunRepository()
	evidenceStore := &workflowEvidenceStore{}
	auditStore := &workflowAuditStore{}
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
		Audit:     auditStore,
		Policy:    policy.NewEngine(),
		Initiator: "unit-test",
		RequestID: "req-test",
		Context:   json.RawMessage(`{"path":"/test"}`),
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
	if got, want := len(result.Audit), 1; got != want {
		t.Fatalf("audit record count = %d, want %d", got, want)
	}
	auditRecord := result.Audit[0]
	if auditRecord.ID == "" {
		t.Fatalf("audit id is empty: %#v", auditRecord)
	}
	if auditRecord.Action != "discovery_command" {
		t.Fatalf("audit action = %q, want discovery_command", auditRecord.Action)
	}
	if auditRecord.Initiator != "unit-test" {
		t.Fatalf("audit initiator = %q, want unit-test", auditRecord.Initiator)
	}
	if auditRecord.RequestID != "req-test" {
		t.Fatalf("audit request_id = %q, want req-test", auditRecord.RequestID)
	}
	if auditRecord.Target != "fixture://junos-mx" {
		t.Fatalf("audit target = %q, want fixture://junos-mx", auditRecord.Target)
	}
	if auditRecord.Profile != ProfileJuniperJunos {
		t.Fatalf("audit profile = %q, want %q", auditRecord.Profile, ProfileJuniperJunos)
	}
	if auditRecord.Task != string(policy.TaskIdentifyDevice) {
		t.Fatalf("audit task = %q, want identify_device", auditRecord.Task)
	}
	if auditRecord.CommandOrAPI != "show version" {
		t.Fatalf("audit command = %q, want show version", auditRecord.CommandOrAPI)
	}
	if auditRecord.EvidenceID != result.Evidence[0].ID {
		t.Fatalf("audit evidence_id = %q, want %q", auditRecord.EvidenceID, result.Evidence[0].ID)
	}
	if auditRecord.StartedAt.IsZero() || auditRecord.CompletedAt.IsZero() {
		t.Fatalf("audit timestamps are not populated: %#v", auditRecord)
	}
	if got, want := len(auditStore.records), 2; got != want {
		t.Fatalf("persisted audit record count = %d, want command plus run record", got)
	}
	if auditStore.records[1].Action != "discovery_run_execute" || auditStore.records[1].Status != "completed" {
		t.Fatalf("run audit = %#v, want completed discovery_run_execute", auditStore.records[1])
	}

	var seedInput struct {
		Audit struct {
			Initiator string `json:"initiator"`
			RequestID string `json:"request_id"`
		} `json:"audit"`
	}
	if err := json.Unmarshal(result.DiscoveryRun.SeedInput, &seedInput); err != nil {
		t.Fatalf("decode seed input: %v", err)
	}
	if seedInput.Audit.Initiator != "unit-test" || seedInput.Audit.RequestID != "req-test" {
		t.Fatalf("seed audit = %#v, want initiator/request", seedInput.Audit)
	}

	var metadata struct {
		Audit struct {
			Initiator string `json:"initiator"`
			Target    string `json:"target"`
			Profile   string `json:"profile"`
		} `json:"audit"`
	}
	if err := json.Unmarshal(evidenceStore.items[0].Metadata, &metadata); err != nil {
		t.Fatalf("decode evidence metadata: %v", err)
	}
	if metadata.Audit.Initiator != "unit-test" || metadata.Audit.Target != "fixture://junos-mx" || metadata.Audit.Profile != ProfileJuniperJunos {
		t.Fatalf("evidence audit metadata = %#v, want initiator/target/profile", metadata.Audit)
	}
}

func TestStartDiscoveryRunMarksFailedWhenCollectorFails(t *testing.T) {
	repo := newWorkflowRunRepository()
	evidenceStore := &workflowEvidenceStore{}
	auditStore := &workflowAuditStore{}
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
		Audit:     auditStore,
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
	if got, want := len(auditStore.records), 1; got != want {
		t.Fatalf("persisted audit record count = %d, want failed run audit", got)
	}
	if auditStore.records[0].Action != "discovery_run_execute" || auditStore.records[0].Status != "failed" {
		t.Fatalf("failed run audit = %#v, want failed discovery_run_execute", auditStore.records[0])
	}
	if auditStore.records[0].ErrorMessage == "" {
		t.Fatalf("failed run audit error is empty: %#v", auditStore.records[0])
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

type workflowAuditStore struct {
	records []audit.Record
}

func (f *workflowAuditStore) CreateRecord(ctx context.Context, params audit.CreateRecordParams) (audit.Record, error) {
	record := audit.Record{
		ID:             fmt.Sprintf("audit-%d", len(f.records)+1),
		Action:         params.Action,
		Initiator:      params.Initiator,
		RequestID:      params.RequestID,
		DiscoveryRunID: params.DiscoveryRunID,
		Target:         params.Target,
		Method:         params.Method,
		Profile:        params.Profile,
		Task:           params.Task,
		CommandOrAPI:   params.CommandOrAPI,
		Status:         params.Status,
		EvidenceID:     params.EvidenceID,
		ErrorMessage:   params.ErrorMessage,
		StartedAt:      params.StartedAt,
		CompletedAt:    params.CompletedAt,
		Context:        params.Context,
	}
	f.records = append(f.records, record)
	return record, nil
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
