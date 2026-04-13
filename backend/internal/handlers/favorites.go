package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
)

type FavoriteHandler struct {
	Favorites models.FavoriteStore
}

func (h FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	listingID := r.PathValue("listing_id")
	if _, err := uuid.Parse(listingID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid listing id"})
		return
	}

	if err := h.Favorites.Add(r.Context(), userID, listingID); err != nil {
		if errors.Is(err, models.ErrListingNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "listing not found"})
			return
		}
		observability.Error(r.Context(), "favorites.add.failed", map[string]any{
			"user_id":    userID,
			"listing_id": listingID,
			"error":      err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add favorite"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "favorited"})
}

func (h FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	listingID := r.PathValue("listing_id")
	if _, err := uuid.Parse(listingID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid listing id"})
		return
	}

	if err := h.Favorites.Remove(r.Context(), userID, listingID); err != nil {
		observability.Error(r.Context(), "favorites.remove.failed", map[string]any{
			"user_id":    userID,
			"listing_id": listingID,
			"error":      err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove favorite"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "removed"})
}

func (h FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	page := 1
	limit := 10
	if raw := r.URL.Query().Get("page"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "page must be a positive integer"})
			return
		}
		page = n
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be a positive integer"})
			return
		}
		limit = n
	}

	res, err := h.Favorites.ListByUser(r.Context(), userID, page, limit)
	if err != nil {
		observability.Error(r.Context(), "favorites.list.failed", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list favorites"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"favorites":   res.Favorites,
		"total":       res.Total,
		"page":        res.Page,
		"limit":       res.Limit,
		"total_pages": res.TotalPages,
	})
}
