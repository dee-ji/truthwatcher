package radiant

import (
	"context"
	"fmt"
	"log/slog"
)

// Dependencies captures external systems required by the Radiant control plane.
type Dependencies struct {
	Archive    Module
	Ideals     Module
	Elsecall   Module
	Shadesmar  Module
	Oathgate   Module
	Highstorm  Module
	Stormlight Module
	Seekers    Module
}

// Module is a minimal initialization contract for Radiant-managed subsystems.
type Module interface {
	Initialize(ctx context.Context) error
}

// Service represents the Radiant control plane runtime.
type Service struct {
	logger *slog.Logger
	deps   Dependencies
}

// NewService constructs a Radiant service with dependency wiring.
func NewService(logger *slog.Logger, deps Dependencies) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		logger: logger,
		deps:   deps,
	}
}

// Start boots Radiant and initializes configured modules.
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("radiant starting")

	for _, item := range s.modules() {
		if item.module == nil {
			continue
		}

		s.logger.Info("initializing module", "module", item.name)
		if err := item.module.Initialize(ctx); err != nil {
			return fmt.Errorf("initialize module %s: %w", item.name, err)
		}
	}

	s.logger.Info("radiant startup complete")
	return nil
}

type moduleEntry struct {
	name   string
	module Module
}

func (s *Service) modules() []moduleEntry {
	return []moduleEntry{
		{name: "archive", module: s.deps.Archive},
		{name: "ideals", module: s.deps.Ideals},
		{name: "elsecall", module: s.deps.Elsecall},
		{name: "shadesmar", module: s.deps.Shadesmar},
		{name: "oathgate", module: s.deps.Oathgate},
		{name: "highstorm", module: s.deps.Highstorm},
		{name: "stormlight", module: s.deps.Stormlight},
		{name: "seekers", module: s.deps.Seekers},
	}
}

// NopModule is a placeholder implementation used during early scaffolding.
type NopModule struct {
	Name string
}

// Initialize satisfies the Module interface.
func (m NopModule) Initialize(context.Context) error {
	return nil
}
