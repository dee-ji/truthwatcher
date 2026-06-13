package assets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var ErrNotFound = errors.New("asset model record not found")

type ConfidenceState string

const (
	StateObserved    ConfidenceState = "observed"
	StateInferred    ConfidenceState = "inferred"
	StateUserSeeded  ConfidenceState = "user_seeded"
	StateConflicting ConfidenceState = "conflicting"
	StateUnknown     ConfidenceState = "unknown"
)

type Asset struct {
	ID               string          `json:"id"`
	Type             string          `json:"type"`
	IdentityKey      string          `json:"identity_key"`
	Vendor           *string         `json:"vendor,omitempty"`
	Model            *string         `json:"model,omitempty"`
	Serial           *string         `json:"serial,omitempty"`
	SystemMAC        *string         `json:"system_mac,omitempty"`
	Confidence       float64         `json:"confidence"`
	ConfidenceReason string          `json:"confidence_reason"`
	State            ConfidenceState `json:"state"`
	Metadata         json.RawMessage `json:"metadata"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type Fact struct {
	ID               string          `json:"id"`
	AssetID          string          `json:"asset_id"`
	Name             string          `json:"name"`
	Value            json.RawMessage `json:"value"`
	Source           string          `json:"source"`
	Confidence       float64         `json:"confidence"`
	ConfidenceReason string          `json:"confidence_reason"`
	State            ConfidenceState `json:"state"`
	EvidenceID       *string         `json:"evidence_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

type Relationship struct {
	ID               string          `json:"id"`
	SourceAssetID    string          `json:"source_asset_id"`
	TargetAssetID    string          `json:"target_asset_id"`
	RelationshipType string          `json:"relationship_type"`
	Confidence       float64         `json:"confidence"`
	ConfidenceReason string          `json:"confidence_reason"`
	State            ConfidenceState `json:"state"`
	EvidenceID       *string         `json:"evidence_id,omitempty"`
	Metadata         json.RawMessage `json:"metadata"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type CreateAssetParams struct {
	Type             string
	IdentityKey      string
	Vendor           *string
	Model            *string
	Serial           *string
	SystemMAC        *string
	Confidence       float64
	ConfidenceReason string
	State            ConfidenceState
	Metadata         json.RawMessage
}

type CreateFactParams struct {
	AssetID          string
	Name             string
	Value            json.RawMessage
	Source           string
	Confidence       float64
	ConfidenceReason string
	State            ConfidenceState
	EvidenceID       *string
}

type CreateRelationshipParams struct {
	SourceAssetID    string
	TargetAssetID    string
	RelationshipType string
	Confidence       float64
	ConfidenceReason string
	State            ConfidenceState
	EvidenceID       *string
	Metadata         json.RawMessage
}

type Repository interface {
	CreateAsset(context.Context, CreateAssetParams) (Asset, error)
	GetAsset(context.Context, string) (Asset, error)
	ListAssets(context.Context) ([]Asset, error)
	CreateFact(context.Context, CreateFactParams) (Fact, error)
	GetFact(context.Context, string) (Fact, error)
	ListFactsByAsset(context.Context, string) ([]Fact, error)
	CreateRelationship(context.Context, CreateRelationshipParams) (Relationship, error)
	GetRelationship(context.Context, string) (Relationship, error)
	ListRelationships(context.Context) ([]Relationship, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) CreateAsset(ctx context.Context, params CreateAssetParams) (Asset, error) {
	if s.repo == nil {
		return Asset{}, fmt.Errorf("asset repository is required")
	}
	if strings.TrimSpace(params.Type) == "" {
		return Asset{}, fmt.Errorf("asset type is required")
	}
	if strings.TrimSpace(params.IdentityKey) == "" {
		return Asset{}, fmt.Errorf("identity_key is required")
	}

	params.Type = normalizeToken(params.Type)
	params.IdentityKey = NormalizeIdentityKey(params.IdentityKey)
	if params.Confidence == 0 && strings.TrimSpace(string(params.State)) == "" {
		params.Confidence = 0.5
	}
	if !validConfidence(params.Confidence) {
		return Asset{}, fmt.Errorf("confidence must be between 0 and 1")
	}
	params.State = defaultConfidenceState(params.State, nil)
	if !params.State.Valid() {
		return Asset{}, fmt.Errorf("invalid confidence state %q", params.State)
	}
	params.ConfidenceReason = defaultConfidenceReason(params.ConfidenceReason, params.State)
	params.Metadata = defaultJSON(params.Metadata)
	if !json.Valid(params.Metadata) {
		return Asset{}, fmt.Errorf("metadata must be valid JSON")
	}
	params.Metadata = AnnotateIdentityMetadata(params.Metadata, IdentityCandidateFromKey(params.Type, params.IdentityKey))

	return s.repo.CreateAsset(ctx, params)
}

func (s Service) GetAsset(ctx context.Context, id string) (Asset, error) {
	if s.repo == nil {
		return Asset{}, fmt.Errorf("asset repository is required")
	}
	if strings.TrimSpace(id) == "" {
		return Asset{}, fmt.Errorf("asset id is required")
	}
	return s.repo.GetAsset(ctx, id)
}

func (s Service) ListAssets(ctx context.Context) ([]Asset, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("asset repository is required")
	}
	return s.repo.ListAssets(ctx)
}

func (s Service) CreateFact(ctx context.Context, params CreateFactParams) (Fact, error) {
	if s.repo == nil {
		return Fact{}, fmt.Errorf("asset repository is required")
	}
	if strings.TrimSpace(params.AssetID) == "" {
		return Fact{}, fmt.Errorf("asset_id is required")
	}
	if strings.TrimSpace(params.Name) == "" {
		return Fact{}, fmt.Errorf("fact name is required")
	}
	if strings.TrimSpace(params.Source) == "" {
		return Fact{}, fmt.Errorf("fact source is required")
	}
	if !validConfidence(params.Confidence) {
		return Fact{}, fmt.Errorf("confidence must be between 0 and 1")
	}

	params.Name = normalizeToken(params.Name)
	params.Source = strings.TrimSpace(params.Source)
	params.State = defaultConfidenceState(params.State, params.EvidenceID)
	if !params.State.Valid() {
		return Fact{}, fmt.Errorf("invalid confidence state %q", params.State)
	}
	params.ConfidenceReason = defaultConfidenceReason(params.ConfidenceReason, params.State)
	if len(params.Value) == 0 {
		return Fact{}, fmt.Errorf("fact value is required")
	}
	if !json.Valid(params.Value) {
		return Fact{}, fmt.Errorf("fact value must be valid JSON")
	}
	existing, err := s.repo.ListFactsByAsset(ctx, params.AssetID)
	if err != nil {
		return Fact{}, err
	}
	if conflict := conflictingFact(existing, params); conflict != nil {
		params.State = StateConflicting
		params.ConfidenceReason = fmt.Sprintf("conflicts with existing fact %s", conflict.ID)
	}

	return s.repo.CreateFact(ctx, params)
}

func (s Service) GetFact(ctx context.Context, id string) (Fact, error) {
	if s.repo == nil {
		return Fact{}, fmt.Errorf("asset repository is required")
	}
	if strings.TrimSpace(id) == "" {
		return Fact{}, fmt.Errorf("fact id is required")
	}
	return s.repo.GetFact(ctx, id)
}

func (s Service) ListFactsByAsset(ctx context.Context, assetID string) ([]Fact, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("asset repository is required")
	}
	if strings.TrimSpace(assetID) == "" {
		return nil, fmt.Errorf("asset id is required")
	}
	return s.repo.ListFactsByAsset(ctx, assetID)
}

func (s Service) CreateRelationship(ctx context.Context, params CreateRelationshipParams) (Relationship, error) {
	if s.repo == nil {
		return Relationship{}, fmt.Errorf("asset repository is required")
	}
	if strings.TrimSpace(params.SourceAssetID) == "" {
		return Relationship{}, fmt.Errorf("source_asset_id is required")
	}
	if strings.TrimSpace(params.TargetAssetID) == "" {
		return Relationship{}, fmt.Errorf("target_asset_id is required")
	}
	if strings.TrimSpace(params.RelationshipType) == "" {
		return Relationship{}, fmt.Errorf("relationship_type is required")
	}
	if !validConfidence(params.Confidence) {
		return Relationship{}, fmt.Errorf("confidence must be between 0 and 1")
	}

	params.RelationshipType = normalizeToken(params.RelationshipType)
	params.State = defaultConfidenceState(params.State, params.EvidenceID)
	if !params.State.Valid() {
		return Relationship{}, fmt.Errorf("invalid confidence state %q", params.State)
	}
	params.ConfidenceReason = defaultConfidenceReason(params.ConfidenceReason, params.State)
	params.Metadata = defaultJSON(params.Metadata)
	if !json.Valid(params.Metadata) {
		return Relationship{}, fmt.Errorf("metadata must be valid JSON")
	}

	return s.repo.CreateRelationship(ctx, params)
}

func (s Service) GetRelationship(ctx context.Context, id string) (Relationship, error) {
	if s.repo == nil {
		return Relationship{}, fmt.Errorf("asset repository is required")
	}
	if strings.TrimSpace(id) == "" {
		return Relationship{}, fmt.Errorf("relationship id is required")
	}
	return s.repo.GetRelationship(ctx, id)
}

func (s Service) ListRelationships(ctx context.Context) ([]Relationship, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("asset repository is required")
	}
	return s.repo.ListRelationships(ctx)
}

func (s Service) ListConflictingFacts(ctx context.Context) ([]Fact, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("asset repository is required")
	}
	items, err := s.repo.ListAssets(ctx)
	if err != nil {
		return nil, err
	}
	var conflicts []Fact
	for _, item := range items {
		facts, err := s.repo.ListFactsByAsset(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		for _, fact := range facts {
			if fact.State == StateConflicting {
				conflicts = append(conflicts, fact)
			}
		}
	}
	return conflicts, nil
}

func (s Service) ListProvisionalIdentityAssets(ctx context.Context) ([]Asset, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("asset repository is required")
	}
	items, err := s.repo.ListAssets(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Asset, 0)
	for _, item := range items {
		candidate := IdentityCandidateForStoredAsset(item)
		if candidate.Strength != IdentityStrengthStrong {
			result = append(result, item)
		}
	}
	return result, nil
}

func MakeIdentityKey(assetType, source, value string) string {
	return strings.Join([]string{
		normalizeToken(assetType),
		normalizeToken(source),
		strings.ToLower(strings.TrimSpace(value)),
	}, ":")
}

func NormalizeIdentityKey(identityKey string) string {
	return strings.ToLower(strings.TrimSpace(identityKey))
}

func defaultJSON(value json.RawMessage) json.RawMessage {
	if strings.TrimSpace(string(value)) == "" {
		return json.RawMessage(`{}`)
	}
	return value
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validConfidence(value float64) bool {
	return value >= 0 && value <= 1
}

func (s ConfidenceState) Valid() bool {
	switch s {
	case StateObserved, StateInferred, StateUserSeeded, StateConflicting, StateUnknown:
		return true
	default:
		return false
	}
}

func defaultConfidenceState(state ConfidenceState, evidenceID *string) ConfidenceState {
	if strings.TrimSpace(string(state)) != "" {
		return ConfidenceState(normalizeToken(string(state)))
	}
	if evidenceID != nil && strings.TrimSpace(*evidenceID) != "" {
		return StateObserved
	}
	return StateInferred
}

func defaultConfidenceReason(reason string, state ConfidenceState) string {
	reason = strings.TrimSpace(reason)
	if reason != "" {
		return reason
	}
	switch state {
	case StateObserved:
		return "directly observed from evidence"
	case StateInferred:
		return "deterministically inferred without direct evidence"
	case StateUserSeeded:
		return "provided by user seed input"
	case StateConflicting:
		return "conflicts with another recorded fact"
	case StateUnknown:
		return "unknown or insufficient evidence"
	default:
		return ""
	}
}

func conflictingFact(existing []Fact, params CreateFactParams) *Fact {
	for i := range existing {
		if existing[i].Name != params.Name {
			continue
		}
		if existing[i].State == StateConflicting {
			continue
		}
		if !jsonEqual(existing[i].Value, params.Value) {
			return &existing[i]
		}
	}
	return nil
}

func jsonEqual(left json.RawMessage, right json.RawMessage) bool {
	var leftValue any
	var rightValue any
	if err := json.Unmarshal(left, &leftValue); err != nil {
		return string(left) == string(right)
	}
	if err := json.Unmarshal(right, &rightValue); err != nil {
		return string(left) == string(right)
	}
	return reflect.DeepEqual(leftValue, rightValue)
}
