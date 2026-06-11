package seeding

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"truthwatcher/internal/assets"
)

const (
	ArchitectureAssetType       = "architecture_context"
	ArchitectureIdentityKey     = "architecture:seed:default"
	SeedSource                  = "user_seeded"
	seedConfidence              = 0.45
	seedConfidenceReason        = "provided by user seed input; useful planning context but not proof"
	architectureFactDescription = "architecture hints are context only and must not be treated as observed network facts"
)

type AssetService interface {
	CreateAsset(context.Context, assets.CreateAssetParams) (assets.Asset, error)
	ListAssets(context.Context) ([]assets.Asset, error)
	CreateFact(context.Context, assets.CreateFactParams) (assets.Fact, error)
	ListFactsByAsset(context.Context, string) ([]assets.Fact, error)
}

type Service struct {
	assets AssetService
}

type Options struct {
	Assets AssetService
}

type Request struct {
	OrganizationNetworkType string   `json:"organization_network_type"`
	KnownASNs               []string `json:"known_asns"`
	KnownRouteReflectors    []string `json:"known_route_reflectors"`
	KnownVendors            []string `json:"known_vendors"`
	KnownEMSSystems         []string `json:"known_ems_systems"`
	KnownServices           []string `json:"known_services"`
	KnownRegionsMarkets     []string `json:"known_regions_markets"`
}

type Result struct {
	Asset   assets.Asset  `json:"asset"`
	Facts   []assets.Fact `json:"facts"`
	Warning string        `json:"warning"`
}

func NewService(opts Options) Service {
	return Service{assets: opts.Assets}
}

func (s Service) SeedArchitecture(ctx context.Context, req Request) (Result, error) {
	if s.assets == nil {
		return Result{}, fmt.Errorf("asset repository is required")
	}
	req = normalizeRequest(req)
	if req.empty() {
		return Result{}, fmt.Errorf("at least one architecture hint is required")
	}

	asset, err := s.architectureAsset(ctx)
	if err != nil {
		return Result{}, err
	}

	facts := make([]assets.Fact, 0, 7)
	for _, item := range factInputs(req) {
		fact, err := s.assets.CreateFact(ctx, assets.CreateFactParams{
			AssetID:          asset.ID,
			Name:             item.name,
			Value:            item.value,
			Source:           SeedSource,
			Confidence:       seedConfidence,
			ConfidenceReason: seedConfidenceReason,
			State:            assets.StateUserSeeded,
		})
		if err != nil {
			return Result{}, err
		}
		facts = append(facts, fact)
	}

	return Result{
		Asset:   asset,
		Facts:   facts,
		Warning: architectureFactDescription,
	}, nil
}

func (s Service) architectureAsset(ctx context.Context) (assets.Asset, error) {
	items, err := s.assets.ListAssets(ctx)
	if err != nil {
		return assets.Asset{}, err
	}
	for _, item := range items {
		if item.IdentityKey == ArchitectureIdentityKey {
			return item, nil
		}
	}

	metadata, err := json.Marshal(map[string]any{
		"seeded_context": true,
		"not_proof":      true,
		"description":    architectureFactDescription,
	})
	if err != nil {
		return assets.Asset{}, err
	}

	return s.assets.CreateAsset(ctx, assets.CreateAssetParams{
		Type:             ArchitectureAssetType,
		IdentityKey:      ArchitectureIdentityKey,
		Confidence:       seedConfidence,
		ConfidenceReason: seedConfidenceReason,
		State:            assets.StateUserSeeded,
		Metadata:         metadata,
	})
}

type factInput struct {
	name  string
	value json.RawMessage
}

func factInputs(req Request) []factInput {
	inputs := []factInput{}
	addString := func(name, value string) {
		if value == "" {
			return
		}
		inputs = append(inputs, factInput{name: name, value: mustJSON(value)})
	}
	addStrings := func(name string, values []string) {
		if len(values) == 0 {
			return
		}
		inputs = append(inputs, factInput{name: name, value: mustJSON(values)})
	}

	addString("organization_network_type", req.OrganizationNetworkType)
	addStrings("known_asns", req.KnownASNs)
	addStrings("known_route_reflectors", req.KnownRouteReflectors)
	addStrings("known_vendors", req.KnownVendors)
	addStrings("known_ems_systems", req.KnownEMSSystems)
	addStrings("known_services", req.KnownServices)
	addStrings("known_regions_markets", req.KnownRegionsMarkets)
	return inputs
}

func normalizeRequest(req Request) Request {
	req.OrganizationNetworkType = normalizeText(req.OrganizationNetworkType)
	req.KnownASNs = normalizeList(req.KnownASNs)
	req.KnownRouteReflectors = normalizeList(req.KnownRouteReflectors)
	req.KnownVendors = normalizeList(req.KnownVendors)
	req.KnownEMSSystems = normalizeList(req.KnownEMSSystems)
	req.KnownServices = normalizeList(req.KnownServices)
	req.KnownRegionsMarkets = normalizeList(req.KnownRegionsMarkets)
	return req
}

func (r Request) empty() bool {
	return r.OrganizationNetworkType == "" && len(r.KnownASNs) == 0 && len(r.KnownRouteReflectors) == 0 && len(r.KnownVendors) == 0 && len(r.KnownEMSSystems) == 0 && len(r.KnownServices) == 0 && len(r.KnownRegionsMarkets) == 0
}

func normalizeList(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = normalizeText(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

func mustJSON(value any) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return data
}
