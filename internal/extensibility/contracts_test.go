package extensibility

import (
	"context"
	"testing"

	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
	"truthwatcher/internal/parser"
	"truthwatcher/internal/planner"
	"truthwatcher/internal/policy"
)

type fakeCollector struct{}

func (fakeCollector) Metadata() Metadata {
	return Metadata{Name: "fake", Kind: KindCollector, ReadOnly: true}
}
func (fakeCollector) Collect(ctx context.Context, target string, profile discovery.Profile, tasks []policy.Task) ([]discovery.CollectedOutput, error) {
	return nil, ctx.Err()
}

type fakeParser struct{}

func (fakeParser) Metadata() Metadata {
	return Metadata{Name: "fake-parser", Kind: KindParser, ReadOnly: true}
}
func (fakeParser) Name() string                                  { return "fake-parser" }
func (fakeParser) Supports(platform string, command string) bool { return true }
func (fakeParser) Parse(ctx context.Context, item evidence.Evidence) (parser.Result, error) {
	return parser.Result{ParserName: "fake-parser", EvidenceID: item.ID}, ctx.Err()
}

type fakeImporter struct{}

func (fakeImporter) Metadata() Metadata {
	return Metadata{Name: "fake-importer", Kind: KindImporter, ReadOnly: true}
}
func (fakeImporter) Import(ctx context.Context, request ImportRequest) (ImportResult, error) {
	return ImportResult{}, ctx.Err()
}

type fakeExporter struct{}

func (fakeExporter) Metadata() Metadata {
	return Metadata{Name: "fake-exporter", Kind: KindExporter, ReadOnly: false}
}
func (fakeExporter) Export(ctx context.Context, request ExportRequest) (ExportResult, error) {
	return ExportResult{}, ctx.Err()
}

type fakeEnricher struct{}

func (fakeEnricher) Metadata() Metadata {
	return Metadata{Name: "fake-enricher", Kind: KindEnricher, ReadOnly: true}
}
func (fakeEnricher) Enrich(ctx context.Context, request EnrichmentRequest) (EnrichmentResult, error) {
	return EnrichmentResult{}, ctx.Err()
}

type fakePlanner struct{}

func (fakePlanner) Metadata() Metadata {
	return Metadata{Name: "fake-planner", Kind: KindPlanner, ReadOnly: true}
}
func (fakePlanner) Plan(ctx context.Context, request planner.Request) (planner.Plan, error) {
	return planner.Plan{ApprovalRequired: true, ExecutionAllowed: false}, ctx.Err()
}

func TestCompileTimeContracts(t *testing.T) {
	var _ Collector = fakeCollector{}
	var _ Parser = fakeParser{}
	var _ Importer = fakeImporter{}
	var _ Exporter = fakeExporter{}
	var _ Enricher = fakeEnricher{}
	var _ Planner = fakePlanner{}
}

func TestPlannerContractDoesNotAuthorizeExecution(t *testing.T) {
	plan, err := fakePlanner{}.Plan(context.Background(), planner.Request{})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if !plan.ApprovalRequired {
		t.Fatal("planner contract returned plan without approval requirement")
	}
	if plan.ExecutionAllowed {
		t.Fatal("planner contract returned executable plan")
	}
}
