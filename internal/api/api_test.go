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
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/graph"
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
	handler := NewHandler(Options{
		Version:       "test-version",
		DiscoveryRuns: &runService,
		Evidence:      &evidenceService,
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
	return assets.Asset{}, nil
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
	return assets.Fact{}, nil
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
	return assets.Relationship{}, nil
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
				ID:          "asset-a",
				Type:        "device",
				IdentityKey: "device:serial:aaa",
				Vendor:      stringPtr("juniper"),
				Serial:      stringPtr("aaa"),
				Metadata:    json.RawMessage(`{}`),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          "asset-b",
				Type:        "device",
				IdentityKey: "device:serial:bbb",
				Vendor:      stringPtr("juniper"),
				Serial:      stringPtr("bbb"),
				Metadata:    json.RawMessage(`{}`),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          "asset-c",
				Type:        "site",
				IdentityKey: "site:code:nyc",
				Vendor:      stringPtr("internal"),
				Metadata:    json.RawMessage(`{}`),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		facts: []assets.Fact{{
			ID:         "fact-a",
			AssetID:    "asset-a",
			Name:       "hostname",
			Value:      json.RawMessage(`"router-a"`),
			Source:     "parser",
			Confidence: 0.95,
			EvidenceID: &evidenceA,
			CreatedAt:  now,
		}},
		relationships: []assets.Relationship{{
			ID:               "relationship-a",
			SourceAssetID:    "asset-a",
			TargetAssetID:    "asset-b",
			RelationshipType: "lldp_neighbor",
			Confidence:       0.9,
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
