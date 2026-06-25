package api

import (
	"encoding/json"

	"truthwatcher/internal/agent"
	"truthwatcher/internal/assets"
	"truthwatcher/internal/audit"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/graph"
	"truthwatcher/internal/parser"
	"truthwatcher/internal/planner"
	"truthwatcher/internal/policy"
	"truthwatcher/internal/seeding"
)

type healthResponse struct {
	Status string `json:"status"`
}
type readinessResponse struct {
	Status string `json:"status"`
}
type versionResponse struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}
type systemInfoResponse struct {
	SystemInfo systemInfo `json:"system_info"`
}
type agentMessageResponse struct {
	AgentMessage agent.Response `json:"agent_message"`
}
type architectureSeedResponse struct {
	ArchitectureSeed seeding.Result `json:"architecture_seed"`
}
type discoveryPlanResponse struct {
	DiscoveryPlan planner.Plan `json:"discovery_plan"`
}
type discoveryRunResponse struct {
	DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
}
type executeDiscoveryRunResponse struct {
	DiscoveryRun discovery.DiscoveryRun `json:"discovery_run"`
	Evidence     []evidence.Evidence    `json:"evidence"`
}
type discoveryRunsResponse struct {
	DiscoveryRuns []discovery.DiscoveryRun `json:"discovery_runs"`
}
type auditRecordsResponse struct {
	AuditRecords []audit.Record `json:"audit_records"`
}
type evidenceListResponse struct {
	Evidence []evidence.Evidence `json:"evidence"`
}
type evidenceResponse struct {
	Evidence evidence.Evidence `json:"evidence"`
}
type parseDiscoveryRunResponse struct {
	ParseResult parser.ParseDiscoveryRunResult `json:"parse_result"`
}
type graphResponse struct {
	Graph graph.Graph `json:"graph"`
}
type assetsResponse struct {
	Assets []assets.Asset `json:"assets"`
}
type assetResponse struct {
	Asset assets.Asset `json:"asset"`
}
type assetHistoryResponse struct {
	Asset   assets.Asset        `json:"asset"`
	History []assetHistoryEvent `json:"history"`
}
type factsResponse struct {
	Facts []assets.Fact `json:"facts"`
}
type relationshipsResponse struct {
	Relationships []assets.Relationship `json:"relationships"`
}
type identityCandidatesResponse struct {
	IdentityCandidates []parser.IdentityCandidate `json:"identity_candidates"`
}
type identityReviewHandoffResponse struct {
	IdentityReviewHandoff parser.IdentityReviewHandoffReport `json:"identity_review_handoff"`
}
type identityCandidateReviewResponse struct {
	IdentityCandidateReview parser.IdentityCandidateReview `json:"identity_candidate_review"`
}

type createDiscoveryRunRequest struct {
	SeedInput json.RawMessage `json:"seed_input"`
}
type executeDiscoveryRunRequest struct {
	Collector   string        `json:"collector"`
	Target      string        `json:"target"`
	Profile     string        `json:"profile"`
	Tasks       []policy.Task `json:"tasks"`
	FixtureRoot string        `json:"fixture_root"`
}
type parseDiscoveryRunRequest struct {
	Platform string `json:"platform"`
}
type reviewIdentityCandidateRequest struct {
	Reviewer  string          `json:"reviewer"`
	Action    string          `json:"action"`
	Rationale string          `json:"rationale"`
	Metadata  json.RawMessage `json:"metadata"`
}
