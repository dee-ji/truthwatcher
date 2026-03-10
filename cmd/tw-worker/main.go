package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/queue"
	"github.com/truthwatcher/truthwatcher/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	q := queue.Queue(queue.NewInMemoryQueue())
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		redisQ := queue.NewRedisQueue(redisAddr, "truthwatcher")
		if err := redisQ.Ping(); err == nil {
			q = redisQ
			logger.Info("tw-worker using redis queue", "addr", redisAddr)
		}
	}

	deploySvc := deploy.NewStubServiceWithDependencies(audit.NewStubService(), intent.NewInMemoryService())
	svc := worker.New(logger, q, deploySvc)
	if err := svc.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("worker exited", "error", err)
	}
}
