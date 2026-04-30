package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

type stubListingReportStore struct {
	createFn       func(ctx context.Context, listingID, reporterID, reason string) (models.Report, error)
	listFn         func(ctx context.Context, status string, page, limit int) (models.ReportsPage, error)
	getFn          func(ctx context.Context, reportID string) (models.Report, error)
	updateStatusFn func(ctx context.Context, reportID, adminID, status, resolutionNote string) (models.Report, error)
	addActionFn    func(ctx context.Context, reportID, adminID, actionType string, payload map[string]any) (models.ModerationAction, error)
}

func (s stubListingReportStore) CreateListingReport(ctx context.Context, listingID, reporterID, reason string) (models.Report, error) {
	if s.createFn == nil {
		return models.Report{ID: "123e4567-e89b-12d3-a456-426614174000", Status: "open"}, nil
	}
	return s.createFn(ctx, listingID, reporterID, reason)
}

func (s stubListingReportStore) ListReports(ctx context.Context, status string, page, limit int) (models.ReportsPage, error) {
	if s.listFn == nil {
		return models.ReportsPage{Reports: []models.Report{}, Page: page, Limit: limit, TotalPages: 1}, nil
	}
	return s.listFn(ctx, status, page, limit)
}

func (s stubListingReportStore) GetReport(ctx context.Context, reportID string) (models.Report, error) {
	if s.getFn == nil {
		return models.Report{ID: reportID, Status: "open"}, nil
	}
	return s.getFn(ctx, reportID)
}

func (s stubListingReportStore) UpdateReportStatus(ctx context.Context, reportID, adminID, status, resolutionNote string) (models.Report, error) {
	if s.updateStatusFn == nil {
		return models.Report{ID: reportID, Status: status}, nil
	}
	return s.updateStatusFn(ctx, reportID, adminID, status, resolutionNote)
}

func (s stubListingReportStore) AddAction(ctx context.Context, reportID, adminID, actionType string, payload map[string]any) (models.ModerationAction, error) {
	if s.addActionFn == nil {
		raw, _ := json.Marshal(payload)
		return models.ModerationAction{ID: "123e4567-e89b-12d3-a456-426614174001", ReportID: reportID, ActorUserID: adminID, ActionType: actionType, ActionPayload: raw}, nil
	}
	return s.addActionFn(ctx, reportID, adminID, actionType, payload)
}

type stubAdminChecker struct {
	isAdmin bool
	err     error
}

func (s stubAdminChecker) IsAdminByID(ctx context.Context, userID string) (bool, error) {
	return s.isAdmin, s.err
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
			createFn: func(ctx context.Context, listingID, reporterID, reason string) (models.Report, error) {
				return models.Report{}, models.ErrListingReportOwnListing
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

func TestListingReportHandlerListAdminForbidden(t *testing.T) {
	const secret = "listing-report-http-test-secret-4"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{},
		Admins:  stubAdminChecker{isAdmin: false},
	}
	wrapped := middleware.Auth(tm, h.ListAdmin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "34343434-3434-3434-3434-343434343434"))
	wrapped(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestListingReportHandlerListAdminInvalidStatus(t *testing.T) {
	const secret = "listing-report-http-test-secret-5"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{},
		Admins:  stubAdminChecker{isAdmin: true},
	}
	wrapped := middleware.Auth(tm, h.ListAdmin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports?status=bad", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "35353535-3535-3535-3535-353535353535"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListingReportHandlerUpdateStatusAdmin(t *testing.T) {
	const secret = "listing-report-http-test-secret-6"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{},
		Admins:  stubAdminChecker{isAdmin: true},
	}
	wrapped := middleware.Auth(tm, h.UpdateStatusAdmin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/reports/123e4567-e89b-12d3-a456-426614174000", strings.NewReader(`{"status":"resolved","resolution_note":"Handled"}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "36363636-3636-3636-3636-363636363636"))
	wrapped(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListingReportHandlerGetAdminOK(t *testing.T) {
	const secret = "listing-report-http-test-secret-8"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{},
		Admins:  stubAdminChecker{isAdmin: true},
	}
	wrapped := middleware.Auth(tm, h.GetAdmin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports/123e4567-e89b-12d3-a456-426614174000", nil)
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "38383838-3838-3838-3838-383838383838"))
	wrapped(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListingReportHandlerListAdminOK(t *testing.T) {
	const secret = "listing-report-http-test-secret-9"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{},
		Admins:  stubAdminChecker{isAdmin: true},
	}
	wrapped := middleware.Auth(tm, h.ListAdmin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports?status=open&page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "39393939-3939-3939-3939-393939393939"))
	wrapped(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListingReportHandlerAddActionAdminOK(t *testing.T) {
	const secret = "listing-report-http-test-secret-10"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{},
		Admins:  stubAdminChecker{isAdmin: true},
	}
	wrapped := middleware.Auth(tm, h.AddActionAdmin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/reports/123e4567-e89b-12d3-a456-426614174000/actions", strings.NewReader(`{"action_type":"note","payload":{"text":"reviewed"}}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "3a3a3a3a-3a3a-3a3a-3a3a-3a3a3a3a3a3a"))
	wrapped(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestListingReportHandlerAddActionAdminInvalidAction(t *testing.T) {
	const secret = "listing-report-http-test-secret-7"
	tm := utils.NewTokenManager(secret)
	h := ListingReportHandler{
		Reports: stubListingReportStore{},
		Admins:  stubAdminChecker{isAdmin: true},
	}
	wrapped := middleware.Auth(tm, h.AddActionAdmin)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/reports/123e4567-e89b-12d3-a456-426614174000/actions", strings.NewReader(`{"action_type":"bad","payload":{}}`))
	req.SetPathValue("id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t, secret, "37373737-3737-3737-3737-373737373737"))
	wrapped(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
