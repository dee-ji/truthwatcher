package graph

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"truthwatcher/internal/assets"
)

type fakeAssetReader struct {
	assets        map[string]assets.Asset
	facts         map[string][]assets.Fact
	relationships []assets.Relationship
}

func (f fakeAssetReader) GetAsset(ctx context.Context, id string) (assets.Asset, error) {
	item, ok := f.assets[id]
	if !ok {
		return assets.Asset{}, assets.ErrNotFound
	}
	return item, nil
}

func (f fakeAssetReader) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	return f.facts[assetID], nil
}

func (f fakeAssetReader) ListRelationships(ctx context.Context) ([]assets.Relationship, error) {
	return f.relationships, nil
}

func TestGetAssetGraphReturnsNodesEdgesAndFacts(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	reader := fakeAssetReader{
		assets: map[string]assets.Asset{
			"asset-a": {
				ID:          "asset-a",
				Type:        "device",
				IdentityKey: "device:serial:aaa",
				Metadata:    json.RawMessage(`{}`),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			"asset-b": {
				ID:          "asset-b",
				Type:        "device",
				IdentityKey: "device:serial:bbb",
				Metadata:    json.RawMessage(`{}`),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		facts: map[string][]assets.Fact{
			"asset-a": {{
				ID:         "fact-a",
				AssetID:    "asset-a",
				Name:       "hostname",
				Value:      json.RawMessage(`"router-a"`),
				Source:     "parser",
				Confidence: 0.95,
				CreatedAt:  now,
			}},
		},
		relationships: []assets.Relationship{{
			ID:               "rel-a-b",
			SourceAssetID:    "asset-a",
			TargetAssetID:    "asset-b",
			RelationshipType: "lldp_neighbor",
			Confidence:       0.9,
			Metadata:         json.RawMessage(`{}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}},
	}

	service := NewService(reader)
	result, err := service.GetAssetGraph(context.Background(), "asset-a")
	if err != nil {
		t.Fatalf("GetAssetGraph: %v", err)
	}

	if got, want := len(result.Nodes), 2; got != want {
		t.Fatalf("node count = %d, want %d", got, want)
	}
	if got, want := len(result.Edges), 1; got != want {
		t.Fatalf("edge count = %d, want %d", got, want)
	}
	if result.Nodes[0].ID != "asset-a" {
		t.Fatalf("root node id = %q, want asset-a", result.Nodes[0].ID)
	}
	if result.Nodes[0].Label != "router-a" {
		t.Fatalf("root node label = %q, want router-a", result.Nodes[0].Label)
	}
	if got, want := len(result.Nodes[0].Facts), 1; got != want {
		t.Fatalf("root facts = %d, want %d", got, want)
	}
	if result.Edges[0].Source != "asset-a" || result.Edges[0].Target != "asset-b" {
		t.Fatalf("edge endpoints = %q -> %q, want asset-a -> asset-b", result.Edges[0].Source, result.Edges[0].Target)
	}
}

func TestPathCandidatesReturnDirectRelationshipCandidates(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	reader := fakeAssetReader{
		assets: map[string]assets.Asset{
			"asset-a": {ID: "asset-a", Type: "device", IdentityKey: "device:serial:aaa", Metadata: json.RawMessage(`{}`), CreatedAt: now, UpdatedAt: now},
			"asset-b": {ID: "asset-b", Type: "device", IdentityKey: "device:serial:bbb", Metadata: json.RawMessage(`{}`), CreatedAt: now, UpdatedAt: now},
		},
		facts: map[string][]assets.Fact{},
		relationships: []assets.Relationship{{
			ID:               "rel-a-b",
			SourceAssetID:    "asset-a",
			TargetAssetID:    "asset-b",
			RelationshipType: "lldp_neighbor",
			Confidence:       0.9,
			Metadata:         json.RawMessage(`{}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}},
	}

	service := NewService(reader)
	candidates, err := service.PathCandidates(context.Background(), "asset-a")
	if err != nil {
		t.Fatalf("PathCandidates: %v", err)
	}

	if got, want := len(candidates), 1; got != want {
		t.Fatalf("candidate count = %d, want %d", got, want)
	}
	if got, want := len(candidates[0].Nodes), 2; got != want {
		t.Fatalf("candidate node count = %d, want %d", got, want)
	}
	if got, want := len(candidates[0].Edges), 1; got != want {
		t.Fatalf("candidate edge count = %d, want %d", got, want)
	}
}

func TestGetAssetGraphMissingAsset(t *testing.T) {
	service := NewService(fakeAssetReader{
		assets: map[string]assets.Asset{},
		facts:  map[string][]assets.Fact{},
	})

	_, err := service.GetAssetGraph(context.Background(), "missing")
	if err != assets.ErrNotFound {
		t.Fatalf("err = %v, want %v", err, assets.ErrNotFound)
	}
}
