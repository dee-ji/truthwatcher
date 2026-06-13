package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/evidence"
)

type ParseStatus string

const (
	ParseStatusParsed  ParseStatus = "parsed"
	ParseStatusSkipped ParseStatus = "skipped"
	ParseStatusFailed  ParseStatus = "failed"
)

type ParseRecord struct {
	ID             string          `json:"id"`
	DiscoveryRunID string          `json:"discovery_run_id"`
	EvidenceID     string          `json:"evidence_id"`
	ParserName     string          `json:"parser_name"`
	Status         ParseStatus     `json:"status"`
	Warnings       json.RawMessage `json:"warnings"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CreateParseResultParams struct {
	DiscoveryRunID string
	EvidenceID     string
	ParserName     string
	Status         ParseStatus
	Warnings       []string
	ErrorMessage   *string
}

type ParseResultRepository interface {
	CreateParseResult(context.Context, CreateParseResultParams) (ParseRecord, error)
	ListParseResultsByDiscoveryRun(context.Context, string) ([]ParseRecord, error)
}

type evidenceLister interface {
	ListEvidenceByDiscoveryRun(context.Context, string) ([]evidence.Evidence, error)
}

type assetStore interface {
	CreateAsset(context.Context, assets.CreateAssetParams) (assets.Asset, error)
	ListAssets(context.Context) ([]assets.Asset, error)
	CreateFact(context.Context, assets.CreateFactParams) (assets.Fact, error)
	ListFactsByAsset(context.Context, string) ([]assets.Fact, error)
	CreateRelationship(context.Context, assets.CreateRelationshipParams) (assets.Relationship, error)
	ListRelationships(context.Context) ([]assets.Relationship, error)
}

type PersistenceOptions struct {
	Evidence     evidenceLister
	Assets       assetStore
	ParseResults ParseResultRepository
	Registry     Registry
}

type PersistenceService struct {
	evidence     evidenceLister
	assets       assetStore
	parseResults ParseResultRepository
	registry     Registry
}

func NewPersistenceService(opts PersistenceOptions) PersistenceService {
	registry := opts.Registry
	if registry.parsers == nil && registry.fallback == nil {
		registry = BuiltInRegistry()
	}
	return PersistenceService{
		evidence:     opts.Evidence,
		assets:       opts.Assets,
		parseResults: opts.ParseResults,
		registry:     registry,
	}
}

type ParseDiscoveryRunParams struct {
	DiscoveryRunID string `json:"discovery_run_id"`
	Platform       string `json:"platform"`
}

type ParseDiscoveryRunResult struct {
	DiscoveryRunID string                `json:"discovery_run_id"`
	EvidenceCount  int                   `json:"evidence_count"`
	ParseResults   []ParseRecord         `json:"parse_results"`
	Assets         []assets.Asset        `json:"assets"`
	Facts          []assets.Fact         `json:"facts"`
	Relationships  []assets.Relationship `json:"relationships"`
	Warnings       []string              `json:"warnings,omitempty"`
}

func (s PersistenceService) ParseDiscoveryRun(ctx context.Context, params ParseDiscoveryRunParams) (ParseDiscoveryRunResult, error) {
	if s.evidence == nil {
		return ParseDiscoveryRunResult{}, fmt.Errorf("evidence repository is required")
	}
	if s.assets == nil {
		return ParseDiscoveryRunResult{}, fmt.Errorf("asset repository is required")
	}
	if s.parseResults == nil {
		return ParseDiscoveryRunResult{}, fmt.Errorf("parse result repository is required")
	}
	params.DiscoveryRunID = strings.TrimSpace(params.DiscoveryRunID)
	if params.DiscoveryRunID == "" {
		return ParseDiscoveryRunResult{}, fmt.Errorf("discovery_run_id is required")
	}
	params.Platform = strings.ToLower(strings.TrimSpace(params.Platform))
	if params.Platform == "" {
		return ParseDiscoveryRunResult{}, fmt.Errorf("platform is required")
	}

	items, err := s.evidence.ListEvidenceByDiscoveryRun(ctx, params.DiscoveryRunID)
	if err != nil {
		return ParseDiscoveryRunResult{}, err
	}

	index, err := newAssetIndex(ctx, s.assets)
	if err != nil {
		return ParseDiscoveryRunResult{}, err
	}

	result := ParseDiscoveryRunResult{
		DiscoveryRunID: params.DiscoveryRunID,
		EvidenceCount:  len(items),
	}
	for _, item := range items {
		parsed, parseErr := s.registry.Parse(ctx, params.Platform, item)
		if parseErr != nil {
			message := parseErr.Error()
			record, err := s.parseResults.CreateParseResult(ctx, CreateParseResultParams{
				DiscoveryRunID: item.DiscoveryRunID,
				EvidenceID:     item.ID,
				ParserName:     s.registry.Select(params.Platform, item.CommandOrAPI).Name(),
				Status:         ParseStatusFailed,
				ErrorMessage:   &message,
			})
			if err != nil {
				return result, err
			}
			result.ParseResults = append(result.ParseResults, record)
			result.Warnings = append(result.Warnings, fmt.Sprintf("parser failed for evidence %s: %s", item.ID, message))
			continue
		}

		created, err := persistResult(ctx, s.assets, index, parsed)
		if err != nil {
			return result, err
		}
		result.Assets = append(result.Assets, created.assets...)
		result.Facts = append(result.Facts, created.facts...)
		result.Relationships = append(result.Relationships, created.relationships...)

		status := ParseStatusParsed
		if resultIsEmpty(parsed) {
			status = ParseStatusSkipped
		}
		record, err := s.parseResults.CreateParseResult(ctx, CreateParseResultParams{
			DiscoveryRunID: item.DiscoveryRunID,
			EvidenceID:     item.ID,
			ParserName:     parsed.ParserName,
			Status:         status,
			Warnings:       parsed.Warnings,
		})
		if err != nil {
			return result, err
		}
		result.ParseResults = append(result.ParseResults, record)
		result.Warnings = append(result.Warnings, parsed.Warnings...)
	}

	return result, nil
}

type assetIndex struct {
	byIdentity map[string]assets.Asset
}

func newAssetIndex(ctx context.Context, store assetStore) (*assetIndex, error) {
	items, err := store.ListAssets(ctx)
	if err != nil {
		return nil, err
	}
	index := &assetIndex{byIdentity: make(map[string]assets.Asset, len(items))}
	for _, item := range items {
		index.byIdentity[assets.NormalizeIdentityKey(item.IdentityKey)] = item
	}
	return index, nil
}

type persistedItems struct {
	assets        []assets.Asset
	facts         []assets.Fact
	relationships []assets.Relationship
}

func persistResult(ctx context.Context, store assetStore, index *assetIndex, result Result) (persistedItems, error) {
	var out persistedItems

	for _, item := range result.DeviceIdentities {
		created, err := ensureAsset(ctx, store, index, assets.CreateAssetParams{
			Type:             item.AssetType,
			IdentityKey:      item.IdentityKey,
			Vendor:           optionalString(item.Vendor),
			Model:            optionalString(item.Model),
			Serial:           optionalString(item.Serial),
			SystemMAC:        optionalString(item.SystemMAC),
			Confidence:       defaultConfidence(item.Confidence),
			ConfidenceReason: "parsed device identity from evidence",
			State:            assets.StateObserved,
			Metadata:         item.Metadata,
		})
		if err != nil {
			return out, err
		}
		if created.ID != "" {
			out.assets = append(out.assets, created)
		}
	}

	for _, item := range result.InventoryComponents {
		created, err := ensureAsset(ctx, store, index, assets.CreateAssetParams{
			Type:             item.AssetType,
			IdentityKey:      item.IdentityKey,
			Vendor:           optionalString(item.Vendor),
			Model:            optionalString(item.Model),
			Serial:           optionalString(item.Serial),
			Confidence:       defaultConfidence(item.Confidence),
			ConfidenceReason: "parsed inventory component from evidence",
			State:            assets.StateObserved,
			Metadata:         item.Metadata,
		})
		if err != nil {
			return out, err
		}
		if created.ID != "" {
			out.assets = append(out.assets, created)
		}
		if strings.TrimSpace(item.ParentIdentityKey) != "" {
			parent, err := ensurePlaceholderAsset(ctx, store, index, item.ParentIdentityKey, assets.StateInferred, 0.45)
			if err != nil {
				return out, err
			}
			relationship, err := createRelationshipIfNew(ctx, store, assets.CreateRelationshipParams{
				SourceAssetID:    parent.ID,
				TargetAssetID:    index.byIdentity[assets.NormalizeIdentityKey(item.IdentityKey)].ID,
				RelationshipType: "contains",
				Confidence:       0.7,
				ConfidenceReason: "parsed component parent from evidence",
				State:            assets.StateObserved,
				EvidenceID:       optionalString(result.EvidenceID),
				Metadata:         json.RawMessage(`{}`),
			})
			if err != nil {
				return out, err
			}
			if relationship.ID != "" {
				out.relationships = append(out.relationships, relationship)
			}
		}
	}

	for _, item := range result.Interfaces {
		created, err := ensureAsset(ctx, store, index, assets.CreateAssetParams{
			Type:             item.AssetType,
			IdentityKey:      item.IdentityKey,
			Confidence:       defaultConfidence(item.Confidence),
			ConfidenceReason: "parsed interface from evidence",
			State:            assets.StateObserved,
			Metadata:         item.Metadata,
		})
		if err != nil {
			return out, err
		}
		if created.ID != "" {
			out.assets = append(out.assets, created)
		}
	}

	for _, item := range result.Facts {
		target, err := ensurePlaceholderAsset(ctx, store, index, item.AssetIdentityKey, assets.StateObserved, defaultConfidence(item.Confidence))
		if err != nil {
			return out, err
		}
		evidenceID := optionalString(firstNonEmpty(item.EvidenceID, result.EvidenceID))
		fact, err := createFactIfNew(ctx, store, assets.CreateFactParams{
			AssetID:          target.ID,
			Name:             item.Name,
			Value:            item.Value,
			Source:           firstNonEmpty(item.Source, result.ParserName),
			Confidence:       defaultConfidence(item.Confidence),
			ConfidenceReason: "parsed fact from evidence",
			State:            assets.StateObserved,
			EvidenceID:       evidenceID,
		})
		if err != nil {
			return out, err
		}
		if fact.ID != "" {
			out.facts = append(out.facts, fact)
		}
	}

	for _, item := range result.Relationships {
		source, err := ensurePlaceholderAsset(ctx, store, index, item.SourceIdentityKey, assets.StateObserved, defaultConfidence(item.Confidence))
		if err != nil {
			return out, err
		}
		target, err := ensurePlaceholderAsset(ctx, store, index, item.TargetIdentityKey, assets.StateObserved, defaultConfidence(item.Confidence))
		if err != nil {
			return out, err
		}
		evidenceID := optionalString(firstNonEmpty(item.EvidenceID, result.EvidenceID))
		relationship, err := createRelationshipIfNew(ctx, store, assets.CreateRelationshipParams{
			SourceAssetID:    source.ID,
			TargetAssetID:    target.ID,
			RelationshipType: item.RelationshipType,
			Confidence:       defaultConfidence(item.Confidence),
			ConfidenceReason: "parsed relationship from evidence",
			State:            assets.StateObserved,
			EvidenceID:       evidenceID,
			Metadata:         item.Metadata,
		})
		if err != nil {
			return out, err
		}
		if relationship.ID != "" {
			out.relationships = append(out.relationships, relationship)
		}
	}

	return out, nil
}

func ensureAsset(ctx context.Context, store assetStore, index *assetIndex, params assets.CreateAssetParams) (assets.Asset, error) {
	key := assets.NormalizeIdentityKey(params.IdentityKey)
	if existing, ok := index.byIdentity[key]; ok {
		return existing, nil
	}
	created, err := store.CreateAsset(ctx, params)
	if err != nil {
		return assets.Asset{}, err
	}
	index.byIdentity[key] = created
	return created, nil
}

func ensurePlaceholderAsset(ctx context.Context, store assetStore, index *assetIndex, identityKey string, state assets.ConfidenceState, confidence float64) (assets.Asset, error) {
	identityKey = assets.NormalizeIdentityKey(identityKey)
	if existing, ok := index.byIdentity[identityKey]; ok {
		return existing, nil
	}
	return ensureAsset(ctx, store, index, assets.CreateAssetParams{
		Type:             assetTypeFromIdentityKey(identityKey),
		IdentityKey:      identityKey,
		Confidence:       confidence,
		ConfidenceReason: "placeholder created from parser output identity reference",
		State:            state,
		Metadata:         json.RawMessage(`{}`),
	})
}

func createFactIfNew(ctx context.Context, store assetStore, params assets.CreateFactParams) (assets.Fact, error) {
	existing, err := store.ListFactsByAsset(ctx, params.AssetID)
	if err != nil {
		return assets.Fact{}, err
	}
	for _, item := range existing {
		if item.Name == strings.ToLower(strings.TrimSpace(params.Name)) &&
			item.Source == strings.TrimSpace(params.Source) &&
			sameOptionalString(item.EvidenceID, params.EvidenceID) &&
			jsonEqual(item.Value, params.Value) {
			return assets.Fact{}, nil
		}
	}
	return store.CreateFact(ctx, params)
}

func createRelationshipIfNew(ctx context.Context, store assetStore, params assets.CreateRelationshipParams) (assets.Relationship, error) {
	existing, err := store.ListRelationships(ctx)
	if err != nil {
		return assets.Relationship{}, err
	}
	relationshipType := strings.ToLower(strings.TrimSpace(params.RelationshipType))
	for _, item := range existing {
		if item.SourceAssetID == params.SourceAssetID &&
			item.TargetAssetID == params.TargetAssetID &&
			item.RelationshipType == relationshipType &&
			sameOptionalString(item.EvidenceID, params.EvidenceID) {
			return assets.Relationship{}, nil
		}
	}
	return store.CreateRelationship(ctx, params)
}

func resultIsEmpty(result Result) bool {
	return len(result.DeviceIdentities) == 0 &&
		len(result.InventoryComponents) == 0 &&
		len(result.Interfaces) == 0 &&
		len(result.Neighbors) == 0 &&
		len(result.BGPPeers) == 0 &&
		len(result.Facts) == 0 &&
		len(result.Relationships) == 0
}

func assetTypeFromIdentityKey(identityKey string) string {
	assetType, _, ok := strings.Cut(identityKey, ":")
	if !ok || strings.TrimSpace(assetType) == "" {
		return "unknown"
	}
	return strings.TrimSpace(assetType)
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func defaultConfidence(value float64) float64 {
	if value == 0 {
		return 0.5
	}
	return value
}

func sameOptionalString(a *string, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func jsonEqual(a json.RawMessage, b json.RawMessage) bool {
	var left any
	var right any
	if json.Unmarshal(a, &left) != nil || json.Unmarshal(b, &right) != nil {
		return string(a) == string(b)
	}
	return fmt.Sprintf("%#v", left) == fmt.Sprintf("%#v", right)
}
