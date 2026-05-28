# Contributing

## Setup

1. Install pinned tools from `AGENTS.md`.
2. `go mod download`
3. `cp .env.example .env`

## Development flow

1. Make focused changes.
2. Run:
   - `task templ:generate`
   - `task lint`
   - `task test`
   - `task build`
3. Add a `CHANGELOG.md` line for user-visible changes.

## Tenancy-critical checklist

- Org-scoped routes must be under `/org/:slug/*`.
- Service methods must enforce `organization = orgID` filtering.
- Update tenancy tests when behavior changes.

## Pull request checklist

- Title format: `<area>: <concise change>`
- Include migration reference if schema changed.
- Keep migrations append-only.
- Do not edit generated files manually.

