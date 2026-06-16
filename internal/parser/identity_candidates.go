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
	IdentityReviewDeferred     IdentityReviewState = "deferred"
	IdentityReviewMoreEvidence IdentityReviewState = "more_evidence_requested"
)

type IdentityReviewAction string

const (
	IdentityReviewActionAccept              IdentityReviewAction = "accept"
	IdentityReviewActionAutoAccept          IdentityReviewAction = "auto_accept"
	IdentityReviewActionReject              IdentityReviewAction = "reject"
	IdentityReviewActionDefer               IdentityReviewAction = "defer"
	IdentityReviewActionRequestMoreEvidence IdentityReviewAction = "request_more_evidence"
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

type IdentityReviewHandoffFilters struct {
	DiscoveryRunID string
	EvidenceID     string
}

type IdentityCandidateRepository interface {
	CreateIdentityCandidate(context.Context, CreateIdentityCandidateParams) (IdentityCandidate, error)
	GetIdentityCandidate(context.Context, string) (IdentityCandidate, error)
	ListIdentityCandidates(context.Context, IdentityCandidateFilters) ([]IdentityCandidate, error)
	ReviewIdentityCandidate(context.Context, ReviewIdentityCandidateParams) (IdentityCandidateReview, error)
	AutoAcceptIdentityCandidate(context.Context, AutoAcceptIdentityCandidateParams) error
	ListIdentityReviewHandoffEntries(context.Context, IdentityReviewHandoffFilters) ([]IdentityReviewHandoffEntry, error)
	ListOrphanedIdentityCandidateReviews(context.Context) ([]IdentityCandidateReview, error)
}

type IdentityCandidateService struct {
	repo IdentityCandidateRepository
}

type ReviewIdentityCandidateParams struct {
	IdentityCandidateID string
	Reviewer            string
	Action              IdentityReviewAction
	Rationale           string
	Metadata            json.RawMessage
}

type AutoAcceptIdentityCandidateParams struct {
	IdentityCandidateID string
	Rationale           string
	Metadata            json.RawMessage
}

type IdentityCandidateReview struct {
	ID                   string               `json:"id"`
	IdentityCandidateID  string               `json:"identity_candidate_id"`
	DiscoveryRunID       string               `json:"discovery_run_id"`
	EvidenceID           string               `json:"evidence_id"`
	Reviewer             string               `json:"reviewer"`
	Action               IdentityReviewAction `json:"action"`
	PreviousReviewState  IdentityReviewState  `json:"previous_review_state"`
	ResultingReviewState IdentityReviewState  `json:"resulting_review_state"`
	Rationale            string               `json:"rationale"`
	Effect               string               `json:"effect"`
	Metadata             json.RawMessage      `json:"metadata"`
	CreatedAt            time.Time            `json:"created_at"`
}

type IdentityReviewHandoffReport struct {
	ReportType              string                       `json:"report_type"`
	Boundary                string                       `json:"boundary"`
	DerivedOutput           bool                         `json:"derived_output"`
	GeneratedAt             time.Time                    `json:"generated_at"`
	Entries                 []IdentityReviewHandoffEntry `json:"entries"`
	Integrity               IdentityReviewIntegrity      `json:"integrity"`
	NonDestructiveGuarantee string                       `json:"non_destructive_guarantee"`
}

type IdentityReviewHandoffEntry struct {
	HandoffStatus       string                   `json:"handoff_status"`
	OutputLabel         string                   `json:"output_label"`
	Candidate           IdentityCandidate        `json:"candidate"`
	LatestReview        *IdentityCandidateReview `json:"latest_review,omitempty"`
	EvidenceReference   IdentityEvidenceRef      `json:"evidence_reference"`
	ParserSource        IdentityParserSource     `json:"parser_source"`
	ReviewSummary       string                   `json:"review_summary"`
	IdentityEffect      string                   `json:"identity_effect"`
	MistsprenIntakeNote string                   `json:"mistspren_intake_note"`
	IntegrityWarnings   []string                 `json:"integrity_warnings,omitempty"`
}

type IdentityEvidenceRef struct {
	EvidenceID     string `json:"evidence_id"`
	DiscoveryRunID string `json:"discovery_run_id"`
	Present        bool   `json:"present"`
}

type IdentityParserSource struct {
	ParserName string          `json:"parser_name"`
	Metadata   json.RawMessage `json:"metadata"`
}

type IdentityReviewIntegrity struct {
	MissingEvidenceReferences int      `json:"missing_evidence_references"`
	OrphanedReviewRecords     int      `json:"orphaned_review_records"`
	UnresolvedPendingEntries  int      `json:"unresolved_pending_entries"`
	Warnings                  []string `json:"warnings,omitempty"`
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

func (s IdentityCandidateService) IdentityReviewHandoffReport(ctx context.Context, filters IdentityReviewHandoffFilters) (IdentityReviewHandoffReport, error) {
	if s.repo == nil {
		return IdentityReviewHandoffReport{}, fmt.Errorf("identity candidate repository is required")
	}
	filters.DiscoveryRunID = strings.TrimSpace(filters.DiscoveryRunID)
	filters.EvidenceID = strings.TrimSpace(filters.EvidenceID)
	entries, err := s.repo.ListIdentityReviewHandoffEntries(ctx, filters)
	if err != nil {
		return IdentityReviewHandoffReport{}, err
	}
	orphaned, err := s.repo.ListOrphanedIdentityCandidateReviews(ctx)
	if err != nil {
		return IdentityReviewHandoffReport{}, err
	}
	report := IdentityReviewHandoffReport{
		ReportType:              "identity_review_handoff",
		Boundary:                "Truthwatcher derived review output for Mistspren intake/workbench review; not an accepted ADR or authoritative Mistspren decision",
		DerivedOutput:           true,
		GeneratedAt:             time.Now().UTC(),
		Entries:                 entries,
		NonDestructiveGuarantee: "report generation is read-only and does not merge canonical assets, rewrite assets.identity_key, or write to Mistspren",
	}
	for i := range report.Entries {
		normalizeIdentityReviewHandoffEntry(&report.Entries[i])
		if !report.Entries[i].EvidenceReference.Present {
			report.Integrity.MissingEvidenceReferences++
		}
		if report.Entries[i].Candidate.ReviewState == IdentityReviewPending {
			report.Integrity.UnresolvedPendingEntries++
		}
	}
	report.Integrity.OrphanedReviewRecords = len(orphaned)
	if report.Integrity.MissingEvidenceReferences > 0 {
		report.Integrity.Warnings = append(report.Integrity.Warnings, "one or more identity candidates reference missing evidence")
	}
	if report.Integrity.OrphanedReviewRecords > 0 {
		report.Integrity.Warnings = append(report.Integrity.Warnings, "one or more identity candidate reviews do not resolve to a candidate")
	}
	if report.Integrity.UnresolvedPendingEntries > 0 {
		report.Integrity.Warnings = append(report.Integrity.Warnings, "one or more identity candidates remain pending and require review")
	}
	return report, nil
}

func (s IdentityCandidateService) ReviewIdentityCandidate(ctx context.Context, params ReviewIdentityCandidateParams) (IdentityCandidateReview, error) {
	if s.repo == nil {
		return IdentityCandidateReview{}, fmt.Errorf("identity candidate repository is required")
	}
	params.IdentityCandidateID = strings.TrimSpace(params.IdentityCandidateID)
	params.Reviewer = strings.TrimSpace(params.Reviewer)
	params.Rationale = strings.TrimSpace(params.Rationale)
	if params.IdentityCandidateID == "" {
		return IdentityCandidateReview{}, fmt.Errorf("identity_candidate_id is required")
	}
	if params.Reviewer == "" {
		return IdentityCandidateReview{}, fmt.Errorf("reviewer is required")
	}
	if !params.Action.Valid() {
		return IdentityCandidateReview{}, fmt.Errorf("invalid identity review action %q", params.Action)
	}
	if params.Action == IdentityReviewActionAutoAccept {
		return IdentityCandidateReview{}, fmt.Errorf("auto_accept is reserved for deterministic parser decisions")
	}
	if params.Rationale == "" {
		return IdentityCandidateReview{}, fmt.Errorf("rationale is required")
	}
	params.Metadata = defaultIdentityCandidateJSON(params.Metadata)
	if !json.Valid(params.Metadata) {
		return IdentityCandidateReview{}, fmt.Errorf("metadata must be valid JSON")
	}
	return s.repo.ReviewIdentityCandidate(ctx, params)
}

func (s IdentityCandidateService) AutoAcceptIdentityCandidate(ctx context.Context, params AutoAcceptIdentityCandidateParams) error {
	if s.repo == nil {
		return fmt.Errorf("identity candidate repository is required")
	}
	params.IdentityCandidateID = strings.TrimSpace(params.IdentityCandidateID)
	params.Rationale = strings.TrimSpace(params.Rationale)
	if params.IdentityCandidateID == "" {
		return fmt.Errorf("identity_candidate_id is required")
	}
	if params.Rationale == "" {
		return fmt.Errorf("rationale is required")
	}
	params.Metadata = defaultIdentityCandidateJSON(params.Metadata)
	if !json.Valid(params.Metadata) {
		return fmt.Errorf("metadata must be valid JSON")
	}
	return s.repo.AutoAcceptIdentityCandidate(ctx, params)
}

func normalizeIdentityReviewHandoffEntry(entry *IdentityReviewHandoffEntry) {
	entry.OutputLabel = "derived_identity_review_output_not_raw_evidence"
	entry.EvidenceReference.EvidenceID = entry.Candidate.EvidenceID
	entry.EvidenceReference.DiscoveryRunID = entry.Candidate.DiscoveryRunID
	entry.ParserSource.ParserName = entry.Candidate.ParserName
	entry.ParserSource.Metadata = defaultIdentityCandidateJSON(entry.Candidate.Metadata)
	if entry.Candidate.ReviewState == IdentityReviewPending {
		entry.HandoffStatus = "unresolved_pending_review"
		entry.ReviewSummary = identityReviewExplanation(entry.Candidate.Metadata, "pending identity candidate requires Truthwatcher review before Mistspren intake")
		entry.IdentityEffect = "no review decision recorded; no canonical asset merge or identity rewrite performed"
		entry.MistsprenIntakeNote = "do not treat this candidate as a reviewed identity decision"
	} else {
		entry.HandoffStatus = "ready_for_mistspren_review"
		entry.ReviewSummary = identityReviewExplanation(entry.Candidate.Metadata, "reviewed identity candidate is ready for Mistspren workbench inspection")
		entry.IdentityEffect = "review state recorded; no canonical asset merge or identity rewrite performed"
		entry.MistsprenIntakeNote = "derived Truthwatcher review output for intake review only"
	}
	if entry.LatestReview != nil {
		entry.ReviewSummary = entry.LatestReview.Rationale
		entry.IdentityEffect = entry.LatestReview.Effect
	}
	if !entry.EvidenceReference.Present {
		entry.IntegrityWarnings = append(entry.IntegrityWarnings, "candidate evidence reference is missing")
	}
}

func identityReviewExplanation(metadata json.RawMessage, fallback string) string {
	var payload map[string]any
	if err := json.Unmarshal(metadata, &payload); err != nil {
		return fallback
	}
	if value, ok := payload["identity_review_explanation"].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func (s IdentityReviewState) Valid() bool {
	switch s {
	case IdentityReviewPending, IdentityReviewAutoAccepted, IdentityReviewAccepted, IdentityReviewRejected, IdentityReviewSuperseded, IdentityReviewDeferred, IdentityReviewMoreEvidence:
		return true
	default:
		return false
	}
}

func (a IdentityReviewAction) Valid() bool {
	switch a {
	case IdentityReviewActionAccept, IdentityReviewActionAutoAccept, IdentityReviewActionReject, IdentityReviewActionDefer, IdentityReviewActionRequestMoreEvidence:
		return true
	default:
		return false
	}
}

func ResultingReviewState(action IdentityReviewAction) IdentityReviewState {
	switch action {
	case IdentityReviewActionAutoAccept:
		return IdentityReviewAutoAccepted
	case IdentityReviewActionAccept:
		return IdentityReviewAccepted
	case IdentityReviewActionReject:
		return IdentityReviewRejected
	case IdentityReviewActionDefer:
		return IdentityReviewDeferred
	case IdentityReviewActionRequestMoreEvidence:
		return IdentityReviewMoreEvidence
	default:
		return ""
	}
}

func IdentityReviewEffect(action IdentityReviewAction) string {
	switch action {
	case IdentityReviewActionAutoAccept:
		return "deterministically auto-accepted evidence-backed strong identity candidate; no canonical asset merge or identity rewrite performed"
	case IdentityReviewActionAccept:
		return "review accepted candidate as evidence-backed identity clue; no canonical asset merge or identity rewrite performed"
	case IdentityReviewActionReject:
		return "review rejected candidate as identity clue; no canonical asset merge or identity rewrite performed"
	case IdentityReviewActionDefer:
		return "review deferred candidate decision; no canonical asset merge or identity rewrite performed"
	case IdentityReviewActionRequestMoreEvidence:
		return "review requested more evidence for candidate; no discovery execution, canonical asset merge, or identity rewrite performed"
	default:
		return ""
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
