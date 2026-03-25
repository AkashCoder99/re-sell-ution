package utils

import (
	"strings"
	"testing"
	"time"
)

func TestTokenManagerCreateAndParse(t *testing.T) {
	m := NewTokenManager("unit-test-secret-key")
	userID := "22222222-2222-2222-2222-222222222222"
	exp := time.Now().Add(2 * time.Hour)

	token, err := m.Create(userID, exp)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT segments, got %d", len(parts))
	}

	got, err := m.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if got != userID {
		t.Fatalf("expected sub %q, got %q", userID, got)
	}
}

func TestTokenManagerParseRejectsTamperedSignature(t *testing.T) {
	m := NewTokenManager("unit-test-secret-key")
	token, err := m.Create("user-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	parts[2] = "bogus"
	tampered := strings.Join(parts, ".")

	if _, err := m.Parse(tampered); err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestTokenManagerParseRejectsExpiredToken(t *testing.T) {
	m := NewTokenManager("unit-test-secret-key")
	token, err := m.Create("user-1", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Parse(token); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestTokenManagerParseRejectsMalformedToken(t *testing.T) {
	m := NewTokenManager("unit-test-secret-key")
	if _, err := m.Parse("not-enough-parts"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFingerprintStableForSameToken(t *testing.T) {
	m := NewTokenManager("unit-test-secret-key")
	token, err := m.Create("user-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	a := Fingerprint(token)
	b := Fingerprint(token)
	if a != b || len(a) != 64 {
		t.Fatalf("unexpected fingerprint: %q", a)
	}
}
