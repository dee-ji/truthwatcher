package extensibility

import (
	"encoding/json"
	"strings"
	"testing"

	"truthwatcher/internal/assets"
)

func TestValidateImportResultWarnsObservedRecordsAreCandidates(t *testing.T) {
	evidenceID := "evidence-a"
	result := ImportResult{
		Candidates: ModelCandidates{
			Assets: []assets.CreateAssetParams{{
				Type:        "device",
				IdentityKey: "device:serial:aaa",
				Confidence:  0.95,
				State:       assets.StateObserved,
				Metadata:    json.RawMessage(`{}`),
			}},
			Facts: []assets.CreateFactParams{{
				AssetID:    "asset-a",
				Name:       "hostname",
				Value:      json.RawMessage(`"router-a"`),
				Source:     "parser",
				Confidence: 0.95,
				State:      assets.StateObserved,
				EvidenceID: &evidenceID,
			}},
			Relationships: []assets.CreateRelationshipParams{{
				SourceAssetID:    "asset-a",
				TargetAssetID:    "asset-b",
				RelationshipType: "lldp_neighbor_of",
				Confidence:       0.8,
				State:            assets.StateObserved,
				EvidenceID:       &evidenceID,
				Metadata:         json.RawMessage(`{}`),
			}},
		},
	}

	warnings, err := ValidateImportResult(result)
	if err != nil {
		t.Fatalf("ValidateImportResult returned error: %v", err)
	}
	joined := strings.Join(warnings, "\n")
	for _, want := range []string{
		"imported observed assets are candidates only",
		"imported observed facts are candidates only",
		"imported observed relationships are candidates only",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("warnings = %#v, want %q", warnings, want)
		}
	}
}

func TestValidateImportResultRejectsInvalidCandidate(t *testing.T) {
	_, err := ValidateImportResult(ImportResult{
		Candidates: ModelCandidates{
			Facts: []assets.CreateFactParams{{
				AssetID:    "asset-a",
				Name:       "hostname",
				Value:      json.RawMessage(`not-json`),
				Source:     "parser",
				Confidence: 0.5,
				State:      assets.StateUserSeeded,
			}},
		},
	})
	if err == nil {
		t.Fatal("ValidateImportResult returned nil error for invalid fact JSON")
	}
	if !strings.Contains(err.Error(), "value must be valid JSON") {
		t.Fatalf("error = %q, want invalid JSON message", err.Error())
	}
}
