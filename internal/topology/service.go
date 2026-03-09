package topology

import (
	"context"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

type Service interface {
	Devices(context.Context) ([]domain.Device, error)
	Links(context.Context) ([]domain.Link, error)
}

type StubService struct{}

func NewStubService() *StubService { return &StubService{} }

func (s *StubService) Devices(context.Context) ([]domain.Device, error) {
	return []domain.Device{{ID: "dev-1", Hostname: "leaf-1", Vendor: "eos"}}, nil
}
func (s *StubService) Links(context.Context) ([]domain.Link, error) {
	return []domain.Link{{ID: "link-1", FromDevice: "leaf-1", ToDevice: "spine-1"}}, nil
}
