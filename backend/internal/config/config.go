package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                                string
	DatabaseURL                         string
	TokenSecret                         string
	TokenExpiryHours                    int
	ListingWriteRateLimitPerIP          int
	ListingWriteRateLimitWindowMinutes  int
	ListingProhibitedWords              []string
	PasswordResetExpiryMinutes          int
	PasswordResetCooldownMinutes        int
	PasswordResetOTPDigits              int
	PasswordResetMaxAttempts            int
	PasswordResetRateLimitPerIP         int
	PasswordResetRateLimitWindowMinutes int
	SMTPHost                            string
	SMTPPort                            string
	SMTPUsername                        string
	SMTPPassword                        string
	SMTPFromEmail                       string
	SMTPFromName                        string
	SMTPSecurity                        string
	SMTPTimeoutSeconds                  int
	CorsOrigin                          string
}

func Load() (Config, error) {
	loadDotEnv(".env")

	expiryHours := 24
	if raw := os.Getenv("TOKEN_EXPIRY_HOURS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		expiryHours = parsed
	}
	listingWriteRateLimitPerIP := 30
	if raw := os.Getenv("LISTING_WRITE_RATE_LIMIT_PER_IP"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		listingWriteRateLimitPerIP = parsed
	}
	listingWriteRateLimitWindowMinutes := 60
	if raw := os.Getenv("LISTING_WRITE_RATE_LIMIT_WINDOW_MINUTES"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		listingWriteRateLimitWindowMinutes = parsed
	}
	listingProhibitedWords := parseCSVList(os.Getenv("LISTING_PROHIBITED_WORDS"))

	passwordResetExpiryMinutes := 15
	if raw := os.Getenv("PASSWORD_RESET_EXPIRY_MINUTES"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		passwordResetExpiryMinutes = parsed
	}
	passwordResetCooldownMinutes := 5
	if raw := os.Getenv("PASSWORD_RESET_COOLDOWN_MINUTES"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		passwordResetCooldownMinutes = parsed
	}
	passwordResetOTPDigits := 6
	if raw := os.Getenv("PASSWORD_RESET_OTP_DIGITS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		passwordResetOTPDigits = parsed
	}
	passwordResetMaxAttempts := 5
	if raw := os.Getenv("PASSWORD_RESET_MAX_ATTEMPTS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		passwordResetMaxAttempts = parsed
	}
	passwordResetRateLimitPerIP := 5
	if raw := os.Getenv("PASSWORD_RESET_RATE_LIMIT_PER_IP"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		passwordResetRateLimitPerIP = parsed
	}
	passwordResetRateLimitWindowMinutes := 60
	if raw := os.Getenv("PASSWORD_RESET_RATE_LIMIT_WINDOW_MINUTES"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		passwordResetRateLimitWindowMinutes = parsed
	}
	smtpTimeoutSeconds := 10
	if raw := os.Getenv("SMTP_TIMEOUT_SECONDS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, err
		}
		smtpTimeoutSeconds = parsed
	}

	cfg := Config{
		Port:                                envOrDefault("PORT", "8080"),
		DatabaseURL:                         os.Getenv("DATABASE_URL"),
		TokenSecret:                         os.Getenv("TOKEN_SECRET"),
		TokenExpiryHours:                    expiryHours,
		ListingWriteRateLimitPerIP:          listingWriteRateLimitPerIP,
		ListingWriteRateLimitWindowMinutes:  listingWriteRateLimitWindowMinutes,
		ListingProhibitedWords:              listingProhibitedWords,
		PasswordResetExpiryMinutes:          passwordResetExpiryMinutes,
		PasswordResetCooldownMinutes:        passwordResetCooldownMinutes,
		PasswordResetOTPDigits:              passwordResetOTPDigits,
		PasswordResetMaxAttempts:            passwordResetMaxAttempts,
		PasswordResetRateLimitPerIP:         passwordResetRateLimitPerIP,
		PasswordResetRateLimitWindowMinutes: passwordResetRateLimitWindowMinutes,
		SMTPHost:                            os.Getenv("SMTP_HOST"),
		SMTPPort:                            envOrDefault("SMTP_PORT", "587"),
		SMTPUsername:                        os.Getenv("SMTP_USERNAME"),
		SMTPPassword:                        os.Getenv("SMTP_PASSWORD"),
		SMTPFromEmail:                       envOrDefault("SMTP_FROM_EMAIL", "no-reply@resellution.local"),
		SMTPFromName:                        envOrDefault("SMTP_FROM_NAME", "ReSellution"),
		SMTPSecurity:                        strings.ToLower(envOrDefault("SMTP_SECURITY", "starttls")),
		SMTPTimeoutSeconds:                  smtpTimeoutSeconds,
		CorsOrigin:                          envOrDefault("CORS_ORIGIN", "http://localhost:5173,http://127.0.0.1:5173"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.TokenSecret == "" {
		return Config{}, errors.New("TOKEN_SECRET is required")
	}
	if err := validateSMTPConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseCSVList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		term := strings.ToLower(strings.TrimSpace(part))
		if term == "" {
			continue
		}
		out = append(out, term)
	}
	return out
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func validateSMTPConfig(cfg Config) error {
	if cfg.SMTPTimeoutSeconds <= 0 {
		return errors.New("SMTP_TIMEOUT_SECONDS must be greater than 0")
	}

	if cfg.SMTPHost == "" {
		if cfg.SMTPUsername != "" || cfg.SMTPPassword != "" {
			return errors.New("SMTP_HOST is required when SMTP_USERNAME or SMTP_PASSWORD is set")
		}
		return nil
	}

	if _, err := mail.ParseAddress(cfg.SMTPFromEmail); err != nil {
		return fmt.Errorf("SMTP_FROM_EMAIL must be a valid email address: %w", err)
	}
	if (cfg.SMTPUsername == "") != (cfg.SMTPPassword == "") {
		return errors.New("SMTP_USERNAME and SMTP_PASSWORD must either both be set or both be empty")
	}

	switch cfg.SMTPSecurity {
	case "starttls", "tls", "none":
		return nil
	default:
		return fmt.Errorf("SMTP_SECURITY must be one of starttls, tls, or none")
	}
}
