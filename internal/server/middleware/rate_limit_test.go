package middleware

import (
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_AllowsUpToLimitThenRejects(t *testing.T) {
	limiter := NewRateLimiter(2, time.Hour)
	mw := limiter.Middleware()

	require.NoError(t, mw(newRequestEvent(t, "/auth/login")))
	require.NoError(t, mw(newRequestEvent(t, "/auth/login")))

	err := mw(newRequestEvent(t, "/auth/login"))
	apiErr := &router.ApiError{}
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, 429, apiErr.Status)
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	current := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return current }
	mw := limiter.Middleware()

	require.NoError(t, mw(newRequestEvent(t, "/auth/login")))
	require.Error(t, mw(newRequestEvent(t, "/auth/login")))

	current = current.Add(2 * time.Minute)
	require.NoError(t, mw(newRequestEvent(t, "/auth/login")))
}

func TestRateLimiter_SeparatesDistinctClients(t *testing.T) {
	limiter := NewRateLimiter(1, time.Hour)
	mw := limiter.Middleware()

	first := newRequestEvent(t, "/auth/login")
	first.Request.Header.Set("X-Forwarded-For", "203.0.113.1")
	require.NoError(t, mw(first))

	second := newRequestEvent(t, "/auth/login")
	second.Request.Header.Set("X-Forwarded-For", "203.0.113.2")
	require.NoError(t, mw(second))

	third := newRequestEvent(t, "/auth/login")
	third.Request.Header.Set("X-Forwarded-For", "203.0.113.1")
	require.Error(t, mw(third))
}
