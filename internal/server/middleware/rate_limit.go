package middleware

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// RateLimiter is a small, dependency-free fixed-window rate limiter keyed by
// client IP. It is intended for protecting a handful of sensitive endpoints
// (e.g. auth) against brute-force and abuse. For multi-instance deployments a
// shared store (Redis) would be required; this in-memory implementation is
// per-process.
type RateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	visitors map[string]*rateWindow
}

type rateWindow struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter creates a limiter allowing limit requests per window per IP.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		limit:    limit,
		window:   window,
		now:      time.Now,
		visitors: make(map[string]*rateWindow),
	}
}

// Middleware returns a PocketBase middleware that rejects requests exceeding the
// configured rate with HTTP 429.
func (rl *RateLimiter) Middleware() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if !rl.allow(clientIP(e)) {
			return e.TooManyRequestsError("rate limit exceeded, please retry shortly", nil)
		}
		return e.Next()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	rl.pruneLocked(now)

	entry, ok := rl.visitors[key]
	if !ok || now.After(entry.resetAt) {
		rl.visitors[key] = &rateWindow{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if entry.count >= rl.limit {
		return false
	}
	entry.count++
	return true
}

// pruneLocked drops expired windows. Caller must hold rl.mu.
func (rl *RateLimiter) pruneLocked(now time.Time) {
	for key, entry := range rl.visitors {
		if now.After(entry.resetAt) {
			delete(rl.visitors, key)
		}
	}
}

func clientIP(e *core.RequestEvent) string {
	req := e.Request
	if forwarded := strings.TrimSpace(req.Header.Get("X-Forwarded-For")); forwarded != "" {
		if first, _, found := strings.Cut(forwarded, ","); found {
			return strings.TrimSpace(first)
		}
		return forwarded
	}
	if realIP := strings.TrimSpace(req.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		return host
	}
	return req.RemoteAddr
}
