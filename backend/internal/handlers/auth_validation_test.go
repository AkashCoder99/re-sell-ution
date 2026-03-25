package handlers

import (
	"strings"
	"testing"

	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

func TestValidateEmail(t *testing.T) {
	if err := validateEmail("user@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := validateEmail("not-an-email"); err == nil {
		t.Fatal("expected error")
	}
	if err := validateEmail("a@" + strings.Repeat("b", 300)); err == nil {
		t.Fatal("expected length error")
	}
}

func TestValidateFullName(t *testing.T) {
	if err := validateFullName("Ab"); err != nil {
		t.Fatal(err)
	}
	if err := validateFullName("A"); err == nil {
		t.Fatal("expected too short")
	}
	if err := validateFullName(strings.Repeat("x", maxFullNameLength+1)); err == nil {
		t.Fatal("expected too long")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := validatePassword("abc12345"); err != nil {
		t.Fatal(err)
	}
	if err := validatePassword("short1"); err == nil {
		t.Fatal("expected too short")
	}
	if err := validatePassword("nodigitsaaaa"); err == nil {
		t.Fatal("expected missing digit")
	}
	if err := validatePassword("12345678"); err == nil {
		t.Fatal("expected missing letter")
	}
}

func TestValidateProfileUpdateRequiresField(t *testing.T) {
	req := &updateProfileRequest{}
	if err := validateProfileUpdate(req); err == nil {
		t.Fatal("expected error when all nil")
	}
}

func TestValidateProfileUpdateTrimsAndValidatesPhotoURL(t *testing.T) {
	u := "  https://cdn.example.com/p.png  "
	req := &updateProfileRequest{PhotoURL: &u}
	if err := validateProfileUpdate(req); err != nil {
		t.Fatal(err)
	}
	if *req.PhotoURL != "https://cdn.example.com/p.png" {
		t.Fatalf("unexpected trim: %q", *req.PhotoURL)
	}

	bad := "ftp://x.com/a"
	req2 := &updateProfileRequest{PhotoURL: &bad}
	if err := validateProfileUpdate(req2); err == nil {
		t.Fatal("expected scheme error")
	}
}

func TestToPublicUser(t *testing.T) {
	city := "Gainesville"
	u := models.User{
		ID:              "id-1",
		Email:           "a@b.com",
		FullName:        "N",
		City:            city,
		Bio:             "bio",
		ProfileImageURL: "https://x/img",
	}
	p := toPublicUser(u)
	if p.ID != u.ID || p.Email != u.Email || p.City != city || p.PhotoURL != u.ProfileImageURL {
		t.Fatalf("unexpected mapping: %+v", p)
	}
}

func TestAuthHandlerCreateToken(t *testing.T) {
	h := AuthHandler{
		TokenManager:     utils.NewTokenManager("secret"),
		TokenExpiryHours: 1,
	}
	tok, err := h.createToken("user-xyz")
	if err != nil {
		t.Fatal(err)
	}
	sub, err := h.TokenManager.Parse(tok)
	if err != nil || sub != "user-xyz" {
		t.Fatalf("parse: %v sub=%q", err, sub)
	}
}

func TestAuthHandlerPasswordResetCooldownMinutes(t *testing.T) {
	h := AuthHandler{PasswordResetCooldownMinutes: -1}
	if h.passwordResetCooldownMinutes() != 0 {
		t.Fatal("expected 0 for negative")
	}
	h.PasswordResetCooldownMinutes = 7
	if h.passwordResetCooldownMinutes() != 7 {
		t.Fatal("expected 7")
	}
}

func TestAuthHandlerPasswordResetOTPDigits(t *testing.T) {
	h := AuthHandler{PasswordResetOTPDigits: 0}
	if h.passwordResetOTPDigits() != 6 {
		t.Fatal("expected default 6")
	}
	h.PasswordResetOTPDigits = 8
	if h.passwordResetOTPDigits() != 8 {
		t.Fatal("expected 8")
	}
}

func TestAuthHandlerPasswordResetMaxAttempts(t *testing.T) {
	h := AuthHandler{PasswordResetMaxAttempts: 0}
	if h.passwordResetMaxAttempts() != 5 {
		t.Fatal("expected default 5")
	}
	h.PasswordResetMaxAttempts = 3
	if h.passwordResetMaxAttempts() != 3 {
		t.Fatal("expected 3")
	}
}

func TestGenerateNumericOTP(t *testing.T) {
	if _, err := generateNumericOTP(0); err == nil {
		t.Fatal("expected error for zero length")
	}
	s, err := generateNumericOTP(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 10 {
		t.Fatalf("len %d", len(s))
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			t.Fatalf("non-digit: %q", s)
		}
	}
}

func TestPasswordResetOTPHashDeterministic(t *testing.T) {
	a := passwordResetOTPHash("123456")
	b := passwordResetOTPHash("123456")
	if a != b || len(a) != 64 {
		t.Fatalf("unexpected hash %q", a)
	}
}
