package models

import (
	"strings"
	"testing"
)

func TestBuildListingSearchFiltersIncludesConditionAndPrice(t *testing.T) {
	categoryID := "123e4567-e89b-12d3-a456-426614174000"
	condition := "good"
	minPrice := 10.0
	maxPrice := 50.0
	params := ListingSearchParams{
		TSQuery:    "bike:*",
		City:       "Gainesville",
		CategoryID: &categoryID,
		Condition:  &condition,
		MinPrice:   &minPrice,
		MaxPrice:   &maxPrice,
	}

	where, args, _ := buildListingSearchFilters(params)
	if !strings.Contains(where, "l.condition =") {
		t.Fatalf("expected condition filter in where: %s", where)
	}
	if !strings.Contains(where, "l.price >=") || !strings.Contains(where, "l.price <=") {
		t.Fatalf("expected price filters in where: %s", where)
	}
	if len(args) != 6 {
		t.Fatalf("expected 6 args (tsquery + city + category + condition + min + max), got %d", len(args))
	}
}

func TestBuildListingSearchOrderClauseSortOptions(t *testing.T) {
	cases := []struct {
		sortBy   string
		contains string
	}{
		{sortBy: "created_at_asc", contains: "l.created_at ASC"},
		{sortBy: "created_at_desc", contains: "l.created_at DESC"},
		{sortBy: "price_asc", contains: "l.price ASC"},
		{sortBy: "price_desc", contains: "l.price DESC"},
	}

	for _, tc := range cases {
		got := buildListingSearchOrderClause(tc.sortBy, "NULL::double precision")
		if !strings.Contains(got, tc.contains) {
			t.Fatalf("sort %s expected %q in %q", tc.sortBy, tc.contains, got)
		}
		if !strings.Contains(got, "l.id ASC") {
			t.Fatalf("sort %s expected stable id tie-breaker in %q", tc.sortBy, got)
		}
	}
}

func TestBuildListingSearchOrderClauseDefaultIncludesRelevance(t *testing.T) {
	got := buildListingSearchOrderClause("relevance", "NULL::double precision")
	if !strings.Contains(got, "ts_rank_cd") {
		t.Fatalf("expected relevance ranking in %q", got)
	}
}
