package extensibility

import (
	"context"
	"encoding/json"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/parser"
	"truthwatcher/internal/planner"
	"truthwatcher/internal/policy"
)

// Kind identifies which boundary a connector implements.
type Kind string

const (
	KindCollector Kind = "collector"
	KindParser    Kind = "parser"
	KindImporter  Kind = "importer"
	KindExporter  Kind = "exporter"
	KindEnricher  Kind = "enricher"
	KindPlanner   Kind = "planner"
)

// Metadata describes a compile-time connector. It is intentionally not a
// runtime plugin manifest; dynamic loading is deferred until contracts prove out.
type Metadata struct {
	Name           string   `json:"name"`
	Kind           Kind     `json:"kind"`
	Version        string   `json:"version"`
	ExternalSystem string   `json:"external_system,omitempty"`
	Capabilities   []string `json:"capabilities,omitempty"`
	ReadOnly       bool     `json:"read_only"`
}

// Connector is the common metadata boundary for every extension point.
type Connector interface {
	Metadata() Metadata
}

// Collector gathers read-only raw outputs. Implementations must still enforce
// policy before collecting and must not write to the network.
type Collector interface {
	Connector
	Collect(ctx context.Context, target string, profile discovery.Profile, tasks []policy.Task) ([]discovery.CollectedOutput, error)
}

// Parser turns stored raw evidence into normalized model candidates without
// writing to the database directly.
type Parser interface {
	Connector
	parser.Parser
}

// Importer reads data from an external system and returns evidence-backed model
// candidates. The kernel decides if and how those candidates are persisted.
type Importer interface {
	Connector
	Import(ctx context.Context, request ImportRequest) (ImportResult, error)
}

// Exporter sends a kernel snapshot to an external system. Exporters must not be
// required by the kernel and should be explicitly invoked.
type Exporter interface {
	Connector
	Export(ctx context.Context, request ExportRequest) (ExportResult, error)
}

// Enricher reads current kernel data and proposes additional evidence-backed
// facts or relationships. Enrichers must not silently overwrite model state.
type Enricher interface {
	Connector
	Enrich(ctx context.Context, request EnrichmentRequest) (EnrichmentResult, error)
}

// Planner proposes safe next discovery steps. A planner must not execute plans.
type Planner interface {
	Connector
	Plan(ctx context.Context, request planner.Request) (planner.Plan, error)
}

// EvidenceCandidate is raw external data before it receives a database ID.
type EvidenceCandidate struct {
	Target       string          `json:"target"`
	Method       string          `json:"method"`
	CommandOrAPI string          `json:"command_or_api"`
	RawOutput    string          `json:"raw_output"`
	ParserName   string          `json:"parser_name,omitempty"`
	CollectedAt  time.Time       `json:"collected_at,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// ModelCandidates are normalized outputs that still require kernel validation
// and persistence. Existing IDs should be left empty unless they refer to known
// kernel records.
type ModelCandidates struct {
	Assets        []assets.CreateAssetParams        `json:"assets,omitempty"`
	Facts         []assets.CreateFactParams         `json:"facts,omitempty"`
	Relationships []assets.CreateRelationshipParams `json:"relationships,omitempty"`
	ParseResults  []parser.Result                   `json:"parse_results,omitempty"`
}

type ImportRequest struct {
	Source        string          `json:"source"`
	Scope         json.RawMessage `json:"scope,omitempty"`
	CredentialRef string          `json:"credential_ref,omitempty"`
	DryRun        bool            `json:"dry_run"`
}

type ImportResult struct {
	Evidence         []EvidenceCandidate        `json:"evidence,omitempty"`
	EvidenceMetadata []ImportedEvidenceMetadata `json:"evidence_metadata,omitempty"`
	Candidates       ModelCandidates            `json:"candidates"`
	Warnings         []string                   `json:"warnings,omitempty"`
}

type ExportRequest struct {
	Destination   string        `json:"destination"`
	Snapshot      GraphSnapshot `json:"snapshot"`
	CredentialRef string        `json:"credential_ref,omitempty"`
	DryRun        bool          `json:"dry_run"`
}

type ExportResult struct {
	Destination string   `json:"destination"`
	Exported    int      `json:"exported"`
	DryRun      bool     `json:"dry_run"`
	Warnings    []string `json:"warnings,omitempty"`
}

type EnrichmentRequest struct {
	Source        string        `json:"source"`
	Snapshot      GraphSnapshot `json:"snapshot"`
	CredentialRef string        `json:"credential_ref,omitempty"`
	DryRun        bool          `json:"dry_run"`
}

type EnrichmentResult struct {
	Evidence   []EvidenceCandidate `json:"evidence,omitempty"`
	Candidates ModelCandidates     `json:"candidates"`
	Warnings   []string            `json:"warnings,omitempty"`
}

// GraphSnapshot is the stable kernel view external exporters/enrichers receive.
type GraphSnapshot struct {
	Assets        []assets.Asset        `json:"assets,omitempty"`
	Facts         []assets.Fact         `json:"facts,omitempty"`
	Relationships []assets.Relationship `json:"relationships,omitempty"`
	Evidence      []evidence.Evidence   `json:"evidence,omitempty"`
}
