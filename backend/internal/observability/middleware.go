package observability

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

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

const requestIDContextKey contextKey = "request_id"

func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDContextKey).(string)
	return id, ok
}

func RequestMetrics(metrics *Metrics, logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		rec.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(rec, r.WithContext(ctx))

		duration := time.Since(start)
		path := r.URL.Path
		if path == "" {
			path = "/"
		}

		metrics.RecordRequest(path, rec.statusCode, duration)

		log.Printf("%s %s -> %d (%s) request_id=%s ip=%s ua=%q",
			r.Method, path, rec.statusCode, duration.Round(time.Millisecond), requestID, r.RemoteAddr, r.UserAgent())

		logger.Info("request",
			map[string]any{
				"request_id":  requestID,
				"method":      r.Method,
				"path":        path,
				"status":      rec.statusCode,
				"duration_ms": duration.Milliseconds(),
				"remote_addr": r.RemoteAddr,
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
