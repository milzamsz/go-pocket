# go-pocket coding rules (extended)

- Do not edit generated artifacts:
  - `*_templ.go`
  - `assets/css/output.css`
  - `assets/css/sources.generated.css`
- Do not commit `.env*` (except `.env.example`) or `pb_data/`.
- Keep handlers thin; business logic belongs in services.
- Use dependency injection through `Deps`; avoid global mutable state.
- For user-visible changes, add one line in `CHANGELOG.md`.
- Before PR: run `task lint && task test` (or `task ci`).

