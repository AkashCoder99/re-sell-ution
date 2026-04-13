package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

type stubNotificationStore struct {
	listByUserFn  func(ctx context.Context, userID string, unreadOnly *bool, page, limit int) (models.NotificationsPage, error)
	markReadFn    func(ctx context.Context, notificationID, userID string) (models.Notification, error)
	markAllReadFn func(ctx context.Context, userID string) (int64, error)
}

func (s stubNotificationStore) ListByUser(ctx context.Context, userID string, unreadOnly *bool, page, limit int) (models.NotificationsPage, error) {
	if s.listByUserFn == nil {
		return models.NotificationsPage{}, nil
	}
	return s.listByUserFn(ctx, userID, unreadOnly, page, limit)
}

func (s stubNotificationStore) MarkRead(ctx context.Context, notificationID, userID string) (models.Notification, error) {
	if s.markReadFn == nil {
		return models.Notification{}, nil
	}
	return s.markReadFn(ctx, notificationID, userID)
}

func (s stubNotificationStore) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	if s.markAllReadFn == nil {
		return 0, nil
	}
	return s.markAllReadFn(ctx, userID)
}

func TestNotificationHandlerListUnauthorized(t *testing.T) {
	h := NotificationHandler{}
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestNotificationHandlerListInvalidPagination(t *testing.T) {
	const secret = "notifications-http-test-secret-1"
	tm := utils.NewTokenManager(secret)
	h := NotificationHandler{}
	wrapped := middleware.Auth(tm, h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?page=abc", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "24242424-2424-2424-2424-242424242424"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestNotificationHandlerListInvalidUnreadFilter(t *testing.T) {
	const secret = "notifications-http-test-secret-2"
	tm := utils.NewTokenManager(secret)
	h := NotificationHandler{}
	wrapped := middleware.Auth(tm, h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?unread=maybe", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "25252525-2525-2525-2525-252525252525"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestNotificationHandlerMarkReadUnauthorized(t *testing.T) {
	h := NotificationHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/x/read", nil)
	req.SetPathValue("id", "bad-id")
	h.MarkRead(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestNotificationHandlerMarkReadInvalidID(t *testing.T) {
	const secret = "notifications-http-test-secret-3"
	tm := utils.NewTokenManager(secret)
	h := NotificationHandler{}
	wrapped := middleware.Auth(tm, h.MarkRead)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/bad-id/read", nil)
	req.SetPathValue("id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "26262626-2626-2626-2626-262626262626"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestNotificationHandlerMarkReadNotFound(t *testing.T) {
	const secret = "notifications-http-test-secret-4"
	tm := utils.NewTokenManager(secret)
	h := NotificationHandler{
		Notifications: stubNotificationStore{
			markReadFn: func(ctx context.Context, notificationID, userID string) (models.Notification, error) {
				return models.Notification{}, models.ErrNotificationNotFound
			},
		},
	}
	wrapped := middleware.Auth(tm, h.MarkRead)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/123e4567-e89b-12d3-a456-426614174000/read", nil)
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "27272727-2727-2727-2727-272727272727"))
	wrapped(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestNotificationHandlerMarkAllReadUnauthorized(t *testing.T) {
	h := NotificationHandler{}
	rec := httptest.NewRecorder()
	h.MarkAllRead(rec, httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/read-all", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
