package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"resellution/backend/internal/config"
	"resellution/backend/internal/db"
	"resellution/backend/internal/handlers"
	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer database.Close()

	userStore := models.UserStore{DB: database}
	tokenManager := utils.NewTokenManager(cfg.TokenSecret)
	var emailSender utils.EmailSender
	if strings.TrimSpace(cfg.SMTPHost) != "" {
		emailSender = utils.SMTPEmailSender{
			Host:      cfg.SMTPHost,
			Port:      cfg.SMTPPort,
			Username:  cfg.SMTPUsername,
			Password:  cfg.SMTPPassword,
			FromEmail: cfg.SMTPFromEmail,
			FromName:  cfg.SMTPFromName,
		}
	}

	authHandler := handlers.AuthHandler{
		Users:                      userStore,
		TokenManager:               tokenManager,
		EmailSender:                emailSender,
		TokenExpiryHours:           cfg.TokenExpiryHours,
		PasswordResetExpiryMinutes: cfg.PasswordResetExpiryMinutes,
		PasswordResetOTPDigits:     cfg.PasswordResetOTPDigits,
		PasswordResetMaxAttempts:   cfg.PasswordResetMaxAttempts,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/password/reset/request", authHandler.RequestPasswordReset)
	mux.HandleFunc("POST /api/v1/auth/password/reset/confirm", authHandler.ConfirmPasswordReset)
	mux.HandleFunc("GET /api/v1/auth/me", middleware.Auth(tokenManager, authHandler.Me))
	mux.HandleFunc("POST /api/v1/auth/logout", middleware.Auth(tokenManager, authHandler.Logout))

	handler := withCORS(cfg.CorsOrigin, mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("backend running on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func withCORS(allowedOriginsCSV string, next http.Handler) http.Handler {
	allowedOrigins := parseAllowedOrigins(allowedOriginsCSV)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			if origin == "" || !isOriginAllowed(origin, allowedOrigins) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseAllowedOrigins(raw string) map[string]struct{} {
	origins := make(map[string]struct{})
	for _, origin := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		origins[trimmed] = struct{}{}
	}
	return origins
}

func isOriginAllowed(origin string, allowed map[string]struct{}) bool {
	_, ok := allowed[origin]
	return ok
}
