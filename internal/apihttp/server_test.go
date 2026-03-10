package apihttp

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/reconcile"
	"github.com/truthwatcher/truthwatcher/internal/topology"
)

func testServer() *Server {
	return New(slog.New(slog.NewTextHandler(os.Stdout, nil)), intent.NewInMemoryService(), topology.NewStubService(), deploy.NewStubService(), reconcile.NewStubService(), audit.NewStubService())
}

func TestHealthz(t *testing.T) {
	s := testServer()
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.Handler().ServeHTTP(r, req)
	if r.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestIntentEndpoints(t *testing.T) {
	s := testServer()
	create := map[string]any{"name": "leaf", "spec": map[string]any{"metadata": map[string]any{"name": "leaf-1"}}}
	b, _ := json.Marshal(create)
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intents", bytes.NewReader(b))
	s.Handler().ServeHTTP(r, req)
	if r.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", r.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(r.Body.Bytes(), &resp)
	id := resp["id"].(string)

	validateRes := httptest.NewRecorder()
	validateReq := httptest.NewRequest(http.MethodPost, "/api/v1/intents/"+id+"/validate", nil)
	s.Handler().ServeHTTP(validateRes, validateReq)
	if validateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", validateRes.Code)
	}

	compileRes := httptest.NewRecorder()
	compileReq := httptest.NewRequest(http.MethodPost, "/api/v1/intents/"+id+"/compile", bytes.NewReader([]byte(`{"vendor":"junos"}`)))
	s.Handler().ServeHTTP(compileRes, compileReq)
	if compileRes.Code != http.StatusAccepted {
		t.Fatalf("expected 202 got %d", compileRes.Code)
	}
}
