package deploy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

var ErrDeploymentNotFound = errors.New("deployment not found")

type Service interface {
	Create(context.Context, string, string) (domain.Deployment, error)
	Get(context.Context, string) (domain.Deployment, error)
}

type StubService struct {
	mu          sync.RWMutex
	deployments map[string]domain.Deployment
	seq         int
}

func NewStubService() *StubService { return &StubService{deployments: map[string]domain.Deployment{}} }
func (s *StubService) Create(_ context.Context, intentID, key string) (domain.Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	id := fmt.Sprintf("dep-%d", s.seq)
	d := domain.Deployment{ID: id, IntentID: intentID, Status: "planned", IdempotencyKey: key, CreatedAt: time.Now().UTC()}
	s.deployments[id] = d
	return d, nil
}
func (s *StubService) Get(_ context.Context, id string) (domain.Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.deployments[id]
	if !ok {
		// TODO: Replace in-memory lookup with Archive-backed deployment repository.
		return domain.Deployment{}, ErrDeploymentNotFound
	}
	return d, nil
}
