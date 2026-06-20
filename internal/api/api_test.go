package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/audit"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/graph"
	"truthwatcher/internal/parser"
	"truthwatcher/internal/seeding"
)

func TestHealthz(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeResponseData[map[string]string](t, response)
	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want ok", body["status"])
	}
}

func TestReadyz(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeResponseData[map[string]string](t, response)
	if body["status"] != "ready" {
		t.Fatalf("status body = %q, want ready", body["status"])
	}
}

func TestVersion(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeResponseData[map[string]string](t, response)
	if body["name"] != "truthwatcher" {
		t.Fatalf("name = %q, want truthwatcher", body["name"])
	}
	if body["version"] != "test-version" {
		t.Fatalf("version = %q, want test-version", body["version"])
	}
}

func TestServesEmbeddedFrontend(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", contentType)
	}
	if !strings.Contains(response.Body.String(), "Truthwatcher") {
		t.Fatalf("body does not contain frontend shell: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "#/discovery-runs") {
		t.Fatalf("body does not contain discovery runs navigation: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "#/assets") {
		t.Fatalf("body does not contain assets navigation: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "#/discovery-plans") {
		t.Fatalf("body does not contain discovery plans navigation: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "#/architecture-seeds") {
		t.Fatalf("body does not contain architecture seeds navigation: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "#/graph") {
		t.Fatalf("body does not contain graph navigation: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "#/ask") {
		t.Fatalf("body does not contain ask navigation: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "#/about") {
		t.Fatalf("body does not contain about navigation: %s", response.Body.String())
	}
}

func TestServesEmbeddedFrontendAsset(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "javascript") {
		t.Fatalf("Content-Type = %q, want javascript", contentType)
	}
	body := response.Body.String()
	if !strings.Contains(body, "/api/v1/discovery-runs/execute") {
		t.Fatalf("body does not contain fake discovery execution endpoint: %s", body)
	}
	if !strings.Contains(body, `collector: "fake"`) {
		t.Fatalf("body does not constrain discovery form to fake collector: %s", body)
	}
	if !strings.Contains(body, "/api/v1/assets/${encodeURIComponent(assetID)}/graph") {
		t.Fatalf("body does not contain graph API fetch path: %s", body)
	}
	if !strings.Contains(body, "renderAssetsView") {
		t.Fatalf("body does not contain asset browser renderer: %s", body)
	}
	if !strings.Contains(body, "renderDiscoveryPlansView") {
		t.Fatalf("body does not contain discovery plan renderer: %s", body)
	}
	if !strings.Contains(body, "/api/v1/discovery-plans") {
		t.Fatalf("body does not contain discovery plans endpoint: %s", body)
	}
	if !strings.Contains(body, "This UI does not execute plans") {
		t.Fatalf("body does not clearly prevent automatic execution: %s", body)
	}
	if !strings.Contains(body, "renderArchitectureSeedsView") {
		t.Fatalf("body does not contain architecture seed renderer: %s", body)
	}
	if !strings.Contains(body, "/api/v1/architecture-seeds") {
		t.Fatalf("body does not contain architecture seeds endpoint: %s", body)
	}
	if !strings.Contains(body, "Seeded hints are context, not observed proof") {
		t.Fatalf("body does not label seeded hints as context: %s", body)
	}
	if !strings.Contains(body, "user_seeded facts with low confidence") {
		t.Fatalf("body does not show seeded facts as low confidence user_seeded data: %s", body)
	}
	if !strings.Contains(body, "/api/v1/assets/${encodeURIComponent(id)}/facts?limit=100") {
		t.Fatalf("body does not contain asset fact API fetch path: %s", body)
	}
	if !strings.Contains(body, "/api/v1/assets/${encodeURIComponent(id)}/relationships?limit=100") {
		t.Fatalf("body does not contain asset relationship API fetch path: %s", body)
	}
	if !strings.Contains(body, "renderGraph") {
		t.Fatalf("body does not contain graph renderer: %s", body)
	}
	if !strings.Contains(body, "/api/v1/evidence/${encodeURIComponent(evidenceID)}") {
		t.Fatalf("body does not contain evidence API fetch path: %s", body)
	}
	if !strings.Contains(body, "Evidence is read-only") {
		t.Fatalf("body does not label evidence as read-only: %s", body)
	}
	if !strings.Contains(body, "copyEvidenceRawOutput") {
		t.Fatalf("body does not contain raw output copy helper: %s", body)
	}
	if !strings.Contains(body, "/api/v1/agent/messages") {
		t.Fatalf("body does not contain agent message endpoint: %s", body)
	}
	if !strings.Contains(body, "Deterministic canned responses only") {
		t.Fatalf("body does not label agent shell as deterministic: %s", body)
	}
	if !strings.Contains(body, "renderAboutView") {
		t.Fatalf("body does not contain about system renderer: %s", body)
	}
	if !strings.Contains(body, "/api/v1/system-info") {
		t.Fatalf("body does not contain system info endpoint: %s", body)
	}
	if !strings.Contains(body, "Evidence before inference") {
		t.Fatalf("body does not contain philosophy copy: %s", body)
	}
}

func TestSystemInfo(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/system-info", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		SystemInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Runtime struct {
				CPUs int `json:"cpus"`
			} `json:"runtime"`
			Disk struct {
				TotalBytes uint64 `json:"total_bytes"`
			} `json:"disk"`
		} `json:"system_info"`
	}](t, response)
	if body.SystemInfo.Name != "truthwatcher" {
		t.Fatalf("name = %q, want truthwatcher", body.SystemInfo.Name)
	}
	if body.SystemInfo.Version != "test-version" {
		t.Fatalf("version = %q, want test-version", body.SystemInfo.Version)
	}
	if body.SystemInfo.Runtime.CPUs <= 0 {
		t.Fatalf("cpus = %d, want positive", body.SystemInfo.Runtime.CPUs)
	}
	if body.SystemInfo.Disk.TotalBytes == 0 {
		t.Fatal("disk total bytes is zero")
	}
}

func TestAgentMessageListsKnownAssets(t *testing.T) {
	assetService := assets.NewService(testAssetRepository())
	discoveryService := discovery.NewService(&fakeDiscoveryRunRepository{
		runs: []discovery.DiscoveryRun{{
			ID:        "11111111-1111-4111-8111-111111111111",
			Status:    discovery.StatusCompleted,
			SeedInput: json.RawMessage(`{"target":"fixture://junos-mx"}`),
			StartedAt: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
			CreatedAt: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		}},
	})
	evidenceService := evidence.NewService(&fakeEvidenceRepository{items: []evidence.Evidence{
		testEvidence("evidence-a"),
		testEvidence("evidence-b"),
	}})
	handler := NewHandler(Options{
		Version:       "test-version",
		Assets:        &assetService,
		DiscoveryRuns: &discoveryService,
		Evidence:      &evidenceService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent/messages", strings.NewReader(`{"message":"list known assets"}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		AgentMessage struct {
			Message  string   `json:"message"`
			Intent   string   `json:"intent"`
			ReadOnly bool     `json:"read_only"`
			Actions  []string `json:"actions"`
		} `json:"agent_message"`
	}](t, response)
	if body.AgentMessage.Intent != "list_known_assets" {
		t.Fatalf("intent = %q, want list_known_assets", body.AgentMessage.Intent)
	}
	if !body.AgentMessage.ReadOnly {
		t.Fatal("agent response is not marked read_only")
	}
	if !strings.Contains(body.AgentMessage.Message, "Known assets: 3") {
		t.Fatalf("message = %q, want known asset count", body.AgentMessage.Message)
	}
	if strings.Contains(strings.Join(body.AgentMessage.Actions, ","), "execute") {
		t.Fatalf("actions include execution: %#v", body.AgentMessage.Actions)
	}
}

func TestAgentMessageRequiresMessage(t *testing.T) {
	assetService := assets.NewService(testAssetRepository())
	discoveryService := discovery.NewService(&fakeDiscoveryRunRepository{})
	evidenceService := evidence.NewService(&fakeEvidenceRepository{})
	handler := NewHandler(Options{
		Version:       "test-version",
		Assets:        &assetService,
		DiscoveryRuns: &discoveryService,
		Evidence:      &evidenceService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent/messages", strings.NewReader(`{"message":""}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	envelope := decodeResponseEnvelope(t, response)
	if envelope.Error == nil || envelope.Error.Message != "message is required" {
		t.Fatalf("error = %#v, want message is required", envelope.Error)
	}
}

func TestCreateDiscoveryPlan(t *testing.T) {
	assetService := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &assetService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-plans", strings.NewReader(`{
		"target": "router-a",
		"method": "ssh",
		"profile": "juniper_junos"
	}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		DiscoveryPlan struct {
			ApprovalRequired bool `json:"approval_required"`
			ExecutionAllowed bool `json:"execution_allowed"`
			Steps            []struct {
				Target           string `json:"target"`
				Method           string `json:"method"`
				Profile          string `json:"profile"`
				Task             string `json:"task"`
				Reason           string `json:"reason"`
				ExpectedEvidence string `json:"expected_evidence"`
				RiskLevel        string `json:"risk_level"`
			} `json:"steps"`
		} `json:"discovery_plan"`
	}](t, response)
	if !body.DiscoveryPlan.ApprovalRequired {
		t.Fatal("approval_required = false, want true")
	}
	if body.DiscoveryPlan.ExecutionAllowed {
		t.Fatal("execution_allowed = true, want false")
	}
	if len(body.DiscoveryPlan.Steps) == 0 {
		t.Fatal("plan steps are empty")
	}
	if body.DiscoveryPlan.Steps[0].Target != "router-a" {
		t.Fatalf("target = %q, want router-a", body.DiscoveryPlan.Steps[0].Target)
	}
	if body.DiscoveryPlan.Steps[0].RiskLevel != "low_read_only" {
		t.Fatalf("risk = %q, want low_read_only", body.DiscoveryPlan.Steps[0].RiskLevel)
	}
}

func TestCreateDiscoveryPlanRejectsScopeExpansion(t *testing.T) {
	assetService := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &assetService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-plans", strings.NewReader(`{
		"target": "10.0.0.0/24",
		"method": "ssh",
		"profile": "juniper_junos"
	}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestCreateArchitectureSeed(t *testing.T) {
	assetService := assets.NewService(&fakeAssetRepository{})
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &assetService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/architecture-seeds", strings.NewReader(`{
		"organization_network_type": "service_provider",
		"known_asns": ["65000"],
		"known_route_reflectors": ["rr1.example.net"],
		"known_vendors": ["juniper"],
		"known_ems_systems": ["ems-a"],
		"known_services": ["l3vpn"],
		"known_regions_markets": ["nyc"]
	}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	body := decodeResponseData[struct {
		ArchitectureSeed seeding.Result `json:"architecture_seed"`
	}](t, response)
	if body.ArchitectureSeed.Asset.State != assets.StateUserSeeded {
		t.Fatalf("asset state = %q, want user_seeded", body.ArchitectureSeed.Asset.State)
	}
	if body.ArchitectureSeed.Warning == "" {
		t.Fatal("seed warning is empty")
	}
	if got, want := len(body.ArchitectureSeed.Facts), 7; got != want {
		t.Fatalf("fact count = %d, want %d", got, want)
	}
	for _, fact := range body.ArchitectureSeed.Facts {
		if fact.Source != seeding.SeedSource {
			t.Fatalf("fact source = %q, want user_seeded", fact.Source)
		}
		if fact.State != assets.StateUserSeeded {
			t.Fatalf("fact state = %q, want user_seeded", fact.State)
		}
	}
}

func TestCreateArchitectureSeedRejectsEmptyHints(t *testing.T) {
	assetService := assets.NewService(&fakeAssetRepository{})
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &assetService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/architecture-seeds", strings.NewReader(`{}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestUnknownAPIPathDoesNotServeFrontend(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/not-real", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	envelope := decodeResponseEnvelope(t, response)
	if envelope.Error == nil || envelope.Error.Message != "api endpoint not found" {
		t.Fatalf("error = %#v, want api endpoint not found", envelope.Error)
	}
}

func TestRequestIDMiddlewarePreservesIncomingID(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", "caller-request-id")
	handler.ServeHTTP(response, request)

	if got := response.Header().Get("X-Request-ID"); got != "caller-request-id" {
		t.Fatalf("X-Request-ID = %q, want caller-request-id", got)
	}
}

func TestRequestLoggingMiddleware(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := NewHandler(Options{
		Version: "test-version",
		Logger:  logger,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(response, request)

	got := logs.String()
	if !strings.Contains(got, "http request") {
		t.Fatalf("logs = %q, want request log", got)
	}
	if !strings.Contains(got, "path=/healthz") {
		t.Fatalf("logs = %q, want request path", got)
	}
	if !strings.Contains(got, "status=200") {
		t.Fatalf("logs = %q, want status", got)
	}
}

func TestPanicRecoveryMiddleware(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := recoverPanic(logger, requestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("X-Request-ID is empty")
	}
	if !strings.Contains(logs.String(), "http panic recovered") {
		t.Fatalf("logs = %q, want panic recovery log", logs.String())
	}
}

func decodeResponseData[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()
	var envelope struct {
		Data     T              `json:"data"`
		Error    *errorEnvelope `json:"error"`
		Metadata map[string]any `json:"metadata"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Error != nil {
		t.Fatalf("unexpected error response: %s", envelope.Error.Message)
	}
	if envelope.Metadata == nil {
		t.Fatal("metadata is nil")
	}
	return envelope.Data
}

func decodeResponseEnvelope(t *testing.T, response *httptest.ResponseRecorder) responseEnvelope {
	t.Helper()
	var envelope responseEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Metadata == nil {
		t.Fatal("metadata is nil")
	}
	return envelope
}

type fakeDiscoveryRunRepository struct {
	runs []discovery.DiscoveryRun
}

func (f *fakeDiscoveryRunRepository) CreateDiscoveryRun(ctx context.Context, params discovery.CreateDiscoveryRunParams) (discovery.DiscoveryRun, error) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	run := discovery.DiscoveryRun{
		ID:        "11111111-1111-4111-8111-111111111111",
		Status:    discovery.StatusPending,
		SeedInput: params.SeedInput,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	f.runs = append([]discovery.DiscoveryRun{run}, f.runs...)
	return run, nil
}

func (f *fakeDiscoveryRunRepository) GetDiscoveryRun(ctx context.Context, id string) (discovery.DiscoveryRun, error) {
	for _, run := range f.runs {
		if run.ID == id {
			return run, nil
		}
	}
	return discovery.DiscoveryRun{}, discovery.ErrNotFound
}

func (f *fakeDiscoveryRunRepository) ListDiscoveryRuns(ctx context.Context) ([]discovery.DiscoveryRun, error) {
	return f.runs, nil
}

func (f *fakeDiscoveryRunRepository) UpdateDiscoveryRunStatus(ctx context.Context, params discovery.UpdateDiscoveryRunStatusParams) (discovery.DiscoveryRun, error) {
	for i := range f.runs {
		if f.runs[i].ID == params.ID {
			f.runs[i].Status = params.Status
			f.runs[i].CompletedAt = params.CompletedAt
			f.runs[i].ErrorMessage = params.ErrorMessage
			f.runs[i].UpdatedAt = time.Date(2026, 6, 10, 12, 1, 0, 0, time.UTC)
			return f.runs[i], nil
		}
	}
	return discovery.DiscoveryRun{}, discovery.ErrNotFound
}

func TestCreateDiscoveryRun(t *testing.T) {
	repo := &fakeDiscoveryRunRepository{}
	service := discovery.NewService(repo)
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-runs", strings.NewReader(`{"seed_input":{"target":"router1"}}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	body := decodeResponseData[struct {
		DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
	}](t, response)
	if body.DiscoveryRun.Status != discovery.StatusPending {
		t.Fatalf("status = %q, want pending", body.DiscoveryRun.Status)
	}
	if !strings.Contains(string(body.DiscoveryRun.SeedInput), "router1") {
		t.Fatalf("seed_input = %s, want router1", body.DiscoveryRun.SeedInput)
	}
}

func TestExecuteDiscoveryRunWithFakeCollector(t *testing.T) {
	runRepo := &fakeDiscoveryRunRepository{}
	runService := discovery.NewService(runRepo)
	evidenceRepo := &fakeEvidenceRepository{}
	evidenceService := evidence.NewService(evidenceRepo)
	auditRepo := &fakeAuditRepository{}
	auditService := audit.NewService(auditRepo)
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &runService,
		Evidence:      &evidenceService,
		Audit:         &auditService,
	})

	requestBody := `{
		"collector": "fake",
		"target": "fixture://junos-mx",
		"profile": "juniper_junos",
		"tasks": ["identify_device"],
		"fixture_root": "../../examples/fixtures"
	}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-runs/execute", strings.NewReader(requestBody))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	body := decodeResponseData[struct {
		DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
		Evidence     []evidence.Evidence    `json:"evidence"`
	}](t, response)
	if body.DiscoveryRun.Status != discovery.StatusCompleted {
		t.Fatalf("status = %q, want completed", body.DiscoveryRun.Status)
	}
	if got, want := len(body.Evidence), 1; got != want {
		t.Fatalf("evidence count = %d, want %d", got, want)
	}
	if body.Evidence[0].CommandOrAPI != "show version" {
		t.Fatalf("command = %q, want show version", body.Evidence[0].CommandOrAPI)
	}

	envelope := decodeResponseEnvelope(t, response)
	audit, ok := envelope.Metadata["audit"].(map[string]any)
	if !ok {
		t.Fatalf("audit metadata = %#v, want object", envelope.Metadata["audit"])
	}
	if audit["initiator"] != "api" {
		t.Fatalf("initiator = %q, want api", audit["initiator"])
	}
	if audit["target"] != "fixture://junos-mx" {
		t.Fatalf("target = %q, want fixture://junos-mx", audit["target"])
	}
	actions, ok := audit["actions"].([]any)
	if !ok {
		t.Fatalf("actions = %#v, want array", audit["actions"])
	}
	if len(actions) != 1 {
		t.Fatalf("action count = %d, want 1", len(actions))
	}
	action, ok := actions[0].(map[string]any)
	if !ok {
		t.Fatalf("action = %#v, want object", actions[0])
	}
	if action["initiator"] != "api" {
		t.Fatalf("action initiator = %q, want api", action["initiator"])
	}
	if action["target"] != "fixture://junos-mx" {
		t.Fatalf("action target = %q, want fixture://junos-mx", action["target"])
	}
	if action["profile"] != "juniper_junos" {
		t.Fatalf("action profile = %q, want juniper_junos", action["profile"])
	}
	if action["command_or_api"] != "show version" {
		t.Fatalf("action command = %q, want show version", action["command_or_api"])
	}
	if action["evidence_id"] == "" {
		t.Fatalf("action evidence_id is empty: %#v", action)
	}
	if action["id"] == "" {
		t.Fatalf("action audit id is empty: %#v", action)
	}
	auditIDs, ok := audit["audit_ids"].([]any)
	if !ok {
		t.Fatalf("audit_ids = %#v, want array", audit["audit_ids"])
	}
	if len(auditIDs) != 1 || auditIDs[0] == "" {
		t.Fatalf("audit_ids = %#v, want persisted command audit id", auditIDs)
	}
	if got, want := len(auditRepo.records), 2; got != want {
		t.Fatalf("persisted audit records = %d, want command plus run record", got)
	}
}

func TestExecuteDiscoveryRunRequiresExplicitProfileAndTasks(t *testing.T) {
	runRepo := &fakeDiscoveryRunRepository{}
	runService := discovery.NewService(runRepo)
	evidenceRepo := &fakeEvidenceRepository{}
	evidenceService := evidence.NewService(evidenceRepo)
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &runService,
		Evidence:      &evidenceService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-runs/execute", strings.NewReader(`{
		"collector": "fake",
		"target": "fixture://junos-mx"
	}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	envelope := decodeResponseEnvelope(t, response)
	if envelope.Error == nil || envelope.Error.Message != "profile is required" {
		t.Fatalf("error = %#v, want profile is required", envelope.Error)
	}
}

func TestExecuteDiscoveryRunRejectsTaskNotInProfile(t *testing.T) {
	runRepo := &fakeDiscoveryRunRepository{}
	runService := discovery.NewService(runRepo)
	evidenceRepo := &fakeEvidenceRepository{}
	evidenceService := evidence.NewService(evidenceRepo)
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &runService,
		Evidence:      &evidenceService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-runs/execute", strings.NewReader(`{
		"collector": "fake",
		"target": "fixture://junos-mx",
		"profile": "juniper_junos",
		"tasks": ["not_allowed"]
	}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestListDiscoveryRuns(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	repo := &fakeDiscoveryRunRepository{
		runs: []discovery.DiscoveryRun{{
			ID:        "11111111-1111-4111-8111-111111111111",
			Status:    discovery.StatusPending,
			SeedInput: json.RawMessage(`{}`),
			StartedAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		}},
	}
	service := discovery.NewService(repo)
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-runs", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeResponseData[struct {
		DiscoveryRuns []discovery.DiscoveryRun `json:"discovery_runs"`
	}](t, response)
	if len(body.DiscoveryRuns) != 1 {
		t.Fatalf("len = %d, want 1", len(body.DiscoveryRuns))
	}
}

func TestGetDiscoveryRun(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	id := "11111111-1111-4111-8111-111111111111"
	repo := &fakeDiscoveryRunRepository{
		runs: []discovery.DiscoveryRun{{
			ID:        id,
			Status:    discovery.StatusPending,
			SeedInput: json.RawMessage(`{}`),
			StartedAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		}},
	}
	service := discovery.NewService(repo)
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-runs/"+id, nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeResponseData[struct {
		DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
	}](t, response)
	if body.DiscoveryRun.ID != id {
		t.Fatalf("id = %q, want %q", body.DiscoveryRun.ID, id)
	}
}

func TestGetDiscoveryRunNotFound(t *testing.T) {
	repo := &fakeDiscoveryRunRepository{}
	service := discovery.NewService(repo)
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-runs/missing", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestDiscoveryRunEndpointsReturnUnavailableWithoutRepository(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-runs", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

type fakeEvidenceRepository struct {
	items []evidence.Evidence
}

type fakeAuditRepository struct {
	records []audit.Record
}

func (f *fakeAuditRepository) CreateAuditRecord(ctx context.Context, params audit.CreateRecordParams) (audit.Record, error) {
	record := audit.Record{
		ID:             "audit-" + string(rune('a'+len(f.records))),
		Action:         params.Action,
		Initiator:      params.Initiator,
		RequestID:      params.RequestID,
		DiscoveryRunID: params.DiscoveryRunID,
		Target:         params.Target,
		Method:         params.Method,
		Profile:        params.Profile,
		Task:           params.Task,
		CommandOrAPI:   params.CommandOrAPI,
		Status:         params.Status,
		EvidenceID:     params.EvidenceID,
		ErrorMessage:   params.ErrorMessage,
		StartedAt:      params.StartedAt,
		CompletedAt:    params.CompletedAt,
		Context:        params.Context,
	}
	f.records = append(f.records, record)
	return record, nil
}

func (f *fakeEvidenceRepository) CreateEvidence(ctx context.Context, params evidence.CreateEvidenceParams) (evidence.Evidence, error) {
	item := evidence.Evidence{
		ID:             "22222222-2222-4222-8222-222222222222",
		DiscoveryRunID: params.DiscoveryRunID,
		Target:         params.Target,
		Method:         params.Method,
		CommandOrAPI:   params.CommandOrAPI,
		RawOutput:      params.RawOutput,
		RawOutputHash:  evidence.HashRawOutput(params.RawOutput),
		CollectedAt:    time.Date(2026, 6, 10, 12, 2, 0, 0, time.UTC),
		Metadata:       params.Metadata,
	}
	f.items = append(f.items, item)
	return item, nil
}

func (f *fakeEvidenceRepository) GetEvidence(ctx context.Context, id string) (evidence.Evidence, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	return evidence.Evidence{}, evidence.ErrNotFound
}

func (f *fakeEvidenceRepository) ListEvidenceByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]evidence.Evidence, error) {
	var result []evidence.Evidence
	for _, item := range f.items {
		if item.DiscoveryRunID == discoveryRunID {
			result = append(result, item)
		}
	}
	return result, nil
}

func TestListEvidenceByDiscoveryRun(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	repo := &fakeEvidenceRepository{
		items: []evidence.Evidence{{
			ID:             "22222222-2222-4222-8222-222222222222",
			DiscoveryRunID: "11111111-1111-4111-8111-111111111111",
			Target:         "router1",
			Method:         "ssh",
			CommandOrAPI:   "show version",
			RawOutput:      "raw output",
			RawOutputHash:  evidence.HashRawOutput("raw output"),
			CollectedAt:    now,
			Metadata:       json.RawMessage(`{}`),
		}},
	}
	service := evidence.NewService(repo)
	handler := NewHandler(Options{
		Version:  "test-version",
		Evidence: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/discovery-runs/11111111-1111-4111-8111-111111111111/evidence", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeResponseData[struct {
		Evidence []evidence.Evidence `json:"evidence"`
	}](t, response)
	if len(body.Evidence) != 1 {
		t.Fatalf("len = %d, want 1", len(body.Evidence))
	}
	if body.Evidence[0].RawOutputHash == "" {
		t.Fatal("raw_output_hash is empty")
	}
}

func TestParseDiscoveryRunEndpoint(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	evidenceService := evidence.NewService(&fakeEvidenceRepository{
		items: []evidence.Evidence{{
			ID:             "evidence-a",
			DiscoveryRunID: runID,
			Target:         "fixture://junos-mx",
			Method:         "fake",
			CommandOrAPI:   "show version",
			RawOutput:      "Hostname: mx-edge-01\nModel: mx480\nJunos: 22.4R3-S2.4\n",
			RawOutputHash:  evidence.HashRawOutput("Hostname: mx-edge-01\nModel: mx480\nJunos: 22.4R3-S2.4\n"),
			CollectedAt:    time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
			Metadata:       json.RawMessage(`{}`),
		}},
	})
	assetService := assets.NewService(&fakeAssetRepository{})
	parseService := parser.NewPersistenceService(parser.PersistenceOptions{
		Evidence:     evidenceService,
		Assets:       assetService,
		ParseResults: &fakeParseResultRepository{},
		Registry:     parser.BuiltInRegistry(),
	})
	handler := NewHandler(Options{
		Version: "test-version",
		Parser:  &parseService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-runs/"+runID+"/parse", strings.NewReader(`{"platform":"junos"}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		ParseResult parser.ParseDiscoveryRunResult `json:"parse_result"`
	}](t, response)
	if body.ParseResult.DiscoveryRunID != runID {
		t.Fatalf("discovery run id = %q, want %q", body.ParseResult.DiscoveryRunID, runID)
	}
	if len(body.ParseResult.Assets) != 1 {
		t.Fatalf("asset count = %d, want 1", len(body.ParseResult.Assets))
	}
	if len(body.ParseResult.Facts) == 0 {
		t.Fatal("expected facts from parser")
	}
	if len(body.ParseResult.ParseResults) != 1 {
		t.Fatalf("parse result count = %d, want 1", len(body.ParseResult.ParseResults))
	}
}

func TestListIdentityCandidates(t *testing.T) {
	service := parser.NewIdentityCandidateService(&fakeIdentityCandidateRepository{
		items: []parser.IdentityCandidate{
			{
				ID:                   "candidate-a",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-a",
				ParserName:           "junos_show_version",
				AssetType:            "device",
				CandidateIdentityKey: "device:hostname:mx-edge-01",
				Strength:             assets.IdentityStrengthProvisional,
				Confidence:           0.55,
				Reason:               "hostname is not globally unique and may change",
				Hostname:             stringPtr("mx-edge-01"),
				ReviewState:          parser.IdentityReviewPending,
				Metadata:             json.RawMessage(`{}`),
				CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:                   "candidate-b",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-b",
				ParserName:           "junos_show_chassis_hardware",
				AssetType:            "chassis",
				CandidateIdentityKey: "chassis:vendor_serial:juniper:jn1234abcdef",
				Strength:             assets.IdentityStrengthStrong,
				Confidence:           0.85,
				Reason:               "identity key uses a durable identifier",
				Serial:               stringPtr("JN1234ABCDEF"),
				ReviewState:          parser.IdentityReviewPending,
				Metadata:             json.RawMessage(`{}`),
				CreatedAt:            time.Date(2026, 6, 10, 12, 1, 0, 0, time.UTC),
			},
		},
	})
	handler := NewHandler(Options{
		Version:            "test-version",
		IdentityCandidates: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/identity-candidates?review_state=pending&strength=strong", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		IdentityCandidates []parser.IdentityCandidate `json:"identity_candidates"`
	}](t, response)
	if got, want := len(body.IdentityCandidates), 1; got != want {
		t.Fatalf("candidate count = %d, want %d", got, want)
	}
	if body.IdentityCandidates[0].ID != "candidate-b" {
		t.Fatalf("candidate id = %q, want candidate-b", body.IdentityCandidates[0].ID)
	}
}

func TestListIdentityCandidateReviewQueue(t *testing.T) {
	service := parser.NewIdentityCandidateService(&fakeIdentityCandidateRepository{
		items: []parser.IdentityCandidate{
			{
				ID:                   "candidate-pending",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-a",
				ParserName:           "junos_show_version",
				AssetType:            "device",
				CandidateIdentityKey: "device:hostname:mx-edge-01",
				Strength:             assets.IdentityStrengthProvisional,
				Confidence:           0.55,
				Reason:               "hostname is not globally unique and may change",
				ReviewState:          parser.IdentityReviewPending,
				Metadata:             json.RawMessage(`{}`),
				CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:                   "candidate-accepted",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-b",
				ParserName:           "junos_show_chassis_hardware",
				AssetType:            "chassis",
				CandidateIdentityKey: "chassis:vendor_serial:juniper:jn1234abcdef",
				Strength:             assets.IdentityStrengthStrong,
				Confidence:           0.85,
				Reason:               "identity key uses a durable identifier",
				ReviewState:          parser.IdentityReviewAccepted,
				Metadata:             json.RawMessage(`{}`),
				CreatedAt:            time.Date(2026, 6, 10, 12, 1, 0, 0, time.UTC),
			},
		},
	})
	handler := NewHandler(Options{
		Version:            "test-version",
		IdentityCandidates: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/identity-candidates/review-queue", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		IdentityCandidates []parser.IdentityCandidate `json:"identity_candidates"`
	}](t, response)
	if got, want := len(body.IdentityCandidates), 1; got != want {
		t.Fatalf("queued candidate count = %d, want %d", got, want)
	}
	if body.IdentityCandidates[0].ID != "candidate-pending" {
		t.Fatalf("candidate id = %q, want candidate-pending", body.IdentityCandidates[0].ID)
	}
}

func TestReviewIdentityCandidateActionsAreAudited(t *testing.T) {
	repo := &fakeIdentityCandidateRepository{
		items: []parser.IdentityCandidate{
			{
				ID:                   "candidate-accept",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-a",
				ParserName:           "junos_show_version",
				AssetType:            "device",
				CandidateIdentityKey: "device:hostname:mx-edge-01",
				Strength:             assets.IdentityStrengthProvisional,
				Confidence:           0.55,
				Reason:               "hostname is not globally unique and may change",
				ReviewState:          parser.IdentityReviewPending,
				Metadata:             json.RawMessage(`{}`),
			},
			{
				ID:                   "candidate-more-evidence",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-b",
				ParserName:           "junos_show_lldp_neighbors",
				AssetType:            "device",
				CandidateIdentityKey: "device:hostname:spine-01",
				Strength:             assets.IdentityStrengthProvisional,
				Confidence:           0.75,
				Reason:               "hostname is not globally unique and may change",
				ReviewState:          parser.IdentityReviewPending,
				Metadata:             json.RawMessage(`{}`),
			},
		},
	}
	service := parser.NewIdentityCandidateService(repo)
	handler := NewHandler(Options{
		Version:            "test-version",
		IdentityCandidates: &service,
	})

	tests := []struct {
		id        string
		body      string
		wantState parser.IdentityReviewState
	}{
		{
			id:        "candidate-accept",
			body:      `{"reviewer":"netops","action":"accept","rationale":"hostname aligns with fixture evidence","metadata":{"ticket":"TW-1"}}`,
			wantState: parser.IdentityReviewAccepted,
		},
		{
			id:        "candidate-more-evidence",
			body:      `{"reviewer":"netops","action":"request_more_evidence","rationale":"neighbor name needs corroborating inventory evidence"}`,
			wantState: parser.IdentityReviewMoreEvidence,
		},
	}

	for _, tt := range tests {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/identity-candidates/"+tt.id+"/review", strings.NewReader(tt.body))
		request.Header.Set("X-Request-ID", "req-review")
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d; body=%s", tt.id, response.Code, http.StatusOK, response.Body.String())
		}

		body := decodeResponseData[struct {
			Review parser.IdentityCandidateReview `json:"identity_candidate_review"`
		}](t, response)
		if body.Review.IdentityCandidateID != tt.id {
			t.Fatalf("review candidate id = %q, want %q", body.Review.IdentityCandidateID, tt.id)
		}
		if body.Review.EvidenceID == "" {
			t.Fatal("review evidence id is empty")
		}
		if body.Review.Reviewer != "netops" {
			t.Fatalf("reviewer = %q, want netops", body.Review.Reviewer)
		}
		if body.Review.ResultingReviewState != tt.wantState {
			t.Fatalf("resulting state = %q, want %q", body.Review.ResultingReviewState, tt.wantState)
		}
		if !strings.Contains(body.Review.Effect, "merge") || !strings.Contains(body.Review.Effect, "identity rewrite") {
			t.Fatalf("effect = %q, want non-destructive merge/rewrite statement", body.Review.Effect)
		}
	}
	if got, want := len(repo.reviews), 2; got != want {
		t.Fatalf("audit review count = %d, want %d", got, want)
	}
	if repo.items[0].ReviewState != parser.IdentityReviewAccepted {
		t.Fatalf("first candidate state = %q, want accepted", repo.items[0].ReviewState)
	}
	if repo.items[1].ReviewState != parser.IdentityReviewMoreEvidence {
		t.Fatalf("second candidate state = %q, want more_evidence_requested", repo.items[1].ReviewState)
	}
}

func TestIdentityReviewHandoffReport(t *testing.T) {
	repo := &fakeIdentityCandidateRepository{
		items: []parser.IdentityCandidate{
			{
				ID:                   "candidate-reviewed",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-reviewed",
				ParserName:           "junos_show_version",
				AssetType:            "device",
				CandidateIdentityKey: "device:hostname:mx-edge-01",
				Strength:             assets.IdentityStrengthProvisional,
				Confidence:           0.55,
				Reason:               "hostname is not globally unique and may change",
				ReviewState:          parser.IdentityReviewAccepted,
				Metadata:             json.RawMessage(`{"identity_review_explanation":"reviewed by operator"}`),
				CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:                   "candidate-auto",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-auto",
				ParserName:           "junos_show_chassis_hardware",
				AssetType:            "chassis",
				CandidateIdentityKey: "chassis:vendor_serial:juniper:jn1234abcdef",
				Strength:             assets.IdentityStrengthStrong,
				Confidence:           0.85,
				Reason:               "identity key uses a durable identifier",
				Serial:               stringPtr("JN1234ABCDEF"),
				ReviewState:          parser.IdentityReviewAutoAccepted,
				Metadata:             json.RawMessage(`{"identity_review_explanation":"auto-accepted no conflict"}`),
				CreatedAt:            time.Date(2026, 6, 10, 12, 1, 0, 0, time.UTC),
			},
			{
				ID:                   "candidate-pending",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-missing",
				ParserName:           "junos_show_lldp_neighbors",
				AssetType:            "device",
				CandidateIdentityKey: "device:hostname:spine-01",
				Strength:             assets.IdentityStrengthProvisional,
				Confidence:           0.65,
				Reason:               "hostname is not globally unique and may change",
				ReviewState:          parser.IdentityReviewPending,
				Metadata:             json.RawMessage(`{"identity_review_explanation":"queued conflict needs review"}`),
				CreatedAt:            time.Date(2026, 6, 10, 12, 2, 0, 0, time.UTC),
			},
		},
		reviews: []parser.IdentityCandidateReview{
			{
				ID:                   "review-reviewed",
				IdentityCandidateID:  "candidate-reviewed",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-reviewed",
				Reviewer:             "netops",
				Action:               parser.IdentityReviewActionAccept,
				PreviousReviewState:  parser.IdentityReviewPending,
				ResultingReviewState: parser.IdentityReviewAccepted,
				Rationale:            "operator accepted hostname identity for handoff context",
				Effect:               parser.IdentityReviewEffect(parser.IdentityReviewActionAccept),
				Metadata:             json.RawMessage(`{"ticket":"TW-1"}`),
				CreatedAt:            time.Date(2026, 6, 10, 13, 0, 0, 0, time.UTC),
			},
			{
				ID:                   "review-auto",
				IdentityCandidateID:  "candidate-auto",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-auto",
				Reviewer:             "parser:auto_acceptance",
				Action:               parser.IdentityReviewActionAutoAccept,
				PreviousReviewState:  parser.IdentityReviewPending,
				ResultingReviewState: parser.IdentityReviewAutoAccepted,
				Rationale:            "auto-accepted because durable identity has no plausible conflict",
				Effect:               parser.IdentityReviewEffect(parser.IdentityReviewActionAutoAccept),
				Metadata:             json.RawMessage(`{"decision_type":"deterministic_auto_acceptance"}`),
				CreatedAt:            time.Date(2026, 6, 10, 13, 1, 0, 0, time.UTC),
			},
			{
				ID:                   "review-orphan",
				IdentityCandidateID:  "candidate-missing",
				DiscoveryRunID:       "11111111-1111-4111-8111-111111111111",
				EvidenceID:           "evidence-reviewed",
				Reviewer:             "netops",
				Action:               parser.IdentityReviewActionReject,
				PreviousReviewState:  parser.IdentityReviewPending,
				ResultingReviewState: parser.IdentityReviewRejected,
				Rationale:            "orphan fixture",
				Effect:               parser.IdentityReviewEffect(parser.IdentityReviewActionReject),
				Metadata:             json.RawMessage(`{}`),
				CreatedAt:            time.Date(2026, 6, 10, 13, 2, 0, 0, time.UTC),
			},
		},
		evidencePresent: map[string]bool{
			"evidence-reviewed": true,
			"evidence-auto":     true,
			"evidence-missing":  false,
		},
	}
	service := parser.NewIdentityCandidateService(repo)
	handler := NewHandler(Options{
		Version:            "test-version",
		IdentityCandidates: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/identity-candidates/handoff-report", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Report parser.IdentityReviewHandoffReport `json:"identity_review_handoff"`
	}](t, response)
	if body.Report.ReportType != "identity_review_handoff" {
		t.Fatalf("report type = %q, want identity_review_handoff", body.Report.ReportType)
	}
	if !body.Report.DerivedOutput {
		t.Fatal("report is not labeled as derived output")
	}
	if !strings.Contains(body.Report.Boundary, "not an accepted ADR") {
		t.Fatalf("boundary = %q, want Mistspren non-authoritative label", body.Report.Boundary)
	}
	if got, want := len(body.Report.Entries), 3; got != want {
		t.Fatalf("entry count = %d, want %d", got, want)
	}

	reviewed := findHandoffEntry(t, body.Report.Entries, "candidate-reviewed")
	if reviewed.LatestReview == nil || reviewed.LatestReview.ID != "review-reviewed" {
		t.Fatalf("reviewed latest review = %#v, want review-reviewed", reviewed.LatestReview)
	}
	if reviewed.HandoffStatus != "ready_for_mistspren_review" {
		t.Fatalf("reviewed handoff status = %q, want ready", reviewed.HandoffStatus)
	}
	if reviewed.EvidenceReference.EvidenceID != "evidence-reviewed" || !reviewed.EvidenceReference.Present {
		t.Fatalf("reviewed evidence ref = %#v, want present evidence-reviewed", reviewed.EvidenceReference)
	}
	if !strings.Contains(reviewed.IdentityEffect, "no canonical asset merge") {
		t.Fatalf("reviewed effect = %q, want non-destructive effect", reviewed.IdentityEffect)
	}

	auto := findHandoffEntry(t, body.Report.Entries, "candidate-auto")
	if auto.LatestReview == nil || auto.LatestReview.Action != parser.IdentityReviewActionAutoAccept {
		t.Fatalf("auto latest review = %#v, want auto_accept", auto.LatestReview)
	}
	if !strings.Contains(auto.ReviewSummary, "auto-accepted") {
		t.Fatalf("auto summary = %q, want auto acceptance rationale", auto.ReviewSummary)
	}

	pending := findHandoffEntry(t, body.Report.Entries, "candidate-pending")
	if pending.HandoffStatus != "unresolved_pending_review" {
		t.Fatalf("pending handoff status = %q, want unresolved", pending.HandoffStatus)
	}
	if pending.LatestReview != nil {
		t.Fatalf("pending latest review = %#v, want nil", pending.LatestReview)
	}
	if pending.EvidenceReference.Present {
		t.Fatal("pending missing evidence reference is marked present")
	}
	if len(pending.IntegrityWarnings) == 0 {
		t.Fatal("pending missing evidence entry has no integrity warnings")
	}

	if body.Report.Integrity.MissingEvidenceReferences != 1 {
		t.Fatalf("missing evidence count = %d, want 1", body.Report.Integrity.MissingEvidenceReferences)
	}
	if body.Report.Integrity.OrphanedReviewRecords != 1 {
		t.Fatalf("orphan count = %d, want 1", body.Report.Integrity.OrphanedReviewRecords)
	}
	if body.Report.Integrity.UnresolvedPendingEntries != 1 {
		t.Fatalf("pending count = %d, want 1", body.Report.Integrity.UnresolvedPendingEntries)
	}
}

func TestIdentityCandidatesEndpointRequiresRepository(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/identity-candidates", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestParseDiscoveryRunEndpointRequiresParserStore(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-runs/11111111-1111-4111-8111-111111111111/parse", strings.NewReader(`{"platform":"junos"}`))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

type fakeParseResultRepository struct {
	records []parser.ParseRecord
}

func (f *fakeParseResultRepository) CreateParseResult(ctx context.Context, params parser.CreateParseResultParams) (parser.ParseRecord, error) {
	warnings, err := json.Marshal(params.Warnings)
	if err != nil {
		return parser.ParseRecord{}, err
	}
	record := parser.ParseRecord{
		ID:             "parse-result-" + params.EvidenceID,
		DiscoveryRunID: params.DiscoveryRunID,
		EvidenceID:     params.EvidenceID,
		ParserName:     params.ParserName,
		Status:         params.Status,
		Warnings:       warnings,
		ErrorMessage:   params.ErrorMessage,
		CreatedAt:      time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	f.records = append(f.records, record)
	return record, nil
}

type fakeIdentityCandidateRepository struct {
	items           []parser.IdentityCandidate
	reviews         []parser.IdentityCandidateReview
	evidencePresent map[string]bool
}

func (f *fakeIdentityCandidateRepository) CreateIdentityCandidate(ctx context.Context, params parser.CreateIdentityCandidateParams) (parser.IdentityCandidate, error) {
	item := parser.IdentityCandidate{
		ID:                   "identity-candidate-created",
		DiscoveryRunID:       params.DiscoveryRunID,
		EvidenceID:           params.EvidenceID,
		ParserName:           params.ParserName,
		AssetType:            params.AssetType,
		CandidateIdentityKey: params.CandidateIdentityKey,
		Strength:             params.Strength,
		Confidence:           params.Confidence,
		Reason:               params.Reason,
		Vendor:               params.Vendor,
		Model:                params.Model,
		Serial:               params.Serial,
		SystemMAC:            params.SystemMAC,
		Hostname:             params.Hostname,
		ProposedAssetID:      params.ProposedAssetID,
		ReviewState:          params.ReviewState,
		Metadata:             params.Metadata,
		CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	f.items = append(f.items, item)
	return item, nil
}

func (f *fakeIdentityCandidateRepository) GetIdentityCandidate(ctx context.Context, id string) (parser.IdentityCandidate, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	return parser.IdentityCandidate{}, assets.ErrNotFound
}

func (f *fakeIdentityCandidateRepository) ListIdentityCandidates(ctx context.Context, filters parser.IdentityCandidateFilters) ([]parser.IdentityCandidate, error) {
	var result []parser.IdentityCandidate
	for _, item := range f.items {
		if filters.DiscoveryRunID != "" && item.DiscoveryRunID != filters.DiscoveryRunID {
			continue
		}
		if filters.EvidenceID != "" && item.EvidenceID != filters.EvidenceID {
			continue
		}
		if filters.ReviewState != "" && item.ReviewState != filters.ReviewState {
			continue
		}
		if filters.Strength != "" && item.Strength != filters.Strength {
			continue
		}
		if filters.CandidateIdentityKey != "" && item.CandidateIdentityKey != filters.CandidateIdentityKey {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (f *fakeIdentityCandidateRepository) ReviewIdentityCandidate(ctx context.Context, params parser.ReviewIdentityCandidateParams) (parser.IdentityCandidateReview, error) {
	for i := range f.items {
		if f.items[i].ID != params.IdentityCandidateID {
			continue
		}
		previous := f.items[i].ReviewState
		resulting := parser.ResultingReviewState(params.Action)
		f.items[i].ReviewState = resulting
		review := parser.IdentityCandidateReview{
			ID:                   "review-" + params.IdentityCandidateID,
			IdentityCandidateID:  params.IdentityCandidateID,
			DiscoveryRunID:       f.items[i].DiscoveryRunID,
			EvidenceID:           f.items[i].EvidenceID,
			Reviewer:             params.Reviewer,
			Action:               params.Action,
			PreviousReviewState:  previous,
			ResultingReviewState: resulting,
			Rationale:            params.Rationale,
			Effect:               parser.IdentityReviewEffect(params.Action),
			Metadata:             params.Metadata,
			CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		}
		f.reviews = append(f.reviews, review)
		return review, nil
	}
	return parser.IdentityCandidateReview{}, assets.ErrNotFound
}

func (f *fakeIdentityCandidateRepository) AutoAcceptIdentityCandidate(ctx context.Context, params parser.AutoAcceptIdentityCandidateParams) error {
	for i := range f.items {
		if f.items[i].ID != params.IdentityCandidateID {
			continue
		}
		if f.items[i].ReviewState != parser.IdentityReviewPending {
			return nil
		}
		previous := f.items[i].ReviewState
		f.items[i].ReviewState = parser.IdentityReviewAutoAccepted
		f.reviews = append(f.reviews, parser.IdentityCandidateReview{
			ID:                   "review-" + params.IdentityCandidateID,
			IdentityCandidateID:  params.IdentityCandidateID,
			DiscoveryRunID:       f.items[i].DiscoveryRunID,
			EvidenceID:           f.items[i].EvidenceID,
			Reviewer:             "parser:auto_acceptance",
			Action:               parser.IdentityReviewActionAutoAccept,
			PreviousReviewState:  previous,
			ResultingReviewState: parser.IdentityReviewAutoAccepted,
			Rationale:            params.Rationale,
			Effect:               parser.IdentityReviewEffect(parser.IdentityReviewActionAutoAccept),
			Metadata:             params.Metadata,
			CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		})
		return nil
	}
	return assets.ErrNotFound
}

func (f *fakeIdentityCandidateRepository) ListIdentityReviewHandoffEntries(ctx context.Context, filters parser.IdentityReviewHandoffFilters) ([]parser.IdentityReviewHandoffEntry, error) {
	var result []parser.IdentityReviewHandoffEntry
	for _, item := range f.items {
		if filters.DiscoveryRunID != "" && item.DiscoveryRunID != filters.DiscoveryRunID {
			continue
		}
		if filters.EvidenceID != "" && item.EvidenceID != filters.EvidenceID {
			continue
		}
		present := true
		if f.evidencePresent != nil {
			present = f.evidencePresent[item.EvidenceID]
		}
		entry := parser.IdentityReviewHandoffEntry{
			Candidate: item,
			EvidenceReference: parser.IdentityEvidenceRef{
				EvidenceID:     item.EvidenceID,
				DiscoveryRunID: item.DiscoveryRunID,
				Present:        present,
			},
			ParserSource: parser.IdentityParserSource{
				ParserName: item.ParserName,
				Metadata:   item.Metadata,
			},
		}
		if review, ok := f.latestReviewForCandidate(item.ID); ok {
			entry.LatestReview = &review
		}
		result = append(result, entry)
	}
	return result, nil
}

func (f *fakeIdentityCandidateRepository) ListOrphanedIdentityCandidateReviews(ctx context.Context) ([]parser.IdentityCandidateReview, error) {
	var result []parser.IdentityCandidateReview
	for _, review := range f.reviews {
		found := false
		for _, item := range f.items {
			if item.ID == review.IdentityCandidateID {
				found = true
				break
			}
		}
		if !found {
			result = append(result, review)
		}
	}
	return result, nil
}

func (f *fakeIdentityCandidateRepository) latestReviewForCandidate(candidateID string) (parser.IdentityCandidateReview, bool) {
	var latest parser.IdentityCandidateReview
	found := false
	for _, review := range f.reviews {
		if review.IdentityCandidateID != candidateID {
			continue
		}
		if !found || review.CreatedAt.After(latest.CreatedAt) || (review.CreatedAt.Equal(latest.CreatedAt) && review.ID > latest.ID) {
			latest = review
			found = true
		}
	}
	return latest, found
}

func (f *fakeParseResultRepository) ListParseResultsByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]parser.ParseRecord, error) {
	var result []parser.ParseRecord
	for _, item := range f.records {
		if item.DiscoveryRunID == discoveryRunID {
			result = append(result, item)
		}
	}
	return result, nil
}

func TestGetEvidence(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	id := "22222222-2222-4222-8222-222222222222"
	repo := &fakeEvidenceRepository{
		items: []evidence.Evidence{{
			ID:             id,
			DiscoveryRunID: "11111111-1111-4111-8111-111111111111",
			Target:         "router1",
			Method:         "ssh",
			CommandOrAPI:   "show version",
			RawOutput:      "raw output",
			RawOutputHash:  evidence.HashRawOutput("raw output"),
			CollectedAt:    now,
			Metadata:       json.RawMessage(`{}`),
		}},
	}
	service := evidence.NewService(repo)
	handler := NewHandler(Options{
		Version:  "test-version",
		Evidence: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/evidence/"+id, nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeResponseData[struct {
		Evidence evidence.Evidence `json:"evidence"`
	}](t, response)
	if body.Evidence.ID != id {
		t.Fatalf("id = %q, want %q", body.Evidence.ID, id)
	}
}

func TestGetEvidenceNotFound(t *testing.T) {
	repo := &fakeEvidenceRepository{}
	service := evidence.NewService(repo)
	handler := NewHandler(Options{
		Version:  "test-version",
		Evidence: &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/evidence/missing", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestEvidenceEndpointsReturnUnavailableWithoutRepository(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/evidence/22222222-2222-4222-8222-222222222222", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

type fakeAssetRepository struct {
	assets        []assets.Asset
	facts         []assets.Fact
	relationships []assets.Relationship
}

func (f *fakeAssetRepository) CreateAsset(ctx context.Context, params assets.CreateAssetParams) (assets.Asset, error) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	item := assets.Asset{
		ID:               "asset-created",
		Type:             params.Type,
		IdentityKey:      params.IdentityKey,
		Vendor:           params.Vendor,
		Model:            params.Model,
		Serial:           params.Serial,
		SystemMAC:        params.SystemMAC,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		Metadata:         params.Metadata,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	f.assets = append(f.assets, item)
	return item, nil
}

func (f *fakeAssetRepository) GetAsset(ctx context.Context, id string) (assets.Asset, error) {
	for _, item := range f.assets {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Asset{}, assets.ErrNotFound
}

func (f *fakeAssetRepository) ListAssets(ctx context.Context) ([]assets.Asset, error) {
	return f.assets, nil
}

func (f *fakeAssetRepository) CreateFact(ctx context.Context, params assets.CreateFactParams) (assets.Fact, error) {
	item := assets.Fact{
		ID:               "fact-created-" + params.Name,
		AssetID:          params.AssetID,
		Name:             params.Name,
		Value:            params.Value,
		Source:           params.Source,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		EvidenceID:       params.EvidenceID,
		CreatedAt:        time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	f.facts = append(f.facts, item)
	return item, nil
}

func (f *fakeAssetRepository) GetFact(ctx context.Context, id string) (assets.Fact, error) {
	for _, item := range f.facts {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Fact{}, assets.ErrNotFound
}

func (f *fakeAssetRepository) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	var result []assets.Fact
	for _, item := range f.facts {
		if item.AssetID == assetID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (f *fakeAssetRepository) CreateRelationship(ctx context.Context, params assets.CreateRelationshipParams) (assets.Relationship, error) {
	item := assets.Relationship{
		ID:               "relationship-created",
		SourceAssetID:    params.SourceAssetID,
		TargetAssetID:    params.TargetAssetID,
		RelationshipType: params.RelationshipType,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		EvidenceID:       params.EvidenceID,
		Metadata:         params.Metadata,
		CreatedAt:        time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	f.relationships = append(f.relationships, item)
	return item, nil
}

func (f *fakeAssetRepository) GetRelationship(ctx context.Context, id string) (assets.Relationship, error) {
	for _, item := range f.relationships {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Relationship{}, assets.ErrNotFound
}

func (f *fakeAssetRepository) ListRelationships(ctx context.Context) ([]assets.Relationship, error) {
	return f.relationships, nil
}

func TestListAssetsWithFiltersAndPagination(t *testing.T) {
	service := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets?type=device&vendor=juniper&limit=1&offset=0", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Assets []assets.Asset `json:"assets"`
	}](t, response)
	if got, want := len(body.Assets), 1; got != want {
		t.Fatalf("asset count = %d, want %d", got, want)
	}
	if body.Assets[0].ID != "asset-a" {
		t.Fatalf("asset id = %q, want asset-a", body.Assets[0].ID)
	}

	envelope := decodeResponseEnvelope(t, response)
	pagination, ok := envelope.Metadata["pagination"].(map[string]any)
	if !ok {
		t.Fatalf("pagination metadata = %#v, want object", envelope.Metadata["pagination"])
	}
	if pagination["limit"] != float64(1) {
		t.Fatalf("limit = %#v, want 1", pagination["limit"])
	}
	if pagination["total"] != float64(2) {
		t.Fatalf("total = %#v, want 2", pagination["total"])
	}
	if pagination["has_next"] != true {
		t.Fatalf("has_next = %#v, want true", pagination["has_next"])
	}
}

func TestListAssetsSupportsFreeTextSearch(t *testing.T) {
	service := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets?q=nyc", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Assets []assets.Asset `json:"assets"`
	}](t, response)
	if got, want := len(body.Assets), 1; got != want {
		t.Fatalf("asset count = %d, want %d", got, want)
	}
	if body.Assets[0].ID != "asset-c" {
		t.Fatalf("asset id = %q, want asset-c", body.Assets[0].ID)
	}
}

func TestListAssetsRejectsInvalidPagination(t *testing.T) {
	service := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets?limit=-1", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestGetAssetHistoryIncludesFactsAndRelationships(t *testing.T) {
	service := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/asset-a/history", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Asset   assets.Asset           `json:"asset"`
		History []historyEventResponse `json:"history"`
	}](t, response)
	if body.Asset.ID != "asset-a" {
		t.Fatalf("asset id = %q, want asset-a", body.Asset.ID)
	}
	if got, want := len(body.History), 3; got != want {
		t.Fatalf("history count = %d, want %d", got, want)
	}
	if !hasHistoryEvent(body.History, "asset_created", "asset-a", "") {
		t.Fatalf("history missing asset_created event: %#v", body.History)
	}
	if !hasHistoryEvent(body.History, "fact_observed", "fact-a", "evidence-a") {
		t.Fatalf("history missing fact_observed evidence event: %#v", body.History)
	}
	if !hasHistoryEvent(body.History, "relationship_observed", "relationship-a", "evidence-b") {
		t.Fatalf("history missing relationship_observed evidence event: %#v", body.History)
	}
}

type historyEventResponse struct {
	EventType      string  `json:"event_type"`
	RecordID       string  `json:"record_id"`
	Name           string  `json:"name"`
	EvidenceID     *string `json:"evidence_id"`
	RelationshipTo *string `json:"relationship_to"`
}

func hasHistoryEvent(events []historyEventResponse, eventType string, recordID string, evidenceID string) bool {
	for _, event := range events {
		if event.EventType != eventType || event.RecordID != recordID {
			continue
		}
		if evidenceID == "" {
			return event.EvidenceID == nil
		}
		return event.EvidenceID != nil && *event.EvidenceID == evidenceID
	}
	return false
}

func TestListProvisionalIdentityAssets(t *testing.T) {
	service := assets.NewService(&fakeAssetRepository{
		assets: []assets.Asset{
			{
				ID:          "asset-strong",
				Type:        "device",
				IdentityKey: "device:vendor_serial:juniper:jn1234",
				Metadata:    json.RawMessage(`{"identity_strength":"strong"}`),
			},
			{
				ID:          "asset-provisional",
				Type:        "device",
				IdentityKey: "device:hostname:mx-edge-01",
				Metadata:    json.RawMessage(`{"identity_strength":"provisional"}`),
			},
		},
	})
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/provisional-identities", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Assets []assets.Asset `json:"assets"`
	}](t, response)
	if got, want := len(body.Assets), 1; got != want {
		t.Fatalf("asset count = %d, want %d", got, want)
	}
	if body.Assets[0].ID != "asset-provisional" {
		t.Fatalf("asset id = %q, want asset-provisional", body.Assets[0].ID)
	}
}

func TestListConflictingFacts(t *testing.T) {
	service := assets.NewService(&fakeAssetRepository{
		assets: []assets.Asset{{ID: "asset-a"}},
		facts: []assets.Fact{
			{ID: "fact-ok", AssetID: "asset-a", Name: "hostname", State: assets.StateObserved},
			{ID: "fact-conflict", AssetID: "asset-a", Name: "hostname", State: assets.StateConflicting},
		},
	})
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/facts/conflicts", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Facts []assets.Fact `json:"facts"`
	}](t, response)
	if got, want := len(body.Facts), 1; got != want {
		t.Fatalf("fact count = %d, want %d", got, want)
	}
	if body.Facts[0].ID != "fact-conflict" {
		t.Fatalf("fact id = %q, want fact-conflict", body.Facts[0].ID)
	}
}

func TestGetAsset(t *testing.T) {
	service := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/asset-a", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Asset assets.Asset `json:"asset"`
	}](t, response)
	if body.Asset.IdentityKey != "device:serial:aaa" {
		t.Fatalf("identity_key = %q, want device:serial:aaa", body.Asset.IdentityKey)
	}
}

func TestListAssetFacts(t *testing.T) {
	service := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/asset-a/facts", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Facts []assets.Fact `json:"facts"`
	}](t, response)
	if got, want := len(body.Facts), 1; got != want {
		t.Fatalf("fact count = %d, want %d", got, want)
	}
	if body.Facts[0].State != assets.StateObserved {
		t.Fatalf("fact state = %q, want %q", body.Facts[0].State, assets.StateObserved)
	}
	if body.Facts[0].ConfidenceReason == "" {
		t.Fatal("fact confidence_reason is empty")
	}
	if body.Facts[0].EvidenceID == nil {
		t.Fatal("fact evidence_id is nil")
	}
}

func TestListAssetRelationships(t *testing.T) {
	service := assets.NewService(testAssetRepository())
	handler := NewHandler(Options{
		Version: "test-version",
		Assets:  &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/asset-a/relationships", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Relationships []assets.Relationship `json:"relationships"`
	}](t, response)
	if got, want := len(body.Relationships), 1; got != want {
		t.Fatalf("relationship count = %d, want %d", got, want)
	}
	if body.Relationships[0].State != assets.StateObserved {
		t.Fatalf("relationship state = %q, want %q", body.Relationships[0].State, assets.StateObserved)
	}
	if body.Relationships[0].EvidenceID == nil {
		t.Fatal("relationship evidence_id is nil")
	}
}

func TestListAssetEvidence(t *testing.T) {
	assetService := assets.NewService(testAssetRepository())
	evidenceService := evidence.NewService(&fakeEvidenceRepository{items: []evidence.Evidence{
		testEvidence("evidence-a"),
		testEvidence("evidence-b"),
	}})
	handler := NewHandler(Options{
		Version:  "test-version",
		Assets:   &assetService,
		Evidence: &evidenceService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/asset-a/evidence?limit=1", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Evidence []evidence.Evidence `json:"evidence"`
	}](t, response)
	if got, want := len(body.Evidence), 1; got != want {
		t.Fatalf("evidence count = %d, want %d", got, want)
	}
}

func TestListFactEvidence(t *testing.T) {
	assetService := assets.NewService(testAssetRepository())
	evidenceService := evidence.NewService(&fakeEvidenceRepository{items: []evidence.Evidence{
		testEvidence("evidence-a"),
	}})
	handler := NewHandler(Options{
		Version:  "test-version",
		Assets:   &assetService,
		Evidence: &evidenceService,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/facts/fact-a/evidence", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Evidence []evidence.Evidence `json:"evidence"`
	}](t, response)
	if got, want := len(body.Evidence), 1; got != want {
		t.Fatalf("evidence count = %d, want %d", got, want)
	}
	if body.Evidence[0].ID != "evidence-a" {
		t.Fatalf("evidence id = %q, want evidence-a", body.Evidence[0].ID)
	}
}

func TestAssetEndpointsReturnUnavailableWithoutRepository(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

type fakeGraphAssetReader struct {
	assets        map[string]assets.Asset
	facts         map[string][]assets.Fact
	relationships []assets.Relationship
}

func testAssetRepository() *fakeAssetRepository {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	evidenceA := "evidence-a"
	evidenceB := "evidence-b"
	return &fakeAssetRepository{
		assets: []assets.Asset{
			{
				ID:               "asset-a",
				Type:             "device",
				IdentityKey:      "device:serial:aaa",
				Vendor:           stringPtr("juniper"),
				Serial:           stringPtr("aaa"),
				Confidence:       0.9,
				ConfidenceReason: "directly observed from evidence",
				State:            assets.StateObserved,
				Metadata:         json.RawMessage(`{}`),
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			{
				ID:               "asset-b",
				Type:             "device",
				IdentityKey:      "device:serial:bbb",
				Vendor:           stringPtr("juniper"),
				Serial:           stringPtr("bbb"),
				Confidence:       0.8,
				ConfidenceReason: "directly observed from evidence",
				State:            assets.StateObserved,
				Metadata:         json.RawMessage(`{}`),
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			{
				ID:               "asset-c",
				Type:             "site",
				IdentityKey:      "site:code:nyc",
				Vendor:           stringPtr("internal"),
				Confidence:       0.5,
				ConfidenceReason: "deterministically inferred without direct evidence",
				State:            assets.StateInferred,
				Metadata:         json.RawMessage(`{}`),
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
		facts: []assets.Fact{{
			ID:               "fact-a",
			AssetID:          "asset-a",
			Name:             "hostname",
			Value:            json.RawMessage(`"router-a"`),
			Source:           "parser",
			Confidence:       0.95,
			ConfidenceReason: "directly observed from evidence",
			State:            assets.StateObserved,
			EvidenceID:       &evidenceA,
			CreatedAt:        now,
		}},
		relationships: []assets.Relationship{{
			ID:               "relationship-a",
			SourceAssetID:    "asset-a",
			TargetAssetID:    "asset-b",
			RelationshipType: "lldp_neighbor",
			Confidence:       0.9,
			ConfidenceReason: "directly observed from evidence",
			State:            assets.StateObserved,
			EvidenceID:       &evidenceB,
			Metadata:         json.RawMessage(`{}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}},
	}
}

func testEvidence(id string) evidence.Evidence {
	return evidence.Evidence{
		ID:             id,
		DiscoveryRunID: "11111111-1111-4111-8111-111111111111",
		Target:         "router1",
		Method:         "ssh",
		CommandOrAPI:   "show version",
		RawOutput:      "raw output",
		RawOutputHash:  evidence.HashRawOutput("raw output"),
		CollectedAt:    time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Metadata:       json.RawMessage(`{}`),
	}
}

func stringPtr(value string) *string {
	return &value
}

func findHandoffEntry(t *testing.T, entries []parser.IdentityReviewHandoffEntry, candidateID string) parser.IdentityReviewHandoffEntry {
	t.Helper()
	for _, entry := range entries {
		if entry.Candidate.ID == candidateID {
			return entry
		}
	}
	t.Fatalf("handoff entry for candidate %q not found in %#v", candidateID, entries)
	return parser.IdentityReviewHandoffEntry{}
}

func (f fakeGraphAssetReader) GetAsset(ctx context.Context, id string) (assets.Asset, error) {
	item, ok := f.assets[id]
	if !ok {
		return assets.Asset{}, assets.ErrNotFound
	}
	return item, nil
}

func (f fakeGraphAssetReader) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	return f.facts[assetID], nil
}

func (f fakeGraphAssetReader) ListRelationships(ctx context.Context) ([]assets.Relationship, error) {
	return f.relationships, nil
}

func TestGetAssetGraph(t *testing.T) {
	service := graph.NewService(testGraphAssetReader())
	handler := NewHandler(Options{
		Version: "test-version",
		Graph:   &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/asset-a/graph", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Graph graph.Graph `json:"graph"`
	}](t, response)
	if got, want := len(body.Graph.Nodes), 2; got != want {
		t.Fatalf("node count = %d, want %d", got, want)
	}
	if got, want := len(body.Graph.Edges), 1; got != want {
		t.Fatalf("edge count = %d, want %d", got, want)
	}
	if body.Graph.Nodes[0].Label != "router-a" {
		t.Fatalf("root label = %q, want router-a", body.Graph.Nodes[0].Label)
	}
}

func TestGetGraphNeighborsRequiresAssetID(t *testing.T) {
	service := graph.NewService(testGraphAssetReader())
	handler := NewHandler(Options{
		Version: "test-version",
		Graph:   &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/graph/neighbors", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestGetGraphNeighbors(t *testing.T) {
	service := graph.NewService(testGraphAssetReader())
	handler := NewHandler(Options{
		Version: "test-version",
		Graph:   &service,
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/graph/neighbors?asset_id=asset-a", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := decodeResponseData[struct {
		Graph graph.Graph `json:"graph"`
	}](t, response)
	if got, want := body.Graph.Edges[0].RelationshipType, "lldp_neighbor"; got != want {
		t.Fatalf("relationship_type = %q, want %q", got, want)
	}
}

func TestGraphEndpointsReturnUnavailableWithoutRepository(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/asset-a/graph", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func testGraphAssetReader() fakeGraphAssetReader {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	return fakeGraphAssetReader{
		assets: map[string]assets.Asset{
			"asset-a": {
				ID:          "asset-a",
				Type:        "device",
				IdentityKey: "device:serial:aaa",
				Metadata:    json.RawMessage(`{}`),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			"asset-b": {
				ID:          "asset-b",
				Type:        "device",
				IdentityKey: "device:serial:bbb",
				Metadata:    json.RawMessage(`{}`),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		facts: map[string][]assets.Fact{
			"asset-a": {{
				ID:         "fact-a",
				AssetID:    "asset-a",
				Name:       "hostname",
				Value:      json.RawMessage(`"router-a"`),
				Source:     "parser",
				Confidence: 0.95,
				CreatedAt:  now,
			}},
		},
		relationships: []assets.Relationship{{
			ID:               "rel-a-b",
			SourceAssetID:    "asset-a",
			TargetAssetID:    "asset-b",
			RelationshipType: "lldp_neighbor",
			Confidence:       0.9,
			Metadata:         json.RawMessage(`{}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}},
	}
}
