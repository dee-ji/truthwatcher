package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/truthwatcher/truthwatcher/internal/apihttp"
	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/reconcile"
	"github.com/truthwatcher/truthwatcher/internal/topology"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	srv := apihttp.New(
		logger,
		intent.NewInMemoryService(),
		topology.NewStubService(),
		deploy.NewStubService(),
		reconcile.NewStubService(),
		audit.NewStubService(),
	)

	// TODO: Read listen address and backing service implementations from config.
	if err := srv.Run(context.Background(), ":8080"); err != nil {
		logger.Error("spanreed exited", "error", err)
	}
}
