package middleware

import (
	"testing"
	"time"
)

func TestFixedWindowRateLimiter(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	limiter := NewFixedWindowRateLimiter(2, time.Minute)
	limiter.now = func() time.Time { return now }

	if !limiter.allow("client") || !limiter.allow("client") {
		t.Fatal("requests within the limit must be allowed")
	}
	if limiter.allow("client") {
		t.Fatal("request above the limit must be rejected")
	}
	if !limiter.allow("other-client") {
		t.Fatal("limits must be isolated by client")
	}

	now = now.Add(time.Minute)
	if !limiter.allow("client") {
		t.Fatal("request must be allowed after the window resets")
	}
}
