package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string][]time.Time
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		entries: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	return rl.AllowLimit(key, rl.limit)
}

func (rl *RateLimiter) AllowLimit(key string, limit int) bool {
	if limit <= 0 {
		return false
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	times := rl.entries[key]
	var valid []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		rl.entries[key] = valid
		return false
	}

	rl.entries[key] = append(valid, now)

	go rl.cleanup()
	return true
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window * 2)
	for key, times := range rl.entries {
		if len(times) > 0 && times[len(times)-1].Before(cutoff) {
			delete(rl.entries, key)
		}
	}
}

func RateLimitMiddleware(rl *RateLimiter, keyFn func(c echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !rl.Allow(keyFn(c)) {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"code":    429,
					"message": "请求过于频繁，请稍后重试",
				})
			}
			return next(c)
		}
	}
}
