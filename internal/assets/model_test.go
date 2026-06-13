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
	assets             []Asset
	facts              []Fact
}

func (f *fakeRepository) CreateAsset(ctx context.Context, params CreateAssetParams) (Asset, error) {
	f.assetParams = params
	return Asset{
		ID:               "asset-1",
		Type:             params.Type,
		IdentityKey:      params.IdentityKey,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		Metadata:         params.Metadata,
	}, nil
}

func (f *fakeRepository) GetAsset(ctx context.Context, id string) (Asset, error) {
	return Asset{ID: id}, nil
}

func (f *fakeRepository) ListAssets(ctx context.Context) ([]Asset, error) {
	if f.assets != nil {
		return f.assets, nil
	}
	return []Asset{{ID: "asset-1"}}, nil
}

func (f *fakeRepository) CreateFact(ctx context.Context, params CreateFactParams) (Fact, error) {
	f.factParams = params
	return Fact{
		ID:               "fact-1",
		AssetID:          params.AssetID,
		Name:             params.Name,
		Value:            params.Value,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
	}, nil
}

func (f *fakeRepository) GetFact(ctx context.Context, id string) (Fact, error) {
	return Fact{ID: id}, nil
}

func (f *fakeRepository) ListFactsByAsset(ctx context.Context, assetID string) ([]Fact, error) {
	if f.facts != nil {
		return f.facts, nil
	}
	return []Fact{{ID: "fact-1", AssetID: assetID}}, nil
}

func (f *fakeRepository) CreateRelationship(ctx context.Context, params CreateRelationshipParams) (Relationship, error) {
	f.relationshipParams = params
	return Relationship{
		ID:               "relationship-1",
		RelationshipType: params.RelationshipType,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
	}, nil
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
	if !json.Valid(repo.assetParams.Metadata) {
		t.Fatalf("Metadata = %q, want valid JSON", repo.assetParams.Metadata)
	}
	if !hasMetadataValue(repo.assetParams.Metadata, "identity_strength", "strong") {
		t.Fatalf("Metadata = %q, want strong identity metadata", repo.assetParams.Metadata)
	}
	if got, want := repo.assetParams.Confidence, 0.5; got != want {
		t.Fatalf("Confidence = %v, want %v", got, want)
	}
	if got, want := repo.assetParams.State, StateInferred; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
	if got, want := repo.assetParams.ConfidenceReason, "deterministically inferred without direct evidence"; got != want {
		t.Fatalf("ConfidenceReason = %q, want %q", got, want)
	}
}

func TestIdentityCandidatePrefersStrongIdentifiers(t *testing.T) {
	candidate := IdentityCandidateForAsset("device", "Juniper", "JN1234", "", "mx-edge-01", "192.0.2.1")

	if candidate.Strength != IdentityStrengthStrong {
		t.Fatalf("strength = %q, want %q", candidate.Strength, IdentityStrengthStrong)
	}
	if got, want := candidate.IdentityKey, "device:vendor_serial:juniper:jn1234"; got != want {
		t.Fatalf("identity key = %q, want %q", got, want)
	}
}

func TestIdentityCandidateMarksHostnameProvisional(t *testing.T) {
	candidate := IdentityCandidateForAsset("device", "", "", "", "mx-edge-01", "")

	if candidate.Strength != IdentityStrengthProvisional {
		t.Fatalf("strength = %q, want %q", candidate.Strength, IdentityStrengthProvisional)
	}
	if got, want := candidate.IdentityKey, "device:hostname:mx-edge-01"; got != want {
		t.Fatalf("identity key = %q, want %q", got, want)
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

func TestCreateFactDefaultsObservedStateWithEvidence(t *testing.T) {
	repo := &fakeRepository{facts: []Fact{}}
	evidenceID := "evidence-1"

	result, err := NewService(repo).CreateFact(context.Background(), CreateFactParams{
		AssetID:    "asset-1",
		Name:       "software_version",
		Value:      json.RawMessage(`"1.0"`),
		Source:     "parser",
		Confidence: 0.9,
		EvidenceID: &evidenceID,
	})
	if err != nil {
		t.Fatalf("CreateFact returned error: %v", err)
	}

	if result.State != StateObserved {
		t.Fatalf("state = %q, want %q", result.State, StateObserved)
	}
	if repo.factParams.ConfidenceReason != "directly observed from evidence" {
		t.Fatalf("confidence reason = %q, want observed default", repo.factParams.ConfidenceReason)
	}
}

func TestCreateFactMarksConflictingValue(t *testing.T) {
	repo := &fakeRepository{
		facts: []Fact{{
			ID:      "fact-existing",
			AssetID: "asset-1",
			Name:    "software_version",
			Value:   json.RawMessage(`"1.0"`),
			State:   StateObserved,
		}},
	}

	result, err := NewService(repo).CreateFact(context.Background(), CreateFactParams{
		AssetID:    "asset-1",
		Name:       "software_version",
		Value:      json.RawMessage(`"2.0"`),
		Source:     "parser",
		Confidence: 0.9,
	})
	if err != nil {
		t.Fatalf("CreateFact returned error: %v", err)
	}

	if result.State != StateConflicting {
		t.Fatalf("state = %q, want %q", result.State, StateConflicting)
	}
	if repo.factParams.ConfidenceReason != "conflicts with existing fact fact-existing" {
		t.Fatalf("confidence reason = %q, want conflict reason", repo.factParams.ConfidenceReason)
	}
}

func TestListConflictingFacts(t *testing.T) {
	repo := &fakeRepository{
		assets: []Asset{{ID: "asset-1"}},
		facts: []Fact{
			{ID: "fact-ok", AssetID: "asset-1", Name: "hostname", State: StateObserved},
			{ID: "fact-conflict", AssetID: "asset-1", Name: "hostname", State: StateConflicting},
		},
	}

	result, err := NewService(repo).ListConflictingFacts(context.Background())
	if err != nil {
		t.Fatalf("ListConflictingFacts returned error: %v", err)
	}
	if got, want := len(result), 1; got != want {
		t.Fatalf("len = %d, want %d", got, want)
	}
	if result[0].ID != "fact-conflict" {
		t.Fatalf("fact id = %q, want fact-conflict", result[0].ID)
	}
}

func TestListProvisionalIdentityAssets(t *testing.T) {
	repo := &fakeRepository{
		assets: []Asset{
			{ID: "strong", Type: "device", IdentityKey: "device:vendor_serial:juniper:jn1234"},
			{ID: "provisional", Type: "device", IdentityKey: "device:hostname:mx-edge-01"},
		},
	}

	result, err := NewService(repo).ListProvisionalIdentityAssets(context.Background())
	if err != nil {
		t.Fatalf("ListProvisionalIdentityAssets returned error: %v", err)
	}
	if got, want := len(result), 1; got != want {
		t.Fatalf("len = %d, want %d", got, want)
	}
	if result[0].ID != "provisional" {
		t.Fatalf("asset id = %q, want provisional", result[0].ID)
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
	if got, want := repo.relationshipParams.State, StateInferred; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
	if got, want := repo.relationshipParams.ConfidenceReason, "deterministically inferred without direct evidence"; got != want {
		t.Fatalf("ConfidenceReason = %q, want %q", got, want)
	}
}

func hasMetadataValue(metadata json.RawMessage, key string, value string) bool {
	var payload map[string]any
	if err := json.Unmarshal(metadata, &payload); err != nil {
		return false
	}
	got, ok := payload[key].(string)
	return ok && got == value
}
