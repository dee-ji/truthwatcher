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

	if component == "radiant" {
		return runRadiant(ctx)
	}

	// TODO: Wire component-specific dependency graphs and runtime lifecycle.
	fmt.Printf("%s starting (bootstrap placeholder)\n", component)
	return nil
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
