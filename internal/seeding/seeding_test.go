package seeding

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"truthwatcher/internal/assets"
)

type fakeAssetStore struct {
	assets []assets.Asset
	facts  []assets.Fact
}

func (f *fakeAssetStore) CreateAsset(ctx context.Context, params assets.CreateAssetParams) (assets.Asset, error) {
	asset := assets.Asset{
		ID:               "asset-seed",
		Type:             params.Type,
		IdentityKey:      params.IdentityKey,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		Metadata:         params.Metadata,
		CreatedAt:        time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC),
	}
	f.assets = append(f.assets, asset)
	return asset, nil
}

func (f *fakeAssetStore) ListAssets(ctx context.Context) ([]assets.Asset, error) {
	return f.assets, nil
}

func (f *fakeAssetStore) CreateFact(ctx context.Context, params assets.CreateFactParams) (assets.Fact, error) {
	fact := assets.Fact{
		ID:               "fact-" + params.Name,
		AssetID:          params.AssetID,
		Name:             params.Name,
		Value:            params.Value,
		Source:           params.Source,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		CreatedAt:        time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC),
	}
	f.facts = append(f.facts, fact)
	return fact, nil
}

func (f *fakeAssetStore) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	var out []assets.Fact
	for _, fact := range f.facts {
		if fact.AssetID == assetID {
			out = append(out, fact)
		}
	}
	return out, nil
}

func TestSeedArchitectureStoresUserSeededFacts(t *testing.T) {
	store := &fakeAssetStore{}
	service := NewService(Options{Assets: store})

	result, err := service.SeedArchitecture(context.Background(), Request{
		OrganizationNetworkType: "service_provider",
		KnownASNs:               []string{"65000", "65000", " 65001 "},
		KnownRouteReflectors:    []string{"rr1.example.net"},
		KnownVendors:            []string{"Juniper"},
		KnownEMSSystems:         []string{"ems-a"},
		KnownServices:           []string{"L3VPN"},
		KnownRegionsMarkets:     []string{"NYC"},
	})
	if err != nil {
		t.Fatalf("SeedArchitecture returned error: %v", err)
	}

	if result.Asset.Type != ArchitectureAssetType {
		t.Fatalf("asset type = %q, want %q", result.Asset.Type, ArchitectureAssetType)
	}
	if result.Asset.State != assets.StateUserSeeded {
		t.Fatalf("asset state = %q, want user_seeded", result.Asset.State)
	}
	if result.Warning == "" {
		t.Fatal("warning is empty")
	}
	if got, want := len(result.Facts), 7; got != want {
		t.Fatalf("fact count = %d, want %d", got, want)
	}
	for _, fact := range result.Facts {
		if fact.Source != SeedSource {
			t.Fatalf("fact source = %q, want %q", fact.Source, SeedSource)
		}
		if fact.State != assets.StateUserSeeded {
			t.Fatalf("fact state = %q, want user_seeded", fact.State)
		}
		if fact.Confidence >= 0.5 {
			t.Fatalf("fact confidence = %v, want below observed confidence", fact.Confidence)
		}
	}

	knownASNs := findFact(t, result.Facts, "known_asns")
	var asns []string
	if err := json.Unmarshal(knownASNs.Value, &asns); err != nil {
		t.Fatalf("decode ASNs: %v", err)
	}
	if got, want := len(asns), 2; got != want {
		t.Fatalf("ASN count = %d, want %d", got, want)
	}
}

func TestSeedArchitectureRejectsEmptyHints(t *testing.T) {
	_, err := NewService(Options{Assets: &fakeAssetStore{}}).SeedArchitecture(context.Background(), Request{})
	if err == nil {
		t.Fatal("SeedArchitecture returned nil error for empty request")
	}
}

func findFact(t *testing.T, facts []assets.Fact, name string) assets.Fact {
	t.Helper()
	for _, fact := range facts {
		if fact.Name == name {
			return fact
		}
	}
	t.Fatalf("fact %q not found", name)
	return assets.Fact{}
}
