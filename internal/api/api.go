package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"

	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
)

type Options struct {
	Version       string
	Logger        *slog.Logger
	DiscoveryRuns *discovery.Service
	Evidence      *evidence.Service
}

type responseEnvelope map[string]any

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
	mux.HandleFunc("GET /api/v1/discovery-runs", handleListDiscoveryRuns(opts.DiscoveryRuns))
	mux.HandleFunc("GET /api/v1/discovery-runs/{id}", handleGetDiscoveryRun(opts.DiscoveryRuns))
	mux.HandleFunc("GET /api/v1/discovery-runs/{id}/evidence", handleListEvidenceByDiscoveryRun(opts.Evidence))
	mux.HandleFunc("GET /api/v1/evidence/{id}", handleGetEvidence(opts.Evidence))

	return recoverPanic(opts.Logger, requestLog(opts.Logger, requestID(mux)))
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, responseEnvelope{
		"status": "ok",
	})
}

func handleReadyz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, responseEnvelope{
		"status": "ready",
	})
}

func handleVersion(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, responseEnvelope{
			"name":    "truthwatcher",
			"version": version,
		})
	}
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

		writeJSON(w, http.StatusCreated, responseEnvelope{"discovery_run": run})
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

		writeJSON(w, http.StatusOK, responseEnvelope{"discovery_runs": runs})
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

		writeJSON(w, http.StatusOK, responseEnvelope{"discovery_run": run})
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

		writeJSON(w, http.StatusOK, responseEnvelope{"evidence": items})
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

		writeJSON(w, http.StatusOK, responseEnvelope{"evidence": item})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload responseEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, responseEnvelope{
		"error": message,
	})
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = "req-" + strconv.FormatUint(atomic.AddUint64(&requestCounter, 1), 10)
		}

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
				http.Error(w, "internal server error", http.StatusInternalServerError)
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
