# go-pocket architecture (agent summary)

## Layers

1. HTTP handlers (`internal/server/handlers/*`) parse input and render templ.
2. Services (`internal/services/*`) own all PocketBase queries and business logic.
3. PocketBase collections + rules enforce persistence and row-level constraints.

Handlers must not call DB methods directly. Add or extend a service instead.

## Tenant isolation model

All org-scoped resources require these three layers:

1. Route + middleware: endpoint under `/org/:slug/*` + `RequireOrgRole`.
2. Service filter: every org-scoped query includes `organization = orgID`.
3. PocketBase rule: collection rules constrained by `organization_members`.

Bypassing one layer is not acceptable without test-backed rationale.

## Billing and email boundaries

- Billing provider: Polar.sh only (`github.com/polarsource/polar-go`).
- Email provider: Resend via internal email service only (`resend-go/v3`).

Do not add Stripe or `net/smtp` usage.

## Assets and frontend runtime

- templ-generated files (`*_templ.go`) are generated artifacts.
- Alpine.js and HTMX are self-hosted in `/assets/js`.
- Tailwind output and generated CSS source files are generated artifacts.

## Deployment

- Single binary in multi-stage Docker build.
- Runtime keeps PocketBase DB at `/app/pb_data` volume.
- Environment is injected by Dokploy or host runtime, not committed files.

