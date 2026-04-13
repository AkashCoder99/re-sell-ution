package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/utils"
)

func TestFavoriteHandlerListUnauthorized(t *testing.T) {
	h := FavoriteHandler{}
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestFavoriteHandlerAddUnauthorized(t *testing.T) {
	h := FavoriteHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/favorites/x", nil)
	req.SetPathValue("listing_id", "bad-id")
	h.Add(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestFavoriteHandlerRemoveUnauthorized(t *testing.T) {
	h := FavoriteHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/favorites/x", nil)
	req.SetPathValue("listing_id", "bad-id")
	h.Remove(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestFavoriteHandlerAddInvalidListingID(t *testing.T) {
	const secret = "favorite-http-test-secret-1"
	tm := utils.NewTokenManager(secret)
	h := FavoriteHandler{}
	wrapped := middleware.Auth(tm, h.Add)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/favorites/x", nil)
	req.SetPathValue("listing_id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "88888888-8888-8888-8888-888888888888"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestFavoriteHandlerRemoveInvalidListingID(t *testing.T) {
	const secret = "favorite-http-test-secret-2"
	tm := utils.NewTokenManager(secret)
	h := FavoriteHandler{}
	wrapped := middleware.Auth(tm, h.Remove)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/favorites/x", nil)
	req.SetPathValue("listing_id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "99999999-9999-9999-9999-999999999999"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestFavoriteHandlerListInvalidPagination(t *testing.T) {
	const secret = "favorite-http-test-secret-3"
	tm := utils.NewTokenManager(secret)
	h := FavoriteHandler{}
	wrapped := middleware.Auth(tm, h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/favorites?page=abc", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "10101010-1010-1010-1010-101010101010"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
