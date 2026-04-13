package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
)

type notificationStore interface {
	ListByUser(ctx context.Context, userID string, unreadOnly *bool, page, limit int) (models.NotificationsPage, error)
	MarkRead(ctx context.Context, notificationID, userID string) (models.Notification, error)
	MarkAllRead(ctx context.Context, userID string) (int64, error)
}

type NotificationHandler struct {
	Notifications notificationStore
}

func (h NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	page, limit, err := parsePaginationParams(r, 1, 20)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	var unreadOnly *bool
	if raw := strings.TrimSpace(r.URL.Query().Get("unread")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unread must be true or false"})
			return
		}
		unreadOnly = &parsed
	}

	res, err := h.Notifications.ListByUser(r.Context(), userID, unreadOnly, page, limit)
	if err != nil {
		observability.Error(r.Context(), "notifications.list.failed", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load notifications"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"notifications": res.Notifications,
		"total":         res.Total,
		"page":          res.Page,
		"limit":         res.Limit,
		"total_pages":   res.TotalPages,
		"unread_total":  res.UnreadTotal,
	})
}

func (h NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	notificationID := strings.TrimSpace(r.PathValue("id"))
	if _, err := uuid.Parse(notificationID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid notification id"})
		return
	}

	notification, err := h.Notifications.MarkRead(r.Context(), notificationID, userID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotificationNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "notification not found"})
			return
		default:
			observability.Error(r.Context(), "notifications.mark_read.failed", map[string]any{
				"user_id":         userID,
				"notification_id": notificationID,
				"error":           err.Error(),
			})
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to mark notification read"})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"notification": notification})
}

func (h NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	updated, err := h.Notifications.MarkAllRead(r.Context(), userID)
	if err != nil {
		observability.Error(r.Context(), "notifications.mark_all_read.failed", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to mark notifications read"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"updated": updated})
}
