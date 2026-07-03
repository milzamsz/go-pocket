package middleware

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// defaultContentSecurityPolicy is intentionally permissive enough to keep the
// self-hosted Alpine.js + HTMX stack working (Alpine evaluates expressions via
// Function(), which requires 'unsafe-eval', and templ emits some inline
// attributes/styles) while still constraining external origins.
const defaultContentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
	"img-src 'self' data: https:; " +
	"font-src 'self' data: https://fonts.gstatic.com; " +
	"connect-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

// SecurityHeaders sets a baseline of defensive HTTP response headers on every
// application response. PocketBase's own admin UI ("/_") and REST API ("/api/")
// are skipped so we don't interfere with their behavior.
func SecurityHeaders() func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		path := e.Request.URL.Path
		if strings.HasPrefix(path, "/_") || strings.HasPrefix(path, "/api/") {
			return e.Next()
		}

		header := e.Response.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("Cross-Origin-Opener-Policy", "same-origin")
		header.Set("X-XSS-Protection", "0")
		if header.Get("Content-Security-Policy") == "" {
			header.Set("Content-Security-Policy", defaultContentSecurityPolicy)
		}
		return e.Next()
	}
}
