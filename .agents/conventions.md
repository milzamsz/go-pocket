# go-pocket conventions (agent quick rules)

## Go and errors

- Target Go `1.26.3`.
- Format with `gofmt -s`.
- Wrap errors with context: `fmt.Errorf("context: %w", err)`.

## templ and UI

- `.templ` files require `templ generate` after edits.
- Never hand-edit generated `*_templ.go`.
- Tailwind utility classes only in templ views.

## Tenancy signatures

- Org-scoped service methods accept `orgID` as explicit argument.
- New org-scoped collections must include `organization` relation.

## Routing

- Tenant routes live under `/org/:slug/*`.
- `/api/*` is reserved for PocketBase built-in API; do not add app routes there.

## Migrations

- Append-only; never edit a migration already merged to `main`.
- Add new migration files for schema evolution.

## Dependency and version policy

- Keep pinned versions from `AGENTS.md`.
- Version bumps go through dedicated PR review.

