package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/discovery"
	"truthwatcher/internal/evidence"
)

type fakeAssets struct {
	assets        []assets.Asset
	facts         []assets.Fact
	relationships []assets.Relationship
}

func (f fakeAssets) GetAsset(ctx context.Context, id string) (assets.Asset, error) {
	for _, item := range f.assets {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Asset{}, assets.ErrNotFound
}

func (f fakeAssets) ListAssets(ctx context.Context) ([]assets.Asset, error) {
	return f.assets, nil
}

func (f fakeAssets) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	var result []assets.Fact
	for _, item := range f.facts {
		if item.AssetID == assetID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (f fakeAssets) ListRelationships(ctx context.Context) ([]assets.Relationship, error) {
	return f.relationships, nil
}

type fakeDiscovery struct {
	runs []discovery.DiscoveryRun
}

func (f fakeDiscovery) GetDiscoveryRun(ctx context.Context, id string) (discovery.DiscoveryRun, error) {
	for _, item := range f.runs {
		if item.ID == id {
			return item, nil
		}
	}
	return discovery.DiscoveryRun{}, discovery.ErrNotFound
}

func (f fakeDiscovery) ListDiscoveryRuns(ctx context.Context) ([]discovery.DiscoveryRun, error) {
	return f.runs, nil
}

type fakeEvidence struct {
	items []evidence.Evidence
}

func (f fakeEvidence) GetEvidence(ctx context.Context, id string) (evidence.Evidence, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	return evidence.Evidence{}, evidence.ErrNotFound
}

func (f fakeEvidence) ListEvidenceByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]evidence.Evidence, error) {
	var result []evidence.Evidence
	for _, item := range f.items {
		if item.DiscoveryRunID == discoveryRunID {
			result = append(result, item)
		}
	}
	return result, nil
}

func TestReplyListsKnownAssets(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "list known assets"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "list_known_assets" {
		t.Fatalf("intent = %q, want list_known_assets", response.Intent)
	}
	if !response.ReadOnly {
		t.Fatal("response is not read-only")
	}
	if !strings.Contains(response.Message, "Known assets: 2") {
		t.Fatalf("message = %q, want asset count", response.Message)
	}
}

func TestReplyExplainsAssetEvidence(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "explain asset evidence for asset-a"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "explain_asset_evidence" {
		t.Fatalf("intent = %q, want explain_asset_evidence", response.Intent)
	}
	if !strings.Contains(response.Message, "show version") {
		t.Fatalf("message = %q, want command summary", response.Message)
	}
}

func TestReplySummarizesDiscoveryRun(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "summarize discovery run run-a"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "summarize_discovery_run" {
		t.Fatalf("intent = %q, want summarize_discovery_run", response.Intent)
	}
	if !strings.Contains(response.Message, "1 evidence records") {
		t.Fatalf("message = %q, want evidence count", response.Message)
	}
}

func TestReplyWhatWeKnowAboutAsset(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "what do we know about router-a"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "what_we_know_about" {
		t.Fatalf("intent = %q, want what_we_know_about", response.Intent)
	}
	if !strings.Contains(response.Message, "hostname=router-a") {
		t.Fatalf("message = %q, want hostname fact", response.Message)
	}
	if !strings.Contains(response.Message, "evidence-a") {
		t.Fatalf("message = %q, want evidence reference", response.Message)
	}
}

func TestReplyShowsNeighbors(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "show neighbors for asset-a"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "show_neighbors" {
		t.Fatalf("intent = %q, want show_neighbors", response.Intent)
	}
	if !strings.Contains(response.Message, "device:serial:bbb") {
		t.Fatalf("message = %q, want neighbor asset", response.Message)
	}
	if !strings.Contains(response.Message, "evidence-a") {
		t.Fatalf("message = %q, want evidence reference", response.Message)
	}
}

func TestReplyWhyAssetExists(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "why do we believe asset-a exists"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "why_asset_exists" {
		t.Fatalf("intent = %q, want why_asset_exists", response.Intent)
	}
	if !strings.Contains(response.Message, "Evidence references") {
		t.Fatalf("message = %q, want evidence references", response.Message)
	}
}

func TestReplyWhatIsUnknown(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "what is unknown"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "what_is_unknown" {
		t.Fatalf("intent = %q, want what_is_unknown", response.Intent)
	}
	if !strings.Contains(response.Message, "model unknown") {
		t.Fatalf("message = %q, want unknown model", response.Message)
	}
}

func TestReplyUnknownAssetDoesNotGuess(t *testing.T) {
	service := testService()

	response, err := service.Reply(context.Background(), Request{Message: "what do we know about not-a-real-router"})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}

	if response.Intent != "what_we_know_about" {
		t.Fatalf("intent = %q, want what_we_know_about", response.Intent)
	}
	if !strings.Contains(response.Message, "I do not know") {
		t.Fatalf("message = %q, want unknown response", response.Message)
	}
}

func TestReplyRejectsEmptyMessage(t *testing.T) {
	_, err := testService().Reply(context.Background(), Request{})
	if err == nil {
		t.Fatal("Reply returned nil error for empty message")
	}
}

func testService() Service {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	evidenceID := "evidence-a"
	return NewService(Options{
		Assets: fakeAssets{
			assets: []assets.Asset{{
				ID:               "asset-a",
				Type:             "device",
				IdentityKey:      "device:serial:aaa",
				Confidence:       0.9,
				ConfidenceReason: "directly observed",
				State:            assets.StateObserved,
				Metadata:         json.RawMessage(`{}`),
				CreatedAt:        now,
				UpdatedAt:        now,
			}, {
				ID:               "asset-b",
				Type:             "device",
				IdentityKey:      "device:serial:bbb",
				Confidence:       0.8,
				ConfidenceReason: "directly observed",
				State:            assets.StateObserved,
				Metadata:         json.RawMessage(`{}`),
				CreatedAt:        now,
				UpdatedAt:        now,
			}},
			facts: []assets.Fact{{
				ID:               "fact-a",
				AssetID:          "asset-a",
				Name:             "hostname",
				Value:            json.RawMessage(`"router-a"`),
				Source:           "parser",
				Confidence:       0.9,
				ConfidenceReason: "directly observed",
				State:            assets.StateObserved,
				EvidenceID:       &evidenceID,
				CreatedAt:        now,
			}},
			relationships: []assets.Relationship{{
				ID:               "relationship-a",
				SourceAssetID:    "asset-a",
				TargetAssetID:    "asset-b",
				RelationshipType: "lldp_neighbor",
				Confidence:       0.9,
				ConfidenceReason: "directly observed",
				State:            assets.StateObserved,
				EvidenceID:       &evidenceID,
				Metadata:         json.RawMessage(`{}`),
				CreatedAt:        now,
				UpdatedAt:        now,
			}},
		},
		Discovery: fakeDiscovery{runs: []discovery.DiscoveryRun{{
			ID:        "run-a",
			Status:    discovery.StatusCompleted,
			SeedInput: json.RawMessage(`{"target":"fixture://junos-mx"}`),
			StartedAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		}}},
		Evidence: fakeEvidence{items: []evidence.Evidence{{
			ID:             evidenceID,
			DiscoveryRunID: "run-a",
			Target:         "fixture://junos-mx",
			Method:         "fake",
			CommandOrAPI:   "show version",
			RawOutput:      "raw output",
			RawOutputHash:  evidence.HashRawOutput("raw output"),
			CollectedAt:    now,
			Metadata:       json.RawMessage(`{}`),
		}}},
	})
}
