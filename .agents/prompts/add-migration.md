# Prompt: add-migration

Create a new PocketBase migration without modifying existing migration files.

Checklist:

1. Create migration with `task migrate:create -- <name>`.
2. Add `up` and `down` logic.
3. Ensure org-scoped collections include `organization` field and rules.
4. Apply locally with `task migrate`.
5. Add tests covering behavior and tenancy safety.

