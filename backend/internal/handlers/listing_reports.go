package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
)

const maxListingReportReasonLen = 500

type listingReportStore interface {
	Create(ctx context.Context, listingID, reporterID, reason string) error
}

type ListingReportHandler struct {
	Reports listingReportStore
}

type createListingReportRequest struct {
	Reason string `json:"reason"`
}

func (h ListingReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	listingID := strings.TrimSpace(r.PathValue("id"))
	if _, err := uuid.Parse(listingID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid listing id"})
		return
	}

	var req createListingReportRequest
	if r.Body != nil && r.ContentLength != 0 {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if utf8.RuneCountInString(req.Reason) > maxListingReportReasonLen {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "reason is too long"})
		return
	}

	if err := h.Reports.Create(r.Context(), listingID, userID, req.Reason); err != nil {
		switch {
		case errors.Is(err, models.ErrListingNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "listing not found"})
			return
		case errors.Is(err, models.ErrListingReportOwnListing):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot report your own listing"})
			return
		default:
			observability.Error(r.Context(), "listing_reports.create.failed", map[string]any{
				"user_id":    userID,
				"listing_id": listingID,
				"error":      err.Error(),
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to report listing"})
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "listing reported"})
}
