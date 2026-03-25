package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestMetricsPropagatesCorrelationIDAndUsesRoutePattern(t *testing.T) {
	metrics := NewMetrics()
	logger := NewLogger()
	mux := http.NewServeMux()

	var correlationID string
	mux.HandleFunc("GET /widgets/{id}", func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		correlationID, ok = CorrelationIDFromContext(r.Context())
		if !ok {
			t.Fatalf("expected correlation id in request context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	handler := RequestMetrics(metrics, logger, mux)
	req := httptest.NewRequest(http.MethodGet, "/widgets/123", nil)
	req.Header.Set(CorrelationIDHeader, "test-correlation-id")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get(CorrelationIDHeader) != "test-correlation-id" {
		t.Fatalf("expected %s response header to echo correlation id", CorrelationIDHeader)
	}
	if rec.Header().Get(RequestIDHeader) != "test-correlation-id" {
		t.Fatalf("expected %s response header to echo correlation id", RequestIDHeader)
	}
	if correlationID != "test-correlation-id" {
		t.Fatalf("expected correlation id in handler context, got %q", correlationID)
	}

	snapshot := metrics.Snapshot()
	requestsByPath, ok := snapshot["requests_by_path"].(map[string]int64)
	if !ok {
		t.Fatalf("expected requests_by_path to be a map[string]int64")
	}
	if requestsByPath["/widgets/{id}"] != 1 {
		t.Fatalf("expected route pattern to be used in metrics, got %#v", requestsByPath)
	}
}

func TestPrometheusMetricsHandlerRendersTextFormat(t *testing.T) {
	metrics := NewMetrics()
	metrics.RecordRequest("/health", http.MethodGet, http.StatusOK, 15)

	req := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	rec := httptest.NewRecorder()

	PrometheusMetricsHandler(metrics).ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusOK {
		t.Fatalf("expected status 200, got %d", got)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType == "" {
		t.Fatalf("expected content type to be set")
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatalf("expected prometheus body output")
	}
	if want := "resellution_http_requests_total"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}
