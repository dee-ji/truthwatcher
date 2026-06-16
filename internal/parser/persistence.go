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
	Evidence           evidenceLister
	Assets             assetStore
	ParseResults       ParseResultRepository
	IdentityCandidates IdentityCandidateRepository
	Registry           Registry
}

type PersistenceService struct {
	evidence           evidenceLister
	assets             assetStore
	parseResults       ParseResultRepository
	identityCandidates IdentityCandidateRepository
	registry           Registry
}

func NewPersistenceService(opts PersistenceOptions) PersistenceService {
	registry := opts.Registry
	if registry.parsers == nil && registry.fallback == nil {
		registry = BuiltInRegistry()
	}
	return PersistenceService{
		evidence:           opts.Evidence,
		assets:             opts.Assets,
		parseResults:       opts.ParseResults,
		identityCandidates: opts.IdentityCandidates,
		registry:           registry,
	}
}

type ParseDiscoveryRunParams struct {
	DiscoveryRunID string `json:"discovery_run_id"`
	Platform       string `json:"platform"`
}

type ParseDiscoveryRunResult struct {
	DiscoveryRunID     string                `json:"discovery_run_id"`
	EvidenceCount      int                   `json:"evidence_count"`
	ParseResults       []ParseRecord         `json:"parse_results"`
	IdentityCandidates []IdentityCandidate   `json:"identity_candidates,omitempty"`
	Assets             []assets.Asset        `json:"assets"`
	Facts              []assets.Fact         `json:"facts"`
	Relationships      []assets.Relationship `json:"relationships"`
	Warnings           []string              `json:"warnings,omitempty"`
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

		candidates, err := s.persistIdentityCandidates(ctx, item.DiscoveryRunID, parsed)
		if err != nil {
			return result, err
		}
		result.IdentityCandidates = append(result.IdentityCandidates, candidates...)

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

func (s PersistenceService) persistIdentityCandidates(ctx context.Context, discoveryRunID string, result Result) ([]IdentityCandidate, error) {
	if s.identityCandidates == nil {
		return nil, nil
	}
	service := NewIdentityCandidateService(s.identityCandidates)
	params := identityCandidatesFromResult(discoveryRunID, result)
	index, err := newAssetIndex(ctx, s.assets)
	if err != nil {
		return nil, err
	}
	created := make([]IdentityCandidate, 0, len(params))
	for _, candidate := range params {
		decision := evaluateIdentityCandidate(candidate, index)
		candidate.ReviewState = IdentityReviewPending
		candidate.Metadata = identityCandidateDecisionMetadata(candidate.Metadata, decision)
		item, err := service.CreateIdentityCandidate(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if decision.AutoAccept {
			if err := service.AutoAcceptIdentityCandidate(ctx, AutoAcceptIdentityCandidateParams{
				IdentityCandidateID: item.ID,
				Rationale:           decision.Explanation,
				Metadata:            identityCandidateAutoReviewMetadata(item, decision),
			}); err != nil {
				return nil, err
			}
			item.ReviewState = IdentityReviewAutoAccepted
			item.Metadata = candidate.Metadata
		}
		created = append(created, item)
	}
	return created, nil
}

type identityCandidateDecision struct {
	AutoAccept  bool
	Rule        string
	Explanation string
	Conflict    string
}

func evaluateIdentityCandidate(candidate CreateIdentityCandidateParams, index *assetIndex) identityCandidateDecision {
	if candidate.Strength != assets.IdentityStrengthStrong {
		return identityCandidateDecision{
			Rule:        "queue_non_strong_candidate",
			Explanation: "queued for review because hostname, name, weak, or provisional identity evidence is not silently authoritative",
		}
	}
	if !hasAutoAcceptableStrongIdentifier(candidate) {
		return identityCandidateDecision{
			Rule:        "queue_strong_identifier_outside_initial_auto_accept_scope",
			Explanation: "queued for review because this strong identity is outside the initial vendor+serial or system-MAC auto-acceptance scope",
		}
	}
	if conflict := candidateAssetConflict(candidate, index); conflict != "" {
		return identityCandidateDecision{
			Rule:        "queue_plausible_canonical_asset_conflict",
			Explanation: "queued for review because existing canonical asset state could plausibly conflict with this identity candidate",
			Conflict:    conflict,
		}
	}
	return identityCandidateDecision{
		AutoAccept:  true,
		Rule:        "auto_accept_strong_no_plausible_conflict",
		Explanation: "auto-accepted because candidate is evidence-backed strong vendor+serial or system-MAC identity with no plausible canonical asset conflict",
	}
}

func hasAutoAcceptableStrongIdentifier(candidate CreateIdentityCandidateParams) bool {
	if candidate.Vendor != nil && candidate.Serial != nil {
		return true
	}
	if candidate.SystemMAC != nil {
		return true
	}
	key := assets.NormalizeIdentityKey(candidate.CandidateIdentityKey)
	return strings.Contains(key, ":vendor_serial:") || strings.Contains(key, ":system_mac:")
}

func candidateAssetConflict(candidate CreateIdentityCandidateParams, index *assetIndex) string {
	if index == nil {
		return ""
	}
	candidateKey := assets.NormalizeIdentityKey(candidate.CandidateIdentityKey)
	for _, item := range index.byIdentity {
		assetKey := assets.NormalizeIdentityKey(item.IdentityKey)
		if assetKey == candidateKey {
			if conflictsWithCanonicalAttributes(candidate, item) {
				return "candidate attributes differ from canonical asset with the same identity key"
			}
			continue
		}
		if candidate.Serial != nil && item.Serial != nil && normalizedComparable(*candidate.Serial) == normalizedComparable(*item.Serial) {
			return "candidate serial is already present on a different canonical asset"
		}
		if candidate.SystemMAC != nil && item.SystemMAC != nil && normalizedComparable(*candidate.SystemMAC) == normalizedComparable(*item.SystemMAC) {
			return "candidate system MAC is already present on a different canonical asset"
		}
	}
	return ""
}

func conflictsWithCanonicalAttributes(candidate CreateIdentityCandidateParams, item assets.Asset) bool {
	return optionalComparableConflict(candidate.Vendor, item.Vendor) ||
		optionalComparableConflict(candidate.Serial, item.Serial) ||
		optionalComparableConflict(candidate.SystemMAC, item.SystemMAC)
}

func optionalComparableConflict(candidate *string, existing *string) bool {
	if candidate == nil || existing == nil {
		return false
	}
	return normalizedComparable(*candidate) != normalizedComparable(*existing)
}

func normalizedComparable(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func identityCandidateDecisionMetadata(metadata json.RawMessage, decision identityCandidateDecision) json.RawMessage {
	payload := map[string]any{}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["identity_review_rule"] = decision.Rule
	payload["identity_review_explanation"] = decision.Explanation
	if decision.AutoAccept {
		payload["identity_review_state"] = IdentityReviewAutoAccepted
	} else {
		payload["identity_review_state"] = IdentityReviewPending
	}
	if decision.Conflict != "" {
		payload["identity_review_conflict"] = decision.Conflict
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return metadata
	}
	return encoded
}

func identityCandidateAutoReviewMetadata(candidate IdentityCandidate, decision identityCandidateDecision) json.RawMessage {
	payload := map[string]any{
		"decision_type":          "deterministic_auto_acceptance",
		"identity_review_rule":   decision.Rule,
		"candidate_identity_key": candidate.CandidateIdentityKey,
		"asset_type":             candidate.AssetType,
		"strength":               candidate.Strength,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return encoded
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
