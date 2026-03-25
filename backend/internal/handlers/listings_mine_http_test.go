package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListingHandlerListMineUnauthorized(t *testing.T) {
	h := ListingHandler{}
	rec := httptest.NewRecorder()
	h.ListMine(rec, httptest.NewRequest(http.MethodGet, "/api/v1/listings/me", nil))
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
