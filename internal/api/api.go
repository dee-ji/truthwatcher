package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"truthwatcher/internal/agent"
	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/graph"
	"truthwatcher/internal/parser"
	"truthwatcher/internal/planner"
	"truthwatcher/internal/policy"
	"truthwatcher/internal/seeding"
	"truthwatcher/web"
)

type Options struct {
	Version       string
	Logger        *slog.Logger
	DiscoveryRuns *discovery.Service
	Evidence      *evidence.Service
	Assets        *assets.Service
	Graph         *graph.Service
	Parser        *parser.PersistenceService
}

type responseEnvelope struct {
	Data     any            `json:"data"`
	Error    *errorEnvelope `json:"error"`
	Metadata map[string]any `json:"metadata"`
}

type errorEnvelope struct {
	Message string `json:"message"`
}

var requestCounter uint64

func NewHandler(opts Options) http.Handler {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(nilWriter{}, nil))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz)
	mux.HandleFunc("GET /api/v1/version", handleVersion(opts.Version))
	mux.HandleFunc("POST /api/v1/discovery-runs", handleCreateDiscoveryRun(opts.DiscoveryRuns))
	mux.HandleFunc("POST /api/v1/discovery-runs/execute", handleExecuteDiscoveryRun(opts.DiscoveryRuns, opts.Evidence))
	mux.HandleFunc("GET /api/v1/discovery-runs", handleListDiscoveryRuns(opts.DiscoveryRuns))
	mux.HandleFunc("GET /api/v1/discovery-runs/{id}", handleGetDiscoveryRun(opts.DiscoveryRuns))
	mux.HandleFunc("GET /api/v1/discovery-runs/{id}/evidence", handleListEvidenceByDiscoveryRun(opts.Evidence))
	mux.HandleFunc("POST /api/v1/discovery-runs/{id}/parse", handleParseDiscoveryRun(opts.Parser))
	mux.HandleFunc("GET /api/v1/evidence/{id}", handleGetEvidence(opts.Evidence))
	mux.HandleFunc("GET /api/v1/assets", handleListAssets(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}", handleGetAsset(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}/facts", handleListAssetFacts(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}/relationships", handleListAssetRelationships(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}/evidence", handleListAssetEvidence(opts.Assets, opts.Evidence))
	mux.HandleFunc("GET /api/v1/facts/{id}/evidence", handleListFactEvidence(opts.Assets, opts.Evidence))
	mux.HandleFunc("GET /api/v1/assets/{id}/graph", handleGetAssetGraph(opts.Graph))
	mux.HandleFunc("GET /api/v1/graph/neighbors", handleGetGraphNeighbors(opts.Graph))
	mux.HandleFunc("POST /api/v1/agent/messages", handleAgentMessage(opts.Assets, opts.DiscoveryRuns, opts.Evidence))
	mux.HandleFunc("POST /api/v1/architecture-seeds", handleCreateArchitectureSeed(opts.Assets))
	mux.HandleFunc("POST /api/v1/discovery-plans", handleCreateDiscoveryPlan(opts.Assets))
	mux.HandleFunc("GET /api/", handleAPINotFound)
	mux.Handle("GET /", web.Handler())

	return recoverPanic(opts.Logger, requestLog(opts.Logger, requestID(mux)))
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func handleReadyz(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func handleVersion(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeData(w, http.StatusOK, map[string]string{
			"name":    "truthwatcher",
			"version": version,
		})
	}
}

func handleAgentMessage(assetService *assets.Service, discoveryService *discovery.Service, evidenceService *evidence.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if assetService == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}
		if discoveryService == nil {
			writeError(w, http.StatusServiceUnavailable, "discovery run repository is not configured")
			return
		}
		if evidenceService == nil {
			writeError(w, http.StatusServiceUnavailable, "evidence repository is not configured")
			return
		}

		var request agent.Request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "request body is required")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}

		service := agent.NewService(agent.Options{
			Assets:    assetService,
			Discovery: discoveryService,
			Evidence:  evidenceService,
		})
		response, err := service.Reply(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]agent.Response{"agent_message": response})
	}
}

func handleCreateArchitectureSeed(assetService *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if assetService == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		var request seeding.Request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "request body is required")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}

		service := seeding.NewService(seeding.Options{Assets: assetService})
		result, err := service.SeedArchitecture(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeData(w, http.StatusCreated, map[string]seeding.Result{"architecture_seed": result})
	}
}

func handleCreateDiscoveryPlan(assetService *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if assetService == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		var request planner.Request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "request body is required")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}

		service := planner.NewService(planner.Options{
			Assets: assetService,
			Policy: policy.NewEngine(),
		})
		plan, err := service.CreatePlan(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]planner.Plan{"discovery_plan": plan})
	}
}

func handleAPINotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "api endpoint not found")
}

func handleCreateDiscoveryRun(service *discovery.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "discovery run repository is not configured")
			return
		}

		var request struct {
			SeedInput json.RawMessage `json:"seed_input"`
		}
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil {
				if !errors.Is(err, io.EOF) {
					writeError(w, http.StatusBadRequest, "invalid JSON request body")
					return
				}
			}
		}

		run, err := service.CreateDiscoveryRun(r.Context(), discovery.CreateDiscoveryRunParams{
			SeedInput: request.SeedInput,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeData(w, http.StatusCreated, map[string]discovery.DiscoveryRun{"discovery_run": run})
	}
}

func handleExecuteDiscoveryRun(discoveryRuns *discovery.Service, evidenceStore *evidence.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if discoveryRuns == nil {
			writeError(w, http.StatusServiceUnavailable, "discovery run repository is not configured")
			return
		}
		if evidenceStore == nil {
			writeError(w, http.StatusServiceUnavailable, "evidence repository is not configured")
			return
		}

		var request struct {
			Collector   string        `json:"collector"`
			Target      string        `json:"target"`
			Profile     string        `json:"profile"`
			Tasks       []policy.Task `json:"tasks"`
			FixtureRoot string        `json:"fixture_root"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "request body is required")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}

		collectorName, target, profile, tasks, err := validateDiscoveryExecutionRequest(request.Collector, request.Target, request.Profile, request.Tasks)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		result, err := discoveryRuns.StartDiscoveryRun(r.Context(), discovery.StartDiscoveryRunParams{
			Seed: discovery.DiscoverySeed{
				Target: target,
				Method: discovery.FakeMethod,
			},
			Profile:   profile,
			Tasks:     tasks,
			Collector: discovery.NewFakeCollector(request.FixtureRoot, policy.NewEngine()),
			Evidence:  evidenceStore,
			Policy:    policy.NewEngine(),
			Initiator: "api",
			RequestID: r.Header.Get("X-Request-ID"),
			Context:   discoveryRequestContext(r),
		})
		if err != nil {
			status := http.StatusInternalServerError
			if result.DiscoveryRun.ID == "" {
				status = http.StatusBadRequest
			}
			writeEnvelope(w, status, responseEnvelope{
				Data: map[string]discovery.DiscoveryRun{
					"discovery_run": result.DiscoveryRun,
				},
				Error:    &errorEnvelope{Message: err.Error()},
				Metadata: discoveryExecutionMetadata(collectorName, target, profile.Name, tasks, result),
			})
			return
		}

		writeEnvelope(w, http.StatusCreated, responseEnvelope{
			Data: map[string]any{
				"discovery_run": result.DiscoveryRun,
				"evidence":      result.Evidence,
			},
			Metadata: discoveryExecutionMetadata(collectorName, target, profile.Name, tasks, result),
		})
	}
}

func handleListDiscoveryRuns(service *discovery.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "discovery run repository is not configured")
			return
		}

		runs, err := service.ListDiscoveryRuns(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string][]discovery.DiscoveryRun{"discovery_runs": runs})
	}
}

func handleGetDiscoveryRun(service *discovery.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "discovery run repository is not configured")
			return
		}

		run, err := service.GetDiscoveryRun(r.Context(), r.PathValue("id"))
		if errors.Is(err, discovery.ErrNotFound) {
			writeError(w, http.StatusNotFound, "discovery run not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]discovery.DiscoveryRun{"discovery_run": run})
	}
}

func handleListEvidenceByDiscoveryRun(service *evidence.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "evidence repository is not configured")
			return
		}

		items, err := service.ListEvidenceByDiscoveryRun(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string][]evidence.Evidence{"evidence": items})
	}
}

func handleParseDiscoveryRun(service *parser.PersistenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "parser persistence is not configured")
			return
		}

		var request struct {
			Platform string `json:"platform"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "request body is required")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}

		result, err := service.ParseDiscoveryRun(r.Context(), parser.ParseDiscoveryRunParams{
			DiscoveryRunID: r.PathValue("id"),
			Platform:       request.Platform,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]parser.ParseDiscoveryRunResult{"parse_result": result})
	}
}

func handleGetEvidence(service *evidence.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "evidence repository is not configured")
			return
		}

		item, err := service.GetEvidence(r.Context(), r.PathValue("id"))
		if errors.Is(err, evidence.ErrNotFound) {
			writeError(w, http.StatusNotFound, "evidence not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]evidence.Evidence{"evidence": item})
	}
}

func handleGetAssetGraph(service *graph.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "graph repository is not configured")
			return
		}

		result, err := service.GetAssetGraph(r.Context(), r.PathValue("id"))
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]graph.Graph{"graph": result})
	}
}

func handleGetGraphNeighbors(service *graph.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "graph repository is not configured")
			return
		}

		assetID := strings.TrimSpace(r.URL.Query().Get("asset_id"))
		if assetID == "" {
			writeError(w, http.StatusBadRequest, "asset_id is required")
			return
		}

		result, err := service.GetNeighbors(r.Context(), assetID)
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]graph.Graph{"graph": result})
	}
}

func validateDiscoveryExecutionRequest(collector, target, profileName string, tasks []policy.Task) (string, string, discovery.Profile, []policy.Task, error) {
	collector = strings.TrimSpace(collector)
	if collector == "" {
		return "", "", discovery.Profile{}, nil, errors.New("collector is required")
	}
	if collector != discovery.FakeMethod {
		return "", "", discovery.Profile{}, nil, errors.New("only fake discovery execution is available through this endpoint")
	}

	target = strings.TrimSpace(target)
	if target == "" {
		return "", "", discovery.Profile{}, nil, errors.New("target is required")
	}
	if !strings.HasPrefix(target, "fixture://") {
		return "", "", discovery.Profile{}, nil, errors.New("fake discovery target must use fixture://")
	}

	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		return "", "", discovery.Profile{}, nil, errors.New("profile is required")
	}
	profile, ok := discovery.BuiltInProfile(profileName)
	if !ok {
		return "", "", discovery.Profile{}, nil, errors.New("unknown discovery profile")
	}

	if len(tasks) == 0 {
		return "", "", discovery.Profile{}, nil, errors.New("at least one discovery task is required")
	}
	engine := policy.NewEngine()
	if err := profile.Validate(engine); err != nil {
		return "", "", discovery.Profile{}, nil, err
	}
	for _, task := range tasks {
		if err := engine.CheckTask(task); err != nil {
			return "", "", discovery.Profile{}, nil, err
		}
		if _, err := profile.CommandsForTask(task); err != nil {
			return "", "", discovery.Profile{}, nil, err
		}
	}

	return collector, target, profile, tasks, nil
}

func discoveryExecutionMetadata(collector, target, profile string, tasks []policy.Task, result discovery.StartDiscoveryRunResult) map[string]any {
	return map[string]any{
		"audit": map[string]any{
			"initiator":      "api",
			"collector":      collector,
			"target":         target,
			"profile":        profile,
			"tasks":          tasks,
			"discovery_run":  result.DiscoveryRun.ID,
			"run_status":     result.DiscoveryRun.Status,
			"evidence_count": len(result.Evidence),
			"actions":        result.Audit,
		},
	}
}

func discoveryRequestContext(r *http.Request) json.RawMessage {
	payload := map[string]string{
		"request_id": r.Header.Get("X-Request-ID"),
		"path":       r.URL.Path,
		"method":     r.Method,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return data
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeEnvelope(w, status, responseEnvelope{
		Data: data,
	})
}

func writeDataWithMetadata(w http.ResponseWriter, status int, data any, metadata map[string]any) {
	writeEnvelope(w, status, responseEnvelope{
		Data:     data,
		Metadata: metadata,
	})
}

func writeEnvelope(w http.ResponseWriter, status int, payload responseEnvelope) {
	if payload.Metadata == nil {
		payload.Metadata = map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeEnvelope(w, status, responseEnvelope{
		Error: &errorEnvelope{Message: message},
	})
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = "req-" + strconv.FormatUint(atomic.AddUint64(&requestCounter, 1), 10)
		}

		r.Header.Set("X-Request-ID", id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", recorder.Header().Get("X-Request-ID"),
		)
	})
}

func recoverPanic(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("http panic recovered",
					"panic", recovered,
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", w.Header().Get("X-Request-ID"),
					"stack", string(debug.Stack()),
				)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type nilWriter struct{}

func (nilWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
