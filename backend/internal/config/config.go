package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                       string
	DatabaseURL                string
	TokenSecret                string
	TokenExpiryHours           int
	PasswordResetExpiryMinutes int
	PasswordResetCooldownMinutes int
	PasswordResetOTPDigits     int
	PasswordResetMaxAttempts   int
	PasswordResetRateLimitPerIP int
	PasswordResetRateLimitWindowMinutes int
	SMTPHost                   string
	SMTPPort                   string
	SMTPUsername               string
	SMTPPassword               string
	SMTPFromEmail              string
	SMTPFromName               string
	CorsOrigin                 string
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

	cfg := Config{
		Port:                       envOrDefault("PORT", "8080"),
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		TokenSecret:                os.Getenv("TOKEN_SECRET"),
		TokenExpiryHours:           expiryHours,
		PasswordResetExpiryMinutes: passwordResetExpiryMinutes,
		PasswordResetCooldownMinutes: passwordResetCooldownMinutes,
		PasswordResetOTPDigits:     passwordResetOTPDigits,
		PasswordResetMaxAttempts:   passwordResetMaxAttempts,
		PasswordResetRateLimitPerIP: passwordResetRateLimitPerIP,
		PasswordResetRateLimitWindowMinutes: passwordResetRateLimitWindowMinutes,
		SMTPHost:                   os.Getenv("SMTP_HOST"),
		SMTPPort:                   envOrDefault("SMTP_PORT", "587"),
		SMTPUsername:               os.Getenv("SMTP_USERNAME"),
		SMTPPassword:               os.Getenv("SMTP_PASSWORD"),
		SMTPFromEmail:              envOrDefault("SMTP_FROM_EMAIL", "no-reply@resellution.local"),
		SMTPFromName:               envOrDefault("SMTP_FROM_NAME", "ReSellution"),
		CorsOrigin:                 envOrDefault("CORS_ORIGIN", "http://localhost:5173,http://127.0.0.1:5173"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.TokenSecret == "" {
		return Config{}, errors.New("TOKEN_SECRET is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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
