package ratelimit

import (
	"testing"
)

func TestIPRateLimiter_Allow(t *testing.T) {
	limiter := NewIPRateLimiter(3, 60)
	ip := "192.168.1.1"

	// First 3 should be allowed
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("request %d: expected allow, got deny", i+1)
		}
	}
	// 4th should be denied
	if limiter.Allow(ip) {
		t.Error("request 4: expected deny, got allow")
	}
}
