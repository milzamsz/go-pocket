# Changelog

## Unreleased

- auth: implement profile update, password change (with current-password verification), and two-factor toggle which were previously silent no-ops; add `users.two_factor_enabled` migration
- security: hash password-reset and email-verification tokens (SHA-256) at rest instead of storing them in plaintext
- billing: replace the non-standard Polar signature check with Standard Webhooks verification (`webhook-id`/`webhook-timestamp`/`webhook-signature`, base64 HMAC, `whsec_` secret)
- security: reject unsigned Polar/Resend webhooks in production when no webhook secret is configured (fail closed)
- security: add baseline security-headers middleware and per-IP rate limiting on auth endpoints
- ops: log configuration warnings at boot for missing production secrets and non-HTTPS base URLs
- ui: unify the design-token system — `.stitch-*` component classes now derive from theme CSS variables (removing a second hardcoded gray palette), and the public/auth and dashboard shells use `primary`/`border`/`card` tokens instead of hardcoded `indigo`/`slate`/hex
- ui: make the dashboard sidebar responsive (off-canvas mobile drawer + hamburger) and remove the hardcoded/contradictory "Free Trial"/"Enterprise Plan" status in favor of the real section label
- a11y: add aria-current to active nav, aria-labels to icon controls, a labeled search field, and visible focus rings across the shells
- docs: add ADR-0005 (Standard Webhooks verification) and docs/UI-AUDIT.md
- ui: start Technical Elegance Stitch port with new `/help`, org product detail/member profile routes, local SortableJS, and org-context-safe member loading
- devx: add production-ready operations scaffold (Taskfile, CI, Docker/Dokploy, env/ignore files, and contributor docs)
- auth: implement Google/GitHub OAuth start/callback flows with secure state cookie verification
- security: enforce authentication middleware on `/app/*` routes
- tenancy: implement invitation token accept/decline flows with repository + service coverage
- admin: replace placeholder admin pages with real superuser-backed users/orgs/analytics/audit data views
- stability: add integration coverage for invalid webhook signatures to ensure unauthorized requests do not mutate state
