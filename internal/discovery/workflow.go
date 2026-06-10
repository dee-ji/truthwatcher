package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"truthwatcher/internal/evidence"
	"truthwatcher/internal/policy"
)

// EvidenceStore is the persistence boundary needed by discovery execution.
type EvidenceStore interface {
	CreateEvidence(context.Context, evidence.CreateEvidenceParams) (evidence.Evidence, error)
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
	Policy    policy.Engine
}

type StartDiscoveryRunResult struct {
	DiscoveryRun DiscoveryRun        `json:"discovery_run"`
	Evidence     []evidence.Evidence `json:"evidence"`
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

	seedInput, err := workflowSeedInput(params.Seed, params.Profile, params.Tasks)
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
		return result, err
	}

	for _, output := range outputs {
		metadata, err := collectedOutputMetadata(output)
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

func workflowSeedInput(seed DiscoverySeed, profile Profile, tasks []policy.Task) (json.RawMessage, error) {
	payload := struct {
		Target  string        `json:"target"`
		Method  string        `json:"method"`
		Profile string        `json:"profile"`
		Tasks   []policy.Task `json:"tasks"`
	}{
		Target:  seed.Target,
		Method:  seed.Method,
		Profile: profile.Name,
		Tasks:   tasks,
	}
	return json.Marshal(payload)
}

func collectedOutputMetadata(output CollectedOutput) (json.RawMessage, error) {
	metadata := struct {
		Collector   string      `json:"collector"`
		Task        policy.Task `json:"task"`
		Profile     string      `json:"profile"`
		Platform    string      `json:"platform"`
		Vendor      string      `json:"vendor"`
		ParserHints []string    `json:"parser_hints"`
	}{
		Collector:   output.Method,
		Task:        output.Task,
		Profile:     output.ProfileName,
		Platform:    output.Platform,
		Vendor:      output.Vendor,
		ParserHints: output.ParserHints,
	}
	return json.Marshal(metadata)
}
