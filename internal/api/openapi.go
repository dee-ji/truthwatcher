package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

type routeDoc struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
}

var documentedRoutes = []routeDoc{
	{Method: http.MethodGet, Path: "/healthz", Summary: "Health check", Tags: []string{"System"}},
	{Method: http.MethodGet, Path: "/readyz", Summary: "Readiness check", Tags: []string{"System"}},
	{Method: http.MethodGet, Path: "/api/v1/version", Summary: "Get API version", Tags: []string{"System"}},
	{Method: http.MethodGet, Path: "/api/v1/system-info", Summary: "Get runtime system information", Tags: []string{"System"}},
	{Method: http.MethodPost, Path: "/api/v1/discovery-runs", Summary: "Create a discovery run", Tags: []string{"Discovery"}},
	{Method: http.MethodPost, Path: "/api/v1/discovery-runs/execute", Summary: "Execute a discovery run", Tags: []string{"Discovery"}},
	{Method: http.MethodGet, Path: "/api/v1/discovery-runs", Summary: "List discovery runs", Tags: []string{"Discovery"}},
	{Method: http.MethodGet, Path: "/api/v1/discovery-runs/{id}", Summary: "Get a discovery run", Tags: []string{"Discovery"}},
	{Method: http.MethodGet, Path: "/api/v1/discovery-runs/{id}/evidence", Summary: "List evidence for a discovery run", Tags: []string{"Discovery", "Evidence"}},
	{Method: http.MethodPost, Path: "/api/v1/discovery-runs/{id}/parse", Summary: "Parse a discovery run", Tags: []string{"Discovery", "Parser"}},
	{Method: http.MethodGet, Path: "/api/v1/identity-candidates", Summary: "List identity candidates", Tags: []string{"Identity"}},
	{Method: http.MethodGet, Path: "/api/v1/identity-candidates/review-queue", Summary: "List pending identity candidates", Tags: []string{"Identity"}},
	{Method: http.MethodGet, Path: "/api/v1/identity-candidates/handoff-report", Summary: "Get identity review handoff report", Tags: []string{"Identity"}},
	{Method: http.MethodPost, Path: "/api/v1/identity-candidates/{id}/review", Summary: "Review an identity candidate", Tags: []string{"Identity"}},
	{Method: http.MethodGet, Path: "/api/v1/evidence/{id}", Summary: "Get evidence", Tags: []string{"Evidence"}},
	{Method: http.MethodGet, Path: "/api/v1/assets", Summary: "List assets", Tags: []string{"Assets"}},
	{Method: http.MethodGet, Path: "/api/v1/assets/provisional-identities", Summary: "List provisional identity assets", Tags: []string{"Assets"}},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}", Summary: "Get an asset", Tags: []string{"Assets"}},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/history", Summary: "Get asset history", Tags: []string{"Assets"}},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/facts", Summary: "List asset facts", Tags: []string{"Assets", "Facts"}},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/relationships", Summary: "List asset relationships", Tags: []string{"Assets"}},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/evidence", Summary: "List evidence for an asset", Tags: []string{"Assets", "Evidence"}},
	{Method: http.MethodGet, Path: "/api/v1/facts/conflicts", Summary: "List conflicting facts", Tags: []string{"Facts"}},
	{Method: http.MethodGet, Path: "/api/v1/facts/{id}/evidence", Summary: "List evidence for a fact", Tags: []string{"Facts", "Evidence"}},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/graph", Summary: "Get an asset graph", Tags: []string{"Graph"}},
	{Method: http.MethodGet, Path: "/api/v1/graph/neighbors", Summary: "Get graph neighbors", Tags: []string{"Graph"}},
	{Method: http.MethodPost, Path: "/api/v1/agent/messages", Summary: "Send a message to the deterministic agent", Tags: []string{"Agent"}},
	{Method: http.MethodPost, Path: "/api/v1/architecture-seeds", Summary: "Create architecture seed hints", Tags: []string{"Planning"}},
	{Method: http.MethodPost, Path: "/api/v1/discovery-plans", Summary: "Create a discovery plan", Tags: []string{"Planning"}},
}

func OpenAPISpec(version string) map[string]any {
	if version == "" {
		version = "dev"
	}
	paths := map[string]any{}
	tagSet := map[string]struct{}{}
	for _, route := range documentedRoutes {
		for _, tag := range route.Tags {
			tagSet[tag] = struct{}{}
		}
		item, _ := paths[route.Path].(map[string]any)
		if item == nil {
			item = map[string]any{}
			paths[route.Path] = item
		}
		operation := map[string]any{"summary": route.Summary, "tags": route.Tags, "responses": standardResponses(route.Method)}
		if strings.Contains(route.Path, "{id}") {
			operation["parameters"] = []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string"}}}
		}
		if route.Method == http.MethodPost {
			operation["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]string{"type": "object"}}}}
		}
		item[strings.ToLower(route.Method)] = operation
	}
	tags := make([]map[string]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, map[string]string{"name": tag})
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i]["name"] < tags[j]["name"] })
	return map[string]any{"openapi": "3.0.3", "info": map[string]string{"title": "Truthwatcher API", "version": version, "description": "Evidence-first network discovery and planning API."}, "servers": []map[string]string{{"url": "/"}}, "tags": tags, "paths": paths, "components": map[string]any{"schemas": map[string]any{"ResponseEnvelope": map[string]any{"type": "object"}}}}
}

func standardResponses(method string) map[string]any {
	responses := map[string]any{"200": responseRef("Successful response"), "400": responseRef("Invalid request"), "500": responseRef("Internal server error")}
	if method == http.MethodPost {
		responses["201"] = responseRef("Created")
	}
	return responses
}

func responseRef(description string) map[string]any {
	return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ResponseEnvelope"}}}}
}

func handleOpenAPIJSON(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(OpenAPISpec(version))
	}
}

func handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><title>Truthwatcher API Docs</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>window.onload=()=>SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui'});</script></body></html>`))
}
