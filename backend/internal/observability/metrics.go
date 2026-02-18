package observability

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	TotalRequests   int64
	RequestsByPath map[string]int64
	RequestsByCode map[int]int64
	TotalDuration  time.Duration
	RequestCount   int64 // for avg latency
}

func NewMetrics() *Metrics {
	return &Metrics{
		RequestsByPath: make(map[string]int64),
		RequestsByCode: make(map[int]int64),
	}
}

func (m *Metrics) RecordRequest(path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.RequestsByPath[path]++
	m.RequestsByCode[statusCode]++
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
		"total_requests":   m.TotalRequests,
		"requests_by_path": copyMapInt64(m.RequestsByPath),
		"requests_by_code": copyMapInt(m.RequestsByCode),
		"avg_latency_ms":   round(avgLatencyMs, 2),
		"uptime_seconds":   int64(time.Since(startTime).Seconds()),
	}
}

var startTime = time.Now()

func copyMapInt64(m map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyMapInt(m map[int]int64) map[int]int64 {
	out := make(map[int]int64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func round(f float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(f*pow+0.5)) / pow
}
