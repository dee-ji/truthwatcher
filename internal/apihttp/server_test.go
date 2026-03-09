package apihttp

import (
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

func TestHealthz(t *testing.T) {
	s := New(slog.New(slog.NewTextHandler(os.Stdout, nil)), intent.NewInMemoryService(), topology.NewStubService(), deploy.NewStubService(), reconcile.NewStubService(), audit.NewStubService())
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.Handler().ServeHTTP(r, req)
	if r.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}
