package reconcile

import (
	"context"
	"fmt"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

type Service interface {
	CreateRun(context.Context) (domain.ReconcileRun, error)
}
type StubService struct{ seq int }

func NewStubService() *StubService { return &StubService{} }
func (s *StubService) CreateRun(context.Context) (domain.ReconcileRun, error) {
	s.seq++
	return domain.ReconcileRun{ID: fmt.Sprintf("recon-%d", s.seq), Status: "queued", CreatedAt: time.Now().UTC()}, nil
}
