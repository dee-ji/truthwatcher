package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/truthwatcher/truthwatcher/internal/apihttp"
	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/queue"
	"github.com/truthwatcher/truthwatcher/internal/reconcile"
	"github.com/truthwatcher/truthwatcher/internal/topology"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	queueBackend := queue.Queue(queue.NewInMemoryQueue())
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		redisQ := queue.NewRedisQueue(redisAddr, "truthwatcher")
		if err := redisQ.Ping(); err == nil {
			queueBackend = redisQ
			logger.Info("tw-server using redis queue", "addr", redisAddr)
		}
	}

	intentSvc := intent.NewInMemoryService()
	auditSvc := audit.NewStubService()
	deploySvc := deploy.NewStubServiceWithDependencies(auditSvc, intentSvc).WithQueue(queueBackend)
	srv := apihttp.New(logger, intentSvc, topology.NewStubService(), deploySvc, reconcile.NewStubService(), auditSvc)
	if err := srv.Run(context.Background(), envOr("API_ADDR", ":8080")); err != nil {
		logger.Error("tw-server exited", "error", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
