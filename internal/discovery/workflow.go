package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"truthwatcher/internal/audit"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/policy"
)

// EvidenceStore is the persistence boundary needed by discovery execution.
type EvidenceStore interface {
	CreateEvidence(context.Context, evidence.CreateEvidenceParams) (evidence.Evidence, error)
}

type AuditStore interface {
	CreateRecord(context.Context, audit.CreateRecordParams) (audit.Record, error)
}

type DiscoverySeed struct {
	Target string `json:"target"`
	Method string `json:"method"`
}

type StartDiscoveryRunParams struct {
	Seed      DiscoverySeed
	Profile   Profile
	Tasks     []policy.Task
	Collector Collector
	Evidence  EvidenceStore
	Audit     AuditStore
	Policy    policy.Engine
	Initiator string
	RequestID string
	Context   json.RawMessage
}

type StartDiscoveryRunResult struct {
	DiscoveryRun DiscoveryRun            `json:"discovery_run"`
	Evidence     []evidence.Evidence     `json:"evidence"`
	Audit        []audit.DiscoveryAction `json:"audit"`
}

// StartDiscoveryRun executes one evidence-first discovery workflow.
func (s Service) StartDiscoveryRun(ctx context.Context, params StartDiscoveryRunParams) (StartDiscoveryRunResult, error) {
	if s.repo == nil {
		return StartDiscoveryRunResult{}, fmt.Errorf("discovery repository is required")
	}
	if params.Evidence == nil {
		return StartDiscoveryRunResult{}, fmt.Errorf("evidence repository is required")
	}
	if params.Collector == nil {
		return StartDiscoveryRunResult{}, fmt.Errorf("collector is required")
	}
	params.Seed.Target = strings.TrimSpace(params.Seed.Target)
	params.Seed.Method = strings.TrimSpace(params.Seed.Method)
	if params.Seed.Target == "" {
		return StartDiscoveryRunResult{}, fmt.Errorf("seed target is required")
	}
	if params.Seed.Method == "" {
		return StartDiscoveryRunResult{}, fmt.Errorf("seed method is required")
	}
	if len(params.Tasks) == 0 {
		return StartDiscoveryRunResult{}, fmt.Errorf("at least one discovery task is required")
	}
	if err := params.Profile.Validate(params.Policy); err != nil {
		return StartDiscoveryRunResult{}, err
	}
	for _, task := range params.Tasks {
		if err := params.Policy.CheckTask(task); err != nil {
			return StartDiscoveryRunResult{}, err
		}
		if _, err := params.Profile.CommandsForTask(task); err != nil {
			return StartDiscoveryRunResult{}, err
		}
	}

	initiator := audit.NormalizeInitiator(strings.TrimSpace(params.Initiator))
	seedInput, err := workflowSeedInput(params.Seed, params.Profile, params.Tasks, initiator, params.RequestID, params.Context)
	if err != nil {
		return StartDiscoveryRunResult{}, err
	}
	run, err := s.CreateDiscoveryRun(ctx, CreateDiscoveryRunParams{SeedInput: seedInput})
	if err != nil {
		return StartDiscoveryRunResult{}, err
	}
	result := StartDiscoveryRunResult{DiscoveryRun: run}

	run, err = s.UpdateDiscoveryRunStatus(ctx, UpdateDiscoveryRunStatusParams{
		ID:     run.ID,
		Status: StatusRunning,
	})
	if err != nil {
		return result, err
	}
	result.DiscoveryRun = run

	outputs, err := params.Collector.Collect(ctx, params.Seed.Target, params.Profile, params.Tasks)
	if err != nil {
		result.DiscoveryRun = s.markRunFailed(ctx, run.ID, err)
		if auditErr := createRunAuditRecord(ctx, params.Audit, params.Seed, params.Profile, initiator, params.RequestID, params.Context, run.ID, audit.StatusFailed, err, run.StartedAt, time.Now().UTC()); auditErr != nil {
			return result, auditErr
		}
		return result, err
	}

	for _, output := range outputs {
		actionStartedAt := time.Now().UTC()
		metadata, err := collectedOutputMetadata(output, initiator, params.RequestID, params.Context, run.ID, "", "started", actionStartedAt, time.Time{})
		if err != nil {
			result.DiscoveryRun = s.markRunFailed(ctx, run.ID, err)
			return result, err
		}
		item, err := params.Evidence.CreateEvidence(ctx, evidence.CreateEvidenceParams{
			DiscoveryRunID: run.ID,
			Target:         output.Target,
			Method:         output.Method,
			CommandOrAPI:   output.Command,
			RawOutput:      output.RawOutput,
			Metadata:       metadata,
		})
		if err != nil {
			result.DiscoveryRun = s.markRunFailed(ctx, run.ID, err)
			return result, err
		}
		result.Evidence = append(result.Evidence, item)
		actionCompletedAt := time.Now().UTC()
		record := auditRecord(output, initiator, params.RequestID, params.Context, run.ID, item.ID, audit.StatusStored, actionStartedAt, actionCompletedAt)
		persisted, err := persistAuditRecord(ctx, params.Audit, record)
		if err != nil {
			result.DiscoveryRun = s.markRunFailed(ctx, run.ID, err)
			return result, err
		}
		record = persisted
		result.Audit = append(result.Audit, record)
	}

	completedAt := time.Now().UTC()
	run, err = s.UpdateDiscoveryRunStatus(ctx, UpdateDiscoveryRunStatusParams{
		ID:          run.ID,
		Status:      StatusCompleted,
		CompletedAt: &completedAt,
	})
	if err != nil {
		return result, err
	}
	if err := createRunAuditRecord(ctx, params.Audit, params.Seed, params.Profile, initiator, params.RequestID, params.Context, run.ID, audit.StatusCompleted, nil, run.StartedAt, completedAt); err != nil {
		return result, err
	}
	result.DiscoveryRun = run
	return result, nil
}

func (s Service) markRunFailed(ctx context.Context, id string, cause error) DiscoveryRun {
	message := cause.Error()
	completedAt := time.Now().UTC()
	run, err := s.UpdateDiscoveryRunStatus(ctx, UpdateDiscoveryRunStatusParams{
		ID:           id,
		Status:       StatusFailed,
		CompletedAt:  &completedAt,
		ErrorMessage: &message,
	})
	if err != nil {
		return DiscoveryRun{ID: id, Status: StatusFailed, ErrorMessage: &message, CompletedAt: &completedAt}
	}
	return run
}

func workflowSeedInput(seed DiscoverySeed, profile Profile, tasks []policy.Task, initiator string, requestID string, context json.RawMessage) (json.RawMessage, error) {
	payload := struct {
		Target  string          `json:"target"`
		Method  string          `json:"method"`
		Profile string          `json:"profile"`
		Tasks   []policy.Task   `json:"tasks"`
		Audit   seedAudit       `json:"audit"`
		Context json.RawMessage `json:"context,omitempty"`
	}{
		Target:  seed.Target,
		Method:  seed.Method,
		Profile: profile.Name,
		Tasks:   tasks,
		Audit: seedAudit{
			Initiator: initiator,
			RequestID: strings.TrimSpace(requestID),
		},
		Context: defaultRawMessage(context),
	}
	return json.Marshal(payload)
}

type seedAudit struct {
	Initiator string `json:"initiator"`
	RequestID string `json:"request_id,omitempty"`
}

func collectedOutputMetadata(output CollectedOutput, initiator string, requestID string, context json.RawMessage, discoveryRunID string, evidenceID string, status string, startedAt time.Time, completedAt time.Time) (json.RawMessage, error) {
	record := auditRecord(output, initiator, requestID, context, discoveryRunID, evidenceID, status, startedAt, completedAt)
	metadata := struct {
		Collector   string                `json:"collector"`
		Task        policy.Task           `json:"task"`
		Profile     string                `json:"profile"`
		Platform    string                `json:"platform"`
		Vendor      string                `json:"vendor"`
		ParserHints []string              `json:"parser_hints"`
		Audit       audit.DiscoveryAction `json:"audit"`
	}{
		Collector:   output.Method,
		Task:        output.Task,
		Profile:     output.ProfileName,
		Platform:    output.Platform,
		Vendor:      output.Vendor,
		ParserHints: output.ParserHints,
		Audit:       record,
	}
	return json.Marshal(metadata)
}

func auditRecord(output CollectedOutput, initiator string, requestID string, context json.RawMessage, discoveryRunID string, evidenceID string, status string, startedAt time.Time, completedAt time.Time) audit.DiscoveryAction {
	return audit.DiscoveryAction{
		Action:         "discovery_command",
		Initiator:      audit.NormalizeInitiator(strings.TrimSpace(initiator)),
		RequestID:      strings.TrimSpace(requestID),
		DiscoveryRunID: discoveryRunID,
		Target:         output.Target,
		Method:         output.Method,
		Profile:        output.ProfileName,
		Task:           string(output.Task),
		CommandOrAPI:   output.Command,
		Status:         status,
		EvidenceID:     evidenceID,
		StartedAt:      startedAt,
		CompletedAt:    completedAt,
		Context:        defaultRawMessage(context),
	}
}

func persistAuditRecord(ctx context.Context, store AuditStore, record audit.DiscoveryAction) (audit.DiscoveryAction, error) {
	if store == nil {
		return record, nil
	}
	persisted, err := store.CreateRecord(ctx, audit.CreateRecordParams{
		Action:         record.Action,
		Initiator:      record.Initiator,
		RequestID:      record.RequestID,
		DiscoveryRunID: record.DiscoveryRunID,
		Target:         record.Target,
		Method:         record.Method,
		Profile:        record.Profile,
		Task:           record.Task,
		CommandOrAPI:   record.CommandOrAPI,
		Status:         record.Status,
		EvidenceID:     record.EvidenceID,
		ErrorMessage:   record.ErrorMessage,
		StartedAt:      record.StartedAt,
		CompletedAt:    record.CompletedAt,
		Context:        record.Context,
	})
	if err != nil {
		return audit.DiscoveryAction{}, err
	}
	return persisted, nil
}

func createRunAuditRecord(ctx context.Context, store AuditStore, seed DiscoverySeed, profile Profile, initiator string, requestID string, context json.RawMessage, discoveryRunID string, status string, cause error, startedAt time.Time, completedAt time.Time) error {
	if store == nil {
		return nil
	}
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	_, err := store.CreateRecord(ctx, audit.CreateRecordParams{
		Action:         "discovery_run_execute",
		Initiator:      initiator,
		RequestID:      requestID,
		DiscoveryRunID: discoveryRunID,
		Target:         seed.Target,
		Method:         seed.Method,
		Profile:        profile.Name,
		Status:         status,
		ErrorMessage:   message,
		StartedAt:      startedAt,
		CompletedAt:    completedAt,
		Context:        context,
	})
	return err
}

func defaultRawMessage(value json.RawMessage) json.RawMessage {
	if strings.TrimSpace(string(value)) == "" {
		return nil
	}
	return value
}
