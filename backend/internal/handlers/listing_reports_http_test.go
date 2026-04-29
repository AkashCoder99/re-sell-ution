package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

type stubListingReportStore struct {
	createFn func(ctx context.Context, listingID, reporterID, reason string) error
}

func (s stubListingReportStore) Create(ctx context.Context, listingID, reporterID, reason string) error {
	if s.createFn == nil {
		return nil
	}
	return s.createFn(ctx, listingID, reporterID, reason)
}

func TestListingReportHandlerCreateUnauthorized(t *testing.T) {
	h := ListingReportHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings/x/report", nil)
	req.SetPathValue("id", "bad-id")
	h.Create(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestListingReportHandlerCreateInvalidListingID(t *testing.T) {
	const secret = "listing-report-http-test-secret-1"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{}
	wrapped := middleware.Auth(tm, h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings/bad-id/report", nil)
	req.SetPathValue("id", "bad-id")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "31313131-3131-3131-3131-313131313131"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListingReportHandlerCreateInvalidJSON(t *testing.T) {
	const secret = "listing-report-http-test-secret-2"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{}
	wrapped := middleware.Auth(tm, h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings/123e4567-e89b-12d3-a456-426614174000/report", strings.NewReader(`{`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "32323232-3232-3232-3232-323232323232"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListingReportHandlerCreateOwnListing(t *testing.T) {
	const secret = "listing-report-http-test-secret-3"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{
			createFn: func(ctx context.Context, listingID, reporterID, reason string) error {
				return models.ErrListingReportOwnListing
			},
		},
	}
	wrapped := middleware.Auth(tm, h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listings/123e4567-e89b-12d3-a456-426614174000/report", strings.NewReader(`{"reason":"spam"}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "33333333-3333-3333-3333-333333333334"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
