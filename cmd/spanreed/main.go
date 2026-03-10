package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/truthwatcher/truthwatcher/internal/apihttp"
	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/authn"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/elsecall"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/rbac"
	"github.com/truthwatcher/truthwatcher/internal/reconcile"
	"github.com/truthwatcher/truthwatcher/internal/topology"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://truthwatcher:truthwatcher@localhost:5432/truthwatcher?sslmode=disable"
	}
	db, _ := sql.Open("postgres", dsn)
	auditSvc := audit.Service(audit.NewStubService())
	intentSvc := intent.Service(intent.NewInMemoryService())
	if db != nil {
		if err := db.Ping(); err == nil {
			auditSvc = audit.NewPostgresService(db)
			intentSvc = intent.NewPostgresService(db, auditSvc, elsecall.NewCompilerService())
			logger.Info("spanreed configured with postgres backend")
		} else {
			logger.Warn("postgres unavailable, using in-memory services", "error", err)
		}
	}

	topologySvc := topology.Service(topology.NewStubService())
	if db != nil {
		if err := db.Ping(); err == nil {
			topologySvc = topology.NewService(topology.NewPostgresRepository(db))
		}
	}

	authConfig := authn.Config{Mode: authn.ModeJWT, LocalDevBypass: true, BypassSubject: "spanreed-local", BypassRoles: []string{"admin"}}
	rbacEval := rbac.NewSimpleEvaluator(rbac.DefaultRoleCatalog())
	srv := apihttp.New(logger, intentSvc, topologySvc, deploy.NewStubServiceWithDependencies(auditSvc, intentSvc), reconcile.NewStubService(), auditSvc, authConfig, rbacEval)
	if err := srv.Run(context.Background(), ":8080"); err != nil {
		logger.Error("spanreed exited", "error", err)
	}
}
