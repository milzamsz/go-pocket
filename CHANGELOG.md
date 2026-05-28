# Changelog

## Unreleased

- ui: start Technical Elegance Stitch port with new `/help`, org product detail/member profile routes, local SortableJS, and org-context-safe member loading
- devx: add production-ready operations scaffold (Taskfile, CI, Docker/Dokploy, env/ignore files, and contributor docs)
- auth: implement Google/GitHub OAuth start/callback flows with secure state cookie verification
- security: enforce authentication middleware on `/app/*` routes
- tenancy: implement invitation token accept/decline flows with repository + service coverage
- admin: replace placeholder admin pages with real superuser-backed users/orgs/analytics/audit data views
- stability: add integration coverage for invalid webhook signatures to ensure unauthorized requests do not mutate state
