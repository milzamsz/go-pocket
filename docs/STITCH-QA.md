# Stitch UI QA Baseline

Date: 2026-05-28

This checklist tracks visual parity between app routes and `docs/.stitch` references.

## Public Surfaces

- `/` aligned to `landing_page_the_ui_backbone_for_enterprise`
- `/blog` aligned to `technical_blog_the_system_log`
- `/blog/{slug}` aligned to `blog_article_scaling_with_technical_elegance`
- `/docs` and `/docs/{path...}` aligned to `documentation_portal`
- `/help` aligned to `help_center_documentation`

## Auth + App Surfaces

- `/auth/login` aligned to `login_authentication`
- `/app/onboarding` aligned to `welcome_onboarding`
- `/app/settings/profile|security|account` aligned to `settings`

## Org Surfaces

- `/org/:slug/` aligned to `dashboard_overview`
- `/org/:slug/products` aligned to `products_list`
- products create modal aligned to `create_new_product_modal`
- `/org/:slug/products/{id}` aligned to `product_details_neo_cyberpunk_widget`
- `/org/:slug/kanban` aligned to `kanban_board`
- `/org/:slug/members` aligned to `team_management`
- `/org/:slug/members/{userID}` aligned to `member_profile_sarah_connor`
- `/org/:slug/settings` aligned to `organization_settings`
- `/org/:slug/billing` aligned to `billing_subscription`

## Interaction Checks

- Theme toggle works on public, auth, app, and org shells.
- Product filters and modal workflow render with Stitch surfaces.
- Product detail and member profile back-navigation works.
- Kanban drag-drop works with local `/assets/js/sortable.min.js`.
- OAuth and auth forms preserve focus visibility and spacing.

## Notes

- Tailwind v4 scanning includes `.templ` and handler references in `assets/css/input.css`.
- Geist + Inter are loaded from Google Fonts in the shared base layout.
- Lucide-style iconography is implemented via local templ icon components in `components/layouts/icons.templ`.
