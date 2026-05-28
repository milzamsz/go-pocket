# Prompt: add-feature

Implement a new feature in go-pocket with this order:

1. Define collection and migration changes.
2. Add/extend service methods (org-scoped methods accept `orgID`).
3. Add handlers under the proper route scope.
4. Add templ views/components as needed.
5. Add/update tests, especially tenancy isolation coverage.
6. Run `task templ:generate`, `task lint`, `task test`, `task build`.

Constraints:

- Billing provider is Polar only.
- Email is Resend via email service.
- Do not bypass service layer from handlers.

