# Prompt: add-page

Add a new page with correct route boundary:

1. Public page: register under `/`.
2. Authenticated app page: register under `/app/*`.
3. Org page: register under `/org/:slug/*` and require org role middleware.

Implementation flow:

- Handler (thin)
- Service calls
- templ render
- tests

If route is `/org/:slug/*`, update `PLAN.md` route table section.

