package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

type Service interface {
	Create(context.Context, string, string) (domain.Deployment, error)
	Get(context.Context, string) (domain.Deployment, error)
}

type StubService struct {
	deployments map[string]domain.Deployment
	seq         int
}

func NewStubService() *StubService { return &StubService{deployments: map[string]domain.Deployment{}} }
func (s *StubService) Create(_ context.Context, intentID, key string) (domain.Deployment, error) {
	s.seq++
	id := fmt.Sprintf("dep-%d", s.seq)
	d := domain.Deployment{ID: id, IntentID: intentID, Status: "planned", IdempotencyKey: key, CreatedAt: time.Now().UTC()}
	s.deployments[id] = d
	return d, nil
}
func (s *StubService) Get(_ context.Context, id string) (domain.Deployment, error) {
	return s.deployments[id], nil
}
