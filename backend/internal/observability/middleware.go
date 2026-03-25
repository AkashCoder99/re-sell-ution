package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

type contextKey string

const (
	correlationIDContextKey contextKey = "correlation_id"
	CorrelationIDHeader     string     = "X-Correlation-ID"
	RequestIDHeader         string     = "X-Request-ID"
	maxCorrelationIDLength  int        = 128
)

func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(correlationIDContextKey).(string)
	return id, ok
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	return CorrelationIDFromContext(ctx)
}

func RequestMetrics(metrics *Metrics, logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logger == nil {
			logger = DefaultLogger()
		}

		start := time.Now()
		correlationID := correlationIDFromRequest(r)
		ctx := context.WithValue(r.Context(), correlationIDContextKey, correlationID)
		req := r.WithContext(ctx)
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		rec.Header().Set(CorrelationIDHeader, correlationID)
		rec.Header().Set(RequestIDHeader, correlationID)

		next.ServeHTTP(rec, req)

		duration := time.Since(start)
		path := normalizedRequestPath(req)
		metrics.RecordRequest(path, req.Method, rec.statusCode, duration)

		logger.InfoContext(req.Context(), "http.request",
			map[string]any{
				"method":      req.Method,
				"path":        path,
				"status":      rec.statusCode,
				"duration_ms": duration.Milliseconds(),
				"remote_addr": req.RemoteAddr,
				"user_agent":  req.UserAgent(),
			},
		)
	})
}

func MetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		snapshot := metrics.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(snapshot)
	}
}

func PrometheusMetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(metrics.PrometheusSnapshot()))
	}
}

func correlationIDFromRequest(r *http.Request) string {
	for _, header := range []string{CorrelationIDHeader, RequestIDHeader} {
		if id, ok := sanitizeCorrelationID(r.Header.Get(header)); ok {
			return id
		}
	}
	return uuid.NewString()
}

func sanitizeCorrelationID(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > maxCorrelationIDLength {
		return "", false
	}

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		if strings.ContainsRune("-_.:/", r) {
			continue
		}
		return "", false
	}

	return value, true
}

func normalizedRequestPath(r *http.Request) string {
	pattern := strings.TrimSpace(r.Pattern)
	if pattern != "" {
		if idx := strings.Index(pattern, " "); idx >= 0 {
			pattern = strings.TrimSpace(pattern[idx+1:])
		}
		if pattern != "" {
			return pattern
		}
	}

	path := strings.TrimSpace(r.URL.Path)
	if path == "" {
		return "/"
	}
	return path
}
