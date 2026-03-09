package apihttp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/reconcile"
	"github.com/truthwatcher/truthwatcher/internal/topology"
	"github.com/truthwatcher/truthwatcher/pkg/version"
)

type Server struct {
	intent    intent.Service
	topology  topology.Service
	deploy    deploy.Service
	reconcile reconcile.Service
	audit     audit.Service
	logger    *slog.Logger
}

func New(logger *slog.Logger, i intent.Service, t topology.Service, d deploy.Service, r reconcile.Service, a audit.Service) *Server {
	return &Server{intent: i, topology: t, deploy: d, reconcile: r, audit: a, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]string{"version": version.Version})
	})
	mux.HandleFunc("/api/v1/intents", s.handleIntents)
	mux.HandleFunc("/api/v1/intents/", s.handleIntentByID)
	mux.HandleFunc("/api/v1/topology/devices", func(w http.ResponseWriter, r *http.Request) {
		out, _ := s.topology.Devices(r.Context())
		writeJSON(w, 200, out)
	})
	mux.HandleFunc("/api/v1/topology/links", func(w http.ResponseWriter, r *http.Request) {
		out, _ := s.topology.Links(r.Context())
		writeJSON(w, 200, out)
	})
	mux.HandleFunc("/api/v1/deployments", s.createDeployment)
	mux.HandleFunc("/api/v1/deployments/", s.getDeployment)
	mux.HandleFunc("/api/v1/reconcile/runs", func(w http.ResponseWriter, r *http.Request) {
		out, _ := s.reconcile.CreateRun(r.Context())
		writeJSON(w, 202, out)
	})
	mux.HandleFunc("/api/v1/audit/events", func(w http.ResponseWriter, r *http.Request) {
		out, _ := s.audit.List(r.Context())
		writeJSON(w, 200, out)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("http", "method", r.Method, "path", r.URL.Path)
		mux.ServeHTTP(w, r)
	})
}
func (s *Server) Run(ctx context.Context, addr string) error {
	h := &http.Server{Addr: addr, Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() { <-ctx.Done(); _ = h.Shutdown(context.Background()) }()
	return h.ListenAndServe()
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func (s *Server) handleIntents(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		out, _ := s.intent.List(r.Context())
		writeJSON(w, 200, out)
		return
	}
	if r.Method == http.MethodPost {
		var in domain.Intent
		_ = json.NewDecoder(r.Body).Decode(&in)
		out, _ := s.intent.Create(r.Context(), in)
		writeJSON(w, 201, out)
		return
	}
	w.WriteHeader(405)
}
func (s *Server) handleIntentByID(w http.ResponseWriter, r *http.Request) {
	tail := strings.TrimPrefix(r.URL.Path, "/api/v1/intents/")
	parts := strings.Split(tail, "/")
	id := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		out, err := s.intent.Get(r.Context(), id)
		if err != nil {
			writeJSON(w, 404, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, out)
		return
	}
	if len(parts) == 2 && parts[1] == "validate" {
		if err := s.intent.Validate(r.Context(), id); err != nil {
			writeJSON(w, 404, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"status": "valid"})
		return
	}
	if len(parts) == 2 && parts[1] == "compile" {
		out, err := s.intent.Compile(r.Context(), id)
		if err != nil {
			writeJSON(w, 404, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 202, map[string]string{"status": out})
		return
	}
	w.WriteHeader(404)
}
func (s *Server) createDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		IntentID       string `json:"intent_id"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	out, _ := s.deploy.Create(r.Context(), req.IntentID, req.IdempotencyKey)
	writeJSON(w, 201, out)
}
func (s *Server) getDeployment(w http.ResponseWriter, r *http.Request) {
	out, _ := s.deploy.Get(r.Context(), strings.TrimPrefix(r.URL.Path, "/api/v1/deployments/"))
	writeJSON(w, 200, out)
}
