package handlers

import (
	"net/http/httptest"
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
