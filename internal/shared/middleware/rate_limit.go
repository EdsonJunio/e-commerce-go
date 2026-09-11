package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"e-commerce-go/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

type FixedWindowRateLimiter struct {
	mu          sync.Mutex
	entries     map[string]rateLimitEntry
	limit       int
	window      time.Duration
	now         func() time.Time
	lastCleanup time.Time
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		entries: make(map[string]rateLimitEntry),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

func (l *FixedWindowRateLimiter) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			c.Header("Retry-After", retryAfterSeconds(l.window))
			response.AbortWithError(c, http.StatusTooManyRequests, "rate_limited", "too many login attempts")
			return
		}
		c.Next()
	}
}

func (l *FixedWindowRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= l.window {
		for entryKey, candidate := range l.entries {
			if !now.Before(candidate.resetTime) {
				delete(l.entries, entryKey)
			}
		}
		l.lastCleanup = now
	}
	entry, exists := l.entries[key]
	if !exists || !now.Before(entry.resetTime) {
		l.entries[key] = rateLimitEntry{count: 1, resetTime: now.Add(l.window)}
		return true
	}
	if entry.count >= l.limit {
		return false
	}
	entry.count++
	l.entries[key] = entry
	return true
}

func retryAfterSeconds(window time.Duration) string {
	seconds := int64(window / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return strconv.FormatInt(seconds, 10)
}
