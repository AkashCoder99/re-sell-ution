package observability

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type requestMetricKey struct {
	Path       string
	Method     string
	StatusCode int
}

type latencyMetricKey struct {
	Path   string
	Method string
}

type Metrics struct {
	mu sync.RWMutex

	TotalRequests       int64
	RequestsByPath      map[string]int64
	RequestsByCode      map[int]int64
	RequestsByMethod    map[string]int64
	RequestsByRoute     map[requestMetricKey]int64
	DurationByRoute     map[latencyMetricKey]time.Duration
	RequestCountByRoute map[latencyMetricKey]int64
	TotalDuration       time.Duration
	RequestCount        int64 // for avg latency
}

func NewMetrics() *Metrics {
	return &Metrics{
		RequestsByPath:      make(map[string]int64),
		RequestsByCode:      make(map[int]int64),
		RequestsByMethod:    make(map[string]int64),
		RequestsByRoute:     make(map[requestMetricKey]int64),
		DurationByRoute:     make(map[latencyMetricKey]time.Duration),
		RequestCountByRoute: make(map[latencyMetricKey]int64),
	}
}

func (m *Metrics) RecordRequest(path, method string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	routeKey := requestMetricKey{Path: path, Method: method, StatusCode: statusCode}
	latencyKey := latencyMetricKey{Path: path, Method: method}

	m.TotalRequests++
	m.RequestsByPath[path]++
	m.RequestsByCode[statusCode]++
	m.RequestsByMethod[method]++
	m.RequestsByRoute[routeKey]++
	m.DurationByRoute[latencyKey] += duration
	m.RequestCountByRoute[latencyKey]++
	m.TotalDuration += duration
	m.RequestCount++
}

func (m *Metrics) Snapshot() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgLatencyMs := float64(0)
	if m.RequestCount > 0 {
		avgLatencyMs = float64(m.TotalDuration.Microseconds()) / float64(m.RequestCount) / 1000
	}

	return map[string]any{
		"total_requests":     m.TotalRequests,
		"requests_by_path":   copyMapStringInt64(m.RequestsByPath),
		"requests_by_code":   copyMapIntInt64(m.RequestsByCode),
		"requests_by_method": copyMapStringInt64(m.RequestsByMethod),
		"avg_latency_ms":     round(avgLatencyMs, 2),
		"uptime_seconds":     int64(time.Since(startTime).Seconds()),
		"routes":             m.routeSnapshots(),
		"latency_by_route":   m.latencySnapshots(),
	}
}

var startTime = time.Now()

func (m *Metrics) PrometheusSnapshot() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var builder strings.Builder
	builder.WriteString("# HELP resellution_http_requests_total Total HTTP requests processed.\n")
	builder.WriteString("# TYPE resellution_http_requests_total counter\n")
	builder.WriteString(fmt.Sprintf("resellution_http_requests_total %d\n", m.TotalRequests))

	builder.WriteString("# HELP resellution_http_requests_by_method_total Total HTTP requests by method.\n")
	builder.WriteString("# TYPE resellution_http_requests_by_method_total counter\n")
	methods := sortedStringKeys(m.RequestsByMethod)
	for _, method := range methods {
		builder.WriteString(fmt.Sprintf(
			"resellution_http_requests_by_method_total{method=\"%s\"} %d\n",
			prometheusLabelValue(method),
			m.RequestsByMethod[method],
		))
	}

	builder.WriteString("# HELP resellution_http_requests_by_status_total Total HTTP requests by status code.\n")
	builder.WriteString("# TYPE resellution_http_requests_by_status_total counter\n")
	statuses := sortedIntKeys(m.RequestsByCode)
	for _, status := range statuses {
		builder.WriteString(fmt.Sprintf(
			"resellution_http_requests_by_status_total{status=\"%s\"} %d\n",
			prometheusLabelValue(fmt.Sprintf("%d", status)),
			m.RequestsByCode[status],
		))
	}

	builder.WriteString("# HELP resellution_http_requests_by_route_total Total HTTP requests by route, method, and status.\n")
	builder.WriteString("# TYPE resellution_http_requests_by_route_total counter\n")
	requestRoutes := sortedRequestMetricKeys(m.RequestsByRoute)
	for _, key := range requestRoutes {
		builder.WriteString(fmt.Sprintf(
			"resellution_http_requests_by_route_total{path=\"%s\",method=\"%s\",status=\"%s\"} %d\n",
			prometheusLabelValue(key.Path),
			prometheusLabelValue(key.Method),
			prometheusLabelValue(fmt.Sprintf("%d", key.StatusCode)),
			m.RequestsByRoute[key],
		))
	}

	builder.WriteString("# HELP resellution_http_request_duration_ms_sum Total HTTP request duration in milliseconds by route and method.\n")
	builder.WriteString("# TYPE resellution_http_request_duration_ms_sum counter\n")
	builder.WriteString("# HELP resellution_http_request_duration_ms_count Total HTTP request count used for latency averages.\n")
	builder.WriteString("# TYPE resellution_http_request_duration_ms_count counter\n")
	builder.WriteString("# HELP resellution_http_request_duration_ms_avg Average HTTP request duration in milliseconds by route and method.\n")
	builder.WriteString("# TYPE resellution_http_request_duration_ms_avg gauge\n")
	latencyRoutes := sortedLatencyMetricKeys(m.DurationByRoute)
	for _, key := range latencyRoutes {
		totalMs := float64(m.DurationByRoute[key].Microseconds()) / 1000
		count := m.RequestCountByRoute[key]
		avgMs := 0.0
		if count > 0 {
			avgMs = totalMs / float64(count)
		}
		builder.WriteString(fmt.Sprintf(
			"resellution_http_request_duration_ms_sum{path=\"%s\",method=\"%s\"} %.2f\n",
			prometheusLabelValue(key.Path),
			prometheusLabelValue(key.Method),
			totalMs,
		))
		builder.WriteString(fmt.Sprintf(
			"resellution_http_request_duration_ms_count{path=\"%s\",method=\"%s\"} %d\n",
			prometheusLabelValue(key.Path),
			prometheusLabelValue(key.Method),
			count,
		))
		builder.WriteString(fmt.Sprintf(
			"resellution_http_request_duration_ms_avg{path=\"%s\",method=\"%s\"} %.2f\n",
			prometheusLabelValue(key.Path),
			prometheusLabelValue(key.Method),
			avgMs,
		))
	}

	builder.WriteString("# HELP resellution_http_uptime_seconds Backend process uptime in seconds.\n")
	builder.WriteString("# TYPE resellution_http_uptime_seconds gauge\n")
	builder.WriteString(fmt.Sprintf("resellution_http_uptime_seconds %d\n", int64(time.Since(startTime).Seconds())))

	return builder.String()
}

func copyMapStringInt64(m map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyMapIntInt64(m map[int]int64) map[int]int64 {
	out := make(map[int]int64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (m *Metrics) routeSnapshots() []map[string]any {
	keys := sortedRequestMetricKeys(m.RequestsByRoute)
	out := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		out = append(out, map[string]any{
			"path":     key.Path,
			"method":   key.Method,
			"status":   key.StatusCode,
			"requests": m.RequestsByRoute[key],
		})
	}
	return out
}

func (m *Metrics) latencySnapshots() []map[string]any {
	keys := sortedLatencyMetricKeys(m.DurationByRoute)
	out := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		count := m.RequestCountByRoute[key]
		avgLatencyMs := 0.0
		if count > 0 {
			avgLatencyMs = float64(m.DurationByRoute[key].Microseconds()) / float64(count) / 1000
		}
		out = append(out, map[string]any{
			"path":           key.Path,
			"method":         key.Method,
			"avg_latency_ms": round(avgLatencyMs, 2),
			"request_count":  count,
		})
	}
	return out
}

func sortedStringKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedIntKeys(m map[int]int64) []int {
	keys := make([]int, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

func sortedRequestMetricKeys(m map[requestMetricKey]int64) []requestMetricKey {
	keys := make([]requestMetricKey, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Path != keys[j].Path {
			return keys[i].Path < keys[j].Path
		}
		if keys[i].Method != keys[j].Method {
			return keys[i].Method < keys[j].Method
		}
		return keys[i].StatusCode < keys[j].StatusCode
	})
	return keys
}

func sortedLatencyMetricKeys(m map[latencyMetricKey]time.Duration) []latencyMetricKey {
	keys := make([]latencyMetricKey, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Path != keys[j].Path {
			return keys[i].Path < keys[j].Path
		}
		return keys[i].Method < keys[j].Method
	})
	return keys
}

func prometheusLabelValue(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, "\n", `\n`, `"`, `\"`)
	return replacer.Replace(value)
}

func round(f float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(f*pow+0.5)) / pow
}
