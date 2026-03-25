package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/utils"
)

func TestListingHandlerCreateUnauthorized(t *testing.T) {
	h := ListingHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings", strings.NewReader(`{}`))
	h.Create(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestListingHandlerCreateInvalidJSONWithAuth(t *testing.T) {
	const secret = "listing-http-test-secret"
	userID := "33333333-3333-3333-3333-333333333333"
	tm := utils.NewTokenManager(secret)
	h := ListingHandler{}
	wrapped := middleware.Auth(tm, h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, userID))
	wrapped(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListingHandlerUpdateInvalidListingID(t *testing.T) {
	const secret = "listing-http-test-secret-2"
	userID := "44444444-4444-4444-4444-444444444444"
	tm := utils.NewTokenManager(secret)
	h := ListingHandler{}
	wrapped := middleware.Auth(tm, h.Update)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/listings/not-a-uuid", strings.NewReader(`{"title":"x"}`))
	req.SetPathValue("id", "not-a-uuid")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, userID))
	wrapped(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListingHandlerPatchStatusInvalidListingID(t *testing.T) {
	const secret = "listing-http-test-secret-3"
	userID := "55555555-5555-5555-5555-555555555555"
	tm := utils.NewTokenManager(secret)
	h := ListingHandler{}
	wrapped := middleware.Auth(tm, h.PatchStatus)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/listings/x/status", strings.NewReader(`{"status":"active"}`))
	req.SetPathValue("id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, userID))
	wrapped(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListingHandlerDeleteInvalidListingID(t *testing.T) {
	const secret = "listing-http-test-secret-4"
	userID := "66666666-6666-6666-6666-666666666666"
	tm := utils.NewTokenManager(secret)
	h := ListingHandler{}
	wrapped := middleware.Auth(tm, h.Delete)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/listings/x", nil)
	req.SetPathValue("id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, userID))
	wrapped(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListingHandlerUploadImageInvalidListingID(t *testing.T) {
	const secret = "listing-http-test-secret-5"
	userID := "77777777-7777-7777-7777-777777777777"
	tm := utils.NewTokenManager(secret)
	h := ListingHandler{}
	wrapped := middleware.Auth(tm, h.UploadImage)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings/x/images", nil)
	req.SetPathValue("id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, userID))
	wrapped(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
