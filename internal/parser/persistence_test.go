package parser

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/evidence"
)

func TestParseDiscoveryRunPersistsParsedModel(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	evidenceItems := []evidence.Evidence{
		persistenceFixtureEvidence(t, "evidence-version", runID, PlatformJunos, CommandShowVersion, "junos-mx", "show_version.txt"),
		persistenceFixtureEvidence(t, "evidence-lldp", runID, PlatformJunos, CommandShowLLDPNeighbors, "junos-mx", "show_lldp_neighbors.txt"),
	}
	assetRepo := &persistenceAssetRepository{}
	assetService := assets.NewService(assetRepo)
	parseRepo := &persistenceParseRepository{}
	service := NewPersistenceService(PersistenceOptions{
		Evidence:     persistenceEvidenceRepository{items: evidenceItems},
		Assets:       assetService,
		ParseResults: parseRepo,
		Registry:     BuiltInRegistry(),
	})

	result, err := service.ParseDiscoveryRun(context.Background(), ParseDiscoveryRunParams{
		DiscoveryRunID: runID,
		Platform:       PlatformJunos,
	})
	if err != nil {
		t.Fatalf("ParseDiscoveryRun returned error: %v", err)
	}

	if got, want := result.EvidenceCount, 2; got != want {
		t.Fatalf("evidence count = %d, want %d", got, want)
	}
	if len(result.Assets) == 0 {
		t.Fatal("no assets were persisted")
	}
	if !hasAssetIdentity(assetRepo.assets, "device:hostname:mx-edge-01") {
		t.Fatalf("device asset missing: %#v", assetRepo.assets)
	}
	if !hasAssetIdentity(assetRepo.assets, "device:hostname:spine-01") {
		t.Fatalf("neighbor asset missing: %#v", assetRepo.assets)
	}
	if got, want := len(result.Facts), 4; got != want {
		t.Fatalf("facts persisted = %d, want %d", got, want)
	}
	if result.Facts[0].EvidenceID == nil || *result.Facts[0].EvidenceID != "evidence-version" {
		t.Fatalf("fact evidence link = %#v, want evidence-version", result.Facts[0].EvidenceID)
	}
	if got, want := len(result.Relationships), 2; got != want {
		t.Fatalf("relationships persisted = %d, want %d", got, want)
	}
	if result.Relationships[0].EvidenceID == nil || *result.Relationships[0].EvidenceID != "evidence-lldp" {
		t.Fatalf("relationship evidence link = %#v, want evidence-lldp", result.Relationships[0].EvidenceID)
	}
	if got, want := len(parseRepo.records), 2; got != want {
		t.Fatalf("parse records = %d, want %d", got, want)
	}
	for _, record := range parseRepo.records {
		if record.Status != ParseStatusParsed {
			t.Fatalf("parse record status = %q, want %q", record.Status, ParseStatusParsed)
		}
	}
}

func TestParseDiscoveryRunPersistsBGPPeerModel(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	evidenceItems := []evidence.Evidence{
		persistenceFixtureEvidence(t, "evidence-bgp", runID, PlatformIOSXR, CommandShowBGPSummary, "iosxr-asr", "show_bgp_summary.txt"),
	}
	assetRepo := &persistenceAssetRepository{}
	assetService := assets.NewService(assetRepo)
	service := NewPersistenceService(PersistenceOptions{
		Evidence:     persistenceEvidenceRepository{items: evidenceItems},
		Assets:       assetService,
		ParseResults: &persistenceParseRepository{},
		Registry:     BuiltInRegistry(),
	})

	_, err := service.ParseDiscoveryRun(context.Background(), ParseDiscoveryRunParams{
		DiscoveryRunID: runID,
		Platform:       PlatformIOSXR,
	})
	if err != nil {
		t.Fatalf("ParseDiscoveryRun returned error: %v", err)
	}

	if !hasAssetIdentity(assetRepo.assets, "routing_context:router_id:198.51.100.10") {
		t.Fatalf("routing context asset missing: %#v", assetRepo.assets)
	}
	if !hasAssetIdentity(assetRepo.assets, "bgp_peer:ip:192.0.2.2") {
		t.Fatalf("bgp peer asset missing: %#v", assetRepo.assets)
	}
	if !hasPersistedFact(assetRepo.facts, "bgp_remote_as", `65002`, "evidence-bgp") {
		t.Fatalf("bgp_remote_as fact missing: %#v", assetRepo.facts)
	}
	if !hasPersistedFact(assetRepo.facts, "bgp_accepted_prefixes", `18`, "evidence-bgp") {
		t.Fatalf("bgp_accepted_prefixes fact missing: %#v", assetRepo.facts)
	}
	if !hasRelationshipType(assetRepo.relationships, "bgp_peer_of", "evidence-bgp") {
		t.Fatalf("bgp_peer_of relationship missing: %#v", assetRepo.relationships)
	}
}

func TestParseDiscoveryRunRecordsSkippedWarnings(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	item := evidence.Evidence{
		ID:             "evidence-unsupported",
		DiscoveryRunID: runID,
		Target:         "fixture://junos-mx",
		Method:         "fake",
		CommandOrAPI:   "show interfaces terse",
		RawOutput:      "unsupported",
		RawOutputHash:  evidence.HashRawOutput("unsupported"),
		CollectedAt:    time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Metadata:       json.RawMessage(`{}`),
	}
	assetService := assets.NewService(&persistenceAssetRepository{})
	parseRepo := &persistenceParseRepository{}
	service := NewPersistenceService(PersistenceOptions{
		Evidence:     persistenceEvidenceRepository{items: []evidence.Evidence{item}},
		Assets:       assetService,
		ParseResults: parseRepo,
		Registry:     BuiltInRegistry(),
	})

	result, err := service.ParseDiscoveryRun(context.Background(), ParseDiscoveryRunParams{
		DiscoveryRunID: runID,
		Platform:       PlatformJunos,
	})
	if err != nil {
		t.Fatalf("ParseDiscoveryRun returned error: %v", err)
	}

	if len(result.Assets) != 0 || len(result.Facts) != 0 || len(result.Relationships) != 0 {
		t.Fatalf("persisted model data for unsupported evidence: %#v", result)
	}
	if got, want := len(parseRepo.records), 1; got != want {
		t.Fatalf("parse records = %d, want %d", got, want)
	}
	record := parseRepo.records[0]
	if record.Status != ParseStatusSkipped {
		t.Fatalf("parse status = %q, want %q", record.Status, ParseStatusSkipped)
	}
	if !strings.Contains(string(record.Warnings), "no parser registered") {
		t.Fatalf("warnings = %s, want no parser registered", record.Warnings)
	}
}

func TestParseDiscoveryRunPersistsIdentityCandidates(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	evidenceItems := []evidence.Evidence{
		persistenceFixtureEvidence(t, "evidence-version", runID, PlatformJunos, CommandShowVersion, "junos-mx", "show_version.txt"),
		persistenceFixtureEvidence(t, "evidence-inventory", runID, PlatformJunos, CommandShowChassisHardware, "junos-mx", "show_chassis_hardware.txt"),
	}
	identityRepo := &persistenceIdentityCandidateRepository{}
	service := NewPersistenceService(PersistenceOptions{
		Evidence:           persistenceEvidenceRepository{items: evidenceItems},
		Assets:             assets.NewService(&persistenceAssetRepository{}),
		ParseResults:       &persistenceParseRepository{},
		IdentityCandidates: identityRepo,
		Registry:           BuiltInRegistry(),
	})

	result, err := service.ParseDiscoveryRun(context.Background(), ParseDiscoveryRunParams{
		DiscoveryRunID: runID,
		Platform:       PlatformJunos,
	})
	if err != nil {
		t.Fatalf("ParseDiscoveryRun returned error: %v", err)
	}

	hostname := findIdentityCandidate(t, identityRepo.items, "device:hostname:mx-edge-01")
	if hostname.EvidenceID != "evidence-version" {
		t.Fatalf("hostname evidence id = %q, want evidence-version", hostname.EvidenceID)
	}
	if hostname.Strength != assets.IdentityStrengthProvisional {
		t.Fatalf("hostname strength = %q, want provisional", hostname.Strength)
	}
	if hostname.ReviewState != IdentityReviewPending {
		t.Fatalf("hostname review state = %q, want pending", hostname.ReviewState)
	}
	if !metadataContains(t, hostname.Metadata, "identity_review_rule", "queue_non_strong_candidate") {
		t.Fatalf("hostname metadata = %s, want queue rule", hostname.Metadata)
	}
	if hostname.Hostname == nil || *hostname.Hostname != "mx-edge-01" {
		t.Fatalf("hostname attribute = %#v, want mx-edge-01", hostname.Hostname)
	}

	strong := findIdentityCandidate(t, identityRepo.items, "chassis:vendor_serial:juniper:jn1234abcdef")
	if strong.EvidenceID != "evidence-inventory" {
		t.Fatalf("strong evidence id = %q, want evidence-inventory", strong.EvidenceID)
	}
	if strong.Strength != assets.IdentityStrengthStrong {
		t.Fatalf("strong strength = %q, want strong", strong.Strength)
	}
	if strong.ReviewState != IdentityReviewAutoAccepted {
		t.Fatalf("strong review state = %q, want auto_accepted", strong.ReviewState)
	}
	if strong.Serial == nil || *strong.Serial != "JN1234ABCDEF" {
		t.Fatalf("strong serial = %#v, want JN1234ABCDEF", strong.Serial)
	}
	if !metadataContains(t, strong.Metadata, "identity_review_rule", "auto_accept_strong_no_plausible_conflict") {
		t.Fatalf("strong metadata = %s, want auto acceptance rule", strong.Metadata)
	}
	review, ok := findIdentityCandidateReview(identityRepo.reviews, strong.ID)
	if !ok {
		t.Fatalf("review audit for %q not found in %#v", strong.ID, identityRepo.reviews)
	}
	if review.Action != IdentityReviewActionAutoAccept {
		t.Fatalf("review action = %q, want auto_accept", review.Action)
	}
	if review.ResultingReviewState != IdentityReviewAutoAccepted {
		t.Fatalf("review state = %q, want auto_accepted", review.ResultingReviewState)
	}
	if !strings.Contains(review.Effect, "no canonical asset merge or identity rewrite") {
		t.Fatalf("review effect = %q, want non-destructive effect", review.Effect)
	}
	if len(result.IdentityCandidates) != len(identityRepo.items) {
		t.Fatalf("result candidates = %d, repo candidates = %d", len(result.IdentityCandidates), len(identityRepo.items))
	}
}

func TestParseDiscoveryRunQueuesConflictingStrongIdentityCandidate(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	item := persistenceFixtureEvidence(t, "evidence-inventory", runID, PlatformJunos, CommandShowChassisHardware, "junos-mx", "show_chassis_hardware.txt")
	assetRepo := &persistenceAssetRepository{
		assets: []assets.Asset{{
			ID:          "asset-existing",
			Type:        "device",
			IdentityKey: "device:hostname:mx-edge-01",
			Serial:      stringPtr("JN1234ABCDEF"),
			Metadata:    json.RawMessage(`{"identity_strength":"provisional"}`),
		}},
	}
	identityRepo := &persistenceIdentityCandidateRepository{}
	service := NewPersistenceService(PersistenceOptions{
		Evidence:           persistenceEvidenceRepository{items: []evidence.Evidence{item}},
		Assets:             assets.NewService(assetRepo),
		ParseResults:       &persistenceParseRepository{},
		IdentityCandidates: identityRepo,
		Registry:           BuiltInRegistry(),
	})

	if _, err := service.ParseDiscoveryRun(context.Background(), ParseDiscoveryRunParams{
		DiscoveryRunID: runID,
		Platform:       PlatformJunos,
	}); err != nil {
		t.Fatalf("ParseDiscoveryRun returned error: %v", err)
	}

	strong := findIdentityCandidate(t, identityRepo.items, "chassis:vendor_serial:juniper:jn1234abcdef")
	if strong.ReviewState != IdentityReviewPending {
		t.Fatalf("strong conflicting review state = %q, want pending", strong.ReviewState)
	}
	if !metadataContains(t, strong.Metadata, "identity_review_rule", "queue_plausible_canonical_asset_conflict") {
		t.Fatalf("strong metadata = %s, want conflict queue rule", strong.Metadata)
	}
	if !strings.Contains(string(strong.Metadata), "different canonical asset") {
		t.Fatalf("strong metadata = %s, want conflict explanation", strong.Metadata)
	}
	if _, ok := findIdentityCandidateReview(identityRepo.reviews, strong.ID); ok {
		t.Fatalf("queued conflicting candidate has auto review audit: %#v", identityRepo.reviews)
	}
	if assetRepo.assets[0].IdentityKey != "device:hostname:mx-edge-01" {
		t.Fatalf("existing identity key = %q, want unchanged hostname identity", assetRepo.assets[0].IdentityKey)
	}
}

func TestParseDiscoveryRunDeduplicatesIdentityCandidates(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	item := persistenceFixtureEvidence(t, "evidence-version", runID, PlatformJunos, CommandShowVersion, "junos-mx", "show_version.txt")
	identityRepo := &persistenceIdentityCandidateRepository{}
	service := NewPersistenceService(PersistenceOptions{
		Evidence:           persistenceEvidenceRepository{items: []evidence.Evidence{item}},
		Assets:             assets.NewService(&persistenceAssetRepository{}),
		ParseResults:       &persistenceParseRepository{},
		IdentityCandidates: identityRepo,
		Registry:           BuiltInRegistry(),
	})

	for i := 0; i < 2; i++ {
		if _, err := service.ParseDiscoveryRun(context.Background(), ParseDiscoveryRunParams{
			DiscoveryRunID: runID,
			Platform:       PlatformJunos,
		}); err != nil {
			t.Fatalf("ParseDiscoveryRun run %d returned error: %v", i+1, err)
		}
	}

	if got, want := countIdentityCandidate(identityRepo.items, "evidence-version", "junos_show_version", "device:hostname:mx-edge-01"), 1; got != want {
		t.Fatalf("candidate count = %d, want %d", got, want)
	}
}

func TestParseDiscoveryRunDoesNotRewriteExistingCanonicalAssetIdentity(t *testing.T) {
	runID := "11111111-1111-4111-8111-111111111111"
	item := persistenceFixtureEvidence(t, "evidence-version", runID, PlatformJunos, CommandShowVersion, "junos-mx", "show_version.txt")
	assetRepo := &persistenceAssetRepository{
		assets: []assets.Asset{{
			ID:          "asset-existing",
			Type:        "device",
			IdentityKey: "device:hostname:mx-edge-01",
			Vendor:      stringPtr("existing-vendor"),
			Serial:      stringPtr("existing-serial"),
			Metadata:    json.RawMessage(`{"identity_strength":"provisional"}`),
		}},
	}
	service := NewPersistenceService(PersistenceOptions{
		Evidence:           persistenceEvidenceRepository{items: []evidence.Evidence{item}},
		Assets:             assets.NewService(assetRepo),
		ParseResults:       &persistenceParseRepository{},
		IdentityCandidates: &persistenceIdentityCandidateRepository{},
		Registry:           BuiltInRegistry(),
	})

	if _, err := service.ParseDiscoveryRun(context.Background(), ParseDiscoveryRunParams{
		DiscoveryRunID: runID,
		Platform:       PlatformJunos,
	}); err != nil {
		t.Fatalf("ParseDiscoveryRun returned error: %v", err)
	}

	if got, want := len(assetRepo.assets), 1; got != want {
		t.Fatalf("asset count = %d, want %d", got, want)
	}
	existing := assetRepo.assets[0]
	if existing.IdentityKey != "device:hostname:mx-edge-01" {
		t.Fatalf("identity key = %q, want unchanged hostname identity", existing.IdentityKey)
	}
	if existing.Vendor == nil || *existing.Vendor != "existing-vendor" {
		t.Fatalf("vendor = %#v, want existing-vendor", existing.Vendor)
	}
	if existing.Serial == nil || *existing.Serial != "existing-serial" {
		t.Fatalf("serial = %#v, want existing-serial", existing.Serial)
	}
}

type persistenceEvidenceRepository struct {
	items []evidence.Evidence
}

func (r persistenceEvidenceRepository) ListEvidenceByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]evidence.Evidence, error) {
	var result []evidence.Evidence
	for _, item := range r.items {
		if item.DiscoveryRunID == discoveryRunID {
			result = append(result, item)
		}
	}
	return result, nil
}

type persistenceParseRepository struct {
	records []ParseRecord
}

func (r *persistenceParseRepository) CreateParseResult(ctx context.Context, params CreateParseResultParams) (ParseRecord, error) {
	warnings, err := json.Marshal(params.Warnings)
	if err != nil {
		return ParseRecord{}, err
	}
	record := ParseRecord{
		ID:             "parse-record-" + params.EvidenceID,
		DiscoveryRunID: params.DiscoveryRunID,
		EvidenceID:     params.EvidenceID,
		ParserName:     params.ParserName,
		Status:         params.Status,
		Warnings:       warnings,
		ErrorMessage:   params.ErrorMessage,
		CreatedAt:      time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	r.records = append(r.records, record)
	return record, nil
}

type persistenceIdentityCandidateRepository struct {
	items   []IdentityCandidate
	reviews []IdentityCandidateReview
}

func (r *persistenceIdentityCandidateRepository) CreateIdentityCandidate(ctx context.Context, params CreateIdentityCandidateParams) (IdentityCandidate, error) {
	for _, item := range r.items {
		if item.EvidenceID == params.EvidenceID &&
			item.ParserName == params.ParserName &&
			item.CandidateIdentityKey == params.CandidateIdentityKey {
			return item, nil
		}
	}
	item := IdentityCandidate{
		ID:                   "identity-candidate-" + string(rune('a'+len(r.items))),
		DiscoveryRunID:       params.DiscoveryRunID,
		EvidenceID:           params.EvidenceID,
		ParserName:           params.ParserName,
		AssetType:            params.AssetType,
		CandidateIdentityKey: params.CandidateIdentityKey,
		Strength:             params.Strength,
		Confidence:           params.Confidence,
		Reason:               params.Reason,
		Vendor:               params.Vendor,
		Model:                params.Model,
		Serial:               params.Serial,
		SystemMAC:            params.SystemMAC,
		Hostname:             params.Hostname,
		ProposedAssetID:      params.ProposedAssetID,
		ReviewState:          params.ReviewState,
		Metadata:             params.Metadata,
		CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	r.items = append(r.items, item)
	return item, nil
}

func (r *persistenceIdentityCandidateRepository) GetIdentityCandidate(ctx context.Context, id string) (IdentityCandidate, error) {
	for _, item := range r.items {
		if item.ID == id {
			return item, nil
		}
	}
	return IdentityCandidate{}, assets.ErrNotFound
}

func (r *persistenceIdentityCandidateRepository) ListIdentityCandidates(ctx context.Context, filters IdentityCandidateFilters) ([]IdentityCandidate, error) {
	var result []IdentityCandidate
	for _, item := range r.items {
		if filters.DiscoveryRunID != "" && item.DiscoveryRunID != filters.DiscoveryRunID {
			continue
		}
		if filters.EvidenceID != "" && item.EvidenceID != filters.EvidenceID {
			continue
		}
		if filters.ReviewState != "" && item.ReviewState != filters.ReviewState {
			continue
		}
		if filters.Strength != "" && item.Strength != filters.Strength {
			continue
		}
		if filters.CandidateIdentityKey != "" && item.CandidateIdentityKey != filters.CandidateIdentityKey {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *persistenceIdentityCandidateRepository) ReviewIdentityCandidate(ctx context.Context, params ReviewIdentityCandidateParams) (IdentityCandidateReview, error) {
	for i := range r.items {
		if r.items[i].ID != params.IdentityCandidateID {
			continue
		}
		previous := r.items[i].ReviewState
		resulting := ResultingReviewState(params.Action)
		r.items[i].ReviewState = resulting
		review := IdentityCandidateReview{
			ID:                   "identity-review-" + params.IdentityCandidateID,
			IdentityCandidateID:  params.IdentityCandidateID,
			DiscoveryRunID:       r.items[i].DiscoveryRunID,
			EvidenceID:           r.items[i].EvidenceID,
			Reviewer:             params.Reviewer,
			Action:               params.Action,
			PreviousReviewState:  previous,
			ResultingReviewState: resulting,
			Rationale:            params.Rationale,
			Effect:               IdentityReviewEffect(params.Action),
			Metadata:             params.Metadata,
			CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		}
		r.reviews = append(r.reviews, review)
		return review, nil
	}
	return IdentityCandidateReview{}, assets.ErrNotFound
}

func (r *persistenceIdentityCandidateRepository) AutoAcceptIdentityCandidate(ctx context.Context, params AutoAcceptIdentityCandidateParams) error {
	for i := range r.items {
		if r.items[i].ID != params.IdentityCandidateID {
			continue
		}
		if r.items[i].ReviewState != IdentityReviewPending {
			return nil
		}
		previous := r.items[i].ReviewState
		r.items[i].ReviewState = IdentityReviewAutoAccepted
		r.reviews = append(r.reviews, IdentityCandidateReview{
			ID:                   "identity-review-" + params.IdentityCandidateID,
			IdentityCandidateID:  params.IdentityCandidateID,
			DiscoveryRunID:       r.items[i].DiscoveryRunID,
			EvidenceID:           r.items[i].EvidenceID,
			Reviewer:             "parser:auto_acceptance",
			Action:               IdentityReviewActionAutoAccept,
			PreviousReviewState:  previous,
			ResultingReviewState: IdentityReviewAutoAccepted,
			Rationale:            params.Rationale,
			Effect:               IdentityReviewEffect(IdentityReviewActionAutoAccept),
			Metadata:             params.Metadata,
			CreatedAt:            time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		})
		return nil
	}
	return assets.ErrNotFound
}

func (r *persistenceIdentityCandidateRepository) ListIdentityReviewHandoffEntries(ctx context.Context, filters IdentityReviewHandoffFilters) ([]IdentityReviewHandoffEntry, error) {
	var result []IdentityReviewHandoffEntry
	for _, item := range r.items {
		if filters.DiscoveryRunID != "" && item.DiscoveryRunID != filters.DiscoveryRunID {
			continue
		}
		if filters.EvidenceID != "" && item.EvidenceID != filters.EvidenceID {
			continue
		}
		entry := IdentityReviewHandoffEntry{
			Candidate: item,
			EvidenceReference: IdentityEvidenceRef{
				EvidenceID:     item.EvidenceID,
				DiscoveryRunID: item.DiscoveryRunID,
				Present:        true,
			},
			ParserSource: IdentityParserSource{
				ParserName: item.ParserName,
				Metadata:   item.Metadata,
			},
		}
		if review, ok := r.latestReviewForCandidate(item.ID); ok {
			entry.LatestReview = &review
		}
		result = append(result, entry)
	}
	return result, nil
}

func (r *persistenceIdentityCandidateRepository) ListOrphanedIdentityCandidateReviews(ctx context.Context) ([]IdentityCandidateReview, error) {
	return nil, nil
}

func (r *persistenceIdentityCandidateRepository) latestReviewForCandidate(candidateID string) (IdentityCandidateReview, bool) {
	var latest IdentityCandidateReview
	found := false
	for _, review := range r.reviews {
		if review.IdentityCandidateID != candidateID {
			continue
		}
		if !found || review.CreatedAt.After(latest.CreatedAt) || (review.CreatedAt.Equal(latest.CreatedAt) && review.ID > latest.ID) {
			latest = review
			found = true
		}
	}
	return latest, found
}

func (r *persistenceParseRepository) ListParseResultsByDiscoveryRun(ctx context.Context, discoveryRunID string) ([]ParseRecord, error) {
	var result []ParseRecord
	for _, item := range r.records {
		if item.DiscoveryRunID == discoveryRunID {
			result = append(result, item)
		}
	}
	return result, nil
}

type persistenceAssetRepository struct {
	assets        []assets.Asset
	facts         []assets.Fact
	relationships []assets.Relationship
}

func (r *persistenceAssetRepository) CreateAsset(ctx context.Context, params assets.CreateAssetParams) (assets.Asset, error) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	item := assets.Asset{
		ID:               "asset-" + string(rune('a'+len(r.assets))),
		Type:             params.Type,
		IdentityKey:      params.IdentityKey,
		Vendor:           params.Vendor,
		Model:            params.Model,
		Serial:           params.Serial,
		SystemMAC:        params.SystemMAC,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		Metadata:         params.Metadata,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	r.assets = append(r.assets, item)
	return item, nil
}

func (r *persistenceAssetRepository) GetAsset(ctx context.Context, id string) (assets.Asset, error) {
	for _, item := range r.assets {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Asset{}, assets.ErrNotFound
}

func (r *persistenceAssetRepository) ListAssets(ctx context.Context) ([]assets.Asset, error) {
	return r.assets, nil
}

func (r *persistenceAssetRepository) CreateFact(ctx context.Context, params assets.CreateFactParams) (assets.Fact, error) {
	item := assets.Fact{
		ID:               "fact-" + string(rune('a'+len(r.facts))),
		AssetID:          params.AssetID,
		Name:             params.Name,
		Value:            params.Value,
		Source:           params.Source,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		EvidenceID:       params.EvidenceID,
		CreatedAt:        time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	r.facts = append(r.facts, item)
	return item, nil
}

func (r *persistenceAssetRepository) GetFact(ctx context.Context, id string) (assets.Fact, error) {
	for _, item := range r.facts {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Fact{}, assets.ErrNotFound
}

func (r *persistenceAssetRepository) ListFactsByAsset(ctx context.Context, assetID string) ([]assets.Fact, error) {
	var result []assets.Fact
	for _, item := range r.facts {
		if item.AssetID == assetID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *persistenceAssetRepository) CreateRelationship(ctx context.Context, params assets.CreateRelationshipParams) (assets.Relationship, error) {
	item := assets.Relationship{
		ID:               "relationship-" + string(rune('a'+len(r.relationships))),
		SourceAssetID:    params.SourceAssetID,
		TargetAssetID:    params.TargetAssetID,
		RelationshipType: params.RelationshipType,
		Confidence:       params.Confidence,
		ConfidenceReason: params.ConfidenceReason,
		State:            params.State,
		EvidenceID:       params.EvidenceID,
		Metadata:         params.Metadata,
		CreatedAt:        time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	r.relationships = append(r.relationships, item)
	return item, nil
}

func (r *persistenceAssetRepository) GetRelationship(ctx context.Context, id string) (assets.Relationship, error) {
	for _, item := range r.relationships {
		if item.ID == id {
			return item, nil
		}
	}
	return assets.Relationship{}, assets.ErrNotFound
}

func (r *persistenceAssetRepository) ListRelationships(ctx context.Context) ([]assets.Relationship, error) {
	return r.relationships, nil
}

func persistenceFixtureEvidence(t *testing.T, id string, runID string, platform string, command string, fixtureDir string, filename string) evidence.Evidence {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "fixtures", fixtureDir, filename))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return evidence.Evidence{
		ID:             id,
		DiscoveryRunID: runID,
		Target:         "fixture://" + platform,
		Method:         "fake",
		CommandOrAPI:   command,
		RawOutput:      string(raw),
		RawOutputHash:  evidence.HashRawOutput(string(raw)),
		CollectedAt:    time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Metadata:       json.RawMessage(`{}`),
	}
}

func hasAssetIdentity(items []assets.Asset, identityKey string) bool {
	for _, item := range items {
		if item.IdentityKey == identityKey {
			return true
		}
	}
	return false
}

func hasPersistedFact(items []assets.Fact, name string, value string, evidenceID string) bool {
	for _, item := range items {
		if item.Name != name || string(item.Value) != value {
			continue
		}
		if item.EvidenceID != nil && *item.EvidenceID == evidenceID {
			return true
		}
	}
	return false
}

func hasRelationshipType(items []assets.Relationship, relationshipType string, evidenceID string) bool {
	for _, item := range items {
		if item.RelationshipType != relationshipType {
			continue
		}
		if item.EvidenceID != nil && *item.EvidenceID == evidenceID {
			return true
		}
	}
	return false
}

func findIdentityCandidate(t *testing.T, items []IdentityCandidate, identityKey string) IdentityCandidate {
	t.Helper()
	for _, item := range items {
		if item.CandidateIdentityKey == identityKey {
			return item
		}
	}
	t.Fatalf("identity candidate %q not found in %#v", identityKey, items)
	return IdentityCandidate{}
}

func countIdentityCandidate(items []IdentityCandidate, evidenceID string, parserName string, identityKey string) int {
	count := 0
	for _, item := range items {
		if item.EvidenceID == evidenceID && item.ParserName == parserName && item.CandidateIdentityKey == identityKey {
			count++
		}
	}
	return count
}

func findIdentityCandidateReview(items []IdentityCandidateReview, candidateID string) (IdentityCandidateReview, bool) {
	for _, item := range items {
		if item.IdentityCandidateID == candidateID {
			return item, true
		}
	}
	return IdentityCandidateReview{}, false
}

func metadataContains(t *testing.T, metadata json.RawMessage, key string, want string) bool {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(metadata, &payload); err != nil {
		t.Fatalf("decode metadata %s: %v", metadata, err)
	}
	got, ok := payload[key].(string)
	return ok && got == want
}

func stringPtr(value string) *string {
	return &value
}
