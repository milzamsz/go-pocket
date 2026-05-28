# Architecture

go-pocket is a single-binary Go SaaS starter built on PocketBase and templUI.

## Runtime shape

- HTTP and auth lifecycle are handled through PocketBase's embedded server.
- App routes are split by scope:
  - public (`/`)
  - authenticated app (`/app/*`)
  - org-scoped (`/org/:slug/*`)
- Services own data access and business logic.

## Tenant isolation

Every org-scoped resource uses:

1. route middleware checks,
2. service query filtering by `orgID`,
3. PocketBase rule constraints.

All three are required.

## External integrations

- Billing: Polar (`polarsource/polar-go`)
- Email: Resend (`resend-go/v3`)

## Build + deploy

- Multi-stage Docker image.
- Runtime persists PocketBase data at `/app/pb_data` volume.
- Deploy on Dokploy or any Docker host.

