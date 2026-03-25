package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"resellution/backend/internal/utils"
)

func TestAuthRejectsMissingAuthorization(t *testing.T) {
	tm := utils.NewTokenManager("test-secret-for-middleware")
	var saw bool
	h := Auth(tm, func(w http.ResponseWriter, r *http.Request) {
		saw = true
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if saw {
		t.Fatal("next handler should not run without Authorization")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRejectsNonBearerHeader(t *testing.T) {
	tm := utils.NewTokenManager("test-secret-for-middleware")
	h := Auth(tm, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not run")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic x")

	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRejectsInvalidToken(t *testing.T) {
	tm := utils.NewTokenManager("test-secret-for-middleware")
	h := Auth(tm, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not run")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")

	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthSetsUserIDInContext(t *testing.T) {
	tm := utils.NewTokenManager("test-secret-for-middleware")
	userID := "11111111-1111-1111-1111-111111111111"
	token, err := tm.Create(userID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	var ctxUserID string
	h := Auth(tm, func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		ctxUserID, ok = UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected user id in context")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	h(httptest.NewRecorder(), req)

	if ctxUserID != userID {
		t.Fatalf("expected user id %q, got %q", userID, ctxUserID)
	}
}

func TestUserIDFromContextMissing(t *testing.T) {
	_, ok := UserIDFromContext(t.Context())
	if ok {
		t.Fatal("expected no user id without middleware")
	}
}
