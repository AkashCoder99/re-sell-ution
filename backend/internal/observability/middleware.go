package observability

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func RequestMetrics(metrics *Metrics, logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		path := r.URL.Path
		if path == "" {
			path = "/"
		}

		metrics.RecordRequest(path, rec.statusCode, duration)

		log.Printf("%s %s -> %d (%s) ip=%s ua=%q",
			r.Method, path, rec.statusCode, duration.Round(time.Millisecond), r.RemoteAddr, r.UserAgent())

		logger.Info("request",
			map[string]any{
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
