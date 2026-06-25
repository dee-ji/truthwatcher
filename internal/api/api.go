package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"truthwatcher/internal/agent"
	"truthwatcher/internal/assets"
	"truthwatcher/internal/audit"
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
	Version            string
	Commit             string
	BuildDate          string
	Logger             *slog.Logger
	DiscoveryRuns      *discovery.Service
	Evidence           *evidence.Service
	Assets             *assets.Service
	Graph              *graph.Service
	Parser             *parser.PersistenceService
	IdentityCandidates *parser.IdentityCandidateService
	Audit              *audit.Service
}

type responseEnvelope struct {
	Data     any            `json:"data"`
	Error    *errorEnvelope `json:"error"`
	Metadata map[string]any `json:"metadata"`
}

type errorEnvelope struct {
	Message string `json:"message"`
}

type systemInfo struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Runtime     runtimeInfo `json:"runtime"`
	Memory      memoryInfo  `json:"memory"`
	Disk        diskInfo    `json:"disk"`
	Build       buildInfo   `json:"build"`
	GeneratedAt time.Time   `json:"generated_at"`
}

type runtimeInfo struct {
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	CPUs       int    `json:"cpus"`
	Goroutines int    `json:"goroutines"`
}

type memoryInfo struct {
	AllocBytes      uint64 `json:"alloc_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	SysBytes        uint64 `json:"sys_bytes"`
	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"`
	HeapSysBytes    uint64 `json:"heap_sys_bytes"`
	NumGC           uint32 `json:"num_gc"`
}

type diskInfo struct {
	Path       string `json:"path"`
	TotalBytes uint64 `json:"total_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
	UsedBytes  uint64 `json:"used_bytes"`
}

type buildInfo struct {
	MainPath  string            `json:"main_path"`
	GoVersion string            `json:"go_version"`
	Settings  map[string]string `json:"settings"`
}

func NewHandler(opts Options) http.Handler {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(nilWriter{}, nil))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz)
	mux.HandleFunc("GET /api/version", handleVersion(opts.Version, opts.Commit, opts.BuildDate))
	mux.HandleFunc("GET /api/v1/version", handleVersion(opts.Version, opts.Commit, opts.BuildDate))
	mux.HandleFunc("GET /openapi.json", handleOpenAPIJSON(opts.Version))
	mux.HandleFunc("GET /docs", handleSwaggerUI)
	mux.HandleFunc("GET /api/v1/system-info", handleSystemInfo(opts.Version))
	mux.HandleFunc("POST /api/v1/discovery-runs", handleCreateDiscoveryRun(opts.DiscoveryRuns))
	mux.HandleFunc("POST /api/v1/discovery-runs/execute", handleExecuteDiscoveryRun(opts.DiscoveryRuns, opts.Evidence, opts.Audit))
	mux.HandleFunc("GET /api/v1/discovery-runs", handleListDiscoveryRuns(opts.DiscoveryRuns))
	mux.HandleFunc("GET /api/v1/discovery-runs/{id}", handleGetDiscoveryRun(opts.DiscoveryRuns))
	mux.HandleFunc("GET /api/v1/discovery-runs/{id}/evidence", handleListEvidenceByDiscoveryRun(opts.Evidence))
	mux.HandleFunc("GET /api/v1/audit-records", handleListAuditRecords(opts.Audit))
	mux.HandleFunc("POST /api/v1/discovery-runs/{id}/parse", handleParseDiscoveryRun(opts.Parser))
	mux.HandleFunc("GET /api/v1/identity-candidates", handleListIdentityCandidates(opts.IdentityCandidates))
	mux.HandleFunc("GET /api/v1/identity-candidates/review-queue", handleListPendingIdentityCandidates(opts.IdentityCandidates))
	mux.HandleFunc("GET /api/v1/identity-candidates/handoff-report", handleIdentityReviewHandoffReport(opts.IdentityCandidates))
	mux.HandleFunc("POST /api/v1/identity-candidates/{id}/review", handleReviewIdentityCandidate(opts.IdentityCandidates))
	mux.HandleFunc("GET /api/v1/evidence/{id}", handleGetEvidence(opts.Evidence))
	mux.HandleFunc("GET /api/v1/assets", handleListAssets(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/provisional-identities", handleListProvisionalIdentityAssets(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}", handleGetAsset(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}/history", handleGetAssetHistory(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}/facts", handleListAssetFacts(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}/relationships", handleListAssetRelationships(opts.Assets))
	mux.HandleFunc("GET /api/v1/assets/{id}/evidence", handleListAssetEvidence(opts.Assets, opts.Evidence))
	mux.HandleFunc("GET /api/v1/facts/conflicts", handleListConflictingFacts(opts.Assets))
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
	writeData(w, http.StatusOK, healthResponse{Status: "ok"})
}

func handleReadyz(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, readinessResponse{Status: "ready"})
}

func handleVersion(version, commit, buildDate string) http.HandlerFunc {
	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	if buildDate == "" {
		buildDate = "unknown"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		writeData(w, http.StatusOK, versionResponse{Name: "truthwatcher", Version: version, Commit: commit, BuildDate: buildDate})
	}
}

func handleSystemInfo(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		info := systemInfo{
			Name:    "truthwatcher",
			Version: version,
			Runtime: runtimeInfo{
				GoVersion:  runtime.Version(),
				OS:         runtime.GOOS,
				Arch:       runtime.GOARCH,
				CPUs:       runtime.NumCPU(),
				Goroutines: runtime.NumGoroutine(),
			},
			Memory: memoryInfo{
				AllocBytes:      mem.Alloc,
				TotalAllocBytes: mem.TotalAlloc,
				SysBytes:        mem.Sys,
				HeapAllocBytes:  mem.HeapAlloc,
				HeapSysBytes:    mem.HeapSys,
				NumGC:           mem.NumGC,
			},
			Disk:        collectDiskInfo("."),
			Build:       collectBuildInfo(),
			GeneratedAt: time.Now().UTC(),
		}

		writeData(w, http.StatusOK, systemInfoResponse{SystemInfo: info})
	}
}

func collectDiskInfo(path string) diskInfo {
	info := diskInfo{Path: path}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return info
	}
	blockSize := uint64(stat.Bsize)
	info.TotalBytes = stat.Blocks * blockSize
	info.FreeBytes = stat.Bavail * blockSize
	if info.TotalBytes >= info.FreeBytes {
		info.UsedBytes = info.TotalBytes - info.FreeBytes
	}
	return info
}

func collectBuildInfo() buildInfo {
	info := buildInfo{Settings: map[string]string{}}
	if build, ok := debug.ReadBuildInfo(); ok {
		info.MainPath = build.Main.Path
		info.GoVersion = build.GoVersion
		for _, setting := range build.Settings {
			info.Settings[setting.Key] = setting.Value
		}
	}
	return info
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

		writeData(w, http.StatusOK, agentMessageResponse{AgentMessage: response})
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

		writeData(w, http.StatusCreated, architectureSeedResponse{ArchitectureSeed: result})
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

		writeData(w, http.StatusOK, discoveryPlanResponse{DiscoveryPlan: plan})
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

		var request createDiscoveryRunRequest
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

		writeData(w, http.StatusCreated, discoveryRunResponse{DiscoveryRun: run})
	}
}

func handleExecuteDiscoveryRun(discoveryRuns *discovery.Service, evidenceStore *evidence.Service, auditStore *audit.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if discoveryRuns == nil {
			writeError(w, http.StatusServiceUnavailable, "discovery run repository is not configured")
			return
		}
		if evidenceStore == nil {
			writeError(w, http.StatusServiceUnavailable, "evidence repository is not configured")
			return
		}

		var request executeDiscoveryRunRequest
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
			Audit:     auditStore,
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
				Data:     discoveryRunResponse{DiscoveryRun: result.DiscoveryRun},
				Error:    &errorEnvelope{Message: err.Error()},
				Metadata: discoveryExecutionMetadata(collectorName, target, profile.Name, tasks, result),
			})
			return
		}

		writeEnvelope(w, http.StatusCreated, responseEnvelope{
			Data:     executeDiscoveryRunResponse{DiscoveryRun: result.DiscoveryRun, Evidence: result.Evidence},
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

		writeData(w, http.StatusOK, discoveryRunsResponse{DiscoveryRuns: runs})
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

		writeData(w, http.StatusOK, discoveryRunResponse{DiscoveryRun: run})
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

		writeData(w, http.StatusOK, evidenceListResponse{Evidence: items})
	}
}

func handleListAuditRecords(service *audit.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "audit repository is not configured")
			return
		}
		filters := audit.ListRecordsFilters{
			DiscoveryRunID: strings.TrimSpace(r.URL.Query().Get("discovery_run_id")),
			EvidenceID:     strings.TrimSpace(r.URL.Query().Get("evidence_id")),
			RequestID:      strings.TrimSpace(r.URL.Query().Get("request_id")),
			Action:         strings.TrimSpace(r.URL.Query().Get("action")),
			Status:         strings.TrimSpace(r.URL.Query().Get("status")),
			Target:         strings.TrimSpace(r.URL.Query().Get("target")),
			Method:         strings.TrimSpace(r.URL.Query().Get("method")),
			Profile:        strings.TrimSpace(r.URL.Query().Get("profile")),
			Limit:          queryInt(r, "limit", 50, 1, 200),
		}
		records, err := service.ListRecords(r.Context(), filters)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		metadata := map[string]any{
			"filters": filters,
			"count":   len(records),
			"limit":   filters.Limit,
		}
		writeDataWithMetadata(w, http.StatusOK, auditRecordsResponse{AuditRecords: records}, metadata)
	}
}

func handleParseDiscoveryRun(service *parser.PersistenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "parser persistence is not configured")
			return
		}

		var request parseDiscoveryRunRequest
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

		writeData(w, http.StatusOK, parseDiscoveryRunResponse{ParseResult: result})
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

		writeData(w, http.StatusOK, evidenceResponse{Evidence: item})
	}
}

func handleGetAssetGraph(service *graph.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "graph repository is not configured")
			return
		}

		depth := queryInt(r, "depth", 1, 1, 2)
		result, err := service.GetAssetGraphWithDepth(r.Context(), r.PathValue("id"), depth)
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, graphResponse{Graph: result})
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

		depth := queryInt(r, "depth", 1, 1, 2)
		result, err := service.GetAssetGraphWithDepth(r.Context(), assetID, depth)
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, graphResponse{Graph: result})
	}
}

func queryInt(r *http.Request, name string, fallback int, minValue int, maxValue int) int {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if parsed < minValue {
		return minValue
	}
	if parsed > maxValue {
		return maxValue
	}
	return parsed
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
	auditIDs := make([]string, 0, len(result.Audit))
	for _, item := range result.Audit {
		if strings.TrimSpace(item.ID) != "" {
			auditIDs = append(auditIDs, item.ID)
		}
	}
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
			"audit_ids":      auditIDs,
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
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			generatedID, err := newRequestID()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to generate request id")
				return
			}
			id = generatedID
		}

		r.Header.Set("X-Request-ID", id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func newRequestID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
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
