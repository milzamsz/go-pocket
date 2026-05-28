# go-pocket

Multi-tenant Go SaaS boilerplate on PocketBase + templUI.

## Stack

- Go `1.26.3`
- PocketBase `v0.38.2`
- templ `v0.3.1020`
- Polar Go SDK `v0.7.0`
- Resend Go SDK `v3.7.0`
- Alpine.js `v3.15.8` + HTMX `v2.0.x` (self-hosted)

## Quick Start

1. Copy env:
   - `cp .env.example .env`
2. Install tools:
   - `go install github.com/a-h/templ/cmd/templ@v0.3.1020`
   - `go install github.com/go-task/task/v3/cmd/task@v3.49.1`
3. Install dependencies:
   - `task deps`
4. Run app:
   - `task dev`

## Daily Commands

- `task templ:generate` - regenerate templ output
- `task lint` - gofmt + go vet
- `task test` - race + coverage tests
- `task build` - build production binary
- `task ci` - local CI parity
- `task seed` - create deterministic local test users
- `task smoke:local` - run HTTP smoke checks against a running local server

## Local Deploy Smoke (Go Run)

1. Start server: `task dev`
2. Seed users: `task seed`
3. Run smoke: `task smoke:local`
4. Verify CSS asset: open `http://127.0.0.1:8090/assets/css/output.css` (must return `200` and non-empty CSS)
5. Verify pages: open `/` and `/app/` and confirm themed styles are applied (not browser-default flat HTML)

`task css:build` now checks, in order: `TAILWINDCSS_BIN`, `tailwindcss` in `PATH`, then local `./tailwindcss`, `./bin/tailwindcss`, `./tools/tailwindcss` (and `.exe` variants on Windows).
If none are present, install Tailwind CLI `v4.3.0` and rerun `task css:build`.

Seed output always prints final effective credentials (requested or fallback):
- superuser admin: `admin@test.com`
- regular user: `user@test.com` (alias also seeded: `use@test.com`)

## Auth Routes (Core)

- `POST /auth/login`, `POST /auth/signup`, `GET /auth/logout`
- `POST /auth/forgot-password`, `POST /auth/reset-password`, `GET /auth/verify-email`
- `GET /auth/oauth/{provider}`, `GET /auth/oauth/{provider}/callback` (`provider`: `google`, `github`)
- OAuth provider credentials are read from:
  - `OAUTH_GOOGLE_CLIENT_ID`, `OAUTH_GOOGLE_CLIENT_SECRET`
  - `OAUTH_GITHUB_CLIENT_ID`, `OAUTH_GITHUB_CLIENT_SECRET`

## Invite Token Flow

- `POST /invite/{token}` accepts an invitation for the authenticated user.
- `POST /invite/{token}/decline` declines and revokes an invitation.

## Admin Routes (Superuser)

- `GET /admin/`
- `GET /admin/users`, `GET /admin/users/{id}`
- `GET /admin/organizations`, `GET /admin/organizations/{id}`
- `GET /admin/analytics`
- `GET /admin/audit`

## CI Gates

CI enforces:

1. `templ generate` with zero uncommitted diff
2. lint
3. tests
4. build

## Deployment

- Docker: [Dockerfile](/C:/Projects/Personal/TEMPLATE/go-pocket/Dockerfile)
- Compose: [docker-compose.yml](/C:/Projects/Personal/TEMPLATE/go-pocket/docker-compose.yml)
- Dokploy: [dokploy.yaml](/C:/Projects/Personal/TEMPLATE/go-pocket/dokploy.yaml)
- Deployment guide: [docs/DEPLOYMENT.md](/C:/Projects/Personal/TEMPLATE/go-pocket/docs/DEPLOYMENT.md)

## Contributor Docs

- [docs/CONTRIBUTING.md](/C:/Projects/Personal/TEMPLATE/go-pocket/docs/CONTRIBUTING.md)
- [docs/ARCHITECTURE.md](/C:/Projects/Personal/TEMPLATE/go-pocket/docs/ARCHITECTURE.md)
- [.agents/architecture.md](/C:/Projects/Personal/TEMPLATE/go-pocket/.agents/architecture.md)
- [.agents/conventions.md](/C:/Projects/Personal/TEMPLATE/go-pocket/.agents/conventions.md)

## First Publish

From a fresh local clone/worktree, publish with:

```bash
git init
git branch -M main
git add .
git commit -m "initial go-pocket saas starter"
git remote add origin git@github.com:milzamsz/go-pocket.git
git push -u origin main
```
