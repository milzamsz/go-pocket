# AGENTS.md — go-pocket

> Multi-tenant Go SaaS boilerplate on PocketBase + templUI.
> Payments: **Polar.sh**. Email: **Resend**. Deploy: **Docker / Dokploy**.
> Read `.agents/architecture.md` and `.agents/conventions.md` before making non-trivial changes.

## Setup commands

- Install Go deps: `go mod download`
- Install templUI components (first time): `templui add "*"`
- Generate templ Go code: `templ generate`
- Run dev server (hot reload): `task dev`
- Run all tests: `task test`
- Lint: `task lint`
- Build production binary: `task build`
- Run pending migrations: `task migrate`
- Create new migration: `task migrate:create <name>`
- Seed dev data: `task seed`

## Required tools (pinned versions — May 2026 baseline)

- **Go 1.26.3** — download from https://go.dev/dl/
- **templ CLI v0.3.1020**: `go install github.com/a-h/templ/cmd/templ@v0.3.1020`
- **Task v3.49.1**: `go install github.com/go-task/task/v3/cmd/task@v3.49.1`
- **Tailwind CSS v4.3.0** standalone binary (download from `tailwindlabs/tailwindcss` releases)
- **templUI CLI v1.9.2**: `go install github.com/templui/templui/cmd/templui@v1.9.2`

## Pinned dependencies

| Package | Version |
|---|---|
| `github.com/pocketbase/pocketbase` | v0.38.2 |
| `github.com/a-h/templ` | v0.3.1020 |
| `github.com/polarsource/polar-go` | v0.7.0 |
| `github.com/resend/resend-go/v3` | v3.7.0 |

Alpine.js v3.15.8 and HTMX v2.0.x are loaded from `/assets/js/` (self-hosted, no CDN).

The Docker build uses `golang:1.26.3-alpine3.23` (builder) → `alpine:3.23.3` (runtime).

**Do not bump these versions without opening a PR** — Dependabot/Renovate handles patch bumps; minor/major requires a human review.

## Code style & conventions

- Go 1.26.3. `gofmt -s`. Errors wrapped with `fmt.Errorf("context: %w", err)`.
- Templ files end in `.templ`; ALWAYS run `templ generate` after editing one. The generated `*_templ.go` files must NOT be edited by hand.
- Tailwind utility classes only — no custom CSS in `.templ` files. Use theme tokens (`bg-background`, `text-foreground`, `text-muted-foreground`, etc.).
- Handlers under `internal/server/handlers/` are thin: parse request → call services → render templ. NO direct DB access.
- Services under `internal/services/` own all PocketBase interactions. Every repository method accepts `orgID` as a non-context argument when the resource is org-scoped.
- Migrations are append-only and immutable. NEVER modify a migration merged to `main`. Add a new one instead.
- One templUI component = one folder under `components/ui/<name>/` containing `<name>.templ`, optional `<name>.go` for props, and optional `<name>.min.js` for interactive scripts.

## Multi-tenancy rules (critical — read before any data change)

- Every business collection has an `organization` relation. Service-layer queries MUST filter by `organization = ?`.
- Endpoints handling tenant data MUST live under `/org/:slug/*` and use the `RequireOrgRole` middleware.
- New collections require: (1) `organization` field, (2) PocketBase rules that join through `organization_members`, (3) service-layer methods that take `orgID`.
- The permission matrix is in `internal/services/tenancy/permissions.go`. Update it (with a test) when adding a new permission.
- The triple isolation layer (service query filter + PB rule + middleware) is mandatory. If you find yourself bypassing any layer, write a test that demonstrates why, then update the matrix.

## Testing instructions

- Run all tests: `task test`
- Run a single package: `go test -race -cover ./internal/services/tenancy/...`
- Multi-tenant isolation tests in `internal/services/tenancy/*_test.go` MUST pass before any tenancy-touching PR is merged.
- Use `testutil.NewTestApp()` for integration tests — it spins up a fresh in-memory PocketBase instance per test.
- Add or update tests for any code change. Aim for ≥70% line coverage.

## Pull request guidelines

- Title format: `<area>: <concise change>` (e.g., `billing: handle Polar trial_ends_at`, `tenancy: enforce seat limit on invite`).
- Run `task lint && task test` before committing.
- Include a one-line `CHANGELOG.md` entry for user-visible changes.
- Reference the migration number for any schema change in the PR description.
- New `/org/:slug/*` route? Add a row to the route table in `PLAN.md` §11.

## Things NOT to do

- Don't import `stripe-go`. Billing is **Polar.sh** — use `github.com/polarsource/polar-go`.
- Don't use `net/smtp` or PocketBase's built-in mailer for outbound email. Use the `email` service which wraps `github.com/resend/resend-go/v3`.
- Don't add new global state. Pass services via the `Deps` struct.
- Don't bypass the service layer. If you find yourself reaching for `app.Dao()` from a handler, write a service method instead.
- Don't introduce a JavaScript framework. We have Alpine.js + HTMX + per-component templUI scripts. That's the stack.
- Don't edit generated files: `*_templ.go`, `assets/css/output.css`, `assets/css/sources.generated.css`.
- Don't commit `.env` files or anything in `pb_data/`.
- Don't ship a route under `/api/*` — that namespace is owned by PocketBase's built-in REST API.

## Deployment

- We deploy via **Dokploy** (open-source PaaS) or plain Docker.
- The `Dockerfile` is a multi-stage build producing a single static binary.
- Production environment variables live in Dokploy's encrypted env store, NOT in committed files.
- Database lives on a Docker volume mounted at `/app/pb_data`. Backups: hourly local + daily S3 via Dokploy's volume backup feature.

## See also

- `.agents/architecture.md` — full architecture summary for agent context
- `.agents/conventions.md` — file naming, error patterns, templ patterns
- `.agents/rules.md` — coding style deep-dive
- `.agents/prompts/` — reusable prompt templates for common tasks (add-feature, add-migration, add-templui-component, add-page, debug)
- `PLAN.md` — full project blueprint
- `docs/` — long-form developer docs and ADRs
