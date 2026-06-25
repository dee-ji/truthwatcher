package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const DefaultInitiator = "unknown"

const (
	StatusStarted   = "started"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusStored    = "stored"
)

type Record struct {
	ID             string          `json:"id,omitempty"`
	Action         string          `json:"action"`
	Initiator      string          `json:"initiator"`
	RequestID      string          `json:"request_id,omitempty"`
	DiscoveryRunID string          `json:"discovery_run_id"`
	Target         string          `json:"target"`
	Method         string          `json:"method"`
	Profile        string          `json:"profile"`
	Task           string          `json:"task,omitempty"`
	CommandOrAPI   string          `json:"command_or_api,omitempty"`
	Status         string          `json:"status"`
	EvidenceID     string          `json:"evidence_id,omitempty"`
	ErrorMessage   string          `json:"error,omitempty"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    time.Time       `json:"completed_at"`
	Context        json.RawMessage `json:"context,omitempty"`
}

type CreateRecordParams struct {
	Action         string
	Initiator      string
	RequestID      string
	DiscoveryRunID string
	Target         string
	Method         string
	Profile        string
	Task           string
	CommandOrAPI   string
	Status         string
	EvidenceID     string
	ErrorMessage   string
	StartedAt      time.Time
	CompletedAt    time.Time
	Context        json.RawMessage
}

type ListRecordsFilters struct {
	DiscoveryRunID string `json:"discovery_run_id,omitempty"`
	EvidenceID     string `json:"evidence_id,omitempty"`
	RequestID      string `json:"request_id,omitempty"`
	Action         string `json:"action,omitempty"`
	Status         string `json:"status,omitempty"`
	Target         string `json:"target,omitempty"`
	Method         string `json:"method,omitempty"`
	Profile        string `json:"profile,omitempty"`
	Limit          int    `json:"limit"`
}

type Repository interface {
	CreateAuditRecord(context.Context, CreateRecordParams) (Record, error)
	ListAuditRecords(context.Context, ListRecordsFilters) ([]Record, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) CreateRecord(ctx context.Context, params CreateRecordParams) (Record, error) {
	if s.repo == nil {
		return Record{}, fmt.Errorf("audit repository is required")
	}
	if strings.TrimSpace(params.Action) == "" {
		return Record{}, fmt.Errorf("audit action is required")
	}
	if strings.TrimSpace(params.Target) == "" {
		return Record{}, fmt.Errorf("audit target is required")
	}
	if strings.TrimSpace(params.Method) == "" {
		return Record{}, fmt.Errorf("audit method is required")
	}
	if strings.TrimSpace(params.Status) == "" {
		return Record{}, fmt.Errorf("audit status is required")
	}
	if params.StartedAt.IsZero() {
		return Record{}, fmt.Errorf("audit started_at is required")
	}
	if params.CompletedAt.IsZero() {
		params.CompletedAt = params.StartedAt
	}
	params.Action = strings.TrimSpace(params.Action)
	params.Initiator = NormalizeInitiator(params.Initiator)
	params.RequestID = strings.TrimSpace(params.RequestID)
	params.DiscoveryRunID = strings.TrimSpace(params.DiscoveryRunID)
	params.Target = RedactSensitiveText(strings.TrimSpace(params.Target))
	params.Method = strings.TrimSpace(params.Method)
	params.Profile = strings.TrimSpace(params.Profile)
	params.Task = strings.TrimSpace(params.Task)
	params.CommandOrAPI = RedactSensitiveText(strings.TrimSpace(params.CommandOrAPI))
	params.Status = strings.TrimSpace(params.Status)
	params.EvidenceID = strings.TrimSpace(params.EvidenceID)
	params.ErrorMessage = RedactSensitiveText(strings.TrimSpace(params.ErrorMessage))
	params.Context = normalizeContext(params.Context)
	return s.repo.CreateAuditRecord(ctx, params)
}

// DiscoveryAction records the minimum safety-relevant audit fields for one
// discovery command/API action. It is kept as an alias so existing discovery API
// metadata remains stable while audit records become persistable.
type DiscoveryAction = Record

func NormalizeInitiator(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultInitiator
	}
	return value
}

func normalizeContext(value json.RawMessage) json.RawMessage {
	if strings.TrimSpace(string(value)) == "" {
		return nil
	}
	if !json.Valid(value) {
		return nil
	}
	return value
}

func (s Service) ListRecords(ctx context.Context, filters ListRecordsFilters) ([]Record, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("audit repository is required")
	}
	filters.DiscoveryRunID = strings.TrimSpace(filters.DiscoveryRunID)
	filters.EvidenceID = strings.TrimSpace(filters.EvidenceID)
	filters.RequestID = strings.TrimSpace(filters.RequestID)
	filters.Action = strings.TrimSpace(filters.Action)
	filters.Status = strings.TrimSpace(filters.Status)
	filters.Target = RedactSensitiveText(strings.TrimSpace(filters.Target))
	filters.Method = strings.TrimSpace(filters.Method)
	filters.Profile = strings.TrimSpace(filters.Profile)
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 200 {
		filters.Limit = 200
	}
	return s.repo.ListAuditRecords(ctx, filters)
}
