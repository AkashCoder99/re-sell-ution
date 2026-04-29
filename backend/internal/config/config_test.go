package config

import "testing"

func TestLoadRejectsPartialSMTPCredentials(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_USERNAME", "mailer")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_EMAIL", "no-reply@example.com")

	if _, err := Load(); err == nil {
		t.Fatalf("expected Load to reject partial SMTP credentials")
	}
}

func TestLoadAcceptsValidSMTPConfig(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USERNAME", "mailer")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("SMTP_FROM_EMAIL", "no-reply@example.com")
	t.Setenv("SMTP_SECURITY", "tls")
	t.Setenv("SMTP_TIMEOUT_SECONDS", "20")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected Load to accept valid SMTP config, got error: %v", err)
	}

	if cfg.SMTPSecurity != "tls" {
		t.Fatalf("expected SMTPSecurity to be tls, got %q", cfg.SMTPSecurity)
	}
	if cfg.SMTPTimeoutSeconds != 20 {
		t.Fatalf("expected SMTPTimeoutSeconds to be 20, got %d", cfg.SMTPTimeoutSeconds)
	}
}

func TestLoadParsesListingModerationConfig(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("LISTING_WRITE_RATE_LIMIT_PER_IP", "12")
	t.Setenv("LISTING_WRITE_RATE_LIMIT_WINDOW_MINUTES", "30")
	t.Setenv("LISTING_PROHIBITED_WORDS", "scam, fake, stolen ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected Load to parse listing moderation config, got error: %v", err)
	}
	if cfg.ListingWriteRateLimitPerIP != 12 {
		t.Fatalf("expected ListingWriteRateLimitPerIP=12, got %d", cfg.ListingWriteRateLimitPerIP)
	}
	if cfg.ListingWriteRateLimitWindowMinutes != 30 {
		t.Fatalf("expected ListingWriteRateLimitWindowMinutes=30, got %d", cfg.ListingWriteRateLimitWindowMinutes)
	}
	if len(cfg.ListingProhibitedWords) != 3 {
		t.Fatalf("expected 3 prohibited words, got %d", len(cfg.ListingProhibitedWords))
	}
	if cfg.ListingProhibitedWords[0] != "scam" || cfg.ListingProhibitedWords[2] != "stolen" {
		t.Fatalf("unexpected prohibited words: %#v", cfg.ListingProhibitedWords)
	}
}

func setBaseEnv(t *testing.T) {
	t.Helper()

	t.Setenv("DATABASE_URL", "postgres://postgres:password@localhost:5432/resellution?sslmode=disable")
	t.Setenv("TOKEN_SECRET", "test-secret")
	for _, key := range []string{
		"LISTING_WRITE_RATE_LIMIT_PER_IP",
		"LISTING_WRITE_RATE_LIMIT_WINDOW_MINUTES",
		"LISTING_PROHIBITED_WORDS",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"SMTP_FROM_EMAIL",
		"SMTP_FROM_NAME",
		"SMTP_SECURITY",
		"SMTP_TIMEOUT_SECONDS",
	} {
		t.Setenv(key, "")
	}
}
