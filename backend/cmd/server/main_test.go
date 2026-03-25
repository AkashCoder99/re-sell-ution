package main

import "testing"

func TestParseAllowedOrigins(t *testing.T) {
	m := parseAllowedOrigins(" http://a.com , ,http://b.com ")
	if len(m) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(m))
	}
	if _, ok := m["http://a.com"]; !ok {
		t.Fatal("missing http://a.com")
	}
	if _, ok := m["http://b.com"]; !ok {
		t.Fatal("missing http://b.com")
	}
}

func TestParseAllowedOriginsEmpty(t *testing.T) {
	if len(parseAllowedOrigins("")) != 0 {
		t.Fatal("expected empty map")
	}
}

func TestIsOriginAllowed(t *testing.T) {
	allowed := map[string]struct{}{"http://localhost:5173": {}}
	if !isOriginAllowed("http://localhost:5173", allowed) {
		t.Fatal("expected allowed")
	}
	if isOriginAllowed("http://evil.com", allowed) {
		t.Fatal("expected denied")
	}
}
