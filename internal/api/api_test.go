package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
