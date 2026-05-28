# Prompt: debug

Debug workflow for go-pocket:

1. Reproduce with exact route and org context.
2. Check app logs and request IDs.
3. Verify middleware path (`/app` vs `/org/:slug`).
4. Validate service-level org filters.
5. Confirm PocketBase collection rules.
6. Add regression test before finalizing fix.

Run `task test` before closing.

