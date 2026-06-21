package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"truthwatcher/internal/assets"
	"truthwatcher/internal/parser"
)

func handleListIdentityCandidates(service *parser.IdentityCandidateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "identity candidate repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		query := r.URL.Query()
		items, err := service.ListIdentityCandidates(r.Context(), parser.IdentityCandidateFilters{
			DiscoveryRunID:       query.Get("discovery_run_id"),
			EvidenceID:           query.Get("evidence_id"),
			ReviewState:          parser.IdentityReviewState(query.Get("review_state")),
			Strength:             assets.IdentityStrength(query.Get("strength")),
			CandidateIdentityKey: query.Get("candidate_identity_key"),
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, identityCandidatesResponse{IdentityCandidates: paged}, metadata)
	}
}

func handleListPendingIdentityCandidates(service *parser.IdentityCandidateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "identity candidate repository is not configured")
			return
		}

		page, err := parsePagination(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		query := r.URL.Query()
		items, err := service.ListIdentityCandidates(r.Context(), parser.IdentityCandidateFilters{
			DiscoveryRunID:       query.Get("discovery_run_id"),
			EvidenceID:           query.Get("evidence_id"),
			ReviewState:          parser.IdentityReviewPending,
			Strength:             assets.IdentityStrength(query.Get("strength")),
			CandidateIdentityKey: query.Get("candidate_identity_key"),
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		paged, metadata := paginate(items, page)

		writeDataWithMetadata(w, http.StatusOK, identityCandidatesResponse{IdentityCandidates: paged}, metadata)
	}
}

func handleIdentityReviewHandoffReport(service *parser.IdentityCandidateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "identity candidate repository is not configured")
			return
		}

		query := r.URL.Query()
		report, err := service.IdentityReviewHandoffReport(r.Context(), parser.IdentityReviewHandoffFilters{
			DiscoveryRunID: query.Get("discovery_run_id"),
			EvidenceID:     query.Get("evidence_id"),
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeData(w, http.StatusOK, identityReviewHandoffResponse{IdentityReviewHandoff: report})
	}
}

func handleReviewIdentityCandidate(service *parser.IdentityCandidateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeError(w, http.StatusServiceUnavailable, "identity candidate repository is not configured")
			return
		}

		var request reviewIdentityCandidateRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "request body is required")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}

		metadata, err := identityReviewMetadata(r, request.Metadata)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		review, err := service.ReviewIdentityCandidate(r.Context(), parser.ReviewIdentityCandidateParams{
			IdentityCandidateID: r.PathValue("id"),
			Reviewer:            request.Reviewer,
			Action:              parser.IdentityReviewAction(request.Action),
			Rationale:           request.Rationale,
			Metadata:            metadata,
		})
		if errors.Is(err, assets.ErrNotFound) {
			writeError(w, http.StatusNotFound, "identity candidate not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeData(w, http.StatusOK, identityCandidateReviewResponse{IdentityCandidateReview: review})
	}
}

func identityReviewMetadata(r *http.Request, input json.RawMessage) (json.RawMessage, error) {
	payload := map[string]any{}
	if len(input) > 0 {
		if !json.Valid(input) {
			return nil, errors.New("metadata must be valid JSON")
		}
		if err := json.Unmarshal(input, &payload); err != nil || payload == nil {
			return nil, errors.New("metadata must be a JSON object")
		}
	}
	payload["request_id"] = r.Header.Get("X-Request-ID")
	payload["path"] = r.URL.Path
	payload["method"] = r.Method
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return data, nil
}
