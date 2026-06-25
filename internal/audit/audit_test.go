package audit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestServiceCreateRecordValidatesAndRedacts(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

	record, err := service.CreateRecord(context.Background(), CreateRecordParams{
		Action:       "discovery_command",
		Initiator:    "",
		Target:       "router-a password=secret",
		Method:       "ssh",
		CommandOrAPI: "show version token=abc123",
		Status:       StatusCompleted,
		StartedAt:    now,
		Context:      json.RawMessage(`{"request_id":"req-a"}`),
	})
	if err != nil {
		t.Fatalf("CreateRecord returned error: %v", err)
	}
	if record.Initiator != DefaultInitiator {
		t.Fatalf("initiator = %q, want default", record.Initiator)
	}
	if strings.Contains(record.Target, "secret") {
		t.Fatalf("target was not redacted: %q", record.Target)
	}
	if strings.Contains(record.CommandOrAPI, "abc123") {
		t.Fatalf("command was not redacted: %q", record.CommandOrAPI)
	}
}

func TestServiceCreateRecordRequiresCoreFields(t *testing.T) {
	_, err := NewService(&fakeRepository{}).CreateRecord(context.Background(), CreateRecordParams{
		Action:    "discovery_command",
		Method:    "ssh",
		Status:    StatusCompleted,
		StartedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("CreateRecord returned nil error")
	}
	if !strings.Contains(err.Error(), "target") {
		t.Fatalf("error = %q, want target validation", err.Error())
	}
}

type fakeRepository struct {
	records []Record
}

func (f *fakeRepository) ListAuditRecords(ctx context.Context, filters ListRecordsFilters) ([]Record, error) {
	return f.records, nil
}

func (f *fakeRepository) CreateAuditRecord(ctx context.Context, params CreateRecordParams) (Record, error) {
	record := Record{
		ID:             "audit-a",
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
