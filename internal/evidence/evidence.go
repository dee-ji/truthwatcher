package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("evidence not found")

type Evidence struct {
	ID             string          `json:"id"`
	DiscoveryRunID string          `json:"discovery_run_id"`
	Target         string          `json:"target"`
	Method         string          `json:"method"`
	CommandOrAPI   string          `json:"command_or_api"`
	RawOutput      string          `json:"raw_output"`
	RawOutputHash  string          `json:"raw_output_hash"`
	ParserName     *string         `json:"parser_name,omitempty"`
	CollectedAt    time.Time       `json:"collected_at"`
	Metadata       json.RawMessage `json:"metadata"`
}

type CreateEvidenceParams struct {
	DiscoveryRunID string
	Target         string
	Method         string
	CommandOrAPI   string
	RawOutput      string
	ParserName     *string
	Metadata       json.RawMessage
}

type Repository interface {
	CreateEvidence(context.Context, CreateEvidenceParams) (Evidence, error)
	GetEvidence(context.Context, string) (Evidence, error)
	ListEvidenceByDiscoveryRun(context.Context, string) ([]Evidence, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) CreateEvidence(ctx context.Context, params CreateEvidenceParams) (Evidence, error) {
	if s.repo == nil {
		return Evidence{}, fmt.Errorf("evidence repository is required")
	}
	if strings.TrimSpace(params.DiscoveryRunID) == "" {
		return Evidence{}, fmt.Errorf("discovery_run_id is required")
	}
	if strings.TrimSpace(params.Target) == "" {
		return Evidence{}, fmt.Errorf("target is required")
	}
	if strings.TrimSpace(params.Method) == "" {
		return Evidence{}, fmt.Errorf("method is required")
	}
	if strings.TrimSpace(params.CommandOrAPI) == "" {
		return Evidence{}, fmt.Errorf("command_or_api is required")
	}

	params.Target = strings.TrimSpace(params.Target)
	params.Method = strings.TrimSpace(params.Method)
	params.CommandOrAPI = strings.TrimSpace(params.CommandOrAPI)

	metadata := strings.TrimSpace(string(params.Metadata))
	if metadata == "" {
		params.Metadata = json.RawMessage(`{}`)
	} else if !json.Valid(params.Metadata) {
		return Evidence{}, fmt.Errorf("metadata must be valid JSON")
	}

	// TODO: add policy-driven redaction hooks before collectors write evidence.
	// The current evidence store intentionally preserves raw output unchanged.
	return s.repo.CreateEvidence(ctx, params)
}

func (s Service) GetEvidence(ctx context.Context, id string) (Evidence, error) {
	if s.repo == nil {
		return Evidence{}, fmt.Errorf("evidence repository is required")
	}
	if strings.TrimSpace(id) == "" {
		return Evidence{}, fmt.Errorf("evidence id is required")
	}
	return s.repo.GetEvidence(ctx, id)
}

func (s Service) ListEvidenceByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]Evidence, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("evidence repository is required")
	}
	if strings.TrimSpace(discoveryRunID) == "" {
		return nil, fmt.Errorf("discovery run id is required")
	}
	return s.repo.ListEvidenceByDiscoveryRun(ctx, discoveryRunID)
}

func HashRawOutput(rawOutput string) string {
	sum := sha256.Sum256([]byte(rawOutput))
	return hex.EncodeToString(sum[:])
}
