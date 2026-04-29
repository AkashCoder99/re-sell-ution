package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"resellution/backend/internal/config"
	"resellution/backend/internal/db"
	"resellution/backend/internal/handlers"
	"resellution/backend/internal/middleware"
	"resellution/backend/internal/models"
	"resellution/backend/internal/observability"
	"resellution/backend/internal/ratelimit"
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
	listingStore := models.ListingStore{DB: database}
	categoryStore := models.CategoryStore{DB: database}
	favoriteStore := models.FavoriteStore{DB: database}
	conversationStore := models.ConversationStore{DB: database}
	notificationStore := models.NotificationStore{DB: database}
	listingReportStore := models.ListingReportStore{DB: database}
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
			Security:  cfg.SMTPSecurity,
			Timeout:   time.Duration(cfg.SMTPTimeoutSeconds) * time.Second,
		}
	}

	listingHandler := handlers.ListingHandler{
		Listings:         listingStore,
		ProhibitedWords: cfg.ListingProhibitedWords,
	}
	categoryHandler := handlers.CategoryHandler{Categories: categoryStore}
	favoriteHandler := handlers.FavoriteHandler{Favorites: favoriteStore}
	conversationHandler := handlers.ConversationHandler{Conversations: conversationStore, EmailSender: emailSender}
	notificationHandler := handlers.NotificationHandler{Notifications: notificationStore}
	listingReportHandler := handlers.ListingReportHandler{Reports: listingReportStore}

	authHandler := handlers.AuthHandler{
		Users:                        userStore,
		TokenManager:                 tokenManager,
		EmailSender:                  emailSender,
		TokenExpiryHours:             cfg.TokenExpiryHours,
		PasswordResetExpiryMinutes:   cfg.PasswordResetExpiryMinutes,
		PasswordResetCooldownMinutes: cfg.PasswordResetCooldownMinutes,
		PasswordResetOTPDigits:       cfg.PasswordResetOTPDigits,
		PasswordResetMaxAttempts:     cfg.PasswordResetMaxAttempts,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"ReSellution API","status":"ok","health":"/health","metrics":"/metrics","dashboard":"/metrics/dashboard","auth_base":"/api/v1/auth"}`))
	})
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./openapi/openapi.yaml")
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<title>ReSellution API Docs</title>
		<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
		<style>
			body { margin: 0; background: #0b1020; }
			#swagger-ui { min-height: 100vh; }
		</style>
	</head>
	<body>
		<div id="swagger-ui"></div>
		<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
		<script>
			window.ui = SwaggerUIBundle({
				url: '/openapi.yaml',
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [SwaggerUIBundle.presets.apis],
				layout: 'BaseLayout'
			})
		</script>
	</body>
</html>`))
	})

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Printf("failed to create upload dir %s: %v", uploadDir, err)
	}
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	metrics := observability.NewMetrics()
	logger := observability.DefaultLogger()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /metrics", observability.MetricsHandler(metrics))
	mux.HandleFunc("GET /metrics/prometheus", observability.PrometheusMetricsHandler(metrics))
	mux.HandleFunc("GET /metrics/dashboard", observability.DashboardHandler())
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	passwordResetRateLimiter := ratelimit.NewIPRateLimiter(cfg.PasswordResetRateLimitPerIP, cfg.PasswordResetRateLimitWindowMinutes)
	mux.HandleFunc("POST /api/v1/auth/password/reset/request", ratelimit.IPRateLimit(passwordResetRateLimiter, authHandler.RequestPasswordReset))
	mux.HandleFunc("POST /api/v1/auth/password/reset/confirm", authHandler.ConfirmPasswordReset)
	mux.HandleFunc("GET /api/v1/auth/me", middleware.Auth(tokenManager, authHandler.Me))
	mux.HandleFunc("PATCH /api/v1/users/me", middleware.Auth(tokenManager, authHandler.UpdateProfile))
	mux.HandleFunc("PUT /api/v1/users/me", middleware.Auth(tokenManager, authHandler.UpdateProfile))
	mux.HandleFunc("DELETE /api/v1/users/me", middleware.Auth(tokenManager, authHandler.DeactivateAccount))
	mux.HandleFunc("POST /api/v1/auth/logout", middleware.Auth(tokenManager, authHandler.Logout))

	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("GET /api/v1/categories/tree", categoryHandler.Tree)
	mux.HandleFunc("GET /api/v1/listings/browse", listingHandler.Browse)
	mux.HandleFunc("GET /api/v1/listings/search", listingHandler.Search)
	listingWriteRateLimiter := ratelimit.NewIPRateLimiter(cfg.ListingWriteRateLimitPerIP, cfg.ListingWriteRateLimitWindowMinutes)
	mux.HandleFunc(
		"POST /api/v1/listings",
		ratelimit.IPRateLimitWithMessage(
			listingWriteRateLimiter,
			"Too many listing write requests. Try again in %d minutes",
			middleware.Auth(tokenManager, listingHandler.Create),
		),
	)
	mux.HandleFunc("GET /api/v1/listings/me", middleware.Auth(tokenManager, listingHandler.ListMine))
	mux.HandleFunc(
		"PATCH /api/v1/listings/{id}",
		ratelimit.IPRateLimitWithMessage(
			listingWriteRateLimiter,
			"Too many listing write requests. Try again in %d minutes",
			middleware.Auth(tokenManager, listingHandler.Update),
		),
	)
	mux.HandleFunc("PATCH /api/v1/listings/{id}/status", middleware.Auth(tokenManager, listingHandler.PatchStatus))
	mux.HandleFunc("DELETE /api/v1/listings/{id}", middleware.Auth(tokenManager, listingHandler.Delete))
	mux.HandleFunc("POST /api/v1/listings/{id}/images", middleware.Auth(tokenManager, listingHandler.UploadImage))
	mux.HandleFunc("POST /api/v1/listings/{id}/report", middleware.Auth(tokenManager, listingReportHandler.Create))
	mux.HandleFunc("GET /api/v1/favorites", middleware.Auth(tokenManager, favoriteHandler.List))
	mux.HandleFunc("GET /api/v1/favorites/{listing_id}", middleware.Auth(tokenManager, favoriteHandler.Status))
	mux.HandleFunc("PUT /api/v1/favorites/{listing_id}", middleware.Auth(tokenManager, favoriteHandler.Add))
	mux.HandleFunc("DELETE /api/v1/favorites/{listing_id}", middleware.Auth(tokenManager, favoriteHandler.Remove))
	mux.HandleFunc("GET /api/v1/chat/conversations", middleware.Auth(tokenManager, conversationHandler.List))
	mux.HandleFunc("POST /api/v1/chat/conversations", middleware.Auth(tokenManager, conversationHandler.Create))
	mux.HandleFunc("GET /api/v1/chat/conversations/{id}", middleware.Auth(tokenManager, conversationHandler.Get))
	mux.HandleFunc("GET /api/v1/chat/conversations/{id}/messages", middleware.Auth(tokenManager, conversationHandler.ListMessages))
	mux.HandleFunc("POST /api/v1/chat/conversations/{id}/messages", middleware.Auth(tokenManager, conversationHandler.SendMessage))
	mux.HandleFunc("PATCH /api/v1/chat/conversations/{id}/read", middleware.Auth(tokenManager, conversationHandler.MarkRead))
	mux.HandleFunc("GET /api/v1/notifications", middleware.Auth(tokenManager, notificationHandler.List))
	mux.HandleFunc("PATCH /api/v1/notifications/{id}/read", middleware.Auth(tokenManager, notificationHandler.MarkRead))
	mux.HandleFunc("PATCH /api/v1/notifications/read-all", middleware.Auth(tokenManager, notificationHandler.MarkAllRead))

	handler := observability.RequestMetrics(metrics, logger, withCORS(cfg.CorsOrigin, mux))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if emailSender == nil {
		logger.Warn("smtp delivery disabled", map[string]any{
			"password_reset_delivery": "application_log",
			"note":                    "set SMTP_HOST and related SMTP_* env vars for email delivery",
		})
	} else {
		logger.Info("smtp delivery enabled", map[string]any{
			"host":     cfg.SMTPHost,
			"port":     cfg.SMTPPort,
			"security": cfg.SMTPSecurity,
			"from":     cfg.SMTPFromEmail,
		})
	}

	logger.Info("backend started", map[string]any{"port": cfg.Port, "url": "http://localhost:" + cfg.Port})
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Expose-Headers", "X-Correlation-ID, X-Request-ID")
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
