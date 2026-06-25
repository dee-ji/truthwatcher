package api

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"truthwatcher/internal/agent"
	"truthwatcher/internal/planner"
	"truthwatcher/internal/seeding"
)

type routeDoc struct {
	Method         string
	Path           string
	Summary        string
	Description    string
	Tags           []string
	RequestSchema  string
	ResponseSchema string
	QueryParams    []parameterDoc
}

type parameterDoc struct {
	Name        string
	Required    bool
	Schema      map[string]any
	Description string
}

var documentedRoutes = []routeDoc{
	{Method: http.MethodGet, Path: "/healthz", Summary: "Health check", Tags: []string{"System"}, ResponseSchema: "HealthResponse"},
	{Method: http.MethodGet, Path: "/readyz", Summary: "Readiness check", Tags: []string{"System"}, ResponseSchema: "ReadinessResponse"},
	{Method: http.MethodGet, Path: "/api/version", Summary: "Get running version", Tags: []string{"System"}, ResponseSchema: "VersionResponse"},
	{Method: http.MethodGet, Path: "/api/v1/version", Summary: "Get API version", Tags: []string{"System"}, ResponseSchema: "VersionResponse"},
	{Method: http.MethodGet, Path: "/api/v1/system-info", Summary: "Get runtime system information", Tags: []string{"System"}, ResponseSchema: "SystemInfoResponse"},
	{Method: http.MethodPost, Path: "/api/v1/discovery-runs", Summary: "Create a discovery run", Tags: []string{"Discovery"}, RequestSchema: "CreateDiscoveryRunRequest", ResponseSchema: "DiscoveryRunResponse"},
	{Method: http.MethodPost, Path: "/api/v1/discovery-runs/execute", Summary: "Execute a discovery run", Tags: []string{"Discovery"}, RequestSchema: "ExecuteDiscoveryRunRequest", ResponseSchema: "ExecuteDiscoveryRunResponse"},
	{Method: http.MethodGet, Path: "/api/v1/discovery-runs", Summary: "List discovery runs", Tags: []string{"Discovery"}, ResponseSchema: "DiscoveryRunsResponse"},
	{Method: http.MethodGet, Path: "/api/v1/discovery-runs/{id}", Summary: "Get a discovery run", Tags: []string{"Discovery"}, ResponseSchema: "DiscoveryRunResponse"},
	{Method: http.MethodGet, Path: "/api/v1/discovery-runs/{id}/evidence", Summary: "List evidence for a discovery run", Tags: []string{"Discovery", "Evidence"}, ResponseSchema: "EvidenceListResponse"},
	{Method: http.MethodGet, Path: "/api/v1/audit-records", Summary: "List read-only audit records", Tags: []string{"Audit"}, ResponseSchema: "AuditRecordsResponse", QueryParams: []parameterDoc{{Name: "discovery_run_id", Schema: stringSchema(), Description: "Filter by discovery run ID"}, {Name: "evidence_id", Schema: stringSchema(), Description: "Filter by evidence ID"}, {Name: "request_id", Schema: stringSchema(), Description: "Filter by request ID"}, {Name: "action", Schema: stringSchema(), Description: "Filter by action"}, {Name: "status", Schema: stringSchema(), Description: "Filter by status"}, {Name: "target", Schema: stringSchema(), Description: "Filter by target substring"}, {Name: "method", Schema: stringSchema(), Description: "Filter by method"}, {Name: "profile", Schema: stringSchema(), Description: "Filter by profile"}, {Name: "limit", Schema: map[string]any{"type": "integer"}, Description: "Maximum records to return (1-200; default 50)"}}},
	{Method: http.MethodPost, Path: "/api/v1/discovery-runs/{id}/parse", Summary: "Parse a discovery run", Tags: []string{"Discovery", "Parser"}, RequestSchema: "ParseDiscoveryRunRequest", ResponseSchema: "ParseDiscoveryRunResponse"},
	{Method: http.MethodGet, Path: "/api/v1/identity-candidates", Summary: "List identity candidates", Tags: []string{"Identity"}, ResponseSchema: "IdentityCandidatesResponse"},
	{Method: http.MethodGet, Path: "/api/v1/identity-candidates/review-queue", Summary: "List pending identity candidates", Tags: []string{"Identity"}, ResponseSchema: "IdentityCandidatesResponse"},
	{Method: http.MethodGet, Path: "/api/v1/identity-candidates/handoff-report", Summary: "Get identity review handoff report", Tags: []string{"Identity"}, ResponseSchema: "IdentityReviewHandoffResponse"},
	{Method: http.MethodPost, Path: "/api/v1/identity-candidates/{id}/review", Summary: "Review an identity candidate", Tags: []string{"Identity"}, RequestSchema: "ReviewIdentityCandidateRequest", ResponseSchema: "IdentityCandidateReviewResponse"},
	{Method: http.MethodGet, Path: "/api/v1/evidence/{id}", Summary: "Get evidence", Tags: []string{"Evidence"}, ResponseSchema: "EvidenceResponse"},
	{Method: http.MethodGet, Path: "/api/v1/assets", Summary: "List assets", Tags: []string{"Assets"}, ResponseSchema: "AssetsResponse", QueryParams: []parameterDoc{{Name: "search", Schema: stringSchema(), Description: "Search asset name, identity key, vendor, model, or serial"}}},
	{Method: http.MethodGet, Path: "/api/v1/assets/provisional-identities", Summary: "List provisional identity assets", Tags: []string{"Assets"}, ResponseSchema: "AssetsResponse"},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}", Summary: "Get an asset", Tags: []string{"Assets"}, ResponseSchema: "AssetResponse"},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/history", Summary: "Get asset history", Tags: []string{"Assets"}, ResponseSchema: "AssetHistoryResponse"},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/facts", Summary: "List asset facts", Tags: []string{"Assets", "Facts"}, ResponseSchema: "FactsResponse"},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/relationships", Summary: "List asset relationships", Tags: []string{"Assets"}, ResponseSchema: "RelationshipsResponse"},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/evidence", Summary: "List evidence for an asset", Tags: []string{"Assets", "Evidence"}, ResponseSchema: "EvidenceListResponse"},
	{Method: http.MethodGet, Path: "/api/v1/facts/conflicts", Summary: "List conflicting facts", Tags: []string{"Facts"}, ResponseSchema: "FactsResponse"},
	{Method: http.MethodGet, Path: "/api/v1/facts/{id}/evidence", Summary: "List evidence for a fact", Tags: []string{"Facts", "Evidence"}, ResponseSchema: "EvidenceListResponse"},
	{Method: http.MethodGet, Path: "/api/v1/assets/{id}/graph", Summary: "Get an asset graph", Tags: []string{"Graph"}, ResponseSchema: "GraphResponse"},
	{Method: http.MethodGet, Path: "/api/v1/graph/neighbors", Summary: "Get graph neighbors", Tags: []string{"Graph"}, ResponseSchema: "GraphResponse"},
	{Method: http.MethodPost, Path: "/api/v1/agent/messages", Summary: "Send a message to the deterministic agent", Tags: []string{"Agent"}, RequestSchema: "AgentRequest", ResponseSchema: "AgentMessageResponse"},
	{Method: http.MethodPost, Path: "/api/v1/architecture-seeds", Summary: "Create architecture seed hints", Tags: []string{"Planning"}, RequestSchema: "ArchitectureSeedRequest", ResponseSchema: "ArchitectureSeedResponse"},
	{Method: http.MethodPost, Path: "/api/v1/discovery-plans", Summary: "Create a discovery plan", Tags: []string{"Planning"}, RequestSchema: "DiscoveryPlanRequest", ResponseSchema: "DiscoveryPlanResponse"},
}

func OpenAPISpec(version string) map[string]any {
	if version == "" {
		version = "dev"
	}
	components := schemaComponents()
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
		operation := map[string]any{"summary": route.Summary, "tags": route.Tags, "responses": standardResponses(route.Method, route.ResponseSchema)}
		params := []map[string]any{}
		if strings.Contains(route.Path, "{id}") {
			params = append(params, map[string]any{"name": "id", "in": "path", "required": true, "schema": stringSchema()})
		}
		for _, qp := range route.QueryParams {
			params = append(params, map[string]any{"name": qp.Name, "in": "query", "required": qp.Required, "schema": qp.Schema, "description": qp.Description})
		}
		if len(params) > 0 {
			operation["parameters"] = params
		}
		if route.RequestSchema != "" {
			operation["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schemaRef(route.RequestSchema)}}}
		}
		item[strings.ToLower(route.Method)] = operation
	}
	tags := make([]map[string]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, map[string]string{"name": tag})
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i]["name"] < tags[j]["name"] })
	return map[string]any{"openapi": "3.0.3", "info": map[string]string{"title": "Truthwatcher API", "version": version, "description": "Evidence-first network discovery and planning API."}, "servers": []map[string]string{{"url": "/"}}, "tags": tags, "paths": paths, "components": map[string]any{"schemas": components}}
}

func standardResponses(method string, responseSchema string) map[string]any {
	responses := map[string]any{"200": responseRef("Successful response", responseSchema), "400": responseRef("Invalid request", "ErrorResponseEnvelope"), "500": responseRef("Internal server error", "ErrorResponseEnvelope")}
	if method == http.MethodPost {
		responses["201"] = responseRef("Created", responseSchema)
	}
	return responses
}

func responseRef(description, schema string) map[string]any {
	return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": schemaRef(schema)}}}
}
func schemaRef(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}
func stringSchema() map[string]any { return map[string]any{"type": "string"} }

func schemaComponents() map[string]any {
	schemas := map[string]any{
		"Error":                 map[string]any{"type": "object", "properties": map[string]any{"message": stringSchema()}, "required": []string{"message"}},
		"ErrorResponseEnvelope": map[string]any{"type": "object", "properties": map[string]any{"data": map[string]any{"nullable": true}, "error": schemaRef("Error"), "metadata": map[string]any{"type": "object", "additionalProperties": true}}, "required": []string{"error", "metadata"}},
	}
	add := func(name string, sample any) {
		schemas[name] = responseEnvelopeSchema(schemaFromType(reflect.TypeOf(sample), schemas))
	}
	addReq := func(name string, sample any) { schemas[name] = schemaFromType(reflect.TypeOf(sample), schemas) }
	add("HealthResponse", healthResponse{})
	add("ReadinessResponse", readinessResponse{})
	add("VersionResponse", versionResponse{})
	add("SystemInfoResponse", systemInfoResponse{})
	add("AgentMessageResponse", agentMessageResponse{})
	add("ArchitectureSeedResponse", architectureSeedResponse{})
	add("DiscoveryPlanResponse", discoveryPlanResponse{})
	add("DiscoveryRunResponse", discoveryRunResponse{})
	add("ExecuteDiscoveryRunResponse", executeDiscoveryRunResponse{})
	add("DiscoveryRunsResponse", discoveryRunsResponse{})
	add("AuditRecordsResponse", auditRecordsResponse{})
	add("EvidenceListResponse", evidenceListResponse{})
	add("EvidenceResponse", evidenceResponse{})
	add("ParseDiscoveryRunResponse", parseDiscoveryRunResponse{})
	add("GraphResponse", graphResponse{})
	add("AssetsResponse", assetsResponse{})
	add("AssetResponse", assetResponse{})
	add("AssetHistoryResponse", assetHistoryResponse{})
	add("FactsResponse", factsResponse{})
	add("RelationshipsResponse", relationshipsResponse{})
	add("IdentityCandidatesResponse", identityCandidatesResponse{})
	add("IdentityReviewHandoffResponse", identityReviewHandoffResponse{})
	add("IdentityCandidateReviewResponse", identityCandidateReviewResponse{})
	addReq("CreateDiscoveryRunRequest", createDiscoveryRunRequest{})
	addReq("ExecuteDiscoveryRunRequest", executeDiscoveryRunRequest{})
	addReq("ParseDiscoveryRunRequest", parseDiscoveryRunRequest{})
	addReq("ReviewIdentityCandidateRequest", reviewIdentityCandidateRequest{})
	addReq("AgentRequest", agent.Request{})
	addReq("ArchitectureSeedRequest", seeding.Request{})
	addReq("DiscoveryPlanRequest", planner.Request{})
	return schemas
}

func responseEnvelopeSchema(dataSchema map[string]any) map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{"data": dataSchema, "error": map[string]any{"nullable": true, "allOf": []any{schemaRef("Error")}}, "metadata": map[string]any{"type": "object", "additionalProperties": true}}, "required": []string{"data", "metadata"}}
}

func schemaFromType(t reflect.Type, schemas map[string]any) map[string]any {
	if t == nil {
		return map[string]any{"nullable": true}
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(json.RawMessage{}) {
		return map[string]any{"type": "object", "additionalProperties": true}
	}
	if t == reflect.TypeOf(time.Time{}) {
		return map[string]any{"type": "string", "format": "date-time"}
	}
	switch t.Kind() {
	case reflect.String:
		return stringSchema()
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": schemaFromType(t.Elem(), schemas)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": schemaFromType(t.Elem(), schemas)}
	case reflect.Interface:
		return map[string]any{"nullable": true}
	case reflect.Struct:
		props := map[string]any{}
		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name, omit := jsonName(f)
			if name == "-" {
				continue
			}
			props[name] = schemaFromType(f.Type, schemas)
			if !omit {
				required = append(required, name)
			}
		}
		s := map[string]any{"type": "object", "properties": props}
		if len(required) > 0 {
			s["required"] = required
		}
		return s
	default:
		return map[string]any{"type": "object"}
	}
}

func jsonName(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name, false
	}
	parts := strings.Split(tag, ",")
	return parts[0], contains(parts[1:], "omitempty")
}
func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
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
