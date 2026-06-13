package extensibility

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/evidence"
)

const fileJSONSchemaVersion = "truthwatcher.file_snapshot.v1"

// EvidenceMetadata is the exported evidence view. It intentionally excludes raw
// output so file exports can share provenance without copying command output.
type EvidenceMetadata struct {
	ID             string          `json:"id,omitempty"`
	DiscoveryRunID string          `json:"discovery_run_id,omitempty"`
	Target         string          `json:"target"`
	Method         string          `json:"method"`
	CommandOrAPI   string          `json:"command_or_api"`
	RawOutputHash  string          `json:"raw_output_hash,omitempty"`
	ParserName     *string         `json:"parser_name,omitempty"`
	CollectedAt    time.Time       `json:"collected_at,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	RawOutput      string          `json:"raw_output,omitempty"`
}

// FileSnapshot is the stable JSON file import/export shape.
type FileSnapshot struct {
	SchemaVersion    string                `json:"schema_version"`
	GeneratedAt      time.Time             `json:"generated_at"`
	Assets           []assets.Asset        `json:"assets,omitempty"`
	Facts            []assets.Fact         `json:"facts,omitempty"`
	Relationships    []assets.Relationship `json:"relationships,omitempty"`
	EvidenceMetadata []EvidenceMetadata    `json:"evidence_metadata,omitempty"`
}

// ImportedEvidenceMetadata preserves exported evidence provenance when raw
// output is not available for direct evidence creation.
type ImportedEvidenceMetadata = EvidenceMetadata

// FileJSONImporter reads Truthwatcher JSON snapshots from local files.
type FileJSONImporter struct{}

func (FileJSONImporter) Metadata() Metadata {
	return Metadata{
		Name:           "file_json_importer",
		Kind:           KindImporter,
		Version:        "v1",
		ExternalSystem: "local_file",
		Capabilities:   []string{"json_import", "evidence_metadata"},
		ReadOnly:       true,
	}
}

func (FileJSONImporter) Import(ctx context.Context, request ImportRequest) (ImportResult, error) {
	if err := ctx.Err(); err != nil {
		return ImportResult{}, err
	}
	path := strings.TrimSpace(request.Source)
	if path == "" {
		return ImportResult{}, fmt.Errorf("source path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ImportResult{}, fmt.Errorf("read import file: %w", err)
	}

	var snapshot FileSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return ImportResult{}, fmt.Errorf("decode import file: %w", err)
	}
	if snapshot.SchemaVersion != fileJSONSchemaVersion {
		return ImportResult{}, fmt.Errorf("unsupported file snapshot schema %q", snapshot.SchemaVersion)
	}

	result := ImportResult{
		Candidates: ModelCandidates{
			Assets:        assetCandidates(snapshot.Assets),
			Facts:         factCandidates(snapshot.Facts),
			Relationships: relationshipCandidates(snapshot.Relationships),
		},
		EvidenceMetadata: evidenceMetadata(snapshot.EvidenceMetadata),
	}
	for _, item := range snapshot.EvidenceMetadata {
		if strings.TrimSpace(item.RawOutput) == "" {
			continue
		}
		result.Evidence = append(result.Evidence, EvidenceCandidate{
			Target:       item.Target,
			Method:       item.Method,
			CommandOrAPI: item.CommandOrAPI,
			RawOutput:    item.RawOutput,
			ParserName:   stringFromPtr(item.ParserName),
			CollectedAt:  item.CollectedAt,
			Metadata:     defaultJSON(item.Metadata),
		})
	}
	if len(snapshot.EvidenceMetadata) > len(result.Evidence) {
		result.Warnings = append(result.Warnings, "evidence metadata was preserved, but raw output was not present for every evidence record")
	}
	return result, nil
}

// FileJSONExporter writes Truthwatcher JSON snapshots to local files.
type FileJSONExporter struct{}

func (FileJSONExporter) Metadata() Metadata {
	return Metadata{
		Name:           "file_json_exporter",
		Kind:           KindExporter,
		Version:        "v1",
		ExternalSystem: "local_file",
		Capabilities:   []string{"json_export", "evidence_metadata"},
		ReadOnly:       true,
	}
}

func (FileJSONExporter) Export(ctx context.Context, request ExportRequest) (ExportResult, error) {
	if err := ctx.Err(); err != nil {
		return ExportResult{}, err
	}
	path := strings.TrimSpace(request.Destination)
	if path == "" {
		return ExportResult{}, fmt.Errorf("destination path is required")
	}

	snapshot := FileSnapshot{
		SchemaVersion:    fileJSONSchemaVersion,
		GeneratedAt:      time.Now().UTC(),
		Assets:           append([]assets.Asset(nil), request.Snapshot.Assets...),
		Facts:            append([]assets.Fact(nil), request.Snapshot.Facts...),
		Relationships:    append([]assets.Relationship(nil), request.Snapshot.Relationships...),
		EvidenceMetadata: metadataFromEvidence(request.Snapshot.Evidence),
	}
	count := len(snapshot.Assets) + len(snapshot.Facts) + len(snapshot.Relationships) + len(snapshot.EvidenceMetadata)
	if request.DryRun {
		return ExportResult{Destination: path, Exported: count, DryRun: true}, nil
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return ExportResult{}, fmt.Errorf("encode export file: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return ExportResult{}, fmt.Errorf("write export file: %w", err)
	}

	return ExportResult{Destination: path, Exported: count}, nil
}

func metadataFromEvidence(items []evidence.Evidence) []EvidenceMetadata {
	out := make([]EvidenceMetadata, 0, len(items))
	for _, item := range items {
		out = append(out, EvidenceMetadata{
			ID:             item.ID,
			DiscoveryRunID: item.DiscoveryRunID,
			Target:         item.Target,
			Method:         item.Method,
			CommandOrAPI:   item.CommandOrAPI,
			RawOutputHash:  item.RawOutputHash,
			ParserName:     item.ParserName,
			CollectedAt:    item.CollectedAt,
			Metadata:       defaultJSON(item.Metadata),
		})
	}
	return out
}

func assetCandidates(items []assets.Asset) []assets.CreateAssetParams {
	out := make([]assets.CreateAssetParams, 0, len(items))
	for _, item := range items {
		out = append(out, assets.CreateAssetParams{
			Type:             item.Type,
			IdentityKey:      item.IdentityKey,
			Vendor:           item.Vendor,
			Model:            item.Model,
			Serial:           item.Serial,
			SystemMAC:        item.SystemMAC,
			Confidence:       item.Confidence,
			ConfidenceReason: item.ConfidenceReason,
			State:            item.State,
			Metadata:         defaultJSON(item.Metadata),
		})
	}
	return out
}

func factCandidates(items []assets.Fact) []assets.CreateFactParams {
	out := make([]assets.CreateFactParams, 0, len(items))
	for _, item := range items {
		out = append(out, assets.CreateFactParams{
			AssetID:          item.AssetID,
			Name:             item.Name,
			Value:            defaultJSON(item.Value),
			Source:           item.Source,
			Confidence:       item.Confidence,
			ConfidenceReason: item.ConfidenceReason,
			State:            item.State,
			EvidenceID:       item.EvidenceID,
		})
	}
	return out
}

func relationshipCandidates(items []assets.Relationship) []assets.CreateRelationshipParams {
	out := make([]assets.CreateRelationshipParams, 0, len(items))
	for _, item := range items {
		out = append(out, assets.CreateRelationshipParams{
			SourceAssetID:    item.SourceAssetID,
			TargetAssetID:    item.TargetAssetID,
			RelationshipType: item.RelationshipType,
			Confidence:       item.Confidence,
			ConfidenceReason: item.ConfidenceReason,
			State:            item.State,
			EvidenceID:       item.EvidenceID,
			Metadata:         defaultJSON(item.Metadata),
		})
	}
	return out
}

func evidenceMetadata(items []EvidenceMetadata) []ImportedEvidenceMetadata {
	out := make([]ImportedEvidenceMetadata, 0, len(items))
	for _, item := range items {
		item.Metadata = defaultJSON(item.Metadata)
		out = append(out, item)
	}
	return out
}

func defaultJSON(value json.RawMessage) json.RawMessage {
	if strings.TrimSpace(string(value)) == "" {
		return json.RawMessage(`{}`)
	}
	return value
}

func stringFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
