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

	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
)

func TestHealthz(t *testing.T) {
	handler := NewHandler(Options{Version: "test-version"})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
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

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
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

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["name"] != "truthwatcher" {
		t.Fatalf("name = %q, want truthwatcher", body["name"])
	}
	if body["version"] != "test-version" {
		t.Fatalf("version = %q, want test-version", body["version"])
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

	var body struct {
		DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
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
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &runService,
		Evidence:      &evidenceService,
	})

	requestBody := `{
		"collector": "fake",
		"target": "fixture://junos-mx",
		"tasks": ["identify_device"],
		"fixture_root": "../../examples/fixtures"
	}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/discovery-runs/execute", strings.NewReader(requestBody))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	var body struct {
		DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
		Evidence     []evidence.Evidence    `json:"evidence"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.DiscoveryRun.Status != discovery.StatusCompleted {
		t.Fatalf("status = %q, want completed", body.DiscoveryRun.Status)
	}
	if got, want := len(body.Evidence), 1; got != want {
		t.Fatalf("evidence count = %d, want %d", got, want)
	}
	if body.Evidence[0].CommandOrAPI != "show version" {
		t.Fatalf("command = %q, want show version", body.Evidence[0].CommandOrAPI)
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

	var body struct {
		DiscoveryRuns []discovery.DiscoveryRun `json:"discovery_runs"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
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

	var body struct {
		DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
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

	var body struct {
		Evidence []evidence.Evidence `json:"evidence"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Evidence) != 1 {
		t.Fatalf("len = %d, want 1", len(body.Evidence))
	}
	if body.Evidence[0].RawOutputHash == "" {
		t.Fatal("raw_output_hash is empty")
	}
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

	var body struct {
		Evidence evidence.Evidence `json:"evidence"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
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
