package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildSafeTSQuery(t *testing.T) {
	normalized, tsQuery, err := buildSafeTSQuery("  Bike + helmet; bike  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if normalized != "bike helmet" {
		t.Fatalf("expected normalized query %q, got %q", "bike helmet", normalized)
	}
	if tsQuery != "bike:* & helmet:*" {
		t.Fatalf("expected tsquery %q, got %q", "bike:* & helmet:*", tsQuery)
	}
}

func TestBuildSafeTSQueryRejectsInvalidInput(t *testing.T) {
	if _, _, err := buildSafeTSQuery("!!!"); err == nil {
		t.Fatal("expected an error for punctuation-only query")
	}
}

func TestParseListingSearchParamsRequiresCoordinatePair(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/listings/search?q=desk&lat=29.6516", nil)

	if _, err := parseListingSearchParams(req); err == nil {
		t.Fatal("expected coordinate pair validation error")
	}
}

func TestParseListingSearchParamsParsesFilters(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"/api/v1/listings/search?q=Bike+Helmet&page=2&limit=5&city=Gainesville&category_id=123e4567-e89b-12d3-a456-426614174000&lat=29.6516&lng=-82.3248&radius_km=15",
		nil,
	)

	params, err := parseListingSearchParams(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if params.Query != "bike helmet" {
		t.Fatalf("expected normalized query %q, got %q", "bike helmet", params.Query)
	}
	if params.TSQuery != "bike:* & helmet:*" {
		t.Fatalf("expected tsquery %q, got %q", "bike:* & helmet:*", params.TSQuery)
	}
	if params.Page != 2 || params.Limit != 5 {
		t.Fatalf("unexpected pagination: page=%d limit=%d", params.Page, params.Limit)
	}
	if params.City != "Gainesville" {
		t.Fatalf("expected city Gainesville, got %q", params.City)
	}
	if params.CategoryID == nil || *params.CategoryID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("expected parsed category id, got %+v", params.CategoryID)
	}
	if params.Latitude == nil || params.Longitude == nil || params.RadiusKM == nil {
		t.Fatal("expected parsed geo filters")
	}
	if *params.RadiusKM != 15 {
		t.Fatalf("expected radius 15, got %v", *params.RadiusKM)
	}
}

func TestSplitSearchTokens(t *testing.T) {
	got := splitSearchTokens("Bike-Helmet_2024")
	if len(got) != 3 || got[0] != "Bike" || got[1] != "Helmet" || got[2] != "2024" {
		t.Fatalf("unexpected tokens: %#v", got)
	}
}

func TestParsePositiveIntQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/?page=3", nil)
	n, err := parsePositiveIntQuery(req, "page", 1)
	if err != nil || n != 3 {
		t.Fatalf("page: %d %v", n, err)
	}
	req2 := httptest.NewRequest("GET", "/?page=0", nil)
	if _, err := parsePositiveIntQuery(req2, "page", 1); err == nil {
		t.Fatal("expected error for page=0")
	}
	req3 := httptest.NewRequest("GET", "/", nil)
	n3, err := parsePositiveIntQuery(req3, "page", 7)
	if err != nil || n3 != 7 {
		t.Fatalf("fallback: %d %v", n3, err)
	}
}

func TestParseOptionalFloatQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/?lat=1.5", nil)
	v, err := parseOptionalFloatQuery(req, "lat")
	if err != nil || v == nil || *v != 1.5 {
		t.Fatalf("lat: %v %v", v, err)
	}
	empty := httptest.NewRequest("GET", "/", nil)
	v2, err := parseOptionalFloatQuery(empty, "lat")
	if err != nil || v2 != nil {
		t.Fatalf("empty: %v %v", v2, err)
	}
	bad := httptest.NewRequest("GET", "/?lat=xx", nil)
	if _, err := parseOptionalFloatQuery(bad, "lat"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestBuildSafeTSQueryRequiresQ(t *testing.T) {
	if _, _, err := buildSafeTSQuery("   "); err == nil {
		t.Fatal("expected empty q error")
	}
}

func TestBuildSafeTSQueryTooLong(t *testing.T) {
	s := strings.Repeat("a", searchQueryMaxLen+1)
	if _, _, err := buildSafeTSQuery(s); err == nil {
		t.Fatal("expected too long")
	}
}

func TestParseListingSearchParamsLimitCap(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/listings/search?q=hello&limit=99", nil)
	if _, err := parseListingSearchParams(req); err == nil {
		t.Fatal("expected limit cap error")
	}
}

func TestParseListingSearchParamsRadiusWithoutCoords(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/listings/search?q=hello&radius_km=10", nil)
	if _, err := parseListingSearchParams(req); err == nil {
		t.Fatal("expected radius without lat/lng error")
	}
}

func TestParseListingSearchParamsInvalidCategoryID(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/listings/search?q=hello&category_id=bad", nil)
	if _, err := parseListingSearchParams(req); err == nil {
		t.Fatal("expected invalid category_id")
	}
}

func TestListingHandlerSearchBadParams(t *testing.T) {
	h := ListingHandler{}
	req := httptest.NewRequest("GET", "/api/v1/listings/search", nil)
	rec := httptest.NewRecorder()
	h.Search(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
