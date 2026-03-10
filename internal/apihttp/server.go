package apihttp

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/authn"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/rbac"
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
	authnMW   authn.Middleware
	rbacEval  rbac.Evaluator
}

func New(
	logger *slog.Logger,
	i intent.Service,
	t topology.Service,
	d deploy.Service,
	r reconcile.Service,
	a audit.Service,
	authnConfig authn.Config,
	rbacEval rbac.Evaluator,
) *Server {
	return &Server{
		intent:    i,
		topology:  t,
		deploy:    d,
		reconcile: r,
		audit:     a,
		logger:    logger,
		authnMW:   authn.NewMiddleware(authnConfig, authn.NewParser(), logger),
		rbacEval:  rbacEval,
	}
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
	mux.HandleFunc("/api/v1/topology/devices", s.handleTopologyDevices)
	mux.HandleFunc("/api/v1/topology/devices/", s.handleTopologyDeviceByID)
	mux.HandleFunc("/api/v1/topology/links", s.handleTopologyLinks)
	mux.HandleFunc("/api/v1/topology/import", s.handleTopologyImport)
	mux.HandleFunc("/api/v1/topology/export", s.handleTopologyExport)
	mux.HandleFunc("/api/v1/topology/query/adjacency", s.handleTopologyAdjacency)
	mux.HandleFunc("/api/v1/deployments", s.createDeployment)
	mux.HandleFunc("/api/v1/deployments/", s.getDeployment)
	mux.HandleFunc("/api/v1/reconcile/runs", s.handleReconcileRuns)
	mux.HandleFunc("/api/v1/reconcile/runs/", s.handleReconcileRunByID)
	mux.HandleFunc("/api/v1/drift/findings", s.handleDriftFindings)
	mux.HandleFunc("/api/v1/audit/events", func(w http.ResponseWriter, r *http.Request) {
		out, _ := s.audit.List(r.Context())
		writeJSON(w, 200, out)
	})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("http", "method", r.Method, "path", r.URL.Path)
		mux.ServeHTTP(w, r)
	})
	return s.authnMW.Wrap(handler)
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
		out, err := s.intent.List(r.Context())
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, out)
		return
	}
	if r.Method == http.MethodPost {
		if !s.requirePermission(w, r, rbac.PermissionIntentWrite) {
			return
		}
		var in domain.Intent
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		out, err := s.intent.Create(r.Context(), in)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
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
	if len(parts) == 2 && parts[1] == "validate" && r.Method == http.MethodPost {
		if !s.requirePermission(w, r, rbac.PermissionIntentWrite) {
			return
		}
		if err := s.intent.Validate(r.Context(), id); err != nil {
			writeJSON(w, 404, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"status": "valid"})
		return
	}
	if len(parts) == 2 && parts[1] == "compile" && r.Method == http.MethodPost {
		if !s.requirePermission(w, r, rbac.PermissionIntentWrite) {
			return
		}
		var req struct {
			Vendor string `json:"vendor"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		out, err := s.intent.Compile(r.Context(), id, req.Vendor)
		if err != nil {
			writeJSON(w, 404, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 202, map[string]string{"status": out})
		return
	}
	w.WriteHeader(404)
}

func topologyFilter(r *http.Request) topology.DeviceFilter {
	q := r.URL.Query()
	return topology.DeviceFilter{Site: q.Get("site"), Vendor: q.Get("vendor"), Platform: q.Get("platform")}
}

func (s *Server) handleTopologyDevices(w http.ResponseWriter, r *http.Request) {
	out, err := s.topology.Devices(r.Context(), topologyFilter(r))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleTopologyDeviceByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/topology/devices/")
	out, err := s.topology.Device(r.Context(), id)
	if err != nil {
		if errors.Is(err, topology.ErrNotFound) {
			writeJSON(w, 404, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleTopologyLinks(w http.ResponseWriter, r *http.Request) {
	out, err := s.topology.Links(r.Context(), topologyFilter(r))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleTopologyImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !s.requirePermission(w, r, rbac.PermissionTopologyWrite) {
		return
	}
	var in domain.TopologySnapshot
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if err := s.topology.Import(r.Context(), in); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 202, map[string]string{"status": "imported"})
}

func (s *Server) handleTopologyExport(w http.ResponseWriter, r *http.Request) {
	out, err := s.topology.Export(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleTopologyAdjacency(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("device_id")
	if id == "" {
		writeJSON(w, 400, map[string]string{"error": "device_id is required"})
		return
	}
	out, err := s.topology.AdjacentDeviceIDs(r.Context(), id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"device_id": id, "adjacent_device_ids": out})
}
func (s *Server) createDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	if !s.requirePermission(w, r, rbac.PermissionDeploymentWrite) {
		return
	}
	var req domain.DeploymentPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	out, err := s.deploy.Create(r.Context(), req)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, out)
}
func (s *Server) getDeployment(w http.ResponseWriter, r *http.Request) {
	out, err := s.deploy.Get(r.Context(), strings.TrimPrefix(r.URL.Path, "/api/v1/deployments/"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleReconcileRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !s.requirePermission(w, r, rbac.PermissionReconcileWrite) {
		return
	}
	var req domain.ReconcileRunRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	out, err := s.reconcile.CreateRun(r.Context(), req)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, out)
}

func (s *Server) handleReconcileRunByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/reconcile/runs/")
	out, err := s.reconcile.GetRun(r.Context(), id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) handleDriftFindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	out, err := s.reconcile.ListFindings(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) requirePermission(w http.ResponseWriter, r *http.Request, permission string) bool {
	if s.rbacEval == nil {
		return true
	}
	if err := s.rbacEval.Evaluate(r.Context(), permission); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return false
	}
	return true
}
