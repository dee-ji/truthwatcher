package extensibility

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/evidence"
)

func TestFileJSONConnectorContracts(t *testing.T) {
	var _ Importer = FileJSONImporter{}
	var _ Exporter = FileJSONExporter{}
}

func TestFileJSONExportOmitsRawEvidenceOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "truthwatcher-export.json")
	parserName := "junos_version"
	snapshot := sampleGraphSnapshot(parserName)

	result, err := FileJSONExporter{}.Export(context.Background(), ExportRequest{
		Destination: path,
		Snapshot:    snapshot,
	})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	if result.Destination != path {
		t.Fatalf("destination = %q, want %q", result.Destination, path)
	}
	if result.Exported != 4 {
		t.Fatalf("exported count = %d, want 4", result.Exported)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if json.Valid(data) == false {
		t.Fatalf("export is not valid JSON: %s", data)
	}
	if strings.Contains(string(data), "raw output should not be exported") {
		t.Fatalf("export leaked raw evidence output: %s", data)
	}

	var exported FileSnapshot
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if exported.SchemaVersion != fileJSONSchemaVersion {
		t.Fatalf("schema = %q, want %q", exported.SchemaVersion, fileJSONSchemaVersion)
	}
	if got, want := len(exported.Assets), 1; got != want {
		t.Fatalf("asset count = %d, want %d", got, want)
	}
	if got, want := len(exported.Facts), 1; got != want {
		t.Fatalf("fact count = %d, want %d", got, want)
	}
	if got, want := len(exported.Relationships), 1; got != want {
		t.Fatalf("relationship count = %d, want %d", got, want)
	}
	if got, want := len(exported.EvidenceMetadata), 1; got != want {
		t.Fatalf("evidence metadata count = %d, want %d", got, want)
	}
	if exported.EvidenceMetadata[0].RawOutput != "" {
		t.Fatalf("raw output = %q, want empty", exported.EvidenceMetadata[0].RawOutput)
	}
	if exported.EvidenceMetadata[0].RawOutputHash == "" {
		t.Fatal("raw output hash is empty")
	}
}

func TestFileJSONImportPreservesMetadataAndSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "truthwatcher-import.json")
	parserName := "junos_version"
	exported := FileSnapshot{
		SchemaVersion: fileJSONSchemaVersion,
		GeneratedAt:   time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
		Assets:        sampleGraphSnapshot(parserName).Assets,
		Facts:         sampleGraphSnapshot(parserName).Facts,
		Relationships: sampleGraphSnapshot(parserName).Relationships,
		EvidenceMetadata: []EvidenceMetadata{{
			ID:             "evidence-a",
			DiscoveryRunID: "run-a",
			Target:         "router-a",
			Method:         "ssh",
			CommandOrAPI:   "show version",
			RawOutputHash:  "hash-a",
			ParserName:     &parserName,
			CollectedAt:    time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
			Metadata:       json.RawMessage(`{"source":"file-test"}`),
		}},
	}
	writeSnapshot(t, path, exported)

	result, err := FileJSONImporter{}.Import(context.Background(), ImportRequest{Source: path})
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}

	if got, want := len(result.Candidates.Assets), 1; got != want {
		t.Fatalf("asset candidate count = %d, want %d", got, want)
	}
	if result.Candidates.Assets[0].IdentityKey != "device:serial:aaa" {
		t.Fatalf("identity key = %q, want device:serial:aaa", result.Candidates.Assets[0].IdentityKey)
	}
	if got, want := len(result.Candidates.Facts), 1; got != want {
		t.Fatalf("fact candidate count = %d, want %d", got, want)
	}
	fact := result.Candidates.Facts[0]
	if fact.Source != "parser" {
		t.Fatalf("fact source = %q, want parser", fact.Source)
	}
	if fact.State != assets.StateObserved {
		t.Fatalf("fact state = %q, want observed", fact.State)
	}
	if fact.EvidenceID == nil || *fact.EvidenceID != "evidence-a" {
		t.Fatalf("fact evidence ID = %#v, want evidence-a", fact.EvidenceID)
	}
	if got, want := len(result.Candidates.Relationships), 1; got != want {
		t.Fatalf("relationship candidate count = %d, want %d", got, want)
	}
	if got, want := len(result.EvidenceMetadata), 1; got != want {
		t.Fatalf("evidence metadata count = %d, want %d", got, want)
	}
	if result.EvidenceMetadata[0].RawOutputHash != "hash-a" {
		t.Fatalf("evidence hash = %q, want hash-a", result.EvidenceMetadata[0].RawOutputHash)
	}
	if len(result.Evidence) != 0 {
		t.Fatalf("evidence candidates = %d, want 0 without raw output", len(result.Evidence))
	}
	if len(result.Warnings) == 0 {
		t.Fatal("warnings are empty; want raw-output absence warning")
	}
}

func TestFileJSONImportCreatesEvidenceCandidateWhenRawOutputExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "truthwatcher-import-raw.json")
	exported := FileSnapshot{
		SchemaVersion: fileJSONSchemaVersion,
		GeneratedAt:   time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
		EvidenceMetadata: []EvidenceMetadata{{
			Target:       "router-a",
			Method:       "ssh",
			CommandOrAPI: "show version",
			RawOutput:    "raw output",
			Metadata:     json.RawMessage(`{"source":"file-test"}`),
		}},
	}
	writeSnapshot(t, path, exported)

	result, err := FileJSONImporter{}.Import(context.Background(), ImportRequest{Source: path})
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if got, want := len(result.Evidence), 1; got != want {
		t.Fatalf("evidence candidate count = %d, want %d", got, want)
	}
	if result.Evidence[0].RawOutput != "raw output" {
		t.Fatalf("raw output = %q, want raw output", result.Evidence[0].RawOutput)
	}
}

func TestFileJSONExportDryRunDoesNotWriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dry-run.json")
	result, err := FileJSONExporter{}.Export(context.Background(), ExportRequest{
		Destination: path,
		Snapshot:    GraphSnapshot{Assets: []assets.Asset{{ID: "asset-a"}}},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	if !result.DryRun {
		t.Fatal("dry run result is false")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote file or returned unexpected stat error: %v", err)
	}
}

func sampleGraphSnapshot(parserName string) GraphSnapshot {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	evidenceID := "evidence-a"
	return GraphSnapshot{
		Assets: []assets.Asset{{
			ID:               "asset-a",
			Type:             "device",
			IdentityKey:      "device:serial:aaa",
			Vendor:           stringPtr("juniper"),
			Serial:           stringPtr("aaa"),
			Confidence:       0.95,
			ConfidenceReason: "directly observed from evidence",
			State:            assets.StateObserved,
			Metadata:         json.RawMessage(`{"role":"pe"}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}},
		Facts: []assets.Fact{{
			ID:               "fact-a",
			AssetID:          "asset-a",
			Name:             "hostname",
			Value:            json.RawMessage(`"router-a"`),
			Source:           "parser",
			Confidence:       0.95,
			ConfidenceReason: "directly observed from evidence",
			State:            assets.StateObserved,
			EvidenceID:       &evidenceID,
			CreatedAt:        now,
		}},
		Relationships: []assets.Relationship{{
			ID:               "relationship-a",
			SourceAssetID:    "asset-a",
			TargetAssetID:    "asset-b",
			RelationshipType: "lldp_neighbor_of",
			Confidence:       0.8,
			ConfidenceReason: "directly observed from evidence",
			State:            assets.StateObserved,
			EvidenceID:       &evidenceID,
			Metadata:         json.RawMessage(`{"protocol":"lldp"}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}},
		Evidence: []evidence.Evidence{{
			ID:             evidenceID,
			DiscoveryRunID: "run-a",
			Target:         "router-a",
			Method:         "ssh",
			CommandOrAPI:   "show version",
			RawOutput:      "raw output should not be exported",
			RawOutputHash:  evidence.HashRawOutput("raw output should not be exported"),
			ParserName:     &parserName,
			CollectedAt:    now,
			Metadata:       json.RawMessage(`{"source":"fixture"}`),
		}},
	}
}

func writeSnapshot(t *testing.T, path string, snapshot FileSnapshot) {
	t.Helper()
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
