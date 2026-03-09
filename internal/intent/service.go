package intent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/domain"
)

var ErrNotFound = errors.New("intent not found")

type Service interface {
	Create(context.Context, domain.Intent) (domain.Intent, error)
	List(context.Context) ([]domain.Intent, error)
	Get(context.Context, string) (domain.Intent, error)
	Validate(context.Context, string) error
	Compile(context.Context, string) (string, error)
}

type InMemoryService struct {
	mu      sync.RWMutex
	intents map[string]domain.Intent
	seq     int
}

func NewInMemoryService() *InMemoryService {
	return &InMemoryService{intents: map[string]domain.Intent{}}
}
func (s *InMemoryService) nextID() string { s.seq++; return fmt.Sprintf("intent-%d", s.seq) }
func (s *InMemoryService) Create(_ context.Context, in domain.Intent) (domain.Intent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	in.ID = s.nextID()
	in.CreatedAt = time.Now().UTC()
	if in.Revision == 0 {
		in.Revision = 1
	}
	s.intents[in.ID] = in
	return in, nil
}
func (s *InMemoryService) List(_ context.Context) ([]domain.Intent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Intent, 0, len(s.intents))
	for _, v := range s.intents {
		out = append(out, v)
	}
	return out, nil
}
func (s *InMemoryService) Get(_ context.Context, id string) (domain.Intent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.intents[id]
	if !ok {
		return domain.Intent{}, ErrNotFound
	}
	return v, nil
}
func (s *InMemoryService) Validate(ctx context.Context, id string) error {
	_, err := s.Get(ctx, id)
	return err
}
func (s *InMemoryService) Compile(ctx context.Context, id string) (string, error) {
	_, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return "compile queued", nil
}
