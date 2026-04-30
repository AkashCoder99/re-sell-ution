package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
)

const (
	maxListingReportReasonLen = 500
	maxResolutionNoteLen      = 1000
	maxModerationPayloadLen   = 4000
)

var allowedReportStatuses = map[string]struct{}{
	"open": {}, "in_review": {}, "resolved": {}, "rejected": {},
}

var allowedModerationActions = map[string]struct{}{
	"assign": {}, "note": {}, "hide_listing": {}, "warn_user": {}, "ban_user": {}, "close_report": {},
}

type listingReportStore interface {
	CreateListingReport(ctx context.Context, listingID, reporterID, reason string) (models.Report, error)
	ListReports(ctx context.Context, status string, page, limit int) (models.ReportsPage, error)
	GetReport(ctx context.Context, reportID string) (models.Report, error)
	UpdateReportStatus(ctx context.Context, reportID, adminID, status, resolutionNote string) (models.Report, error)
	AddAction(ctx context.Context, reportID, adminID, actionType string, payload map[string]any) (models.ModerationAction, error)
}

type adminChecker interface {
	IsAdminByID(ctx context.Context, userID string) (bool, error)
}

type ListingReportHandler struct {
	Reports listingReportStore
	Admins  adminChecker
}

type createListingReportRequest struct {
	Reason string `json:"reason"`
}

type updateReportStatusRequest struct {
	Status         string `json:"status"`
	ResolutionNote string `json:"resolution_note"`
}

type createModerationActionRequest struct {
	ActionType string         `json:"action_type"`
	Payload    map[string]any `json:"payload"`
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

	report, err := h.Reports.CreateListingReport(r.Context(), listingID, userID, req.Reason)
	if err != nil {
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

	writeJSON(w, http.StatusCreated, map[string]any{"message": "listing reported", "report": report})
}

func (h ListingReportHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	adminID, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	_ = adminID

	status, err := parseReportStatusFilter(r.URL.Query().Get("status"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	page, err := parsePositiveIntParam(r.URL.Query().Get("page"), 1)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "page must be a positive integer"})
		return
	}
	limit, err := parsePositiveIntParam(r.URL.Query().Get("limit"), 20)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be a positive integer"})
		return
	}

	res, err := h.Reports.ListReports(r.Context(), status, page, limit)
	if err != nil {
		observability.Error(r.Context(), "listing_reports.admin_list.failed", map[string]any{"error": err.Error()})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load reports"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"reports":     res.Reports,
		"total":       res.Total,
		"page":        res.Page,
		"limit":       res.Limit,
		"total_pages": res.TotalPages,
	})
}

func (h ListingReportHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	reportID := strings.TrimSpace(r.PathValue("id"))
	if _, err := uuid.Parse(reportID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid report id"})
		return
	}

	report, err := h.Reports.GetReport(r.Context(), reportID)
	if err != nil {
		if errors.Is(err, models.ErrReportNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "report not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load report"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]models.Report{"report": report})
}

func (h ListingReportHandler) UpdateStatusAdmin(w http.ResponseWriter, r *http.Request) {
	adminID, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	reportID := strings.TrimSpace(r.PathValue("id"))
	if _, err := uuid.Parse(reportID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid report id"})
		return
	}

	var req updateReportStatusRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	status, err := normalizeReportStatus(req.Status)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	req.ResolutionNote = strings.TrimSpace(req.ResolutionNote)
	if utf8.RuneCountInString(req.ResolutionNote) > maxResolutionNoteLen {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resolution_note is too long"})
		return
	}

	report, err := h.Reports.UpdateReportStatus(r.Context(), reportID, adminID, status, req.ResolutionNote)
	if err != nil {
		if errors.Is(err, models.ErrReportNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "report not found"})
			return
		}
		observability.Error(r.Context(), "listing_reports.admin_update_status.failed", map[string]any{
			"report_id": reportID,
			"admin_id":  adminID,
			"error":     err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update report"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]models.Report{"report": report})
}

func (h ListingReportHandler) AddActionAdmin(w http.ResponseWriter, r *http.Request) {
	adminID, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	reportID := strings.TrimSpace(r.PathValue("id"))
	if _, err := uuid.Parse(reportID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid report id"})
		return
	}

	var req createModerationActionRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxModerationPayloadLen))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	actionType, err := normalizeModerationAction(req.ActionType)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.Payload == nil {
		req.Payload = map[string]any{}
	}

	action, err := h.Reports.AddAction(r.Context(), reportID, adminID, actionType, req.Payload)
	if err != nil {
		if errors.Is(err, models.ErrReportNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "report not found"})
			return
		}
		observability.Error(r.Context(), "listing_reports.admin_add_action.failed", map[string]any{
			"report_id": reportID,
			"admin_id":  adminID,
			"error":     err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add action"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]models.ModerationAction{"action": action})
}

func (h ListingReportHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", false
	}
	if h.Admins == nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
		return "", false
	}
	isAdmin, err := h.Admins.IsAdminByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return "", false
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to verify admin access"})
		return "", false
	}
	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
		return "", false
	}
	return userID, true
}

func parseReportStatusFilter(raw string) (string, error) {
	status := strings.TrimSpace(raw)
	if status == "" || status == "all" {
		return status, nil
	}
	return normalizeReportStatus(status)
}

func normalizeReportStatus(raw string) (string, error) {
	status := strings.TrimSpace(raw)
	if _, ok := allowedReportStatuses[status]; !ok {
		return "", errors.New("invalid status")
	}
	return status, nil
}

func normalizeModerationAction(raw string) (string, error) {
	action := strings.TrimSpace(raw)
	if _, ok := allowedModerationActions[action]; !ok {
		return "", errors.New("invalid action_type")
	}
	return action, nil
}

func parsePositiveIntParam(raw string, fallback int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, errors.New("invalid positive integer")
	}
	return value, nil
}
