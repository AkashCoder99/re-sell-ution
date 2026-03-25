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

func setBaseEnv(t *testing.T) {
	t.Helper()

	t.Setenv("DATABASE_URL", "postgres://postgres:password@localhost:5432/resellution?sslmode=disable")
	t.Setenv("TOKEN_SECRET", "test-secret")
	for _, key := range []string{
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
