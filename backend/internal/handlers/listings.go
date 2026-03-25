package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
)

type ListingHandler struct {
	Listings models.ListingStore
}

const (
	minListingTitleLen       = 1
	maxListingTitleLen       = 200
	minListingDescriptionLen = 1
	maxListingDescriptionLen = 5000
	maxListingCityLen        = 200
	maxListingStateLen       = 100
	maxImageURLLen           = 2048
	maxListingPrice          = 999999999.99
)

var allowedConditions = map[string]struct{}{
	"new": {}, "like_new": {}, "good": {}, "fair": {}, "poor": {},
}

var allowedStatuses = map[string]struct{}{
	"active": {}, "reserved": {}, "sold": {},
}

func (h ListingHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	cats, err := h.Listings.ListCategories(r.Context())
	if err != nil {
		observability.Error(r.Context(), "listings.categories.failed", map[string]any{"error": err.Error()})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load categories"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"categories": cats})
}

type createListingRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Condition   string   `json:"condition"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	City        string   `json:"city"`
	State       string   `json:"state"`
	CategoryID  *string  `json:"category_id"`
	ImageURLs   []string `json:"image_urls"`
}

func (h ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req createListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	in := models.ListingCreate{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Condition:   strings.TrimSpace(req.Condition),
		Price:       req.Price,
		Currency:    strings.TrimSpace(strings.ToUpper(req.Currency)),
		City:        strings.TrimSpace(req.City),
		State:       strings.TrimSpace(req.State),
		ImageURLs:   req.ImageURLs,
	}
	if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		cid := strings.TrimSpace(*req.CategoryID)
		in.CategoryID = &cid
	}

	if err := validateListingCreate(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if in.CategoryID != nil {
		ok, err := h.Listings.CategoryExists(r.Context(), *in.CategoryID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to validate category"})
			return
		}
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
			return
		}
	}

	for _, u := range in.ImageURLs {
		if err := validateImageURL(strings.TrimSpace(u)); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	listing, err := h.Listings.Create(r.Context(), userID, userID, in)
	if err != nil {
		observability.Error(r.Context(), "listings.create.failed", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create listing"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"listing": listing})
}

func (h ListingHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	page := 1
	limit := 10
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	res, err := h.Listings.ListBySeller(r.Context(), userID, status, page, limit)
	if err != nil {
		observability.Error(r.Context(), "listings.list_mine.failed", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load listings"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"listings":    res.Listings,
		"total":       res.Total,
		"page":        res.Page,
		"limit":       res.Limit,
		"total_pages": res.TotalPages,
	})
}

type updateListingRequest struct {
	Title        *string         `json:"title"`
	Description  *string         `json:"description"`
	Condition    *string         `json:"condition"`
	Price        *float64        `json:"price"`
	Currency     *string         `json:"currency"`
	City         *string         `json:"city"`
	State        *string         `json:"state"`
	CategoryJSON json.RawMessage `json:"category_id"`
}

func (h ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid listing id"})
		return
	}

	var req updateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	patch := models.ListingPatch{
		Title:       req.Title,
		Description: req.Description,
		Condition:   req.Condition,
		Price:       req.Price,
		Currency:    req.Currency,
		City:        req.City,
		State:       req.State,
	}
	if req.CategoryJSON != nil {
		patch.CategoryIDSet = true
		raw := strings.TrimSpace(string(req.CategoryJSON))
		if raw == "null" || raw == "" {
			patch.CategoryID = nil
		} else {
			var s string
			if err := json.Unmarshal(req.CategoryJSON, &s); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
				return
			}
			s = strings.TrimSpace(s)
			patch.CategoryID = &s
		}
	}

	if patch.Title == nil && patch.Description == nil && patch.Condition == nil && patch.Price == nil &&
		patch.Currency == nil && patch.City == nil && patch.State == nil && !patch.CategoryIDSet {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one field is required"})
		return
	}

	if patch.Condition != nil {
		c := strings.TrimSpace(*patch.Condition)
		if _, ok := allowedConditions[c]; !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid condition"})
			return
		}
		*patch.Condition = c
	}

	if err := validateListingPatch(&patch); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if patch.CategoryIDSet && patch.CategoryID != nil && *patch.CategoryID != "" {
		ok, err := h.Listings.CategoryExists(r.Context(), *patch.CategoryID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to validate category"})
			return
		}
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
			return
		}
	}

	listing, err := h.Listings.Update(r.Context(), id, userID, userID, patch)
	if err != nil {
		if errors.Is(err, models.ErrListingNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "listing not found"})
			return
		}
		observability.Error(r.Context(), "listings.update.failed", map[string]any{
			"user_id":    userID,
			"listing_id": id,
			"error":      err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update listing"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"listing": listing})
}

type patchStatusRequest struct {
	Status       string  `json:"status"`
	SoldToUserID *string `json:"sold_to_user_id"`
}

func (h ListingHandler) PatchStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid listing id"})
		return
	}

	var req patchStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	status := strings.TrimSpace(req.Status)
	if _, ok := allowedStatuses[status]; !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
	}

	var soldTo *string
	if req.SoldToUserID != nil {
		s := strings.TrimSpace(*req.SoldToUserID)
		if s != "" {
			if _, err := uuid.Parse(s); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid sold_to_user_id"})
				return
			}
			soldTo = &s
		}
	}

	listing, err := h.Listings.UpdateStatus(r.Context(), id, userID, userID, status, soldTo)
	if err != nil {
		if errors.Is(err, models.ErrListingNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "listing not found"})
			return
		}
		observability.Error(r.Context(), "listings.patch_status.failed", map[string]any{
			"user_id":    userID,
			"listing_id": id,
			"error":      err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update status"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"listing": listing})
}

func (h ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid listing id"})
		return
	}

	if err := h.Listings.SoftDelete(r.Context(), id, userID, userID); err != nil {
		if errors.Is(err, models.ErrListingNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "listing not found"})
			return
		}
		observability.Error(r.Context(), "listings.delete.failed", map[string]any{
			"user_id":    userID,
			"listing_id": id,
			"error":      err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete listing"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func validateListingCreate(in *models.ListingCreate) error {
	if utf8.RuneCountInString(in.Title) < minListingTitleLen {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(in.Title) > maxListingTitleLen {
		return errors.New("title is too long")
	}
	if utf8.RuneCountInString(in.Description) < minListingDescriptionLen {
		return errors.New("description is required")
	}
	if utf8.RuneCountInString(in.Description) > maxListingDescriptionLen {
		return errors.New("description is too long")
	}
	if _, ok := allowedConditions[in.Condition]; !ok {
		return errors.New("invalid condition")
	}
	if in.Price < 0 || in.Price > maxListingPrice {
		return errors.New("invalid price")
	}
	if in.Currency == "" {
		in.Currency = "INR"
	}
	if len(in.Currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}
	if utf8.RuneCountInString(in.City) < 1 {
		return errors.New("city is required")
	}
	if utf8.RuneCountInString(in.City) > maxListingCityLen {
		return errors.New("city is too long")
	}
	if utf8.RuneCountInString(in.State) > maxListingStateLen {
		return errors.New("state is too long")
	}
	if in.CategoryID != nil && *in.CategoryID != "" {
		if _, err := uuid.Parse(*in.CategoryID); err != nil {
			return errors.New("invalid category_id")
		}
	}
	return nil
}

func validateListingPatch(p *models.ListingPatch) error {
	if p.Title != nil {
		t := strings.TrimSpace(*p.Title)
		if utf8.RuneCountInString(t) < minListingTitleLen {
			return errors.New("title is required")
		}
		if utf8.RuneCountInString(t) > maxListingTitleLen {
			return errors.New("title is too long")
		}
		*p.Title = t
	}
	if p.Description != nil {
		d := strings.TrimSpace(*p.Description)
		if utf8.RuneCountInString(d) < minListingDescriptionLen {
			return errors.New("description is required")
		}
		if utf8.RuneCountInString(d) > maxListingDescriptionLen {
			return errors.New("description is too long")
		}
		*p.Description = d
	}
	if p.Price != nil {
		if *p.Price < 0 || *p.Price > maxListingPrice {
			return errors.New("invalid price")
		}
	}
	if p.Currency != nil {
		c := strings.TrimSpace(strings.ToUpper(*p.Currency))
		if len(c) != 3 {
			return errors.New("currency must be a 3-letter code")
		}
		*p.Currency = c
	}
	if p.City != nil {
		c := strings.TrimSpace(*p.City)
		if utf8.RuneCountInString(c) < 1 {
			return errors.New("city is required")
		}
		if utf8.RuneCountInString(c) > maxListingCityLen {
			return errors.New("city is too long")
		}
		*p.City = c
	}
	if p.State != nil {
		s := strings.TrimSpace(*p.State)
		if utf8.RuneCountInString(s) > maxListingStateLen {
			return errors.New("state is too long")
		}
		*p.State = s
	}
	if p.CategoryIDSet && p.CategoryID != nil && *p.CategoryID != "" {
		if _, err := uuid.Parse(*p.CategoryID); err != nil {
			return errors.New("invalid category_id")
		}
	}
	return nil
}

func validateImageURL(s string) error {
	if s == "" {
		return errors.New("image URL cannot be empty")
	}
	if len(s) > maxImageURLLen {
		return errors.New("image URL is too long")
	}
	parsed, err := url.ParseRequestURI(s)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("image URL must be a valid absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("image URL must use http or https")
	}
	return nil
}
