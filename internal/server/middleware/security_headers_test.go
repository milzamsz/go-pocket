package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders_SetsBaselineHeaders(t *testing.T) {
	e := newRequestEvent(t, "/app")

	err := SecurityHeaders()(e)
	require.NoError(t, err)

	header := e.Response.Header()
	require.Equal(t, "nosniff", header.Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", header.Get("X-Frame-Options"))
	require.Equal(t, "strict-origin-when-cross-origin", header.Get("Referrer-Policy"))
	require.NotEmpty(t, header.Get("Content-Security-Policy"))
}

func TestSecurityHeaders_SkipsPocketBaseRoutes(t *testing.T) {
	e := newRequestEvent(t, "/api/collections/users/records")

	err := SecurityHeaders()(e)
	require.NoError(t, err)

	require.Empty(t, e.Response.Header().Get("Content-Security-Policy"))
	require.Empty(t, e.Response.Header().Get("X-Content-Type-Options"))
}

func TestSecurityHeaders_DoesNotOverrideExistingCSP(t *testing.T) {
	e := newRequestEvent(t, "/app")
	e.Response.Header().Set("Content-Security-Policy", "default-src 'none'")

	err := SecurityHeaders()(e)
	require.NoError(t, err)

	require.Equal(t, "default-src 'none'", e.Response.Header().Get("Content-Security-Policy"))
}
