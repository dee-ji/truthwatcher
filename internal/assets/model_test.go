package assets

import (
	"context"
	"encoding/json"
	"testing"
)

type fakeRepository struct {
	assetParams        CreateAssetParams
	factParams         CreateFactParams
	relationshipParams CreateRelationshipParams
}

func (f *fakeRepository) CreateAsset(ctx context.Context, params CreateAssetParams) (Asset, error) {
	f.assetParams = params
	return Asset{ID: "asset-1", Type: params.Type, IdentityKey: params.IdentityKey, Metadata: params.Metadata}, nil
}

func (f *fakeRepository) GetAsset(ctx context.Context, id string) (Asset, error) {
	return Asset{ID: id}, nil
}

func (f *fakeRepository) ListAssets(ctx context.Context) ([]Asset, error) {
	return []Asset{{ID: "asset-1"}}, nil
}

func (f *fakeRepository) CreateFact(ctx context.Context, params CreateFactParams) (Fact, error) {
	f.factParams = params
	return Fact{ID: "fact-1", AssetID: params.AssetID, Name: params.Name, Value: params.Value}, nil
}

func (f *fakeRepository) GetFact(ctx context.Context, id string) (Fact, error) {
	return Fact{ID: id}, nil
}

func (f *fakeRepository) ListFactsByAsset(ctx context.Context, assetID string) ([]Fact, error) {
	return []Fact{{ID: "fact-1", AssetID: assetID}}, nil
}

func (f *fakeRepository) CreateRelationship(ctx context.Context, params CreateRelationshipParams) (Relationship, error) {
	f.relationshipParams = params
	return Relationship{ID: "relationship-1", RelationshipType: params.RelationshipType}, nil
}

func (f *fakeRepository) GetRelationship(ctx context.Context, id string) (Relationship, error) {
	return Relationship{ID: id}, nil
}

func (f *fakeRepository) ListRelationships(ctx context.Context) ([]Relationship, error) {
	return []Relationship{{ID: "relationship-1"}}, nil
}

func TestMakeIdentityKey(t *testing.T) {
	got := MakeIdentityKey(" Device ", " Serial ", " ABC123 ")
	want := "device:serial:abc123"
	if got != want {
		t.Fatalf("identity key = %q, want %q", got, want)
	}
}

func TestCreateAssetNormalizesIdentityAndDefaultsMetadata(t *testing.T) {
	repo := &fakeRepository{}

	_, err := NewService(repo).CreateAsset(context.Background(), CreateAssetParams{
		Type:        " Device ",
		IdentityKey: " DEVICE:SERIAL:ABC123 ",
	})
	if err != nil {
		t.Fatalf("CreateAsset returned error: %v", err)
	}

	if got, want := repo.assetParams.Type, "device"; got != want {
		t.Fatalf("Type = %q, want %q", got, want)
	}
	if got, want := repo.assetParams.IdentityKey, "device:serial:abc123"; got != want {
		t.Fatalf("IdentityKey = %q, want %q", got, want)
	}
	if got, want := string(repo.assetParams.Metadata), "{}"; got != want {
		t.Fatalf("Metadata = %q, want %q", got, want)
	}
}

func TestCreateFactValidatesJSONAndConfidence(t *testing.T) {
	_, err := NewService(&fakeRepository{}).CreateFact(context.Background(), CreateFactParams{
		AssetID:    "asset-1",
		Name:       "software_version",
		Value:      json.RawMessage(`"1.0"`),
		Source:     "parser",
		Confidence: 1.5,
	})
	if err == nil {
		t.Fatal("CreateFact returned nil error for invalid confidence")
	}

	_, err = NewService(&fakeRepository{}).CreateFact(context.Background(), CreateFactParams{
		AssetID:    "asset-1",
		Name:       "software_version",
		Value:      json.RawMessage(`{`),
		Source:     "parser",
		Confidence: 0.9,
	})
	if err == nil {
		t.Fatal("CreateFact returned nil error for invalid JSON")
	}
}

func TestCreateRelationshipDefaultsMetadata(t *testing.T) {
	repo := &fakeRepository{}

	_, err := NewService(repo).CreateRelationship(context.Background(), CreateRelationshipParams{
		SourceAssetID:    "asset-1",
		TargetAssetID:    "asset-2",
		RelationshipType: " LLDP_NEIGHBOR_OF ",
		Confidence:       0.8,
	})
	if err != nil {
		t.Fatalf("CreateRelationship returned error: %v", err)
	}

	if got, want := repo.relationshipParams.RelationshipType, "lldp_neighbor_of"; got != want {
		t.Fatalf("RelationshipType = %q, want %q", got, want)
	}
	if got, want := string(repo.relationshipParams.Metadata), "{}"; got != want {
		t.Fatalf("Metadata = %q, want %q", got, want)
	}
}
