# go-pocket-dashboard — Admin Dashboard Starter Blueprint

> A production-ready open-source **admin dashboard starter** built with **Go + PocketBase + templUI + HTMX/Alpine**, mirroring the page set, feature set, and developer ergonomics of [Kiranism/next-shadcn-dashboard-starter](https://github.com/Kiranism/next-shadcn-dashboard-starter) — translated from Next.js 16 / React 19 / shadcn-ui to a server-rendered Go stack.

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Kiranism → go-pocket Mapping (1:1)](#2-kiranism--go-pocket-mapping-11)
3. [Design Philosophy](#3-design-philosophy)
4. [Technology Stack (pinned)](#4-technology-stack-pinned)
5. [High-Level Architecture](#5-high-level-architecture)
6. [Feature-Based Folder Structure](#6-feature-based-folder-structure)
7. [Pages Catalog (mirrors Kiranism's pages)](#7-pages-catalog-mirrors-kiranisms-pages)
8. [Authentication Strategy](#8-authentication-strategy)
9. [Layout System (Sidebar / Topbar / Content)](#9-layout-system-sidebar--topbar--content)
10. [UI Component Library (templUI port)](#10-ui-component-library-templui-port)
11. [Data Tables (TanStack-style, server-driven)](#11-data-tables-tanstack-style-server-driven)
12. [Forms (React-Hook-Form + Zod equivalent)](#12-forms-react-hook-form--zod-equivalent)
13. [Type-Safe Search Params (Nuqs equivalent)](#13-type-safe-search-params-nuqs-equivalent)
14. [State Management (Zustand equivalent)](#14-state-management-zustand-equivalent)
15. [Command Palette (kbar equivalent)](#15-command-palette-kbar-equivalent)
16. [Charts & Analytics (Recharts equivalent)](#16-charts--analytics-recharts-equivalent)
17. [Kanban Board (dnd-kit equivalent)](#17-kanban-board-dnd-kit-equivalent)
18. [Multi-Tenant Workspaces & RBAC](#18-multi-tenant-workspaces--rbac)
19. [Billing & Feature Gating](#19-billing--feature-gating)
20. [Theme System (6+ themes, tweakcn equivalent)](#20-theme-system-6-themes-tweakcn-equivalent)
21. [Error Tracking (Sentry)](#21-error-tracking-sentry)
22. [Infobar Component](#22-infobar-component)
23. [Configuration & Environment](#23-configuration--environment)
24. [Development Workflow](#24-development-workflow)
25. [Build, Asset Pipeline & Embed](#25-build-asset-pipeline--embed)
26. [Deployment (Docker / Dokploy)](#26-deployment-docker--dokploy)
27. [Observability, Security & Performance](#27-observability-security--performance)
28. [Testing Strategy](#28-testing-strategy)
29. [Phase-by-Phase Implementation Roadmap](#29-phase-by-phase-implementation-roadmap)
30. [AI-Assisted Development (AGENTS.md & .agents/)](#30-ai-assisted-development-agentsmd--agents)
31. [Cleanup Guide (remove demo data)](#31-cleanup-guide-remove-demo-data)
32. [Appendix: Key Code Snippets](#32-appendix-key-code-snippets)

---

## 1. Executive Summary

**go-pocket-dashboard** is a Go-native answer to the most popular open-source Next.js admin starter. It replaces the entire JavaScript-first toolchain (React 19, Clerk, TanStack Table, React Hook Form, Zustand, Nuqs, kbar, Recharts, dnd-kit) with **server-rendered equivalents** built on Go 1.26.3, PocketBase v0.38.2, templ v0.3.1020, templUI v1.9.2, Tailwind CSS v4.3.0, HTMX v2.0.x, and Alpine.js v3.15.8.

Every page in Kiranism's starter has a 1:1 counterpart here. Every feature (RBAC navigation, plan-gated routes, multi-tenant workspaces, type-safe search params, command palette, kanban DnD, six themes) is reproduced with **zero JavaScript framework runtime** — just templ-rendered HTML, HTMX swaps, and minimal Alpine for component-local interactivity.

### What you get out of the box

- One Go binary, one Docker image. Deploy via Dokploy or any Docker host.
- A **feature-based folder structure** (matching Kiranism's `src/features/<feature>/` convention) translated to `internal/features/<feature>/`.
- Pre-built admin layout (collapsible sidebar, sticky topbar, breadcrumbs, content area).
- Analytics overview with cards + charts (Chart.js via templUI).
- TanStack-style data tables with **server-side** search, filter, sort, pagination — backed by HTMX partial swaps and type-safe URL params.
- React-Hook-Form-style forms with `go-playground/validator` schemas, inline error rendering, and HTMX submission.
- Multi-tenant workspaces with role-based navigation filtering.
- Billing + plan-gated routes (Polar.sh).
- Six themes shipped (Default, Slate, Stone, Zinc, Neutral, Rose) via OKLCH CSS variables, runtime-switchable.
- Command palette (⌘K) with fuzzy search, keyboard nav, action dispatch.
- Kanban board with drag-and-drop (SortableJS + HTMX) and server persistence.
- Sentry error tracking, infobar component, not-found and global-error pages.
- `AGENTS.md` + `.agents/` for AI coding agents from day one.

### Who it's for

Backend-leaning teams who love shadcn-quality dashboards but don't want to maintain a Next.js app, a CI/CD lane for Node, a separate auth provider, or three different bundlers. **One binary. One language. Same polish.**

### Non-goals

- Not a port of Clerk. We integrate Clerk optionally; the default auth is PocketBase native.
- Not a Next.js-API-compatibility shim. We translate concepts (parallel routes → HTMX hx-swap-oob; server actions → form posts), we don't emulate.
- Not a SPA. The dashboard is server-rendered HTML with strategic islands of interactivity.

---

## 2. Kiranism → go-pocket Mapping (1:1)

The following table is the contract for "1:1." Every row maps a Kiranism concept to a go-pocket-dashboard equivalent.

| Kiranism (Next.js + React) | go-pocket-dashboard (Go + templUI) | Notes |
|---|---|---|
| Next.js 16 App Router | PocketBase + Echo router (`se.Router`) | Single binary; routes registered in `internal/server/routes.go`. |
| React 19 + TypeScript | Go 1.26.3 + templ v0.3.1020 (type-safe templates) | Server-rendered HTML, no hydration. |
| shadcn/ui | templUI v1.9.2 (40+ components) | Same component set; same OKLCH theme tokens. |
| Tailwind CSS v4 | Tailwind CSS v4.3.0 (standalone CLI) | Identical. |
| Clerk (auth + orgs + billing) | PocketBase auth + custom orgs/billing | Clerk is **optional** — see §8 for both paths. |
| Clerk `<OrganizationList />` | Custom workspaces page using templUI's Card/Table | Same UX. |
| Clerk `<OrganizationProfile />` | Custom team management page | Members, invitations, settings tabs. |
| Clerk `<PricingTable />` + Clerk Billing | Polar.sh integration | Per-workspace subscriptions; webhook → PB collections. |
| Clerk `<Protect>` (plan-gating) | `middleware.RequirePlan("pro")` | Server-side check + fallback render. |
| TanStack Table | Custom Go data-table service + templUI Table component | Server-side everything. |
| Dice table | Same — folded into the data-table service. |
| React Hook Form + Zod | templ forms + `go-playground/validator/v10` | Field-level errors rendered inline via HTMX. |
| Zustand (client state) | Alpine.js stores + server-persisted cookies for cross-request state | Plus PocketBase realtime for live data. |
| Nuqs (search params) | `internal/searchparams/` package — type-safe URL state | Parse/encode helpers per page. |
| kbar (cmd+k) | Custom Alpine + HTMX command palette using templUI Dialog | Keyboard nav, fuzzy search, action dispatch. |
| Recharts | Chart.js v4 via templUI's Charts component | Same chart types (Area, Bar, Line, Pie). |
| dnd-kit | SortableJS + HTMX `hx-post` on drop | Server persists order. |
| Sentry SDK for Next.js | `getsentry/sentry-go` v0.32.0 | Same DSN model. |
| ESLint + Prettier + Husky | `golangci-lint` + `gofmt` + `pre-commit` | Same intent. |
| Next.js `loading.tsx` | templUI Skeleton + HTMX `hx-indicator` | Suspense-like UX. |
| Next.js `error.tsx` | Echo error handler + dedicated error templ | Sentry hooks the same way. |
| Parallel routes | HTMX `hx-swap-oob` for independent fragment loads | Each card/chart loads in parallel via lazy GETs. |
| Server Actions | Plain `POST /...` Echo handlers | Submit + redirect or return updated fragment. |
| tweakcn (theme presets) | Six predefined OKLCH variable sets in `assets/css/themes/*.css` | Same six themes by name. |
| Clerk Organizations RBAC | `organization_members.role` + permission matrix | Client-side nav filtering driven by server-rendered hint attrs. |
| Bun | Go modules + Task | No JS package manager in production. |
| Vercel deploy | Dokploy + Docker | Same git-push-to-deploy DX. |

This table is the **single source of truth** for parity. Any new feature added upstream gets mapped here first.

---

## 3. Design Philosophy

Six principles, drawn from templUI's shadcn-inspired ethos, PocketBase's minimalism, and the lessons of Kiranism's starter:

1. **You own the code.** Every component, every handler, every collection schema lives in your repo. No hidden magic, no vendor lock-in.
2. **One binary to rule them all.** Embedded PocketBase, embedded assets, embedded migrations, embedded content. `go build` produces a single 25-35 MB executable.
3. **Server-first, JS-second.** templ renders on the server. Alpine.js handles component-local interactivity (≤5 lines per island). HTMX handles partial swaps. JavaScript is a progressive enhancement.
4. **Type safety where it matters.** templ catches HTML/Go type errors at compile time. `go-playground/validator` catches form schema errors at boot. Search-param structs catch URL contract drift at compile time.
5. **CSP-strict by default.** templUI ships without `eval`, without inline scripts. Every JS file is hashed and served from `/assets/js/`. Inline init scripts carry a nonce.
6. **Feature-based, not layer-based.** Like Kiranism's `src/features/<feature>/`, we use `internal/features/<feature>/{handlers,services,components,schemas,utils}`. New features land in one folder, not seven.

---

## 4. Technology Stack (pinned)

### Go modules

| Package | Version | Purpose |
|---|---|---|
| `github.com/pocketbase/pocketbase` | v0.38.2 | Embedded backend: auth, DB, storage, realtime, admin UI. |
| `github.com/a-h/templ` | v0.3.1020 | Type-safe HTML templates. |
| `github.com/polarsource/polar-go` | v0.7.0 | Billing (workspace subscriptions, plan-gating). |
| `github.com/resend/resend-go/v3` | v3.7.0 | Transactional email (invitations, reset, etc.). |
| `github.com/go-playground/validator/v10` | v10.24.0 | Form schema validation (Zod equivalent). |
| `github.com/getsentry/sentry-go` | v0.32.0 | Error tracking. |
| `github.com/nicksnyder/go-i18n/v2` | v2.6.0 | Optional i18n. |
| `github.com/spf13/pflag` | v1.0.6 | CLI flag parsing. |
| `github.com/stretchr/testify` | v1.10.0 | Testing helpers. |

Tooling versions (installed separately):

| Tool | Version |
|---|---|
| Go | **1.26.3** |
| templ CLI | v0.3.1020 |
| Task | v3.49.1 |
| Tailwind CSS standalone | v4.3.0 |
| templUI CLI | v1.9.2 |
| Alpine.js (browser) | v3.15.8 |
| HTMX (browser) | v2.0.x |
| Chart.js (browser) | v4.4.x |
| SortableJS (browser) | v1.15.x |
| Docker builder | `golang:1.26.3-alpine3.23` |
| Docker runtime | `alpine:3.23.3` |

### Why this stack vs. Kiranism's

| Concern | Kiranism (Node.js) | go-pocket-dashboard | Winner |
|---|---|---|---|
| Cold-start latency | ~200ms (Vercel serverless) | ~10ms (warm binary) | Go |
| Memory footprint | 100-200 MB per worker | 30-60 MB total | Go |
| Build complexity | Bun, Next, ESLint, Prettier, Husky | Go + Task + Tailwind CLI | Go |
| Type safety (templates) | TSX (good, but client-only) | templ (compile-time HTML safety) | Go |
| Ecosystem of UI primitives | shadcn (vast) | templUI v1.9.2 (40+, growing) | Tie — same shape |
| Auth ergonomics | Clerk (best-in-class UI, paid) | PocketBase + UI we own | Tie — different trade-offs |
| Single-binary deploy | No | Yes | Go |

---

## 5. High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                Single Go Binary (go-pocket-dashboard)                │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │                HTTP Layer (Echo via PocketBase)               │   │
│  │                                                               │   │
│  │  /auth/*           /dashboard/*         /api/*       /_/      │   │
│  │  (login/signup)    (admin app)          (PB REST)    (PB UI)  │   │
│  │  /webhooks/*       /dashboard/exclusive (plan-gated)          │   │
│  └─────────────┬──────────────────────────────────┬──────────────┘   │
│                │                                   │                 │
│  ┌─────────────▼─────────────────┐   ┌────────────▼──────────────┐   │
│  │     Feature Modules           │   │   PocketBase Core         │   │
│  │     internal/features/        │   │   - Collections           │   │
│  │     - overview/               │   │   - Event hooks           │   │
│  │     - products/               │   │   - Auth (email/OAuth)    │   │
│  │     - kanban/                 │   │   - Realtime SSE          │   │
│  │     - workspaces/             │   │   - File storage          │   │
│  │     - billing/                │   │                           │   │
│  │     - profile/                │   │                           │   │
│  │     - exclusive/              │   │                           │   │
│  └─────────────┬─────────────────┘   └────────────┬──────────────┘   │
│                │                                   │                 │
│  ┌─────────────▼───────────────────────────────────▼─────────────┐   │
│  │           templ Component Library (components/)               │   │
│  │   ui/ (40+ templUI ports)  layout/ (sidebar, topbar, ...)    │   │
│  │   pages/ (full pages)      icons/ (Lucide-via-templ)         │   │
│  └─────────────────────────────┬─────────────────────────────────┘   │
│                                │                                     │
│  ┌─────────────────────────────▼─────────────────────────────────┐   │
│  │              Embedded Assets (//go:embed)                     │   │
│  │   CSS · JS · fonts · images · email templates · migrations   │   │
│  └─────────────────────────────┬─────────────────────────────────┘   │
│                                │                                     │
│  ┌─────────────────────────────▼─────────────────────────────────┐   │
│  │              SQLite (pb_data/data.db + auxiliary.db)          │   │
│  └───────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼────────────────┐
            ▼               ▼                ▼
        Polar.sh API    Resend API       Sentry
        (billing)       (email)          (error tracking)
```

### Request lifecycle (a typical dashboard render)

1. Browser hits `https://dash.example.com/dashboard/products?q=widget&page=2&sort=-created`.
2. PocketBase's Echo router matches `/dashboard/products` to `features/products/handlers.List`.
3. Middleware chain: request ID → logger → auth (cookie → user record) → workspace resolver → RBAC (`RequireRole("member")`).
4. Handler parses search params via `searchparams.Parse[ProductListParams](r)` — type-safe URL state.
5. Handler calls `services.Products.List(ctx, workspaceID, params)` which returns `([]Product, totalCount, error)`.
6. Handler renders `pages.ProductsList(viewModel)` which composes `components/layout/dashboard.templ` + `features/products/components/table.templ` (templUI Table inside).
7. If the request is `HX-Request: true`, only the table fragment is streamed; otherwise the full page renders.

### Parallel route equivalent

Kiranism uses Next.js parallel routes to load each Overview card independently. We achieve identical UX with **HTMX lazy GETs**:

```html
<!-- Overview page renders skeleton placeholders that immediately fetch their content -->
<div hx-get="/dashboard/_partials/overview/revenue-card" hx-trigger="load" hx-swap="outerHTML">
  <Skeleton class="h-32" />
</div>
<div hx-get="/dashboard/_partials/overview/sales-card" hx-trigger="load" hx-swap="outerHTML">
  <Skeleton class="h-32" />
</div>
<!-- ...four such cards load in parallel; each has its own error boundary handler -->
```

Result: independent loading, error handling, and rendering — exactly what parallel routes deliver, no framework needed.

---

## 6. Feature-Based Folder Structure

```
go-pocket-dashboard/
├── main.go
├── go.mod                              # module github.com/milzamsz/go-pocket
├── go.sum
├── Taskfile.yml
├── .env.example
├── .gitignore
├── .dockerignore
├── Dockerfile
├── docker-compose.yml
├── LICENSE
├── README.md
├── CHANGELOG.md
├── AGENTS.md
├── .agents/
│   ├── rules.md
│   ├── architecture.md
│   ├── conventions.md
│   ├── prompts/
│   │   ├── add-feature.md
│   │   ├── add-page.md
│   │   ├── add-data-table.md
│   │   ├── add-form.md
│   │   ├── add-chart.md
│   │   └── add-command-action.md
│   └── tools/
│       ├── claude.md                   # symlink → AGENTS.md
│       ├── cursor.mdc                  # symlink → AGENTS.md
│       ├── aider.yml
│       └── gemini.json
│
├── cmd/
│   └── tools/
│       └── seed/main.go                # Seed demo data (products, kanban, etc.)
│
├── internal/
│   ├── app/
│   │   ├── app.go                      # Bootstrap PocketBase + services
│   │   └── lifecycle.go
│   ├── config/
│   │   └── config.go
│   │
│   ├── server/
│   │   ├── routes.go                   # Top-level route registration
│   │   ├── errors.go                   # Global error handler (Sentry-aware)
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── workspace.go            # Resolves active workspace
│   │   │   ├── rbac.go                 # RequireRole / RequirePermission
│   │   │   ├── plan.go                 # RequirePlan (plan-gating)
│   │   │   ├── csrf.go
│   │   │   ├── ratelimit.go
│   │   │   ├── requestid.go
│   │   │   ├── logger.go
│   │   │   └── recover.go
│   │   └── partials.go                 # Routes for HTMX-only fragments
│   │
│   ├── features/                       # ★ FEATURE-BASED MODULES (mirrors Kiranism's src/features/)
│   │   ├── overview/                   # /dashboard
│   │   │   ├── handlers.go
│   │   │   ├── service.go              # Aggregations: total_revenue, signups, ...
│   │   │   ├── components/
│   │   │   │   ├── area_chart.templ
│   │   │   │   ├── bar_chart.templ
│   │   │   │   ├── pie_chart.templ
│   │   │   │   ├── recent_sales.templ
│   │   │   │   └── stat_card.templ
│   │   │   └── schemas/
│   │   │       └── range.go            # ?range=7d|30d|90d
│   │   ├── products/                   # /dashboard/products
│   │   │   ├── handlers.go             # List, New, Create, Edit, Update, Delete
│   │   │   ├── service.go              # CRUD + search/filter/sort
│   │   │   ├── components/
│   │   │   │   ├── table.templ        # Server-side data table
│   │   │   │   ├── table_row.templ    # HTMX-swappable row
│   │   │   │   ├── filters.templ      # Faceted filters
│   │   │   │   ├── form.templ         # Create/edit form
│   │   │   │   └── delete_dialog.templ
│   │   │   └── schemas/
│   │   │       ├── product.go         # ProductCreate / ProductUpdate (validator)
│   │   │       └── list_params.go     # Type-safe ?q&page&pageSize&sort&category
│   │   ├── kanban/                     # /dashboard/kanban
│   │   │   ├── handlers.go
│   │   │   ├── service.go              # Columns + cards CRUD + reordering
│   │   │   ├── components/
│   │   │   │   ├── board.templ
│   │   │   │   ├── column.templ
│   │   │   │   └── card.templ
│   │   │   └── schemas/
│   │   │       └── reorder.go          # Reorder request body
│   │   ├── workspaces/                 # /dashboard/workspaces
│   │   │   ├── handlers.go
│   │   │   ├── service.go              # CRUD orgs, members, invitations
│   │   │   ├── components/
│   │   │   │   ├── list.templ
│   │   │   │   ├── switcher.templ      # Topbar dropdown
│   │   │   │   ├── team_table.templ
│   │   │   │   ├── invite_dialog.templ
│   │   │   │   └── settings_tabs.templ
│   │   │   └── schemas/
│   │   │       ├── workspace.go
│   │   │       └── invite.go
│   │   ├── billing/                    # /dashboard/billing
│   │   │   ├── handlers.go             # Checkout, portal, webhook
│   │   │   ├── service.go              # Polar.sh wrapper
│   │   │   ├── components/
│   │   │   │   ├── pricing_table.templ
│   │   │   │   ├── current_plan.templ
│   │   │   │   └── invoices_table.templ
│   │   │   └── schemas/
│   │   │       └── checkout.go
│   │   ├── profile/                    # /dashboard/profile
│   │   │   ├── handlers.go
│   │   │   ├── service.go
│   │   │   ├── components/
│   │   │   │   ├── account_tab.templ
│   │   │   │   ├── security_tab.templ
│   │   │   │   └── sessions_tab.templ
│   │   │   └── schemas/
│   │   │       └── profile.go
│   │   ├── exclusive/                  # /dashboard/exclusive (plan-gated)
│   │   │   ├── handlers.go
│   │   │   └── components/
│   │   │       └── upgrade_cta.templ
│   │   └── auth/                       # /auth/*
│   │       ├── handlers.go             # Sign-in, sign-up, OAuth, reset, verify
│   │       ├── service.go
│   │       ├── components/
│   │       │   ├── signin_form.templ
│   │       │   ├── signup_form.templ
│   │       │   ├── oauth_buttons.templ
│   │       │   └── magic_link_form.templ
│   │       └── schemas/
│   │           ├── credentials.go
│   │           └── reset.go
│   │
│   ├── domain/                         # Pure structs / enums / errors (no PB)
│   │   ├── user.go
│   │   ├── workspace.go
│   │   ├── product.go
│   │   ├── kanban.go
│   │   ├── role.go
│   │   ├── plan.go
│   │   └── errors.go
│   │
│   ├── searchparams/                   # ★ Type-safe URL state (Nuqs equivalent)
│   │   ├── parser.go                   # Parse[T]() / Encode[T]()
│   │   ├── codecs.go                   # Int, IntDefault, String, Enum, CSV, Bool
│   │   └── parser_test.go
│   │
│   ├── command/                        # ★ Command palette registry (kbar equivalent)
│   │   ├── registry.go                 # Action registration + RBAC filtering
│   │   ├── search.go                   # Fuzzy match scoring
│   │   ├── handler.go                  # GET /command/search HTMX endpoint
│   │   └── actions.go                  # Built-in actions
│   │
│   ├── ui/
│   │   ├── flash.go                    # Server-side flash messages → toast
│   │   └── pagination.go               # Pagination URL builders
│   │
│   ├── rbac/                           # ★ Role/permission matrix
│   │   ├── permissions.go
│   │   ├── matrix.go
│   │   └── matrix_test.go
│   │
│   ├── plans/                          # Plan catalog + feature gates
│   │   ├── catalog.go                  # Free / Pro / Team / Enterprise
│   │   └── gates.go                    # HasFeature("analytics_export")
│   │
│   ├── theme/                          # ★ Theme registry (6+ themes)
│   │   ├── registry.go
│   │   └── themes.go                   # Default, Slate, Stone, Zinc, Neutral, Rose
│   │
│   ├── monitoring/
│   │   ├── sentry.go
│   │   └── slog.go
│   │
│   ├── i18n/
│   │   └── i18n.go
│   │
│   ├── version/
│   │   └── version.go
│   │
│   └── testutil/
│       ├── app.go                      # Spawn test PB instance
│       └── fixtures.go
│
├── components/                         # All shared templ components
│   ├── ui/                             # 1:1 port of templUI (40+)
│   │   ├── accordion/ ... tooltip/     # (see full list in §10)
│   │   └── ...
│   ├── layout/                         # ★ Admin shell
│   │   ├── dashboard.templ             # Full shell (sidebar + topbar + main)
│   │   ├── auth.templ                  # Centered auth layout
│   │   ├── sidebar.templ
│   │   ├── sidebar_item.templ
│   │   ├── sidebar_group.templ
│   │   ├── topbar.templ
│   │   ├── breadcrumb.templ
│   │   ├── workspace_switcher.templ
│   │   ├── user_menu.templ
│   │   ├── theme_picker.templ
│   │   ├── command_trigger.templ       # ⌘K button in topbar
│   │   ├── command_palette.templ       # The palette itself
│   │   ├── infobar.templ               # ★ Helpful tips / status messages
│   │   ├── notifications_bell.templ
│   │   └── footer.templ
│   ├── icons/
│   │   └── icons.templ                 # Lucide icons as templ funcs
│   └── pages/                          # Full page templates per route
│       ├── overview.templ
│       ├── products_list.templ
│       ├── products_new.templ
│       ├── products_edit.templ
│       ├── kanban.templ
│       ├── workspaces_list.templ
│       ├── workspaces_team.templ
│       ├── billing.templ
│       ├── exclusive.templ
│       ├── exclusive_locked.templ      # Fallback when plan is insufficient
│       ├── profile.templ
│       ├── signin.templ
│       ├── signup.templ
│       ├── forgot_password.templ
│       ├── reset_password.templ
│       ├── not_found.templ             # 404
│       ├── global_error.templ          # 500 with Sentry event ID
│       └── unauthorized.templ          # 401/403
│
├── assets/
│   ├── css/
│   │   ├── input.css                   # Tailwind entry + theme tokens
│   │   ├── output.css                  # Generated
│   │   ├── sources.generated.css       # Generated for import workflow
│   │   └── themes/
│   │       ├── default.css
│   │       ├── slate.css
│   │       ├── stone.css
│   │       ├── zinc.css
│   │       ├── neutral.css
│   │       └── rose.css
│   ├── js/
│   │   ├── alpine.min.js
│   │   ├── htmx.min.js
│   │   ├── htmx-sse.min.js             # For Server-Sent Events
│   │   ├── chart.umd.min.js            # Chart.js
│   │   ├── sortable.min.js             # SortableJS (kanban DnD)
│   │   ├── fuzzy.min.js                # Tiny fuzzy matcher for command palette
│   │   └── <templui-per-component>.min.js
│   ├── img/
│   │   ├── logo.svg
│   │   ├── logo-dark.svg
│   │   ├── favicon.ico
│   │   ├── og.png
│   │   ├── avatars/                    # Seeded demo avatars
│   │   └── empty_states/               # SVG illustrations
│   ├── fonts/
│   │   └── inter/...
│   └── embed.go                        # //go:embed all
│
├── migrations/
│   ├── 1700000000_init_users.go
│   ├── 1700000050_init_workspaces.go    # workspaces + members + invitations
│   ├── 1700000100_init_products.go
│   ├── 1700000200_init_kanban.go
│   ├── 1700000300_init_subscriptions.go
│   ├── 1700000400_init_invoices.go
│   ├── 1700000500_init_audit_log.go
│   └── seed/
│       ├── seed.go                     # Wiring
│       ├── products.go
│       ├── kanban.go
│       └── users.go
│
├── pb_data/                            # Runtime (gitignored)
│
├── docs/                               # Repo docs
│   ├── ARCHITECTURE.md
│   ├── DEPLOYMENT.md
│   ├── CONTRIBUTING.md
│   ├── clerk_setup.md                  # Optional Clerk path
│   ├── cleanup.md                      # See §31
│   └── adr/
│
└── scripts/
    ├── install.sh
    └── deploy.sh
```

### Why "feature-based"

Kiranism's killer pattern is co-located feature folders. We adopt it verbatim. Every new domain entity becomes a folder under `internal/features/<name>/` with five predictable subdirectories: `handlers.go`, `service.go`, `components/`, `schemas/`, `utils/`. Discoverability for humans and AI agents both jumps an order of magnitude.

---

## 7. Pages Catalog (mirrors Kiranism's pages)

Every page in Kiranism's starter has a matched page here. Routes, status codes, layouts, and key UX elements are documented.

| # | Kiranism page | Path | Handler | Layout | Key UX |
|---|---|---|---|---|---|
| 1 | Sign-in | `GET /auth/signin` | `auth.SignInPage` | `layout.Auth` | Email/password + OAuth + magic link toggle |
| 2 | Sign-up | `GET /auth/signup` | `auth.SignUpPage` | `layout.Auth` | Same form, terms-of-service checkbox |
| 3 | Forgot password | `GET /auth/forgot-password` | `auth.ForgotPasswordPage` | `layout.Auth` | Email submit, success state |
| 4 | Reset password | `GET /auth/reset-password` | `auth.ResetPasswordPage` | `layout.Auth` | Token from query string |
| 5 | Verify email | `GET /auth/verify-email` | `auth.VerifyEmailPage` | `layout.Auth` | Resend link, status polling via SSE |
| 6 | Dashboard Overview | `GET /dashboard` | `overview.Page` | `layout.Dashboard` | Stat cards + 4 charts (parallel-loaded), recent sales |
| 7 | Products list | `GET /dashboard/products` | `products.List` | `layout.Dashboard` | Data table, search, filters, sort, pagination, bulk actions |
| 8 | Product new | `GET /dashboard/products/new` | `products.NewPage` | `layout.Dashboard` | RHF-style form with image upload |
| 9 | Product edit | `GET /dashboard/products/:id` | `products.EditPage` | `layout.Dashboard` | Same form, pre-filled, delete button |
| 10 | Kanban | `GET /dashboard/kanban` | `kanban.Page` | `layout.Dashboard` | Drag-and-drop board, add column/card, server persistence |
| 11 | Workspaces list | `GET /dashboard/workspaces` | `workspaces.List` | `layout.Dashboard` | Card grid, create-workspace dialog, set active |
| 12 | Team management | `GET /dashboard/workspaces/team` | `workspaces.Team` | `layout.Dashboard` | Members table, invite dialog, role select, security & danger tabs |
| 13 | Billing & Plans | `GET /dashboard/billing` | `billing.Page` | `layout.Dashboard` | Current plan card, pricing table, invoices history |
| 14 | Exclusive page | `GET /dashboard/exclusive` | `exclusive.Page` | `layout.Dashboard` | Locked behind `middleware.RequirePlan("pro")`; shows upgrade CTA otherwise |
| 15 | Profile | `GET /dashboard/profile` | `profile.Page` | `layout.Dashboard` | Tabs: account, security, sessions, danger zone |
| 16 | Not Found | (any unmatched) | `errors.NotFound` | `layout.Dashboard` (when authed) or plain | 404 illustration + back-to-dashboard |
| 17 | Global Error | (panic recovered) | `errors.Global` | plain | 500 with Sentry event ID + report button |

### HTMX-only fragment routes

For lazy/parallel loading. Not listed in user-facing nav.

| Path | Returns |
|---|---|
| `GET /dashboard/_partials/overview/:card` | Single stat-card or chart fragment |
| `GET /dashboard/_partials/products/table` | Just the products table tbody |
| `GET /dashboard/_partials/products/filters/apply` | Filter chip updates |
| `POST /dashboard/_partials/kanban/reorder` | Persists card move; returns updated board state |
| `GET /command/search?q=...` | Command palette result list |
| `GET /notifications/dropdown` | Notifications panel |

### Auth route group

All auth routes share `layout.Auth` (centered card, no sidebar). Group middleware: `RedirectIfAuthenticated`.

### Dashboard route group

All `/dashboard/*` routes share `layout.Dashboard` (sidebar + topbar). Group middleware (in order):

1. `RequireAuth` — redirects unauthenticated to `/auth/signin`.
2. `ResolveActiveWorkspace` — reads `active_workspace` cookie, falls back to user's first membership.
3. `LoadRBACContext` — attaches the user's role + permission set to the request.
4. `LoadTheme` — reads `theme` cookie, applies the right CSS theme file.
5. `Telemetry` — tags Sentry scope with `user_id`, `workspace_id`, `plan`.

---

## 8. Authentication Strategy

We support **two auth paths**, selectable at boot via env var `AUTH_PROVIDER=pocketbase|clerk` (default `pocketbase`). This matches Kiranism's "swap Clerk if you want" ethos.

### Path A — PocketBase native (default)

- Email + password (PB built-in).
- OAuth2 (Google, GitHub, Apple, Microsoft) configured at boot.
- Magic link (OTP via email).
- MFA (TOTP, PB v0.38.2 supports it).
- JWT issued by PB, stored in an HttpOnly Secure SameSite=Lax cookie (`pb_auth`).
- Sign-in/sign-up/forgot/reset pages all live in `internal/features/auth/`.

### Path B — Clerk (drop-in)

For teams who love Clerk's hosted UI, we ship a Clerk integration:

- `AUTH_PROVIDER=clerk` activates the Clerk middleware (`clerk-sdk-go/v2`).
- Sign-in/sign-up pages redirect to Clerk's hosted pages.
- Profile page renders a thin wrapper around Clerk's `<UserProfile />` (via `<iframe>` or hosted redirect).
- PocketBase users are auto-provisioned from Clerk JWT on first request (email match or `external_id`).
- All other features (workspaces, billing, kanban, products) keep working unchanged.

Documented in `docs/clerk_setup.md` — mirrors Kiranism's same-named doc.

### Session model (both paths)

`internal/server/middleware/auth.go` produces the same `c.Get("user")` record regardless of provider, so downstream code is provider-agnostic.

### Route guards

```go
// Public routes — none required
e.Router.GET("/", marketing.HomeHandler)

// Auth routes — must NOT be authenticated
authPublic := e.Router.Group("/auth").BindFunc(middleware.RedirectIfAuthenticated)
authPublic.GET("/signin", auth.SignInPage)
authPublic.POST("/signin", auth.SignIn)
// ...

// Dashboard — authenticated + workspace + RBAC
dash := e.Router.Group("/dashboard").BindFunc(
    middleware.RequireAuth(pb),
    middleware.ResolveActiveWorkspace(pb),
    middleware.LoadRBACContext,
    middleware.LoadTheme,
)
dash.GET("", overview.Page(d))
dash.GET("/products", products.List(d))
// ...

// Exclusive — plan-gated
dash.GET("/exclusive", exclusive.Page(d), middleware.RequirePlan("pro"))
```

---

## 9. Layout System (Sidebar / Topbar / Content)

The dashboard layout is a 1:1 visual port of Kiranism's: collapsible left sidebar, sticky top header, scrollable content area, optional infobar.

### `layout.Dashboard` structure

```
┌──────────────────────────────────────────────────────────────────────┐
│ Topbar:  ⌘K  •  Workspace Switcher  •  Notifications  •  Theme  •  Avatar │
├──────────┬───────────────────────────────────────────────────────────┤
│ Sidebar  │ Breadcrumb  ·  Page Title                                 │
│          │                                                            │
│ Overview │ [ Infobar — optional contextual tip ]                     │
│ Products │                                                            │
│ Kanban   │ < Page Content via { children... } >                      │
│ Workspaces                                                            │
│ Billing  │                                                            │
│ Profile  │                                                            │
│          │                                                            │
│ Settings │                                                            │
│          │                                                            │
│ Collapse │                                                            │
└──────────┴───────────────────────────────────────────────────────────┘
```

### Sidebar

- Two states: expanded (240px) and collapsed (64px, icon-only).
- State persisted in `sidebar_collapsed` cookie + `localStorage` for SSR/CSR coherence.
- Nav items rendered from a server-side declaration — **filtered by RBAC** before render so unauthorized items never reach the DOM.
- Active item detection by path prefix in the templ component.
- Mobile: sidebar becomes a `@sheet.Sheet()` drawer (templUI).

### Topbar

- Left: hamburger (mobile) + command palette trigger (`⌘K` chip).
- Center: workspace switcher dropdown (templUI Dropdown).
- Right: notifications bell, theme picker (popover with 6 swatches), avatar menu.
- Sticky top, blurred background (`backdrop-blur` Tailwind class).

### Infobar

A dedicated thin band beneath the topbar that shows contextual tips, banner-level status, or upgrade nudges. Mirrors Kiranism's "Infobar component". Server-rendered, dismissible per session (cookie-stored).

```templ
@layout.Infobar(layout.InfobarProps{
    Variant: "info", // info | warning | success
    Icon:    icons.Sparkles,
    Title:   "You're on the Free plan",
    CTA:     layout.InfobarCTA{Label: "Upgrade to Pro", Href: "/dashboard/billing"},
    DismissID: "infobar-upgrade-free-2026-05",
})
```

### Mobile-first details

- Touch targets ≥44px.
- Sidebar drawer opens via Alpine on hamburger click.
- Tables collapse to card lists below `md` breakpoint.
- Topbar pruned: theme picker moves into avatar menu under `sm`.

---

## 10. UI Component Library (templUI port)

We use templUI v1.9.2 directly via the CLI:

```bash
templui init
templui add "*"
```

This copies all 40+ components into `components/ui/` under your repo. You own the code; updates are explicit.

**All shipped components** (same list as the SaaS plan):
Accordion, Alert, Aspect Ratio, Avatar, Badge, Breadcrumb, Button, Calendar, Card, Carousel, Charts, Checkbox, Collapsible, Copy Button, Date Picker, Dialog, Dropdown, Form, Icon, Input, Input OTP, Label, Pagination, Popover, Progress, Radio, Rating, Select Box, Separator, Sheet, Sidebar, Skeleton, Slider, Switch, Table, Tabs, Tags Input, Textarea, Time Picker, Toast, Tooltip.

### Dashboard-specific composite components

In addition to the bare templUI primitives, we add `components/layout/` composites tuned for dashboard UX:

| Composite | Underlying primitives | Use case |
|---|---|---|
| `StatCard` | Card + Skeleton + Icon | KPI tiles on overview page |
| `PageHeader` | Breadcrumb + heading + actions slot | Top of every dashboard page |
| `DataTable` | Table + Input + Select + Pagination | Generic CRUD list page scaffold |
| `EmptyState` | Card + Icon + Button | First-run/empty queries |
| `ConfirmDialog` | Dialog + Button (destructive variant) | Delete confirmations |
| `FormField` | Label + Input + (error message) | Form rows with validation surface |
| `TabbedSettings` | Tabs + Card | Profile, team, workspace settings |
| `KanbanColumn` / `KanbanCard` | Card + Badge + Avatar group | Kanban board pieces |

These composites live in `components/layout/` (not in `components/ui/`) so they don't accidentally collide with future templUI CLI updates.

### Theme tokens

We use templUI's OKLCH-based theme variables verbatim (`--background`, `--foreground`, `--primary`, `--card`, `--popover`, `--muted`, `--accent`, `--destructive`, `--border`, `--input`, `--ring`, sidebar tokens). Six themes are shipped (see §20).

---

## 11. Data Tables (TanStack-style, server-driven)

Kiranism uses TanStack Table for client-side sort/filter/pagination + Nuqs for URL state. We deliver the **same UX** but server-driven — preferable because:

- No client-side data shipping for large tables.
- Search/filter state is bookmarkable and shareable via URL.
- Works without JavaScript (forms degrade gracefully).
- HTMX swaps the table body in 30-80ms — feels instant.

### The `DataTable` contract

A generic `DataTable[T]` composite that any feature instantiates with:

```go
// internal/features/products/components/table.templ
templ ProductsTable(rows []*domain.Product, params schemas.ProductListParams, total int) {
    @layout.DataTable(layout.DataTableProps{
        Columns: []layout.Column{
            {Key: "name", Label: "Name", Sortable: true},
            {Key: "category", Label: "Category", Sortable: true, FilterOptions: products.CategoryOptions()},
            {Key: "price", Label: "Price", Sortable: true, Align: "right"},
            {Key: "stock", Label: "Stock", Sortable: true, Align: "right"},
            {Key: "created", Label: "Created", Sortable: true},
            {Key: "_actions", Label: "", Width: "60px"},
        },
        SearchPlaceholder: "Search products...",
        SearchQuery:       params.Q,
        Page:              params.Page,
        PageSize:          params.PageSize,
        Total:             total,
        Sort:              params.Sort,
        BasePath:          "/dashboard/products",
        TargetID:          "products-table",
        BulkActions: []layout.BulkAction{
            {Label: "Delete", Action: "/dashboard/products/bulk-delete", Variant: "destructive"},
            {Label: "Export", Action: "/dashboard/products/export"},
        },
    }) {
        for _, p := range rows {
            @ProductRow(p)
        }
    }
}
```

### Server-side capabilities

- **Search**: `?q=` is debounced (300ms) on the client and posted via HTMX. Server runs `LIKE` against indexed fields.
- **Filter**: faceted filters (chips). Multi-select uses `?category=a&category=b` (CSV-decoded).
- **Sort**: prefix `-` for descending: `?sort=-created`. Server validates against an allowlist of sortable columns.
- **Pagination**: `?page=2&pageSize=20`. Server returns total count for "X-Total" header + pagination component.
- **Bulk actions**: checkboxes select rows; bulk action posts the array of IDs.
- **Row actions**: per-row "More" dropdown with View/Edit/Delete (templUI Dropdown).

### HTMX wiring

The table body has `hx-get` on filter/search/sort/page changes, targeting the same `<tbody>` element. The whole page never reloads.

```html
<input
  name="q"
  hx-get="/dashboard/products"
  hx-trigger="input changed delay:300ms"
  hx-target="#products-table"
  hx-select="#products-table tbody"
  hx-push-url="true"
  placeholder="Search products..."
/>
```

### Sorting visuals

Column headers show an arrow when active; clicking toggles asc→desc→none.

### Pagination component

Uses templUI's `Pagination` primitive, with prev/next/numbered page links. URL builder is centralized in `internal/ui/pagination.go`.

### Loading + error states

- `hx-indicator` shows a templUI Skeleton in the tbody while fetching.
- 5xx responses render a flash toast + retry button.

---

## 12. Forms (React-Hook-Form + Zod equivalent)

Kiranism uses `react-hook-form` for state and `zod` for schema validation. Our Go equivalent:

- **Schema:** Go struct + `go-playground/validator/v10` tags.
- **State:** Standard HTML form (HTMX handles partial submissions).
- **Validation:** Server-side primary; client-side mirror for instant feedback.
- **Error surface:** Per-field error rendering via HTMX OOB (out-of-band) swaps.

### Schema declaration

```go
// internal/features/products/schemas/product.go
package schemas

type ProductCreate struct {
    Name     string  `form:"name"     validate:"required,min=2,max=120"`
    Category string  `form:"category" validate:"required,oneof=electronics apparel home toys"`
    Price    int64   `form:"price"    validate:"required,gte=0"`     // cents
    Stock    int     `form:"stock"    validate:"required,gte=0"`
    Active   bool    `form:"active"`
    Image    string  `form:"image"    validate:"omitempty,url"`
    Tags     []string `form:"tags"`
}
```

### Form rendering

```templ
templ ProductForm(form schemas.ProductCreate, errors map[string]string) {
    <form hx-post="/dashboard/products" hx-target="#product-form" hx-swap="outerHTML" id="product-form" class="space-y-6">
        @layout.FormField(layout.FormFieldProps{
            Label: "Name", Name: "name", Required: true, Error: errors["Name"],
        }) {
            @input.Input(input.Props{ID: "name", Name: "name", Value: form.Name, Required: true})
        }
        @layout.FormField(layout.FormFieldProps{
            Label: "Category", Name: "category", Required: true, Error: errors["Category"],
        }) {
            @selectbox.SelectBox(selectbox.Props{Name: "category", Options: products.CategoryOptions(), Value: form.Category})
        }
        // ... more fields ...
        @button.Button(button.Props{Type: "submit"}) { Save product }
    </form>
}
```

### Handler pattern

```go
func Create(d *server.Deps) echo.HandlerFunc {
    return func(c echo.Context) error {
        var form schemas.ProductCreate
        if err := c.Bind(&form); err != nil { return badRequest(c, err) }

        if errs := d.Validator.StructCtx(c.Request().Context(), &form); errs != nil {
            return c.Render(http.StatusUnprocessableEntity, ProductForm(form, fieldErrors(errs)))
        }

        if _, err := d.Products.Create(c.Request().Context(), c.Get("workspace_id").(string), &form); err != nil {
            return internalError(c, err)
        }

        // HTMX-aware redirect
        c.Response().Header().Set("HX-Redirect", "/dashboard/products?created=1")
        return c.NoContent(http.StatusOK)
    }
}
```

### Live validation (optional)

For "type-as-you-go" feedback: each field can opt-in by adding `hx-post="/dashboard/products/validate-field"` on `blur`. The endpoint validates that one field and returns just the error message swap. Minimal cost, big UX win.

### File uploads

`<input type="file">` posted via HTMX `hx-encoding="multipart/form-data"`. Bound directly into PocketBase's file field via `c.FormFile("image")`.

---

## 13. Type-Safe Search Params (Nuqs equivalent)

Nuqs gives Next.js apps a "Zustand for the URL." Our equivalent is `internal/searchparams/`:

### Per-page schema

```go
// internal/features/products/schemas/list_params.go
package schemas

type ProductListParams struct {
    Q        string   `param:"q"`
    Page     int      `param:"page"      default:"1"   min:"1"`
    PageSize int      `param:"pageSize"  default:"20"  oneof:"10 20 50 100"`
    Sort     string   `param:"sort"      default:"-created"  allowed:"name -name price -price created -created"`
    Category []string `param:"category"  csv:"true"`
    Active   *bool    `param:"active"`           // tri-state filter
}
```

### Parse + encode

```go
params, err := searchparams.Parse[schemas.ProductListParams](c.Request())
// → typed struct, defaults applied, validations enforced

href := searchparams.Encode("/dashboard/products", params.WithPage(3))
// → "/dashboard/products?q=widget&page=3&pageSize=20&sort=-created&category=apparel,toys"
```

### Why this matters

- **One source of truth** for URL contract per page.
- **Compile-time** breakage on typos (no stringly-typed URL keys scattered around).
- **Default values** keep URLs short for common views.
- **Allowlists** prevent arbitrary user input from reaching SQL ORDER BY.

The `searchparams` package is unit-tested with table-driven cases covering encoding, decoding, defaults, validation errors, and edge cases (empty CSV, multi-value bool, etc.).

---

## 14. State Management (Zustand equivalent)

Kiranism uses Zustand for client-side state (kanban board local state, sidebar collapse, theme preview, etc.). Our equivalents:

| Kiranism Zustand use case | go-pocket equivalent |
|---|---|
| Sidebar collapsed | `sidebar_collapsed` cookie + Alpine `x-data` boolean |
| Theme preview before save | Alpine `x-data` on theme picker; on commit, POST to `/api/preferences/theme` |
| Kanban board (optimistic) | SortableJS handles DOM; HTMX persists order on drop with optimistic UI |
| Command palette open/close | Alpine `x-show` driven by `Cmd/Ctrl+K` listener |
| Multi-step form draft | Server-side draft (PocketBase `form_drafts` collection) keyed by user/session |
| Notifications panel | Server SSE stream → HTMX `hx-ext="sse"` swaps into the dropdown |

### Alpine.js stores

For the few genuinely client-only states (sidebar, palette open, theme preview), we use Alpine's `Alpine.store(...)` global state:

```html
<script nonce="...">
  document.addEventListener('alpine:init', () => {
    Alpine.store('ui', {
      sidebarCollapsed: localStorage.getItem('sidebar') === 'collapsed',
      commandOpen: false,
      toggleSidebar() {
        this.sidebarCollapsed = !this.sidebarCollapsed;
        localStorage.setItem('sidebar', this.sidebarCollapsed ? 'collapsed' : 'expanded');
        document.cookie = `sidebar=${this.sidebarCollapsed ? 'collapsed' : 'expanded'}; path=/; max-age=31536000; SameSite=Lax`;
      },
    });
  });
</script>
```

This 20-line shim replaces ~50 lines of Zustand boilerplate.

### Why not Zustand-in-Go?

Server-rendered apps don't need it. Most "state" is either (a) URL params (handled by §13), (b) cookies (HTTP-native), or (c) database (PocketBase).

---

## 15. Command Palette (kbar equivalent)

The `⌘K` palette is a signature feature of the Kiranism starter. Our version:

### Server-side action registry

```go
// internal/command/registry.go
type Action struct {
    ID          string
    Label       string
    Subtitle    string
    Icon        string                       // Lucide icon name
    Category    string                       // "Navigation" | "Workspace" | "Theme" | ...
    Keywords    []string                     // For fuzzy matching
    Shortcut    []string                     // ["g", "p"] for "go to products"
    URL         string                       // If set, navigate to this URL
    HXPost      string                       // If set, HTMX-post here (e.g., switch theme)
    Permission  rbac.Permission              // Optional gate
    PlanGate    string                       // Optional "pro" / "team" gate
}

// Registered at boot, in feature init() funcs:
command.Register(Action{
    ID: "nav.products", Label: "Go to Products", Category: "Navigation",
    Icon: "package", URL: "/dashboard/products", Shortcut: []string{"g", "p"},
})
command.Register(Action{
    ID: "products.new", Label: "Create new product", Category: "Products",
    Icon: "plus", URL: "/dashboard/products/new", Permission: rbac.ProductsCreate,
})
command.Register(Action{
    ID: "theme.set.slate", Label: "Switch to Slate theme", Category: "Theme",
    Icon: "palette", HXPost: "/api/preferences/theme?value=slate",
})
```

### UI

A templUI `@dialog.Dialog()` activated by Alpine on `Cmd/Ctrl+K`. Input field + scrollable result list. Each result row shows icon, label, subtitle, shortcut.

### Search endpoint

```
GET /command/search?q=prod
```

Returns ranked results as an HTMX fragment. Scoring: substring match (weighted by position) + keyword match + recency boost (recently used actions float up).

### Keyboard navigation

- `↑`/`↓` to move selection.
- `Enter` to invoke.
- `Esc` to close.
- Within the input, typing immediately filters.
- For URL actions, navigation is browser-native; for HXPost actions, HTMX fires and palette closes on success.

### RBAC + plan filtering

The registry filters actions before sending them — users never see (and can never invoke) actions they're not permitted to perform.

---

## 16. Charts & Analytics (Recharts equivalent)

Kiranism's Overview uses Recharts (Area, Bar, Line, Pie). We use **Chart.js v4** via templUI's Charts component.

### Why Chart.js

- Self-hosted (one ~70KB minified file, embedded in our binary).
- Zero React dependency — Chart.js runs on a raw `<canvas>`.
- templUI already ships a `@charts.Chart()` wrapper.

### Data flow

```
Overview handler
   ↓ (PocketBase aggregation queries)
overview.Service.RevenueSeries(workspaceID, range) → []domain.TimePoint
   ↓
templ component injects JSON into <script type="application/json" id="chart-revenue-data">
   ↓
Chart.js reads JSON, renders into <canvas id="chart-revenue">
```

### Parallel-loading pattern

Each chart is its own HTMX fragment, loaded on `hx-trigger="load"`:

```html
<div hx-get="/dashboard/_partials/overview/revenue-chart?range=30d"
     hx-trigger="load"
     hx-swap="outerHTML">
  @ui.Skeleton(class: "h-72")
</div>
```

This is our analog to Next.js parallel routes — each card loads independently, has its own error boundary, and renders progressively.

### Range selector

```
[ 7 days | 30 days | 90 days | 12 months ]   (Tabs)
```

Triggers `hx-get` on all charts in parallel.

### Recent sales panel

Server-rendered list of last 10 transactions. Avatar + name + amount. Auto-refreshes every 60s via `hx-trigger="every 60s"`.

---

## 17. Kanban Board (dnd-kit equivalent)

Kiranism's kanban uses `dnd-kit` for drag-and-drop with Zustand for state. Ours:

### Stack

- **DOM library**: SortableJS v1.15.x (~30KB, zero deps, mature).
- **Persistence**: HTMX POST on drop → server updates `kanban_cards.position` and `kanban_cards.column`.
- **Optimistic UI**: SortableJS keeps the DOM moved; if the server returns error, HTMX swaps back to truth.

### Data model

```
kanban_columns: id, workspace, name, position, color
kanban_cards: id, workspace, column, title, description, position, assignees, due_date, labels[]
```

`position` is a sparse integer; on drop we reposition (we could later swap to fractional indexing if reorders cluster).

### Wire-up

```templ
templ KanbanColumn(col *domain.KanbanColumn, cards []*domain.KanbanCard) {
    <div class="w-72 shrink-0 rounded-lg bg-muted p-3"
         data-column-id={ col.ID }>
        <h3 class="mb-3 font-semibold">{ col.Name }</h3>
        <div id={ "column-" + col.ID } class="space-y-2 min-h-[40px]"
             data-sortable-group="kanban">
            for _, card := range cards {
                @KanbanCard(card)
            }
        </div>
    </div>
}
```

```js
// initialized in components/layout/dashboard.templ (after Sortable.min.js loaded)
document.querySelectorAll('[data-sortable-group="kanban"]').forEach(el => {
  new Sortable(el, {
    group: 'kanban',
    animation: 150,
    onEnd: e => {
      htmx.ajax('POST', '/dashboard/_partials/kanban/reorder', {
        values: {
          card_id: e.item.dataset.cardId,
          to_column: e.to.closest('[data-column-id]').dataset.columnId,
          to_index: e.newIndex,
        },
        target: '#kanban-board',
        swap: 'outerHTML',
      });
    },
  });
});
```

### Server handler

```go
func Reorder(d *server.Deps) echo.HandlerFunc {
    return func(c echo.Context) error {
        var req schemas.ReorderRequest
        if err := c.Bind(&req); err != nil { return badRequest(c, err) }
        if err := d.Validator.Struct(&req); err != nil { return badRequest(c, err) }

        if err := d.Kanban.Move(c.Request().Context(), c.Get("workspace_id").(string), req.CardID, req.ToColumn, req.ToIndex); err != nil {
            return internalError(c, err)
        }

        cols, cards, err := d.Kanban.Board(c.Request().Context(), c.Get("workspace_id").(string))
        if err != nil { return internalError(c, err) }
        return c.Render(http.StatusOK, KanbanBoard(cols, cards))
    }
}
```

### Add/edit/delete

Cards opened in a templUI `@sheet.Sheet()` side panel with the same form pattern from §12. Columns managed via a "+" button at the right of the board.

---

## 18. Multi-Tenant Workspaces & RBAC

Kiranism leans on Clerk Organizations for this. We provide native equivalents that work even on the PocketBase auth path.

### Collections (per-workspace tenancy)

| Collection | Purpose |
|---|---|
| `workspaces` | id, slug, name, logo, owner, polar_customer_id, plan, subscription_status, seats_used, seats_limit, settings |
| `workspace_members` | id, workspace, user, role, joined_at |
| `workspace_invitations` | id, workspace, email, role, token, invited_by, expires_at, accepted_at, revoked_at |
| (every business collection) | + `workspace` relation |

### Roles

| Role | Capabilities |
|---|---|
| `owner` | Everything; one per workspace; transferable; manages billing |
| `admin` | Manage members + invitations; settings; view billing; cannot delete workspace |
| `member` | Full read/write on domain data; no member/billing |
| `viewer` | Read-only |

### Permission matrix

```go
// internal/rbac/matrix.go
var matrix = map[Permission]map[Role]bool{
    PermProductsRead:    {Owner: true, Admin: true, Member: true, Viewer: true},
    PermProductsCreate:  {Owner: true, Admin: true, Member: true},
    PermProductsUpdate:  {Owner: true, Admin: true, Member: true},
    PermProductsDelete:  {Owner: true, Admin: true},
    PermKanbanWrite:     {Owner: true, Admin: true, Member: true},
    PermBillingView:     {Owner: true, Admin: true},
    PermBillingManage:   {Owner: true, Admin: true},
    PermMembersInvite:   {Owner: true, Admin: true},
    PermMembersRemove:   {Owner: true, Admin: true},
    PermWorkspaceDelete: {Owner: true},
}
```

### RBAC-driven sidebar

Sidebar nav declared once on the server. Before render, we filter out items the active user can't access. So a `viewer` doesn't see "Billing" or "Settings → Members" in the nav at all.

```go
sidebar.Item{
    Path: "/dashboard/billing",
    Label: "Billing",
    Icon: "credit-card",
    Permission: rbac.PermBillingView,
}
```

### Active workspace resolution

Same as the go-pocket multi-tenant plan: URL segment `/workspaces/:slug/...` (where applicable) → `active_workspace` cookie → `users.last_active_workspace` → first membership.

For this dashboard, since the URL doesn't always carry a workspace slug, we lean primarily on the cookie. The workspace switcher in the topbar updates it via `POST /api/workspaces/switch`.

### Invitation flow

1. Admin opens Team page, clicks Invite, enters email + role.
2. Server creates `workspace_invitations` row with 32-byte token, 7-day expiry.
3. Email sent via Resend with link `/invite/:token`.
4. Recipient lands on `/invite/:token` → sign-in or sign-up → membership row created → redirect to `/dashboard?workspace=<slug>`.

### Tenant isolation — three layers (non-negotiable)

1. **Service layer**: every repository method takes `workspaceID`; queries filter by it.
2. **PocketBase rules**: every business collection's `listRule`/`viewRule`/etc. joins through `workspace_members`.
3. **Middleware**: `ResolveActiveWorkspace` + `RequireRole` reject cross-workspace access at the HTTP edge.

---

## 19. Billing & Feature Gating

Polar.sh handles per-workspace billing (mirrors the SaaS plan §12.5). The dashboard surfaces:

### Billing page (`/dashboard/billing`)

- **Current plan card** with seats used / seats limit, renewal date, status badge.
- **Pricing table**: 4 tiers (Free / Pro / Team / Enterprise), monthly/annual toggle, "Most popular" badge, CTA per row.
- **Invoices table**: server-driven data table with date, amount, status, hosted invoice link.
- **Customer portal button**: opens Polar-hosted portal in new tab.

### Plan-gating

```go
dash.GET("/exclusive", exclusive.Page(d), middleware.RequirePlan("pro"))
```

`RequirePlan` reads the active workspace's `plan` field and:

- If `plan ∈ allowed`: continues.
- Else: renders `pages.ExclusiveLocked()` — a friendly upgrade CTA. Mirrors Kiranism's `<Protect>` fallback.

### Feature flags per plan

```go
// internal/plans/gates.go
var planFeatures = map[string][]string{
    "free":       {"basic_products", "basic_kanban"},
    "pro":        {"basic_products", "basic_kanban", "analytics_export", "exclusive_page"},
    "team":       {"basic_products", "basic_kanban", "analytics_export", "exclusive_page", "audit_log", "sso"},
    "enterprise": {"*"},
}

func HasFeature(plan, feature string) bool { /* ... */ }
```

`HasFeature` is used in templ to hide entire UI regions (export buttons, advanced filters) on insufficient plans.

### Webhook handler

`POST /webhooks/polar` — HMAC-verified, dispatches `subscription.created/updated/canceled`, `order.paid`, etc. Updates `subscriptions`, `invoices`, denormalizes `plan` + `subscription_status` onto `workspaces`.

---

## 20. Theme System (6+ themes, tweakcn equivalent)

Kiranism uses [tweakcn](https://tweakcn.com/) to generate theme presets. We ship six themes built on the same OKLCH token model:

| Theme | Vibe |
|---|---|
| Default | Stock templUI / shadcn |
| Slate | Cool gray, blue-tinted |
| Stone | Warm gray |
| Zinc | Neutral metallic gray |
| Neutral | Pure grayscale |
| Rose | Warm rose primary on neutral surfaces |

Each theme is a single CSS file under `assets/css/themes/<name>.css` that overrides the OKLCH variables. Switching themes:

1. User clicks theme picker in topbar.
2. Alpine immediately applies a preview by swapping the `<link rel="stylesheet">` href.
3. On confirm, `POST /api/preferences/theme?value=slate` persists to user record + cookie.
4. Next request: `LoadTheme` middleware reads cookie and renders the right `<link>` href.

### No FOUC

Theme is determined on the server before rendering, so the right stylesheet ships in the initial HTML. Dark/light is layered on top via the `.dark` class (handled separately by the dark-mode toggle).

### Adding a new theme

```bash
# Drop a new file under assets/css/themes/
# Add the theme to internal/theme/themes.go
# It immediately appears in the topbar picker
```

The plan recommends sourcing new theme presets from tweakcn.com and pasting the OKLCH variables verbatim.

---

## 21. Error Tracking (Sentry)

Mirrors Kiranism's Sentry integration:

- `getsentry/sentry-go` v0.32.0 initialized at boot with `SENTRY_DSN`.
- Recover middleware captures panics, attaches `request_id`, `user_id`, `workspace_id`, `plan` to the Sentry scope.
- HTTP 5xx responses also reported.
- `pages.GlobalError` renders the friendly error page with the Sentry event ID, so users can give support a single token.

```templ
templ GlobalError(eventID string) {
    @layout.Auth() {  // borrow the simpler centered layout
        <div class="space-y-4 text-center">
            @icons.AlertTriangle(icons.Props{Size: 48, Class: "mx-auto text-destructive"})
            <h1 class="text-2xl font-semibold">Something went wrong</h1>
            <p class="text-muted-foreground">We've been notified. Try again in a moment.</p>
            if eventID != "" {
                <p class="text-xs text-muted-foreground">Reference: <code>{ eventID }</code></p>
            }
            @button.Button(button.Props{Href: "/dashboard", Variant: "outline"}) { Back to dashboard }
        </div>
    }
}
```

### Replay (privacy-aware)

Sentry's session replay is **not** enabled by default. Setup doc explains how to enable client-side replay if desired, including PII masking.

---

## 22. Infobar Component

A dedicated thin band beneath the topbar for contextual messages. Server-rendered, dismissible:

```templ
@layout.Infobar(layout.InfobarProps{
    Variant:   "info",                     // info | warning | success | destructive
    Icon:      "info",
    Title:     "Welcome to the Pro trial",
    Message:   "Your trial ends in 7 days.",
    CTA:       layout.InfobarCTA{Label: "Upgrade", Href: "/dashboard/billing"},
    Dismissible: true,
    DismissID:   "trial-cta-2026-05",       // Persisted per user (cookie)
})
```

### When to use

- Trial expiry warnings.
- Maintenance windows.
- Welcome guidance on a fresh workspace.
- New-feature announcements.
- Permission elevation prompts.

### Dismissal persistence

A cookie keyed by `DismissID` ensures the same infobar doesn't reappear after the user dismisses it. Reset by changing the `DismissID` (e.g., increment a version suffix).

---

## 23. Configuration & Environment

```bash
# App
APP_ENV=development               # development | staging | production
APP_NAME=go-pocket-dashboard
APP_URL=http://localhost:8090
APP_PORT=8090
APP_SECRET=change-me-32-bytes-min

# Auth
AUTH_PROVIDER=pocketbase          # pocketbase | clerk

# PocketBase
PB_DATA_DIR=./pb_data
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=change-me

# Clerk (only if AUTH_PROVIDER=clerk)
CLERK_PUBLISHABLE_KEY=
CLERK_SECRET_KEY=
CLERK_WEBHOOK_SECRET=

# Resend
RESEND_API_KEY=re_xxxxxxxxxxxxxxxx
RESEND_FROM="Acme <hello@yourdomain.com>"

# Polar.sh
POLAR_SERVER=sandbox              # sandbox | production
POLAR_ORGANIZATION_ACCESS_TOKEN=polar_oat_xxxxxxxxxxxxxxxxx
POLAR_WEBHOOK_SECRET=whsec_xxx
POLAR_PRICE_PRO_MONTHLY=price_xxx
POLAR_PRICE_PRO_YEARLY=price_xxx
POLAR_PRICE_TEAM_MONTHLY=price_xxx
POLAR_PRICE_TEAM_YEARLY=price_xxx
POLAR_PRICE_ENTERPRISE_MONTHLY=price_xxx

# OAuth (used if AUTH_PROVIDER=pocketbase)
OAUTH_GOOGLE_CLIENT_ID=
OAUTH_GOOGLE_CLIENT_SECRET=
OAUTH_GITHUB_CLIENT_ID=
OAUTH_GITHUB_CLIENT_SECRET=

# Sentry
SENTRY_DSN=
SENTRY_ENVIRONMENT=development
SENTRY_TRACES_SAMPLE_RATE=0.1

# Theme
DEFAULT_THEME=default             # default | slate | stone | zinc | neutral | rose
DEFAULT_COLOR_MODE=system         # light | dark | system

# Feature flags
FEATURE_KANBAN=true
FEATURE_COMMAND_PALETTE=true
FEATURE_BILLING=true
FEATURE_CLERK_INTEGRATION=false

# Observability
LOG_LEVEL=info
LOG_FORMAT=text
```

---

## 24. Development Workflow

### One-time setup

```bash
git clone https://github.com/milzamsz/go-pocket myapp
cd myapp
cp .env.example .env
go version                                              # 1.26.3
go install github.com/a-h/templ/cmd/templ@v0.3.1020
go install github.com/go-task/task/v3/cmd/task@v3.49.1
go install github.com/templui/templui/cmd/templui@v1.9.2
# Download Tailwind CSS v4.3.0 standalone binary from tailwindlabs/tailwindcss releases
task setup            # Runs migrations + seeds demo data + creates PB admin
```

### Daily loop

```bash
task dev              # Parallel: templ --watch + tailwindcss --watch + go run .
```

Open `http://localhost:8090/dashboard` to see the running app.

### Taskfile (excerpt)

```yaml
version: "3"

tasks:
  setup:
    cmds:
      - go mod tidy
      - templui init
      - templui add "*"
      - task: migrate
      - task: seed

  dev:
    cmds:
      - task --parallel tailwind templ

  tailwind:
    cmds:
      - tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

  templ:
    cmds:
      - templ generate --watch --proxy="http://localhost:8090" --cmd="go run ./main.go" --open-browser=false

  build:
    cmds:
      - templ generate
      - tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify
      - CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version={{.VERSION}}" -o ./bin/go-pocket-dashboard .
    vars:
      VERSION:
        sh: git describe --tags --always --dirty

  test:
    cmds:
      - go test ./... -race -cover

  lint:
    cmds:
      - go vet ./...
      - gofmt -l -s .
      - templ fmt --check .

  templui:update:
    cmds:
      - templui --installed add

  seed:
    cmds:
      - go run ./cmd/tools/seed

  migrate:
    cmds:
      - go run . migrate up
```

---

## 25. Build, Asset Pipeline & Embed

### Embedded files

```go
// assets/embed.go
package assets

import "embed"

//go:embed all:css all:js all:img all:fonts
var Assets embed.FS
```

```go
// migrations are auto-registered via PB's migration package
// content (none for this dashboard starter) — could embed docs later if desired
```

### Build steps

1. `templ generate` → `*.templ` to `*_templ.go`.
2. `tailwindcss -i input.css -o output.css --minify` → 30-50 KB CSS.
3. `go build -ldflags="-s -w -X main.Version=$(git describe)"` → 25-35 MB binary.

### Cache busting

Every asset URL carries `?v=<sha256-prefix>` computed once at boot from the embedded FS.

### Per-component templUI scripts

Interactive templUI components (datepicker, dropdown, dialog, popover, toast, sheet, etc.) each ship a `*.min.js` file. We load them via `@<component>.Script()` calls in the base layout's `<head>`. The browser caches them aggressively (immutable, hashed filenames).

---

## 26. Deployment (Docker / Dokploy)

Same story as the SaaS plan — one binary, one Docker image, two paths.

### Dockerfile (multi-stage)

```dockerfile
# syntax=docker/dockerfile:1.7

FROM golang:1.26.3-alpine3.23 AS builder
RUN apk add --no-cache git curl \
    && curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss \
    && go install github.com/a-h/templ/cmd/templ@v0.3.1020

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN templ generate \
 && tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /out/go-pocket-dashboard .

FROM alpine:3.23.3
RUN apk add --no-cache ca-certificates tzdata sqlite \
 && adduser -D -u 10001 gopocket
WORKDIR /app
COPY --from=builder --chown=gopocket:gopocket /out/go-pocket-dashboard /app/go-pocket-dashboard
USER gopocket
EXPOSE 8090
VOLUME ["/app/pb_data"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8090/healthz || exit 1
ENTRYPOINT ["/app/go-pocket-dashboard"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
```

### Dokploy

The recommended path: 10-step Dokploy walkthrough identical to the SaaS plan. Git-push deploys, auto Let's Encrypt, S3 volume backups, scheduled jobs via UI (e.g., `clean-expired-invitations` daily), rolling restarts (zero-downtime).

### Plain Docker

```bash
docker build -t go-pocket-dashboard:latest .
docker run -d --name gpd --env-file .env -p 8090:8090 -v gpd-data:/app/pb_data --restart unless-stopped go-pocket-dashboard:latest
```

### Vercel / Next.js compatibility note

Kiranism's starter deploys to Vercel. We can't, because we need persistent storage for SQLite. Use Dokploy, Fly.io, Railway, or any VPS instead.

---

## 27. Observability, Security & Performance

### Logging

`slog` to stdout (text dev, JSON prod). PocketBase's own logs persist in `auxiliary.db`. Request ID middleware ties everything together.

### Metrics

Optional `/metrics` Prometheus endpoint, gated by admin auth. Counters by route/status, signups, table queries, kanban moves.

### Security headers (Caddy/Traefik or in-app)

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'; script-src 'self' 'nonce-<random>'; style-src 'self'; img-src 'self' data: blob:; connect-src 'self' https://api.polar.sh https://sandbox-api.polar.sh https://api.resend.com https://o<sentry-id>.ingest.sentry.io
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: interest-cohort=(), camera=(), microphone=()
```

### Other security

- CSRF on all state-changing routes (double-submit cookie pattern).
- Rate limiting on `/auth/*` (5/min/IP) and `/command/search` (60/min/IP).
- bcrypt cost 12 for passwords (PB default).
- Open redirect protection on `?return_to=`.
- All form posts validated server-side regardless of client validation.

### Performance budget

- TTFB on `/dashboard` < 100ms (single SQLite query, in-memory plan/RBAC).
- `output.css` < 50KB gzipped.
- Total JS on a typical authenticated page < 120KB (Alpine + HTMX + Chart.js + 3-4 templUI scripts + Sortable on kanban only).
- Lighthouse 95+ on auth pages, 90+ on dashboard (chart canvas counts against LCP).

---

## 28. Testing Strategy

### Levels

1. **Unit** — services, schema validation, `searchparams` codec, RBAC matrix.
2. **Integration** — features end-to-end via `tests.NewTestApp()` in-memory PB.
3. **HTTP** — handler tests with `httptest`; assert HTML fragments via `PuerkitoBio/goquery`.
4. **E2E (optional)** — Playwright suite covering signin → create product → kanban move → upgrade → exclusive page.

### Coverage targets

- Services & utilities: 80%+
- Handlers: 60%+ (focus critical paths)
- Overall: 70%+

### Multi-tenant isolation suite

A dedicated `internal/features/*_isolation_test.go` per feature: spawns two workspaces with two users, attempts cross-workspace access, asserts 403/empty. **Required to pass on every PR.**

### CI

GitHub Actions runs lint + tests on every push and PR. Matrix on `linux/amd64` + `linux/arm64` for build.

---

## 29. Phase-by-Phase Implementation Roadmap

Total estimate: ~28 focused days solo, ~12 days for two devs.

### Phase 0 — Scaffold (Day 1)

- `go mod init github.com/milzamsz/go-pocket`.
- Add deps (PB, templ, Polar, Resend, Sentry, validator, etc.).
- `templui init && templui add "*"`.
- Tailwind config + `input.css` with templUI tokens + 6 theme files.
- `Taskfile.yml`, `.env.example`, `Dockerfile`, `docker-compose.yml`.
- `AGENTS.md` + `.agents/` directory.
- `main.go` boots PocketBase with a stub `OnServe` handler.

**Deliverable:** `task dev` works, `/healthz` returns OK.

### Phase 1 — Layout shell (Days 2-4)

- Base layout (`layout.Auth`, `layout.Dashboard`).
- Sidebar (collapsible, RBAC-aware).
- Topbar (workspace switcher placeholder, notifications stub, theme picker, avatar menu).
- Breadcrumb.
- Infobar component.
- Empty `/dashboard` page renders the shell.
- Dark mode + 6 themes.

**Deliverable:** Beautiful empty shell with full chrome.

### Phase 2 — Auth (Days 5-7)

- Migration: users with extra fields (system_role, theme, locale, last_active_workspace).
- Sign-in / sign-up / forgot / reset / verify pages and handlers.
- OAuth wiring (Google + GitHub).
- Magic link.
- Optional Clerk path (env-toggled).

**Deliverable:** Full auth surface; `/dashboard` redirects to `/auth/signin` when unauthed.

### Phase 3 — Workspaces & RBAC (Days 8-10)

- Migration: workspaces + workspace_members + workspace_invitations.
- Workspaces list page.
- Workspace switcher in topbar.
- Team management page (members table, invite dialog, role select).
- Permission matrix + RBAC middleware.
- Sidebar nav filtering by permission.

**Deliverable:** Multi-tenant skeleton, invite-and-join works.

### Phase 4 — Products feature (Days 11-13)

- Migration: products.
- List page with full DataTable (search, filter, sort, pagination, bulk delete).
- Create / Edit / Delete flows with templ forms + validator.
- Image upload.
- `searchparams` package complete.

**Deliverable:** First fully-featured CRUD feature; sets the pattern for everything after.

### Phase 5 — Overview + Charts (Days 14-15)

- Migration: analytics_events (optional) or aggregated queries.
- Stat cards.
- 4 charts (Area revenue, Bar sales-by-category, Line signups, Pie distribution).
- Recent sales list (auto-refresh).
- Range selector with parallel HTMX loads.

**Deliverable:** Beautiful overview that loads progressively.

### Phase 6 — Kanban (Days 16-17)

- Migration: kanban_columns + kanban_cards.
- Board page rendering columns + cards.
- SortableJS + HTMX reorder.
- Add/edit/delete column + card in Sheet drawer.
- Seed demo board.

**Deliverable:** Working kanban with persistent DnD.

### Phase 7 — Billing + Exclusive (Days 18-20)

- Polar service.
- Pricing table page.
- Current plan card + invoices table.
- Customer portal redirect.
- Webhook handler at `/webhooks/polar`.
- Plan-gating middleware.
- `/dashboard/exclusive` page with locked fallback.

**Deliverable:** End-to-end billing, plan-gated routes.

### Phase 8 — Profile + Settings (Day 21)

- Profile page with 4 tabs (account, security, sessions, danger).
- Avatar upload, password change, 2FA toggle, locale + theme persistence.

**Deliverable:** Profile parity with Clerk's UserProfile.

### Phase 9 — Command Palette (Day 22)

- Action registry + RBAC/plan filtering.
- ⌘K Dialog + Alpine wiring.
- Fuzzy search endpoint.
- Built-in actions (nav + theme switch + workspace switch + sign-out).

**Deliverable:** Working ⌘K with first ~30 actions.

### Phase 10 — Email + transactional (Day 23)

- Resend service.
- Templates (welcome, verify, reset, invite, payment_failed).
- Premailer for CSS inlining.

**Deliverable:** All emails sending and styled.

### Phase 11 — Errors + Not Found + Sentry (Day 24)

- Global error handler + Sentry init.
- `pages.NotFound` + `pages.GlobalError`.
- Recover middleware tags Sentry scope.

**Deliverable:** Production-grade error UX.

### Phase 12 — Tests + observability (Days 25-26)

- Service tests, handler tests, multi-tenant isolation suite.
- slog wiring, request ID, `/healthz`, `/metrics`.

**Deliverable:** ≥70% coverage, observable in prod.

### Phase 13 — Polish + deploy (Days 27-28)

- README + docs (clerk_setup.md, cleanup.md, ARCHITECTURE.md).
- Production Dockerfile finalized + GitHub Actions image push.
- Dokploy deploy to staging then production.
- Demo deploy at `dashboard.go-pocket.dev`.

**Deliverable:** Public v1.0 release.

---

## 30. AI-Assisted Development (AGENTS.md & .agents/)

Same approach as the SaaS plan. `AGENTS.md` at the repo root covers setup, conventions, things-NOT-to-do, and links into `.agents/` for richer context.

### Dashboard-specific AGENTS.md highlights

```markdown
# AGENTS.md — go-pocket-dashboard

> Admin dashboard starter built on Go + PocketBase + templUI.
> Pinned versions: Go 1.26.3, PocketBase v0.38.2, templ v0.3.1020, templUI v1.9.2, Tailwind v4.3.0.

## Feature folder pattern

Every new feature lives in `internal/features/<name>/` with:
- `handlers.go` — HTTP handlers (thin)
- `service.go` — Business logic (talks to PocketBase)
- `components/*.templ` — Feature-specific templ components
- `schemas/*.go` — Validator structs (form schemas + search-param structs)

NEVER spread a feature across `internal/handlers/<name>.go` + `internal/services/<name>.go` + `components/<name>/`. Keep it co-located.

## Data tables

Use the generic `@layout.DataTable()` composite. Server-side everything. URL state via `internal/searchparams`.

## Forms

Schema = Go struct + validator tags. Render with `@layout.FormField()`. HTMX-submit with `hx-post` + `hx-swap="outerHTML"` and re-render the form with errors on validation failure.

## Command palette

Register actions in your feature's `init()` via `command.Register(...)`. Always set the appropriate `Permission` and (if relevant) `PlanGate`.

## Things NOT to do

- Don't use React, Next.js patterns, or client-side routing. We're server-rendered.
- Don't ship Recharts. Use Chart.js via templUI's Charts component.
- Don't ship dnd-kit. Use SortableJS for kanban.
- Don't bypass `internal/searchparams` — every URL parameter goes through a typed schema.
- Don't add a new theme without registering it in `internal/theme/themes.go`.
- Don't use Clerk components in templ files even if `AUTH_PROVIDER=clerk`. The auth flow is handled centrally; UI stays templUI-native.
```

### `.agents/prompts/` for this starter

- `add-feature.md` — "Create a new feature module" (folder scaffold + first handler/service/page).
- `add-page.md` — "Add a new page under /dashboard/*" (route, handler, page templ, sidebar nav entry).
- `add-data-table.md` — "Add a server-driven data table" (DataTable props + schema + service method).
- `add-form.md` — "Add a form with validation" (schema + form templ + handler + error rendering).
- `add-chart.md` — "Add a chart to overview" (service method + chart templ + parallel HTMX load).
- `add-command-action.md` — "Register a new command palette action".
- `add-kanban-customization.md` — "Customize kanban cards/columns".

---

## 31. Cleanup Guide (remove demo data)

Mirrors Kiranism's `__CLEANUP__/cleanup.md`. After cloning, before customizing for your domain:

1. **Demo seed data:** Delete `cmd/tools/seed/` and remove the `task seed` invocation from `task setup`.
2. **Sample features:** If you don't need products or kanban, delete the corresponding `internal/features/<name>/` and `migrations/170XXX_init_<name>.go`. Remove sidebar entries.
3. **Themes:** Keep `default.css`; delete the other five if you only want one.
4. **Sample workspaces:** The seeded "Acme" and "Globex" workspaces can be removed via PB admin UI or by editing `cmd/tools/seed/users.go`.
5. **README:** Replace the README with your own product description.
6. **Branding:** Update `assets/img/logo.svg`, `favicon.ico`, `og.png`. Adjust the `APP_NAME` env var.

Track checklist in `docs/cleanup.md`.

---

## 32. Appendix: Key Code Snippets

### `main.go`

```go
package main

import (
    "log"

    "github.com/milzamsz/go-pocket/internal/app"
    "github.com/pocketbase/pocketbase"

    _ "github.com/milzamsz/go-pocket/migrations"
)

func main() {
    pb := pocketbase.New()

    if err := app.Bootstrap(pb); err != nil {
        log.Fatal(err)
    }
    if err := pb.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### `internal/searchparams/parser.go` (essence)

```go
package searchparams

import (
    "net/http"
    "net/url"
    "reflect"
    "strconv"
    "strings"
)

// Parse decodes URL query into a typed struct using `param:"..."` tags,
// applies `default:"..."` values, and validates `allowed:"..."` allowlists.
func Parse[T any](r *http.Request) (T, error) {
    var out T
    q := r.URL.Query()
    if err := decodeInto(reflect.ValueOf(&out).Elem(), q); err != nil {
        return out, err
    }
    return out, nil
}

// Encode produces a URL from a struct, omitting fields equal to defaults.
func Encode[T any](base string, params T) string {
    v := url.Values{}
    encodeFrom(reflect.ValueOf(params), v)
    if len(v) == 0 { return base }
    return base + "?" + v.Encode()
}

// ...decodeInto + encodeFrom omitted for brevity; the package supports:
// - int / int64 / string / bool / []string (CSV)
// - default:"..." / min:"..." / max:"..." / oneof:"a b c" / allowed:"a b c"
// - pointer fields for tri-state filters
```

### `internal/server/routes.go` (skeleton)

```go
func RegisterRoutes(r *core.Router, pb *pocketbase.PocketBase, d *Deps) {
    // Global
    r.BindFunc(middleware.RequestID, middleware.Logger, middleware.Recover, middleware.SecurityHeaders)

    // Assets
    registerAssets(r, d)

    // Public marketing/landing (optional)
    r.GET("/", marketingHome(d))

    // Auth
    authPub := r.Group("/auth").BindFunc(middleware.RedirectIfAuthenticated(pb))
    authPub.GET("/signin", auth.SignInPage(d))
    authPub.POST("/signin", auth.SignIn(pb, d))
    authPub.GET("/signup", auth.SignUpPage(d))
    authPub.POST("/signup", auth.SignUp(pb, d))
    authPub.GET("/forgot-password", auth.ForgotPasswordPage(d))
    authPub.POST("/forgot-password", auth.ForgotPassword(pb, d))
    authPub.GET("/reset-password", auth.ResetPasswordPage(d))
    authPub.POST("/reset-password", auth.ResetPassword(pb, d))
    authPub.GET("/oauth/{provider}", auth.OAuthRedirect(pb))
    authPub.GET("/oauth/{provider}/callback", auth.OAuthCallback(pb, d))
    r.GET("/auth/logout", auth.Logout(pb), middleware.RequireAuth(pb))

    // Dashboard
    dash := r.Group("/dashboard").BindFunc(
        middleware.RequireAuth(pb),
        middleware.ResolveActiveWorkspace(pb),
        middleware.LoadRBACContext,
        middleware.LoadTheme,
        middleware.Telemetry,
    )
    dash.GET("", overview.Page(d))
    dash.GET("/products", products.List(d))
    dash.GET("/products/new", products.NewPage(d), middleware.RequirePermission(rbac.PermProductsCreate))
    dash.POST("/products", products.Create(d), middleware.RequirePermission(rbac.PermProductsCreate))
    dash.GET("/products/{id}", products.EditPage(d))
    dash.POST("/products/{id}", products.Update(d), middleware.RequirePermission(rbac.PermProductsUpdate))
    dash.DELETE("/products/{id}", products.Delete(d), middleware.RequirePermission(rbac.PermProductsDelete))
    dash.GET("/kanban", kanban.Page(d))
    dash.GET("/workspaces", workspaces.List(d))
    dash.GET("/workspaces/team", workspaces.Team(d))
    dash.GET("/billing", billing.Page(d), middleware.RequirePermission(rbac.PermBillingView))
    dash.GET("/exclusive", exclusive.Page(d), middleware.RequirePlan("pro"))
    dash.GET("/profile", profile.Page(d))

    // HTMX-only fragments
    parts := r.Group("/dashboard/_partials").BindFunc(
        middleware.RequireAuth(pb),
        middleware.ResolveActiveWorkspace(pb),
        middleware.LoadRBACContext,
        middleware.RequireHXRequest,
    )
    parts.GET("/overview/{card}", overview.Card(d))
    parts.GET("/products/table", products.TableFragment(d))
    parts.POST("/kanban/reorder", kanban.Reorder(d))

    // Command palette
    r.GET("/command/search", command.Search(d), middleware.RequireAuth(pb))

    // Webhooks
    r.POST("/webhooks/polar", webhooks.Polar(d))
    r.POST("/webhooks/resend", webhooks.Resend(d))

    // Errors
    r.GET("/healthz", func(e *core.RequestEvent) error { return e.String(200, "ok") })
}
```

### `components/layout/dashboard.templ` (essence)

```templ
package layout

import (
    "github.com/milzamsz/go-pocket/components/ui/dialog"
    "github.com/milzamsz/go-pocket/components/ui/dropdown"
    "github.com/milzamsz/go-pocket/components/ui/popover"
    "github.com/milzamsz/go-pocket/components/ui/sheet"
    "github.com/milzamsz/go-pocket/components/ui/toast"
)

type DashboardProps struct {
    Title      string
    Crumbs     []Crumb
    Infobar    *InfobarProps
    Actions    templ.Component       // Optional right-aligned actions
    Theme      string
    ColorMode  string                 // light | dark
    SidebarCollapsed bool
}

templ Dashboard(p DashboardProps) {
    @Base(BaseMeta{ Title: p.Title + " · Acme", Theme: p.Theme, ColorMode: p.ColorMode }) {
        <div class="flex min-h-screen" x-data="{ sidebar: $store.ui.sidebarCollapsed }">
            @Sidebar(SidebarProps{ Collapsed: p.SidebarCollapsed })
            <div class="flex flex-1 flex-col">
                @Topbar()
                if p.Infobar != nil {
                    @Infobar(*p.Infobar)
                }
                <main class="flex-1 overflow-y-auto p-6">
                    @PageHeader(PageHeaderProps{ Title: p.Title, Crumbs: p.Crumbs, Actions: p.Actions })
                    { children... }
                </main>
            </div>
        </div>
        @CommandPalette()
        @toast.Script()
        @dialog.Script()
        @dropdown.Script()
        @popover.Script()
        @sheet.Script()
    }
}
```

### `internal/features/products/handlers.go` (essence)

```go
package products

import (
    "net/http"

    "github.com/milzamsz/go-pocket/components/pages"
    "github.com/milzamsz/go-pocket/internal/features/products/schemas"
    "github.com/milzamsz/go-pocket/internal/searchparams"
    "github.com/milzamsz/go-pocket/internal/server"

    "github.com/labstack/echo/v5"
)

func List(d *server.Deps) echo.HandlerFunc {
    return func(c echo.Context) error {
        params, err := searchparams.Parse[schemas.ProductListParams](c.Request())
        if err != nil { return badRequest(c, err) }

        workspaceID := c.Get("workspace_id").(string)
        rows, total, err := d.Products.List(c.Request().Context(), workspaceID, params)
        if err != nil { return internalError(c, err) }

        if isHX(c) {
            return c.Render(http.StatusOK, ProductsTable(rows, params, total))
        }
        return c.Render(http.StatusOK, pages.ProductsList(rows, params, total))
    }
}

func Create(d *server.Deps) echo.HandlerFunc {
    return func(c echo.Context) error {
        var form schemas.ProductCreate
        if err := c.Bind(&form); err != nil { return badRequest(c, err) }
        if errs := d.Validator.Struct(&form); errs != nil {
            return c.Render(http.StatusUnprocessableEntity, ProductForm(form, fieldErrors(errs)))
        }
        workspaceID := c.Get("workspace_id").(string)
        if _, err := d.Products.Create(c.Request().Context(), workspaceID, &form); err != nil {
            return internalError(c, err)
        }
        c.Response().Header().Set("HX-Redirect", "/dashboard/products?created=1")
        return c.NoContent(http.StatusOK)
    }
}
```

---

## Closing notes

This blueprint is a true 1:1 port — every page, every feature, every developer ergonomic from Kiranism's Next.js starter has a documented Go equivalent above. Where the Next.js world reaches for a library (Recharts, dnd-kit, kbar, Zustand, Nuqs, TanStack Table, react-hook-form), this plan names the Go/HTMX/Alpine replacement and explains the wiring.

When implementation begins, follow the 28-day roadmap in §29 and keep the §2 mapping table updated as parity changes. Resist the urge to deviate from the feature-folder convention — discoverability is the silent superpower of this stack.

Ship something. Then iterate.
