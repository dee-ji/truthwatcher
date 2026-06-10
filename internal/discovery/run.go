package discovery

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("discovery run not found")

type RunStatus string

const (
	StatusPending   RunStatus = "pending"
	StatusRunning   RunStatus = "running"
	StatusCompleted RunStatus = "completed"
	StatusFailed    RunStatus = "failed"
	StatusCanceled  RunStatus = "canceled"
)

type DiscoveryRun struct {
	ID           string          `json:"id"`
	Status       RunStatus       `json:"status"`
	SeedInput    json.RawMessage `json:"seed_input"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type CreateDiscoveryRunParams struct {
	SeedInput json.RawMessage
}

type UpdateDiscoveryRunStatusParams struct {
	ID           string
	Status       RunStatus
	CompletedAt  *time.Time
	ErrorMessage *string
}

type Repository interface {
	CreateDiscoveryRun(context.Context, CreateDiscoveryRunParams) (DiscoveryRun, error)
	GetDiscoveryRun(context.Context, string) (DiscoveryRun, error)
	ListDiscoveryRuns(context.Context) ([]DiscoveryRun, error)
	UpdateDiscoveryRunStatus(context.Context, UpdateDiscoveryRunStatusParams) (DiscoveryRun, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) CreateDiscoveryRun(ctx context.Context, params CreateDiscoveryRunParams) (DiscoveryRun, error) {
	if s.repo == nil {
		return DiscoveryRun{}, fmt.Errorf("discovery repository is required")
	}

	seedInput := strings.TrimSpace(string(params.SeedInput))
	if seedInput == "" {
		params.SeedInput = json.RawMessage(`{}`)
	} else if !json.Valid(params.SeedInput) {
		return DiscoveryRun{}, fmt.Errorf("seed_input must be valid JSON")
	}

	return s.repo.CreateDiscoveryRun(ctx, params)
}

func (s Service) GetDiscoveryRun(ctx context.Context, id string) (DiscoveryRun, error) {
	if s.repo == nil {
		return DiscoveryRun{}, fmt.Errorf("discovery repository is required")
	}
	if strings.TrimSpace(id) == "" {
		return DiscoveryRun{}, fmt.Errorf("discovery run id is required")
	}
	return s.repo.GetDiscoveryRun(ctx, id)
}

func (s Service) ListDiscoveryRuns(ctx context.Context) ([]DiscoveryRun, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("discovery repository is required")
	}
	return s.repo.ListDiscoveryRuns(ctx)
}

func (s Service) UpdateDiscoveryRunStatus(ctx context.Context, params UpdateDiscoveryRunStatusParams) (DiscoveryRun, error) {
	if s.repo == nil {
		return DiscoveryRun{}, fmt.Errorf("discovery repository is required")
	}
	if strings.TrimSpace(params.ID) == "" {
		return DiscoveryRun{}, fmt.Errorf("discovery run id is required")
	}
	if !params.Status.Valid() {
		return DiscoveryRun{}, fmt.Errorf("invalid discovery run status %q", params.Status)
	}
	return s.repo.UpdateDiscoveryRunStatus(ctx, params)
}

func (s RunStatus) Valid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCanceled:
		return true
	default:
		return false
	}
}

func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	encoded := make([]byte, 32)
	hex.Encode(encoded, b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:32]), nil
}
