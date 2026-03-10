package apihttp

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/truthwatcher/truthwatcher/internal/audit"
	"github.com/truthwatcher/truthwatcher/internal/deploy"
	"github.com/truthwatcher/truthwatcher/internal/domain"
	"github.com/truthwatcher/truthwatcher/internal/intent"
	"github.com/truthwatcher/truthwatcher/internal/reconcile"
	"github.com/truthwatcher/truthwatcher/internal/state"
	"github.com/truthwatcher/truthwatcher/internal/topology"
)

func testServer() *Server {
	topo := topology.NewService(topology.NewInMemoryRepository())
	intentSvc := intent.NewInMemoryService()
	auditSvc := audit.NewStubService()
	_ = topo.Import(context.Background(), domain.TopologySnapshot{Devices: []domain.Device{{ID: "d1", Hostname: "leaf-1"}}})
	stateSvc := state.NewService(state.NewInMemoryRepository())
	reconcileSvc := reconcile.NewService(reconcile.NewInMemoryRepository(), intentSvc, stateSvc, auditSvc)
	return New(slog.New(slog.NewTextHandler(os.Stdout, nil)), intentSvc, topo, deploy.NewStubServiceWithDependencies(auditSvc, intentSvc), reconcileSvc, auditSvc)
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

func TestTopologyIntegrationEndpoints(t *testing.T) {
	s := testServer()
	fixture := domain.TopologySnapshot{
		Vendors:    []domain.Vendor{{ID: "v1", Name: "eos"}},
		Platforms:  []domain.Platform{{ID: "p1", VendorID: "v1", Name: "7050"}},
		Sites:      []domain.Site{{ID: "s1", Name: "dc1"}},
		Devices:    []domain.Device{{ID: "d1", Hostname: "leaf1", VendorID: "v1", PlatformID: "p1", SiteID: "s1"}, {ID: "d2", Hostname: "spine1", VendorID: "v1", PlatformID: "p1", SiteID: "s1"}},
		Interfaces: []domain.Interface{{ID: "i1", DeviceID: "d1", Name: "eth1"}, {ID: "i2", DeviceID: "d2", Name: "eth1"}},
		Links:      []domain.Link{{ID: "l1", AInterfaceID: "i1", ZInterfaceID: "i2"}},
	}
	b, _ := json.Marshal(fixture)
	importReq := httptest.NewRequest(http.MethodPost, "/api/v1/topology/import", bytes.NewReader(b))
	importRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(importRes, importReq)
	if importRes.Code != http.StatusAccepted {
		t.Fatalf("import expected 202 got %d", importRes.Code)
	}

	devRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(devRes, httptest.NewRequest(http.MethodGet, "/api/v1/topology/devices?site=dc1", nil))
	if devRes.Code != http.StatusOK || !bytes.Contains(devRes.Body.Bytes(), []byte("leaf1")) {
		t.Fatalf("devices endpoint failed: code=%d body=%s", devRes.Code, devRes.Body.String())
	}

	deviceRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(deviceRes, httptest.NewRequest(http.MethodGet, "/api/v1/topology/devices/d1", nil))
	if deviceRes.Code != http.StatusOK || !bytes.Contains(deviceRes.Body.Bytes(), []byte("adjacent_device_ids")) {
		t.Fatalf("device detail failed: code=%d body=%s", deviceRes.Code, deviceRes.Body.String())
	}
}

func TestDeploymentEndpoints(t *testing.T) {
	s := testServer()

	createIntentBody := []byte(`{"name":"deployable","spec":{"metadata":{"name":"deployable"}}}`)
	intentRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(intentRes, httptest.NewRequest(http.MethodPost, "/api/v1/intents", bytes.NewReader(createIntentBody)))
	if intentRes.Code != http.StatusCreated {
		t.Fatalf("intent create expected 201 got %d", intentRes.Code)
	}
	var intentResp map[string]any
	_ = json.Unmarshal(intentRes.Body.Bytes(), &intentResp)
	intentID := intentResp["id"].(string)

	compileRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(compileRes, httptest.NewRequest(http.MethodPost, "/api/v1/intents/"+intentID+"/compile", bytes.NewReader([]byte(`{"vendor":"junos"}`))))
	if compileRes.Code != http.StatusAccepted {
		t.Fatalf("intent compile expected 202 got %d", compileRes.Code)
	}

	reqBody := []byte(`{"intent_id":"` + intentID + `","idempotency_key":"deploy-1","mode":"dry-run","targets":["leaf-1","leaf-2"],"batch_size":1,"canary_targets":1,"require_manual_approval":true}`)
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/deployments", bytes.NewReader(reqBody)))
	if res.Code != http.StatusCreated {
		t.Fatalf("deployment create expected 201 got %d body=%s", res.Code, res.Body.String())
	}
	var dep domain.Deployment
	_ = json.Unmarshal(res.Body.Bytes(), &dep)
	if dep.Mode != "dry-run" {
		t.Fatalf("expected dry-run mode got %s", dep.Mode)
	}

	getRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(getRes, httptest.NewRequest(http.MethodGet, "/api/v1/deployments/"+dep.ID, nil))
	if getRes.Code != http.StatusOK {
		t.Fatalf("deployment get expected 200 got %d", getRes.Code)
	}
}

func TestDeploymentCreateValidation(t *testing.T) {
	s := testServer()
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/deployments", bytes.NewReader([]byte(`{"intent_id":"intent-1"}`))))
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}

func TestReconcileEndpoints(t *testing.T) {
	s := testServer()

	createIntentBody := []byte(`{"name":"recon","spec":{"metadata":{"name":"leaf-1"}}}`)
	intentRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(intentRes, httptest.NewRequest(http.MethodPost, "/api/v1/intents", bytes.NewReader(createIntentBody)))
	if intentRes.Code != http.StatusCreated {
		t.Fatalf("intent create expected 201 got %d", intentRes.Code)
	}
	var intentResp map[string]any
	_ = json.Unmarshal(intentRes.Body.Bytes(), &intentResp)
	intentID := intentResp["id"].(string)
	s.Handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/intents/"+intentID+"/compile", bytes.NewReader([]byte(`{"vendor":"junos"}`))))

	runRes := httptest.NewRecorder()
	body := []byte(`{"intent_id":"` + intentID + `"}`)
	s.Handler().ServeHTTP(runRes, httptest.NewRequest(http.MethodPost, "/api/v1/reconcile/runs", bytes.NewReader(body)))
	if runRes.Code != http.StatusAccepted {
		t.Fatalf("reconcile run expected 202 got %d body=%s", runRes.Code, runRes.Body.String())
	}
	var run domain.ReconcileRun
	_ = json.Unmarshal(runRes.Body.Bytes(), &run)

	getRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(getRes, httptest.NewRequest(http.MethodGet, "/api/v1/reconcile/runs/"+run.ID, nil))
	if getRes.Code != http.StatusOK {
		t.Fatalf("reconcile get expected 200 got %d", getRes.Code)
	}

	findingsRes := httptest.NewRecorder()
	s.Handler().ServeHTTP(findingsRes, httptest.NewRequest(http.MethodGet, "/api/v1/drift/findings", nil))
	if findingsRes.Code != http.StatusOK {
		t.Fatalf("drift findings expected 200 got %d", findingsRes.Code)
	}
}
