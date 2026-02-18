package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipEntry struct {
	count      int
	windowFrom time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*ipEntry
	limit    int
	window   time.Duration
}

func NewIPRateLimiter(limit int, windowMinutes int) *IPRateLimiter {
	if limit <= 0 {
		limit = 5
	}
	if windowMinutes <= 0 {
		windowMinutes = 60
	}
	return &IPRateLimiter{
		entries: make(map[string]*ipEntry),
		limit:   limit,
		window:  time.Duration(windowMinutes) * time.Minute,
	}
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.entries[ip]
	if !ok || now.Sub(e.windowFrom) >= l.window {
		l.entries[ip] = &ipEntry{count: 1, windowFrom: now}
		return true
	}
	if e.count >= l.limit {
		return false
	}
	e.count++
	return true
}

func (l *IPRateLimiter) RemainingMinutes(ip string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[ip]
	if !ok {
		return 0
	}
	elapsed := time.Since(e.windowFrom)
	if elapsed >= l.window {
		return 0
	}
	return int((l.window - elapsed).Minutes())
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func IPRateLimit(limiter *IPRateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !limiter.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			mins := limiter.RemainingMinutes(ip)
			if mins < 1 {
				mins = 1
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"Too many password reset requests. Try again in %d minutes"}`, mins)))
			return
		}
		next.ServeHTTP(w, r)
	}
}
