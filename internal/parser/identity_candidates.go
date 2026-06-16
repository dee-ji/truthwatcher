package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"truthwatcher/internal/assets"
)

type IdentityReviewState string

const (
	IdentityReviewPending      IdentityReviewState = "pending"
	IdentityReviewAutoAccepted IdentityReviewState = "auto_accepted"
	IdentityReviewAccepted     IdentityReviewState = "accepted"
	IdentityReviewRejected     IdentityReviewState = "rejected"
	IdentityReviewSuperseded   IdentityReviewState = "superseded"
)

type IdentityCandidate struct {
	ID                   string                  `json:"id"`
	DiscoveryRunID       string                  `json:"discovery_run_id"`
	EvidenceID           string                  `json:"evidence_id"`
	ParserName           string                  `json:"parser_name"`
	AssetType            string                  `json:"asset_type"`
	CandidateIdentityKey string                  `json:"candidate_identity_key"`
	Strength             assets.IdentityStrength `json:"strength"`
	Confidence           float64                 `json:"confidence"`
	Reason               string                  `json:"reason"`
	Vendor               *string                 `json:"vendor,omitempty"`
	Model                *string                 `json:"model,omitempty"`
	Serial               *string                 `json:"serial,omitempty"`
	SystemMAC            *string                 `json:"system_mac,omitempty"`
	Hostname             *string                 `json:"hostname,omitempty"`
	ProposedAssetID      *string                 `json:"proposed_asset_id,omitempty"`
	ReviewState          IdentityReviewState     `json:"review_state"`
	Metadata             json.RawMessage         `json:"metadata"`
	CreatedAt            time.Time               `json:"created_at"`
}

type CreateIdentityCandidateParams struct {
	DiscoveryRunID       string
	EvidenceID           string
	ParserName           string
	AssetType            string
	CandidateIdentityKey string
	Strength             assets.IdentityStrength
	Confidence           float64
	Reason               string
	Vendor               *string
	Model                *string
	Serial               *string
	SystemMAC            *string
	Hostname             *string
	ProposedAssetID      *string
	ReviewState          IdentityReviewState
	Metadata             json.RawMessage
}

type IdentityCandidateFilters struct {
	DiscoveryRunID       string
	EvidenceID           string
	ReviewState          IdentityReviewState
	Strength             assets.IdentityStrength
	CandidateIdentityKey string
}

type IdentityCandidateRepository interface {
	CreateIdentityCandidate(context.Context, CreateIdentityCandidateParams) (IdentityCandidate, error)
	ListIdentityCandidates(context.Context, IdentityCandidateFilters) ([]IdentityCandidate, error)
}

type IdentityCandidateService struct {
	repo IdentityCandidateRepository
}

func NewIdentityCandidateService(repo IdentityCandidateRepository) IdentityCandidateService {
	return IdentityCandidateService{repo: repo}
}

func (s IdentityCandidateService) CreateIdentityCandidate(ctx context.Context, params CreateIdentityCandidateParams) (IdentityCandidate, error) {
	if s.repo == nil {
		return IdentityCandidate{}, fmt.Errorf("identity candidate repository is required")
	}
	params.DiscoveryRunID = strings.TrimSpace(params.DiscoveryRunID)
	params.EvidenceID = strings.TrimSpace(params.EvidenceID)
	params.ParserName = strings.TrimSpace(params.ParserName)
	params.AssetType = strings.ToLower(strings.TrimSpace(params.AssetType))
	params.CandidateIdentityKey = assets.NormalizeIdentityKey(params.CandidateIdentityKey)
	params.Reason = strings.TrimSpace(params.Reason)
	params.Vendor = cleanStringPtr(params.Vendor)
	params.Model = cleanStringPtr(params.Model)
	params.Serial = cleanStringPtr(params.Serial)
	params.SystemMAC = cleanStringPtr(params.SystemMAC)
	params.Hostname = cleanStringPtr(params.Hostname)
	params.ProposedAssetID = cleanStringPtr(params.ProposedAssetID)
	if params.DiscoveryRunID == "" {
		return IdentityCandidate{}, fmt.Errorf("discovery_run_id is required")
	}
	if params.EvidenceID == "" {
		return IdentityCandidate{}, fmt.Errorf("evidence_id is required")
	}
	if params.ParserName == "" {
		return IdentityCandidate{}, fmt.Errorf("parser_name is required")
	}
	if params.AssetType == "" {
		return IdentityCandidate{}, fmt.Errorf("asset_type is required")
	}
	if params.CandidateIdentityKey == "" {
		return IdentityCandidate{}, fmt.Errorf("candidate_identity_key is required")
	}
	if !validIdentityStrength(params.Strength) {
		return IdentityCandidate{}, fmt.Errorf("invalid identity strength %q", params.Strength)
	}
	if !validCandidateConfidence(params.Confidence) {
		return IdentityCandidate{}, fmt.Errorf("confidence must be between 0 and 1")
	}
	if params.Reason == "" {
		return IdentityCandidate{}, fmt.Errorf("reason is required")
	}
	if params.ReviewState == "" {
		params.ReviewState = IdentityReviewPending
	}
	if !params.ReviewState.Valid() {
		return IdentityCandidate{}, fmt.Errorf("invalid identity review state %q", params.ReviewState)
	}
	params.Metadata = defaultIdentityCandidateJSON(params.Metadata)
	if !json.Valid(params.Metadata) {
		return IdentityCandidate{}, fmt.Errorf("metadata must be valid JSON")
	}
	return s.repo.CreateIdentityCandidate(ctx, params)
}

func (s IdentityCandidateService) ListIdentityCandidates(ctx context.Context, filters IdentityCandidateFilters) ([]IdentityCandidate, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("identity candidate repository is required")
	}
	filters.DiscoveryRunID = strings.TrimSpace(filters.DiscoveryRunID)
	filters.EvidenceID = strings.TrimSpace(filters.EvidenceID)
	filters.CandidateIdentityKey = assets.NormalizeIdentityKey(filters.CandidateIdentityKey)
	if filters.ReviewState != "" {
		filters.ReviewState = IdentityReviewState(strings.ToLower(strings.TrimSpace(string(filters.ReviewState))))
		if !filters.ReviewState.Valid() {
			return nil, fmt.Errorf("invalid identity review state %q", filters.ReviewState)
		}
	}
	if filters.Strength != "" {
		filters.Strength = assets.IdentityStrength(strings.ToLower(strings.TrimSpace(string(filters.Strength))))
		if !validIdentityStrength(filters.Strength) {
			return nil, fmt.Errorf("invalid identity strength %q", filters.Strength)
		}
	}
	return s.repo.ListIdentityCandidates(ctx, filters)
}

func (s IdentityReviewState) Valid() bool {
	switch s {
	case IdentityReviewPending, IdentityReviewAutoAccepted, IdentityReviewAccepted, IdentityReviewRejected, IdentityReviewSuperseded:
		return true
	default:
		return false
	}
}

func identityCandidatesFromResult(discoveryRunID string, result Result) []CreateIdentityCandidateParams {
	candidates := make([]CreateIdentityCandidateParams, 0, len(result.DeviceIdentities)+len(result.InventoryComponents)+len(result.Interfaces)+len(result.Neighbors))
	for _, item := range result.DeviceIdentities {
		candidates = append(candidates, candidateFromAssetRef(discoveryRunID, result.EvidenceID, result.ParserName, item.AssetRef, item.Vendor, item.Model, item.Serial, item.SystemMAC, item.Hostname, item.Metadata))
	}
	for _, item := range result.InventoryComponents {
		candidates = append(candidates, candidateFromAssetRef(discoveryRunID, result.EvidenceID, result.ParserName, item.AssetRef, item.Vendor, item.Model, item.Serial, "", "", item.Metadata))
	}
	for _, item := range result.Interfaces {
		candidates = append(candidates, candidateFromAssetRef(discoveryRunID, result.EvidenceID, result.ParserName, item.AssetRef, "", "", "", item.MACAddress, "", item.Metadata))
	}
	for _, item := range result.Neighbors {
		if strings.TrimSpace(item.RemoteIdentityKey) == "" {
			continue
		}
		candidate := assets.IdentityCandidateFromKey("device", item.RemoteIdentityKey)
		confidence := defaultConfidence(item.Confidence)
		candidates = append(candidates, CreateIdentityCandidateParams{
			DiscoveryRunID:       strings.TrimSpace(discoveryRunID),
			EvidenceID:           strings.TrimSpace(result.EvidenceID),
			ParserName:           strings.TrimSpace(result.ParserName),
			AssetType:            candidate.AssetType,
			CandidateIdentityKey: candidate.IdentityKey,
			Strength:             candidate.Strength,
			Confidence:           confidence,
			Reason:               candidate.Reason,
			Hostname:             optionalString(item.RemoteSystemName),
			ReviewState:          IdentityReviewPending,
			Metadata:             item.Metadata,
		})
	}
	return candidates
}

func candidateFromAssetRef(discoveryRunID string, resultEvidenceID string, parserName string, ref AssetRef, vendor string, model string, serial string, systemMAC string, hostname string, metadata json.RawMessage) CreateIdentityCandidateParams {
	identity := assets.IdentityCandidateFromKey(ref.AssetType, ref.IdentityKey)
	return CreateIdentityCandidateParams{
		DiscoveryRunID:       strings.TrimSpace(discoveryRunID),
		EvidenceID:           firstNonEmpty(ref.EvidenceID, resultEvidenceID),
		ParserName:           strings.TrimSpace(parserName),
		AssetType:            identity.AssetType,
		CandidateIdentityKey: identity.IdentityKey,
		Strength:             identity.Strength,
		Confidence:           defaultConfidence(ref.Confidence),
		Reason:               identity.Reason,
		Vendor:               optionalString(vendor),
		Model:                optionalString(model),
		Serial:               optionalString(serial),
		SystemMAC:            optionalString(systemMAC),
		Hostname:             optionalString(hostname),
		ReviewState:          IdentityReviewPending,
		Metadata:             metadata,
	}
}

func validIdentityStrength(value assets.IdentityStrength) bool {
	switch value {
	case assets.IdentityStrengthStrong, assets.IdentityStrengthProvisional, assets.IdentityStrengthWeak:
		return true
	default:
		return false
	}
}

func validCandidateConfidence(value float64) bool {
	return value >= 0 && value <= 1
}

func defaultIdentityCandidateJSON(value json.RawMessage) json.RawMessage {
	if strings.TrimSpace(string(value)) == "" {
		return json.RawMessage(`{}`)
	}
	return value
}

func cleanStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}
