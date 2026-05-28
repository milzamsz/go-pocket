# Deployment

## Prerequisites

- Docker engine (or Dokploy-managed Docker host)
- Environment values configured in platform secrets

## Local container run

1. Create env file: `cp .env.example .env`
2. Build and run: `docker compose up --build`
3. Open: `http://localhost:8090`

## Production Docker

1. Build image:
   - `docker build -t go-pocket:prod .`
2. Run with persistent volume:
   - `docker run -d --name go-pocket -p 8090:8090 --env-file .env -v go-pocket-data:/app/pb_data go-pocket:prod`

## Dokploy

1. Import repository.
2. Use `Dockerfile` build.
3. Mount persistent volume to `/app/pb_data`.
4. Set env variables from `.env.example` keys.
5. Configure health check path `/healthz` on port `8090`.

### Required Secrets For Current Core Scope

- `POLAR_ACCESS_TOKEN`
- `POLAR_WEBHOOK_SECRET`
- `RESEND_API_KEY`
- `RESEND_WEBHOOK_SECRET`
- `OAUTH_GOOGLE_CLIENT_ID`
- `OAUTH_GOOGLE_CLIENT_SECRET`
- `OAUTH_GITHUB_CLIENT_ID`
- `OAUTH_GITHUB_CLIENT_SECRET`

### Recommended Base URL

- Set both `APP_URL` and `APP_BASE_URL` to your public app URL (for OAuth callback and email links).

## Staging Smoke Checklist

1. Hit `GET /healthz` and confirm `200`.
2. Log in through password flow and confirm authenticated `GET /app/` returns `200`.
3. Trigger one invite flow (`POST /org/:slug/members/invite`) and confirm redirect + invitation record.
4. Verify OAuth redirect endpoints return redirects:
   - `GET /auth/oauth/google`
   - `GET /auth/oauth/github`
5. Send intentionally invalid webhook signatures and confirm `401`:
   - `POST /webhooks/polar`
   - `POST /webhooks/resend`
6. As superuser, confirm admin pages load:
   - `/admin/users`
   - `/admin/organizations`
   - `/admin/analytics`
   - `/admin/audit`

## Local Smoke Workflow

1. Start app locally: `go run . serve --http=0.0.0.0:8090`
2. In another terminal, seed deterministic local users: `go run . seed`
3. Run smoke checks: `go run . smoke-local`

The seed command prints final credentials and whether requested or fallback password was used.

## Backups

Back up the `/app/pb_data` volume on a schedule (hourly local, daily offsite recommended).
