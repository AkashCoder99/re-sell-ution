package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"resellution/backend/internal/utils"
)

func TestAuthHandlerRegisterInvalidJSON(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	h.Register(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerRegisterMissingFields(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	h.Register(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerLoginInvalidJSON(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not-json`))
	h.Login(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerLoginMissingCredentials(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"","password":""}`))
	h.Login(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerRequestPasswordResetInvalidJSON(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	h.RequestPasswordReset(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerConfirmPasswordResetInvalidJSON(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	h.ConfirmPasswordReset(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandlerMeUnauthorized(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	h.Me(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandlerUpdateProfileUnauthorized(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"full_name":"Test User"}`))
	h.UpdateProfile(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandlerLogoutUnauthorized(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	h.Logout(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandlerDeactivateUnauthorized(t *testing.T) {
	h := AuthHandler{}
	rec := httptest.NewRecorder()
	h.DeactivateAccount(rec, httptest.NewRequest(http.MethodDelete, "/", nil))
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func bearerToken(t *testing.T, secret, userID string) string {
	t.Helper()
	tm := utils.NewTokenManager(secret)
	tok, err := tm.Create(userID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}
