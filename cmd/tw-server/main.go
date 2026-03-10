package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/truthwatcher/truthwatcher/internal/apihttp"
	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/authn"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/queue"
	"github.com/truthwatcher/truthwatcher/internal/rbac"
	"github.com/truthwatcher/truthwatcher/internal/reconcile"
	"github.com/truthwatcher/truthwatcher/internal/state"
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
	stateSvc := state.NewService(state.NewInMemoryRepository())
	reconcileSvc := reconcile.NewService(reconcile.NewInMemoryRepository(), intentSvc, stateSvc, auditSvc)
	authConfig := authn.Config{
		Mode:           authn.ModeJWT,
		LocalDevBypass: envOr("AUTH_LOCAL_DEV_BYPASS", "true") == "true",
		BypassSubject:  envOr("AUTH_BYPASS_SUBJECT", "local-dev"),
		BypassRoles:    []string{envOr("AUTH_BYPASS_ROLE", "admin")},
	}
	if err := authConfig.Validate(); err != nil {
		logger.Error("invalid auth configuration", "error", err)
		os.Exit(1)
	}
	rbacEval := rbac.NewSimpleEvaluator(rbac.DefaultRoleCatalog())
	srv := apihttp.New(logger, intentSvc, topology.NewStubService(), deploySvc, reconcileSvc, auditSvc, authConfig, rbacEval)
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
