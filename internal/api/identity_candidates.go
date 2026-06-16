package api

import (
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

		writeDataWithMetadata(w, http.StatusOK, map[string][]parser.IdentityCandidate{"identity_candidates": paged}, metadata)
	}
}
