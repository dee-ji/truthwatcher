package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/evidence"
)

const (
	defaultPageLimit = 100
	maxPageLimit     = 500
)

type pagination struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Count   int  `json:"count"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

type assetFilters struct {
	Type        string
	Vendor      string
	Serial      string
	IdentityKey string
	Query       string
}

type assetHistoryEvent struct {
	EventType      string                 `json:"event_type"`
	AssetID        string                 `json:"asset_id"`
	RecordID       string                 `json:"record_id"`
	Name           string                 `json:"name"`
	State          assets.ConfidenceState `json:"state"`
	Confidence     float64                `json:"confidence"`
	EvidenceID     *string                `json:"evidence_id,omitempty"`
	RelationshipTo *string                `json:"relationship_to,omitempty"`
	OccurredAt     string                 `json:"occurred_at"`
	Details        any                    `json:"details,omitempty"`
}

func handleListAssets(service *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		filters := parseAssetFilters(r)

		items, err := service.ListAssets(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = filterAssets(items, filters)
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, map[string][]assets.Asset{"assets": paged}, metadata)
	}
}

func handleGetAsset(service *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		item, err := service.GetAsset(r.Context(), r.PathValue("id"))
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]assets.Asset{"asset": item})
	}
}

func handleGetAssetHistory(service *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		assetID := r.PathValue("id")
		item, err := service.GetAsset(r.Context(), assetID)
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		facts, err := service.ListFactsByAsset(r.Context(), assetID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		relationships, err := service.ListRelationships(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		events := assetHistoryEvents(item, facts, filterRelationshipsForAsset(relationships, assetID))
		writeData(w, http.StatusOK, map[string]any{
			"asset":   item,
			"history": events,
		})
	}
}

func handleListProvisionalIdentityAssets(service *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		items, err := service.ListProvisionalIdentityAssets(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, map[string][]assets.Asset{"assets": paged}, metadata)
	}
}

func handleListAssetFacts(service *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		assetID := r.PathValue("id")
		if _, err := service.GetAsset(r.Context(), assetID); errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		items, err := service.ListFactsByAsset(r.Context(), assetID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, map[string][]assets.Fact{"facts": paged}, metadata)
	}
}

func handleListAssetRelationships(service *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		assetID := r.PathValue("id")
		if _, err := service.GetAsset(r.Context(), assetID); errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		relationships, err := service.ListRelationships(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items := filterRelationshipsForAsset(relationships, assetID)
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, map[string][]assets.Relationship{"relationships": paged}, metadata)
	}
}

func handleListConflictingFacts(service *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		items, err := service.ListConflictingFacts(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, map[string][]assets.Fact{"facts": paged}, metadata)
	}
}

func handleListAssetEvidence(assetService *assets.Service, evidenceService *evidence.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if assetService == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}
		if evidenceService == nil {
			writeError(w, http.StatusServiceUnavailable, "evidence repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		assetID := r.PathValue("id")
		if _, err := assetService.GetAsset(r.Context(), assetID); errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "asset not found")
			return
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		items, err := evidenceForAsset(r, assetService, evidenceService, assetID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, map[string][]evidence.Evidence{"evidence": paged}, metadata)
	}
}

func handleListFactEvidence(assetService *assets.Service, evidenceService *evidence.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if assetService == nil {
			writeError(w, http.StatusServiceUnavailable, "asset repository is not configured")
			return
		}
		if evidenceService == nil {
			writeError(w, http.StatusServiceUnavailable, "evidence repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		fact, err := assetService.GetFact(r.Context(), r.PathValue("id"))
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "fact not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		items := []evidence.Evidence{}
		if fact.EvidenceID != nil {
			item, err := evidenceService.GetEvidence(r.Context(), *fact.EvidenceID)
			if errors.Is(err, evidence.ErrNotFound) {
				writeError(w, http.StatusNotFound, "evidence not found")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			items = append(items, item)
		}
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, map[string][]evidence.Evidence{"evidence": paged}, metadata)
	}
}

func parsePagination(r *http.Request) (pagination, error) {
	limit, err := parseNonNegativeInt(r.URL.Query().Get("limit"), defaultPageLimit, "limit")
	if err != nil {
		return pagination{}, err
	}
	if limit == 0 {
		return pagination{}, fmt.Errorf("limit must be greater than 0")
	}
	if limit > maxPageLimit {
		return pagination{}, fmt.Errorf("limit must be less than or equal to %d", maxPageLimit)
	}

	offset, err := parseNonNegativeInt(r.URL.Query().Get("offset"), 0, "offset")
	if err != nil {
		return pagination{}, err
	}

	return pagination{Limit: limit, Offset: offset}, nil
}

func parseNonNegativeInt(raw string, defaultValue int, field string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", field)
	}
	return value, nil
}

func parseAssetFilters(r *http.Request) assetFilters {
	query := r.URL.Query()
	return assetFilters{
		Type:        strings.TrimSpace(query.Get("type")),
		Vendor:      strings.TrimSpace(query.Get("vendor")),
		Serial:      strings.TrimSpace(query.Get("serial")),
		IdentityKey: strings.TrimSpace(query.Get("identity_key")),
		Query:       strings.TrimSpace(query.Get("q")),
	}
}

func filterAssets(items []assets.Asset, filters assetFilters) []assets.Asset {
	if filters == (assetFilters{}) {
		return items
	}

	filtered := make([]assets.Asset, 0, len(items))
	for _, item := range items {
		if filters.Type != "" && !strings.EqualFold(item.Type, filters.Type) {
			continue
		}
		if filters.Vendor != "" && !stringPtrEqualFold(item.Vendor, filters.Vendor) {
			continue
		}
		if filters.Serial != "" && !stringPtrEqualFold(item.Serial, filters.Serial) {
			continue
		}
		if filters.IdentityKey != "" && !strings.EqualFold(item.IdentityKey, filters.IdentityKey) {
			continue
		}
		if filters.Query != "" && !assetMatchesQuery(item, filters.Query) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func assetMatchesQuery(item assets.Asset, query string) bool {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return true
	}
	values := []string{item.ID, item.Type, item.IdentityKey}
	for _, value := range []*string{item.Vendor, item.Model, item.Serial, item.SystemMAC} {
		if value != nil {
			values = append(values, *value)
		}
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(strings.TrimSpace(value)), needle) {
			return true
		}
	}
	return false
}

func assetHistoryEvents(item assets.Asset, facts []assets.Fact, relationships []assets.Relationship) []assetHistoryEvent {
	events := []assetHistoryEvent{{
		EventType:  "asset_created",
		AssetID:    item.ID,
		RecordID:   item.ID,
		Name:       item.IdentityKey,
		State:      item.State,
		Confidence: item.Confidence,
		OccurredAt: item.CreatedAt.Format(time.RFC3339),
		Details: map[string]string{
			"asset_type": item.Type,
		},
	}}
	for _, fact := range facts {
		events = append(events, assetHistoryEvent{
			EventType:  "fact_observed",
			AssetID:    item.ID,
			RecordID:   fact.ID,
			Name:       fact.Name,
			State:      fact.State,
			Confidence: fact.Confidence,
			EvidenceID: fact.EvidenceID,
			OccurredAt: fact.CreatedAt.Format(time.RFC3339),
			Details: map[string]json.RawMessage{
				"value": fact.Value,
			},
		})
	}
	for _, relationship := range relationships {
		other := relationship.TargetAssetID
		if relationship.TargetAssetID == item.ID {
			other = relationship.SourceAssetID
		}
		events = append(events, assetHistoryEvent{
			EventType:      "relationship_observed",
			AssetID:        item.ID,
			RecordID:       relationship.ID,
			Name:           relationship.RelationshipType,
			State:          relationship.State,
			Confidence:     relationship.Confidence,
			EvidenceID:     relationship.EvidenceID,
			RelationshipTo: &other,
			OccurredAt:     relationship.CreatedAt.Format(time.RFC3339),
			Details: map[string]string{
				"source_asset_id": relationship.SourceAssetID,
				"target_asset_id": relationship.TargetAssetID,
			},
		})
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].OccurredAt > events[j].OccurredAt
	})
	return events
}

func stringPtrEqualFold(value *string, want string) bool {
	return value != nil && strings.EqualFold(strings.TrimSpace(*value), want)
}

func filterRelationshipsForAsset(items []assets.Relationship, assetID string) []assets.Relationship {
	filtered := make([]assets.Relationship, 0, len(items))
	for _, item := range items {
		if item.SourceAssetID == assetID || item.TargetAssetID == assetID {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func evidenceForAsset(r *http.Request, assetService *assets.Service, evidenceService *evidence.Service, assetID string) ([]evidence.Evidence, error) {
	ids := map[string]struct{}{}

	facts, err := assetService.ListFactsByAsset(r.Context(), assetID)
	if err != nil {
		return nil, err
	}
	for _, fact := range facts {
		if fact.EvidenceID != nil {
			ids[*fact.EvidenceID] = struct{}{}
		}
	}

	relationships, err := assetService.ListRelationships(r.Context())
	if err != nil {
		return nil, err
	}
	for _, relationship := range filterRelationshipsForAsset(relationships, assetID) {
		if relationship.EvidenceID != nil {
			ids[*relationship.EvidenceID] = struct{}{}
		}
	}

	orderedIDs := make([]string, 0, len(ids))
	for id := range ids {
		orderedIDs = append(orderedIDs, id)
	}
	sort.Strings(orderedIDs)

	items := make([]evidence.Evidence, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		item, err := evidenceService.GetEvidence(r.Context(), id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func paginate[T any](items []T, page pagination) ([]T, map[string]any) {
	total := len(items)
	if page.Offset >= total {
		page.Count = 0
		page.Total = total
		page.HasNext = false
		return []T{}, paginationMetadata(page)
	}

	end := page.Offset + page.Limit
	if end > total {
		end = total
	}
	paged := items[page.Offset:end]
	page.Count = len(paged)
	page.Total = total
	page.HasNext = end < total

	return paged, paginationMetadata(page)
}

func paginationMetadata(page pagination) map[string]any {
	return map[string]any{
		"pagination": page,
	}
}
