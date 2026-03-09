package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/truthwatcher/truthwatcher/internal/radiant"
)

// Run provides a shared entrypoint for Truthwatcher command binaries.
func Run(ctx context.Context, component string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	switch component {
	case "radiant":
		return runRadiant(ctx)
	case "highstorm", "stormlight", "seekers":
		// TODO: Implement component-specific runtime graphs and dependency injection.
		fmt.Printf("%s starting (bootstrap placeholder)\n", component)
		return nil
	default:
		return fmt.Errorf("unknown component %q", component)
	}
}

func runRadiant(ctx context.Context) error {
	service := radiant.NewService(slog.Default(), radiant.Dependencies{
		Archive:    radiant.NopModule{Name: "archive"},
		Ideals:     radiant.NopModule{Name: "ideals"},
		Elsecall:   radiant.NopModule{Name: "elsecall"},
		Shadesmar:  radiant.NopModule{Name: "shadesmar"},
		Oathgate:   radiant.NopModule{Name: "oathgate"},
		Highstorm:  radiant.NopModule{Name: "highstorm"},
		Stormlight: radiant.NopModule{Name: "stormlight"},
		Seekers:    radiant.NopModule{Name: "seekers"},
	})

	return service.Start(ctx)
}
