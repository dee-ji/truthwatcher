package audit

import (
	"context"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

type Service interface {
	List(context.Context) ([]domain.AuditEvent, error)
}

type StubService struct{}

func NewStubService() *StubService { return &StubService{} }
func (s *StubService) List(context.Context) ([]domain.AuditEvent, error) {
	return []domain.AuditEvent{{ID: "evt-1", Actor: "system", Action: "bootstrap", CreatedAt: time.Now().UTC()}}, nil
}
