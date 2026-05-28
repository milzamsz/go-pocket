# go-pocket — Comprehensive Boilerplate Blueprint

> A production-ready, **multi-tenant** Go SaaS starter, 1:1 in spirit with [goilerplate.com](https://goilerplate.com), built on **PocketBase** (instead of Postgres) and a **direct port of [templUI](https://github.com/templui/templui)** components. Payments via **Polar.sh**, email via **Resend**, deploys via **Docker / Dokploy**, instrumented for AI-assisted development with **AGENTS.md**.

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Design Philosophy](#2-design-philosophy)
3. [Technology Stack](#3-technology-stack)
4. [High-Level Architecture](#4-high-level-architecture)
5. [PocketBase Integration Strategy](#5-pocketbase-integration-strategy)
6. [Project Folder Structure](#6-project-folder-structure)
7. [Data Model & PocketBase Collections](#7-data-model--pocketbase-collections)
8. [Multi-Tenancy Model (Organizations)](#8-multi-tenancy-model-organizations)
9. [Authentication & Authorization](#9-authentication--authorization)
10. [UI Component Library (templUI port)](#10-ui-component-library-templui-port)
11. [Page & Route Inventory](#11-page--route-inventory)
12. [Feature Modules](#12-feature-modules)
13. [Configuration & Environment](#13-configuration--environment)
14. [Development Workflow](#14-development-workflow)
15. [Build, Asset Pipeline & Embed](#15-build-asset-pipeline--embed)
16. [Deployment (Docker / Dokploy)](#16-deployment-docker--dokploy)
17. [Observability, Security & Performance](#17-observability-security--performance)
18. [Testing Strategy](#18-testing-strategy)
19. [Phase-by-Phase Implementation Roadmap](#19-phase-by-phase-implementation-roadmap)
20. [Documentation Site (1:1 with goilerplate.com/docs)](#20-documentation-site-11-with-goilerplatecomdocs)
21. [AI-Assisted Development (AGENTS.md & .agents/)](#21-ai-assisted-development-agentsmd--agents)
22. [Appendix: Key Code Snippets](#22-appendix-key-code-snippets)
23. [Admin Dashboard Starter Parity Plan (templUI)](#23-admin-dashboard-starter-parity-plan-templui)

---

## 1. Executive Summary

**go-pocket** is an opinionated, batteries-included Go boilerplate for shipping production **multi-tenant** SaaS applications. It mirrors the developer experience of [Goilerplate](https://goilerplate.com) (Go + templ + Alpine + Tailwind + Echo) but replaces the traditional Postgres + GORM stack with **PocketBase** as the embedded data layer — yielding a single-binary, single-file-database, container-deployable application that still scales to thousands of tenants.

### What you get out of the box

- One Go binary, one Docker image. Deploy via **Dokploy** (open-source PaaS) or any Docker host in minutes.
- A 1:1 port of all **40+ templUI components** (Accordion → Tooltip) plus marketing/landing primitives in the spirit of Goilerplate.
- First-class **PocketBase** integration: collections, hooks, OAuth, file storage, real-time, migrations, admin dashboard — all from Go code.
- **Multi-tenant from day one**: organizations, members, invitations, role-based per-org permissions, org-scoped data isolation, org-scoped billing.
- Full SaaS feature set: auth, billing (**Polar.sh** — open-source, MoR), transactional email (**Resend**), admin panel, blog, docs site, dark mode, i18n, SEO, HTMX + Alpine interactivity.
- A documentation site (`/docs`) that mirrors the structure of `goilerplate.com/docs` for instant familiarity.
- **AI-ready**: ships with `AGENTS.md` + `.agents/` directory so Cursor, Claude Code, Aider, Codex, Gemini CLI, and Copilot can contribute productively from commit #1.

### Who it's for

Solo founders and small teams who want **Go's reliability + PocketBase's simplicity + Goilerplate's polish + true multi-tenancy**, without paying for a license and without surrendering ownership of their stack.

### Non-goals

- Not a microservices framework. Single binary by design.
- Not a "config-driven" framework. You write Go; the boilerplate composes itself.
- Not opinionated about your domain logic — only about the plumbing.

---

## 2. Design Philosophy

Five principles, borrowed directly from templUI's shadcn-inspired ethos and PocketBase's minimalism:

1. **You own the code.** Like shadcn/ui and templUI, every component, every handler, every collection schema lives in your repo. No hidden magic, no vendor lock-in.
2. **One binary to rule them all.** PocketBase's superpower is its self-contained Go binary. We don't break it. CSS, JS, templates, migrations, assets — all embedded with `//go:embed`.
3. **Server-first, JS-second.** Templ renders on the server. Alpine.js handles micro-interactions. HTMX handles partial swaps. JavaScript is a progressive enhancement, never a prerequisite.
4. **CSP-compliant by default.** templUI ships without `eval`, without inline scripts. We keep that posture — every JS file is hashed and served from `/assets/js/`.
5. **Convention over configuration.** Folder structure, naming, route patterns, collection schemas — all consistent. Once you know one feature, you know them all.

---

## 3. Technology Stack

| Layer | Choice | Why |
|---|---|---|
| **Language** | **Go 1.26.3** | Latest stable; modern stdlib, `slog`, `go:embed`, range-over-func iterators. |
| **Backend framework** | [PocketBase](https://github.com/pocketbase/pocketbase) **v0.38.2** (embedded as Go framework) | One binary, built-in auth/storage/realtime, uses Echo internally — no second router needed. |
| **HTTP router** | PocketBase's built-in Echo router (`se.Router`) | Zero extra dependency. Goilerplate uses Echo; PocketBase IS Echo under the hood. |
| **Template engine** | [a-h/templ](https://templ.guide) **v0.3.1020** | Type-safe Go-native components; what templUI is built on. |
| **CSS** | Tailwind CSS **v4.3.0** (standalone CLI) | templUI requirement; OKLCH-based theme tokens. |
| **UI components** | [templUI](https://github.com/templui/templui) **v1.9.2** | 40+ shadcn-style components; direct port into this repo via `templui add "*"`. |
| **Interactivity (client)** | **Alpine.js v3.15.8** + **HTMX v2.0.x** | Goilerplate's stack; templUI scripts assume vanilla DOM. |
| **Database** | SQLite (via PocketBase, `modernc.org/sqlite` pure-Go) | No CGO, cross-compile freely; WAL mode for concurrent reads. |
| **Auth** | PocketBase native (email/password, OAuth2, OTP, MFA) | Battle-tested; we add only UI + flows. |
| **Tenancy** | Custom `organizations` + `organization_members` + `invitations` collections | Org-scoped data isolation, per-org roles, invitation links. |
| **Payments** | [Polar.sh](https://polar.sh) (`github.com/polarsource/polar-go` **v0.7.0**) | Open-source MoR (handles VAT/tax). Subscriptions, customer portal, webhooks → PocketBase collections. Per-org billing. |
| **Email** | [Resend](https://resend.com) (`github.com/resend/resend-go/v3` **v3.7.0**) | Modern transactional API; React/templ-friendly templates; deliverability built-in. |
| **File storage** | PocketBase's filesystem (local or S3-compatible) | Per-field, with thumbnails, signed URLs. |
| **Task runner** | [go-task/task](https://taskfile.dev) | templUI's standard; replaces Makefiles. |
| **Hot reload (dev)** | `templ generate --watch` + `tailwindcss --watch` | Same as templUI. |
| **Logging** | Go stdlib `log/slog` + PocketBase logs | Structured logs; PB stores in `auxiliary.db`. |
| **i18n** | `go-i18n/v2` + per-locale `*.toml` | Lightweight; ICU pluralization. |
| **Migrations** | PocketBase Go migrations (`migrations/` package) | Versioned, reversible, embedded. |
| **Testing** | `testing` + `testify` + `testcontainers` (optional) | PocketBase has a `tests` package for fixtures. |

### What we deliberately exclude

- **No GORM / sqlc / ent.** PocketBase wraps all data access; you use `app.FindRecordById`, `app.RecordQuery`, etc.
- **No Node.js at runtime.** Tailwind CLI is a single binary; templ generates Go code. Zero Node deps in production.
- **No Redis** (until you actually need it). PocketBase + SQLite handles sessions, jobs, and queues for small/medium scale.

### Pinned versions (May 2026 baseline)

Reference `go.mod` for the boilerplate:

```go
module github.com/milzamsz/go-pocket

go 1.26.3

require (
    github.com/pocketbase/pocketbase v0.38.2
    github.com/a-h/templ v0.3.1020
    github.com/polarsource/polar-go v0.7.0
    github.com/resend/resend-go/v3 v3.7.0
    github.com/vanng822/go-premailer v1.24.0
    github.com/nicksnyder/go-i18n/v2 v2.6.0
    github.com/getsentry/sentry-go v0.32.0
    github.com/stretchr/testify v1.10.0
    github.com/spf13/pflag v1.0.6
    golang.org/x/crypto v0.34.0
)
```

Tooling versions (installed separately):

| Tool | Version | Install |
|---|---|---|
| Go | 1.26.3 | `https://go.dev/dl/` |
| templ CLI | v0.3.1020 | `go install github.com/a-h/templ/cmd/templ@v0.3.1020` |
| Task | v3.49.1 | `go install github.com/go-task/task/v3/cmd/task@v3.49.1` |
| Tailwind CSS standalone | v4.3.0 | Download binary from `tailwindlabs/tailwindcss` releases |
| templUI CLI | v1.9.2 | `go install github.com/templui/templui/cmd/templui@v1.9.2` |
| Alpine.js (loaded in browser) | v3.15.8 | Served from `/assets/js/alpine.min.js` |
| HTMX (loaded in browser) | v2.0.x | Served from `/assets/js/htmx.min.js` |
| Docker base image | `golang:1.26.3-alpine3.23` build → `alpine:3.23.3` runtime | — |

> **Bumping policy:** Pin to exact versions in `go.mod`. Renovate/Dependabot bumps go through PR + CI. Patch bumps are auto-mergeable after green CI; minor/major bumps need a human review.

---

## 4. High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Single Go Binary (go-pocket)                 │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │                  HTTP Layer (Echo via PocketBase)             │   │
│  │                                                               │   │
│  │  /                  /docs              /api/*    /_/         │   │
│  │  (marketing)        (component docs)   (PB REST) (PB Admin)  │   │
│  │  /app/*  /org/:slug/*  /auth/*  /webhooks/polar  /webhooks/resend │
│  │  (dashboard)        (login/signup)                           │   │
│  └─────────────┬──────────────────────────────────┬─────────────┘   │
│                │                                   │                 │
│  ┌─────────────▼─────────────────┐   ┌────────────▼──────────────┐   │
│  │   Templ Components Layer      │   │   PocketBase Core         │   │
│  │   (templUI port + app pages)  │   │   - Collections           │   │
│  │   - components/ui/* (40+)     │   │   - Event hooks           │   │
│  │   - components/marketing/*    │   │   - Auth (email/OAuth)    │   │
│  │   - components/app/*          │   │   - Realtime SSE          │   │
│  │   - layouts/*                 │   │   - File storage          │   │
│  └─────────────┬─────────────────┘   └────────────┬──────────────┘   │
│                │                                   │                 │
│  ┌─────────────▼───────────────────────────────────▼─────────────┐   │
│  │              Domain / Service Layer (internal/)               │   │
│  │  auth · billing · email · storage · admin · blog · feature… │   │
│  └────────────────────────────┬──────────────────────────────────┘   │
│                               │                                      │
│  ┌────────────────────────────▼──────────────────────────────────┐   │
│  │              Embedded Assets (//go:embed)                     │   │
│  │  CSS · JS · fonts · images · email templates · migrations    │   │
│  └────────────────────────────┬──────────────────────────────────┘   │
│                               │                                      │
│  ┌────────────────────────────▼──────────────────────────────────┐   │
│  │              SQLite (pb_data/data.db + auxiliary.db)          │   │
│  └───────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────┘
                               │
                ┌──────────────┼──────────────┐
                ▼              ▼              ▼
            Polar.sh API    Resend API    S3 (optional)
            (billing)       (email)       (storage)
```

### Request lifecycle

1. Browser hits `https://app.example.com/dashboard`.
2. PocketBase's Echo router matches the route registered in `internal/server/routes.go`.
3. Middleware chain runs: request ID → logger → auth (reads JWT cookie, hydrates user record) → CSRF → rate limit.
4. Handler queries PocketBase collections via `app.FindRecordById`, builds a view model.
5. Handler renders a `templ` component which composes templUI primitives + page-specific markup.
6. Response streams HTML; `output.css` and per-component `*.min.js` are loaded from `/assets/`.

---

## 5. PocketBase Integration Strategy

### The chosen approach: **Embedded with a clean service layer**

You asked "what's best for scalability & best practice" — here's the recommendation and rationale:

**Embed PocketBase as a Go framework** (the [Go Overview](https://pocketbase.io/docs/go-overview/) pattern), and **wrap all data access in a thin service layer** under `internal/services/*`. This gives you the best of both worlds:

| Concern | How we solve it |
|---|---|
| **Single-binary deploy** | ✅ PocketBase is imported as a Go module; everything compiles to one executable. |
| **Scalability** | ✅ PocketBase + WAL-mode SQLite handles 10k+ concurrent reads on a $5 VPS. For write-heavy workloads, route writes through a single goroutine and use `LiteFS` or upgrade to managed Postgres later — the service layer absorbs the swap. |
| **Best-practice abstraction** | ✅ Handlers never touch `app.Dao()` directly. They call `services.User.Find(ctx, id)`, which internally uses PocketBase. Swap implementations without rewriting handlers. |
| **Testability** | ✅ Services accept an interface (`UserRepository`); tests use an in-memory fake. PocketBase itself provides `tests.NewTestApp()` for integration tests. |
| **Avoid coupling pitfall** | ✅ We treat PocketBase as our "database driver", not our application framework. Business logic lives in services, not in PB hooks (hooks are thin glue only). |

### Hook usage philosophy

PocketBase hooks (`OnRecordCreate`, `OnRecordUpdate`, `OnServe`, etc.) are used **only** for:

- Route registration (`OnServe`).
- Side effects tightly coupled to a record's lifecycle (e.g., send welcome email on `OnRecordCreate("users")`).
- Cross-cutting validation that must run before persistence.

Everything else lives in services, which is **the** unit of business logic.

### Scaling path

| Stage | Setup | Capacity |
|---|---|---|
| MVP | Single binary on a $5 VPS, local SQLite | 1k-10k MAU |
| Growth | Add LiteFS for SQLite replication; CDN in front of `/assets/*` | 10k-100k MAU |
| Scale | Migrate writes to managed Postgres via service-layer swap; keep PB for auth/storage | 100k+ MAU |

The service-layer abstraction makes step 3 surgical, not catastrophic.

---

## 6. Project Folder Structure

```
go-pocket/
├── main.go                          # Entry point: pocketbase.New() + register routes/hooks
├── go.mod                           # module github.com/milzamsz/go-pocket
├── go.sum
├── Taskfile.yml                     # task dev | task build | task migrate
├── .env.example                     # Documented env vars
├── .gitignore
├── .dockerignore
├── Dockerfile                       # Production multi-stage build (single static binary)
├── docker-compose.yml               # Local + Dokploy-compatible compose
├── dokploy.yaml                     # Optional Dokploy deployment manifest
├── LICENSE
├── README.md
├── CHANGELOG.md
├── AGENTS.md                        # Top-level instructions for AI coding agents
├── .agents/                         # AI agent rules, prompts, and tool configs
│   ├── rules.md                     # Shared rules (style, testing, PR conventions)
│   ├── architecture.md              # Architecture summary for agent context
│   ├── conventions.md               # File naming, error patterns, templ patterns
│   ├── prompts/                     # Reusable agent prompts
│   │   ├── add-feature.md
│   │   ├── add-migration.md
│   │   ├── add-templui-component.md
│   │   ├── add-page.md
│   │   └── debug.md
│   └── tools/                       # Per-tool config (compat layer)
│       ├── claude.md                # → symlinked from CLAUDE.md
│       ├── cursor.mdc               # → symlinked from .cursorrules
│       ├── aider.yml                # → .aider.conf.yml
│       └── gemini.json              # → .gemini/settings.json
│
├── cmd/
│   └── tools/                       # One-off CLI commands (e.g., seed, hash-password)
│       └── seed/main.go
│
├── internal/                        # All private application code
│   ├── app/                         # PocketBase instance bootstrap
│   │   ├── app.go                   # NewApp(): wires services, hooks, routes
│   │   └── lifecycle.go             # OnBootstrap, OnTerminate handlers
│   │
│   ├── config/
│   │   ├── config.go                # Struct + envconfig parsing
│   │   └── flags.go                 # CLI flag overrides
│   │
│   ├── server/
│   │   ├── routes.go                # All route registrations in one place
│   │   ├── middleware/
│   │   │   ├── auth.go              # Cookie/JWT → user record
│   │   │   ├── csrf.go
│   │   │   ├── ratelimit.go
│   │   │   ├── requestid.go
│   │   │   ├── logger.go
│   │   │   └── recover.go
│   │   └── handlers/                # HTTP handlers (thin; delegate to services)
│   │       ├── marketing/
│   │       │   ├── home.go
│   │       │   ├── pricing.go
│   │       │   ├── blog.go
│   │       │   └── docs.go
│   │       ├── auth/
│   │       │   ├── login.go
│   │       │   ├── signup.go
│   │       │   ├── oauth.go
│   │       │   ├── reset.go
│   │       │   └── verify.go
│   │       ├── app/                 # Authenticated app (dashboard) handlers
│   │       │   ├── dashboard.go
│   │       │   ├── settings.go
│   │       │   ├── onboarding.go    # Create first org on first login
│   │       │   └── org_switcher.go  # Switch active org
│   │       ├── org/                 # ★ Org-scoped handlers (require active org)
│   │       │   ├── overview.go
│   │       │   ├── members.go       # List, invite, remove, change role
│   │       │   ├── invitations.go   # Accept/decline invitation links
│   │       │   ├── billing.go       # Per-org Polar subscription
│   │       │   ├── settings.go      # Org name, slug, logo, danger zone
│   │       │   └── audit.go         # Org-scoped audit log
│   │       ├── admin/               # Custom admin UI (separate from PB's /_/)
│   │       │   ├── users.go
│   │       │   ├── analytics.go
│   │       │   └── settings.go
│   │       └── webhooks/
│   │           └── polar.go         # Polar.sh webhook (signature-verified)
│   │
│   ├── services/                    # Domain services — the ONLY layer that touches PocketBase
│   │   ├── users/
│   │   │   ├── service.go           # User.Find / Create / Update / Delete
│   │   │   ├── repository.go        # Interface + PB implementation
│   │   │   └── dto.go
│   │   ├── auth/
│   │   │   ├── service.go           # Login, signup, OAuth, sessions
│   │   │   ├── password.go
│   │   │   └── tokens.go
│   │   ├── tenancy/                 # ★ Multi-tenant core
│   │   │   ├── service.go           # Org CRUD, member mgmt, role checks
│   │   │   ├── invitations.go       # Invite by email, accept/decline, expiry
│   │   │   ├── switching.go         # Active-org cookie + middleware helpers
│   │   │   ├── permissions.go       # Org-role × resource matrix
│   │   │   └── repository.go        # Interface + PB implementation
│   │   ├── billing/                 # Polar.sh integration
│   │   │   ├── service.go           # Polar subscription orchestration
│   │   │   ├── polar_client.go      # Wraps polargo SDK
│   │   │   ├── webhook_handler.go   # Polar webhook signature + dispatcher
│   │   │   ├── checkout.go          # Create checkout sessions
│   │   │   ├── customer_portal.go   # Create customer sessions
│   │   │   └── plans.go             # Product/price catalog
│   │   ├── email/                   # Resend integration
│   │   │   ├── service.go           # Resend client + send/batch/queue
│   │   │   ├── resend_client.go     # Wraps resend-go/v3
│   │   │   ├── templates.go         # Renders templ email components
│   │   │   └── templates/
│   │   │       ├── welcome.templ
│   │   │       ├── verify.templ
│   │   │       ├── reset.templ
│   │   │       ├── invite.templ              # Org invitation
│   │   │       ├── subscription_started.templ
│   │   │       ├── payment_failed.templ
│   │   │       └── weekly_digest.templ
│   │   ├── storage/
│   │   │   └── service.go           # Wraps PocketBase filesystem; signed URLs
│   │   ├── content/                 # Blog + docs (file-based MDX-like)
│   │   │   ├── service.go
│   │   │   └── parser.go
│   │   └── analytics/
│   │       └── service.go           # Track events → analytics_events collection
│   │
│   ├── domain/                      # Pure structs and enums (no PB dependency)
│   │   ├── user.go
│   │   ├── organization.go
│   │   ├── membership.go
│   │   ├── invitation.go
│   │   ├── subscription.go
│   │   ├── plan.go
│   │   └── errors.go
│   │
│   ├── i18n/
│   │   ├── i18n.go                  # Loader + middleware
│   │   └── locales/
│   │       ├── en.toml
│   │       └── id.toml
│   │
│   ├── version/
│   │   └── version.go               # Set via ldflags at build time
│   │
│   └── testutil/
│       ├── app.go                   # Spawns a test PocketBase instance
│       └── fixtures.go
│
├── components/                      # ALL templ components (UI + pages)
│   ├── ui/                          # 1:1 port of templUI
│   │   ├── accordion/
│   │   │   ├── accordion.templ
│   │   │   └── accordion.go         # Helpers
│   │   ├── alert/
│   │   ├── aspectratio/
│   │   ├── avatar/
│   │   ├── badge/
│   │   ├── breadcrumb/
│   │   ├── button/
│   │   ├── calendar/
│   │   ├── card/
│   │   ├── carousel/
│   │   ├── charts/
│   │   ├── checkbox/
│   │   ├── collapsible/
│   │   ├── copybutton/
│   │   ├── datepicker/
│   │   ├── dialog/
│   │   ├── dropdown/
│   │   ├── form/
│   │   ├── icon/
│   │   ├── input/
│   │   ├── inputotp/
│   │   ├── label/
│   │   ├── pagination/
│   │   ├── popover/
│   │   ├── progress/
│   │   ├── radio/
│   │   ├── rating/
│   │   ├── selectbox/
│   │   ├── separator/
│   │   ├── sheet/
│   │   ├── sidebar/
│   │   ├── skeleton/
│   │   ├── slider/
│   │   ├── switch/
│   │   ├── table/
│   │   ├── tabs/
│   │   ├── tagsinput/
│   │   ├── textarea/
│   │   ├── timepicker/
│   │   ├── toast/
│   │   └── tooltip/
│   ├── marketing/                   # Goilerplate-style marketing primitives
│   │   ├── hero.templ
│   │   ├── features.templ
│   │   ├── pricing_table.templ
│   │   ├── testimonial.templ
│   │   ├── faq.templ
│   │   ├── cta.templ
│   │   ├── logo_cloud.templ
│   │   ├── footer.templ
│   │   └── navbar.templ
│   ├── app/                         # Authenticated app shells
│   │   ├── sidebar_nav.templ
│   │   ├── topbar.templ
│   │   ├── breadcrumb.templ
│   │   ├── user_menu.templ
│   │   ├── org_switcher.templ       # ★ Active-org dropdown in topbar
│   │   ├── member_avatar_group.templ
│   │   ├── role_badge.templ
│   │   └── empty_state.templ
│   ├── icons/                       # Lucide icons as templ components (auto-generated)
│   │   └── icons.templ
│   ├── pages/                       # Full-page templates
│   │   ├── marketing/
│   │   │   ├── home.templ
│   │   │   ├── pricing.templ
│   │   │   ├── about.templ
│   │   │   ├── contact.templ
│   │   │   ├── blog_index.templ
│   │   │   ├── blog_post.templ
│   │   │   ├── docs_layout.templ
│   │   │   └── docs_page.templ
│   │   ├── auth/
│   │   │   ├── login.templ
│   │   │   ├── signup.templ
│   │   │   ├── forgot_password.templ
│   │   │   ├── reset_password.templ
│   │   │   └── verify_email.templ
│   │   ├── app/
│   │   │   ├── dashboard.templ
│   │   │   ├── onboarding_create_org.templ   # ★ First-time org setup
│   │   │   ├── settings_profile.templ        # User-level settings only
│   │   │   ├── settings_security.templ
│   │   │   ├── settings_account.templ        # Email, locale, theme
│   │   │   └── notifications.templ
│   │   ├── org/                              # ★ Org-scoped pages
│   │   │   ├── overview.templ
│   │   │   ├── members.templ                 # Members table + invite modal
│   │   │   ├── invitations_pending.templ
│   │   │   ├── billing.templ                 # Plans + current subscription
│   │   │   ├── billing_invoices.templ
│   │   │   ├── settings_general.templ        # Name, slug, logo
│   │   │   ├── settings_danger.templ         # Transfer ownership, delete org
│   │   │   ├── invite_accept.templ           # /invite/:token landing
│   │   │   └── audit_log.templ
│   │   ├── admin/
│   │   │   ├── users.templ
│   │   │   ├── analytics.templ
│   │   │   └── settings.templ
│   │   └── errors/
│   │       ├── 404.templ
│   │       ├── 500.templ
│   │       └── maintenance.templ
│   └── layouts/
│       ├── base.templ               # <html>, <head>, scripts, fonts
│       ├── marketing.templ
│       ├── auth.templ
│       ├── app.templ                # Sidebar + topbar shell (no active org)
│       ├── org.templ                # ★ Org-scoped shell (with org switcher)
│       └── admin.templ
│
├── assets/                          # Source assets (compiled/embedded into binary)
│   ├── css/
│   │   ├── input.css                # Tailwind entry (templUI theme tokens)
│   │   ├── output.css               # Generated; embedded
│   │   └── sources.generated.css    # Generated by Taskfile in import workflow
│   ├── js/                          # Per-component .min.js (from templUI add)
│   │   ├── datepicker.min.js
│   │   ├── calendar.min.js
│   │   ├── dropdown.min.js
│   │   ├── ... (one per interactive component)
│   │   ├── alpine.min.js
│   │   └── htmx.min.js
│   ├── img/
│   │   ├── logo.svg
│   │   ├── favicon.ico
│   │   ├── og.png
│   │   └── ...
│   ├── fonts/
│   │   └── inter/...                # Self-hosted, CSP-friendly
│   └── embed.go                     # //go:embed all:./...
│
├── migrations/                      # PocketBase Go migrations
│   ├── 1700000000_init_users.go
│   ├── 1700000050_init_organizations.go     # ★ orgs, members, invitations
│   ├── 1700000100_init_subscriptions.go     # Org-scoped
│   ├── 1700000200_init_invoices.go          # Org-scoped
│   ├── 1700000300_init_posts.go             # Blog
│   ├── 1700000400_init_analytics.go
│   ├── 1700000500_init_audit_log.go         # Org-scoped
│   └── seed/
│       └── seed.go                  # Dev fixtures (incl. demo org + members)
│
├── content/                         # File-based blog & docs (MDX-like)
│   ├── blog/
│   │   ├── 2026-05-15-launch.md
│   │   └── ...
│   └── docs/
│       ├── 01-getting-started.md
│       ├── 02-architecture.md
│       ├── 03-authentication.md
│       ├── 04-components.md
│       ├── 05-billing.md
│       ├── 06-deployment.md
│       └── ...
│
├── pb_data/                         # Runtime PocketBase data (gitignored)
│   ├── data.db
│   ├── auxiliary.db
│   └── storage/
│
├── pb_public/                       # Static fallback files served by PB (optional)
│
├── docs/                            # Repository docs (developer-facing)
│   ├── ARCHITECTURE.md
│   ├── DEPLOYMENT.md
│   ├── CONTRIBUTING.md
│   └── adr/                         # Architecture Decision Records
│
└── scripts/
    ├── install.sh                   # One-line bootstrap
    └── deploy.sh                    # rsync to VPS + systemctl restart
```

### Why this structure mirrors templUI's

templUI's repo uses `components/`, `internal/`, `cmd/`, `assets/`, `utils/`, `static/`, `Taskfile.yml`, `.env.example`, `Dockerfile`, `docker-compose.yml`. We adopt the same top-level layout so anyone coming from templUI feels at home, while adding SaaS-specific folders (`services/`, `migrations/`, `content/`).

---

## 7. Data Model & PocketBase Collections

Every collection is defined as a Go migration under `migrations/`. The schema below maps domain entities to PocketBase collections. **Tenancy note:** every business collection (subscriptions, invoices, posts, audit_log, analytics_events, plus your future domain collections) carries an `organization` relation — this is the column we filter by in PB access rules and service-layer queries to guarantee tenant isolation.

### `users` (auth collection — extends PB built-in)

| Field | Type | Notes |
|---|---|---|
| `id` | text (PB default) | 15-char unique |
| `email` | email (PB default) | Unique, indexed |
| `emailVisibility` | bool | PB default |
| `verified` | bool | PB default; used for gated routes |
| `password` | password (PB default) | bcrypt managed by PB |
| `name` | text | Display name |
| `avatar` | file (single) | Optional, max 2MB, image types |
| `system_role` | select | `user`, `admin` — **platform-level only** (e.g., Anthropic-style staff). Per-org permissions live in `organization_members`. |
| `locale` | text | `en`, `id`, ... |
| `theme` | select | `light`, `dark`, `system` |
| `last_active_organization` | relation → organizations | Pointer to the org rendered after login; defaults to first membership. |
| `last_seen_at` | date | Updated by middleware |
| `created` / `updated` | date | PB defaults |

### `organizations` ★

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `slug` | text | Unique, URL-safe (`acme-co`); used in routes and Polar metadata. |
| `name` | text | Display name |
| `logo` | file (single) | Optional, max 1MB |
| `owner` | relation → users | Required; mirrored in `organization_members` with role `owner`. Cannot be removed. |
| `polar_customer_id` | text | Set on first checkout; the org *is* the Polar customer. |
| `plan` | select | `free`, `pro`, `team`, `enterprise` — denormalized for fast UI checks. |
| `subscription_status` | select | `trialing`, `active`, `past_due`, `canceled`, `none` — denormalized. |
| `trial_ends_at` | date | Optional |
| `seats_used` | number | Computed: count of active members |
| `seats_limit` | number | Per-plan ceiling; enforced server-side at invite time. |
| `settings` | json | Free-form: feature flags, integrations, branding overrides. |
| `created` / `updated` | date | |

> **Personal orgs.** Every new user gets a **personal organization** auto-created on first login (slug derived from email, name `"<user>'s workspace"`). This eliminates "you need to create an org before doing anything" UX dead-ends and matches the GitHub / Linear / Vercel pattern.

### `organization_members` ★

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `organization` | relation → organizations | Cascade delete |
| `user` | relation → users | Cascade delete |
| `role` | select | `owner`, `admin`, `member`, `viewer` |
| `joined_at` | date | |
| `created` / `updated` | date | |

> **Composite uniqueness:** `(organization, user)` must be unique — enforced via a unique index in the migration.

### `invitations` ★

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `organization` | relation → organizations | Cascade delete |
| `email` | email | Indexed |
| `role` | select | `admin`, `member`, `viewer` — not `owner` (owner can only be transferred, never invited as). |
| `token` | text | Random 32-byte URL-safe; indexed, unique. |
| `invited_by` | relation → users | |
| `expires_at` | date | Default: now + 7 days |
| `accepted_at` | date | Nullable |
| `revoked_at` | date | Nullable |
| `created` | date | |

### `subscriptions` (per-org)

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `organization` | relation → organizations | **Required, cascade delete.** Replaces the old `user` link — billing is per-org, not per-user. |
| `polar_subscription_id` | text | Unique |
| `polar_product_id` | text | |
| `polar_price_id` | text | |
| `plan` | select | `free`, `pro`, `team`, `enterprise` |
| `status` | select | `trialing`, `active`, `past_due`, `canceled`, `incomplete`, `unpaid` |
| `current_period_start` | date | |
| `current_period_end` | date | |
| `cancel_at_period_end` | bool | |
| `metadata` | json | Free-form |

### `invoices` (per-org, read-only mirror of Polar)

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `organization` | relation → organizations | |
| `polar_order_id` | text | Unique |
| `amount_paid` | number | Cents |
| `currency` | text | ISO 4217 |
| `status` | select | `paid`, `open`, `void`, `refunded` |
| `hosted_invoice_url` | url | Polar-hosted invoice link |
| `paid_at` | date | |

### `posts` (blog — platform-level, not org-scoped)

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `slug` | text | Unique, URL-safe |
| `title` | text | |
| `excerpt` | text | |
| `body` | editor / text | Markdown |
| `cover` | file | Optional |
| `author` | relation → users | |
| `tags` | json (array) | |
| `published` | bool | |
| `published_at` | date | |

> **Note:** Docs (`/docs/*`) are file-based under `content/docs/` — *not* a PB collection — because they version with the repo. Blog posts and the marketing surface are **platform-level** and not org-scoped.

### `audit_log` (per-org)

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `organization` | relation → organizations | Nullable for platform-level events (signup, etc.) |
| `actor` | relation → users | Nullable (system actions) |
| `action` | text | `user.login`, `org.member.invited`, `billing.upgraded`, ... |
| `target_collection` | text | |
| `target_id` | text | |
| `metadata` | json | IP, UA, before/after diff |
| `created` | date | Indexed |

### `analytics_events` (lightweight, optional org link)

| Field | Type | Notes |
|---|---|---|
| `id` | text | |
| `event` | text | `page_view`, `signup`, `upgrade` |
| `user` | relation → users | Nullable for anon |
| `organization` | relation → organizations | Nullable for marketing-page events |
| `session_id` | text | |
| `path` | text | |
| `referrer` | text | |
| `metadata` | json | |
| `created` | date | TTL-pruned after 90 days |

### API access rules — tenant-aware

Every org-scoped collection uses PocketBase rule expressions that join through `organization_members` to enforce isolation. Examples:

```text
// subscriptions: only members of the org can read; only owner/admin can write
listRule:   organization.members_via_organization.user ?= @request.auth.id
viewRule:   organization.members_via_organization.user ?= @request.auth.id
createRule: organization.members_via_organization.user ?= @request.auth.id &&
            organization.members_via_organization.role ?~ "owner|admin"
updateRule: organization.members_via_organization.user ?= @request.auth.id &&
            organization.members_via_organization.role ?~ "owner|admin"
deleteRule: organization.owner = @request.auth.id

// organization_members: any member can read the membership list; only admin/owner can write
listRule:   organization.members_via_organization.user ?= @request.auth.id
createRule: organization.members_via_organization.user ?= @request.auth.id &&
            organization.members_via_organization.role ?~ "owner|admin"
```

These rules are the **last line of defense**. The service layer always pre-filters by `organization = activeOrg`, so even a misconfigured rule cannot leak data across tenants.

---

## 8. Multi-Tenancy Model (Organizations)

go-pocket is **multi-tenant from commit #1**. The unit of isolation is the **organization** — a workspace with members, billing, and data. Every authenticated user belongs to ≥1 organization (their personal one is auto-created), can belong to many, and switches between them via a topbar dropdown.

### Mental model

```
User ──── many-to-many (via organization_members) ──── Organization
  │                                                         │
  │ last_active_organization                                │ owns subscriptions, invoices,
  │ (pointer for default route)                             │ audit_log, and all domain data
```

- A user without an org **cannot exist** (impossible state). On first signup, an `OnRecordAfterCreateSuccess("users")` hook creates a personal org and a `(user, owner)` membership.
- A user with no `last_active_organization` falls back to the first membership ordered by `joined_at`.
- Billing is **per-organization**. The personal org gets a free Polar customer record lazily — only when the user upgrades.

### Active-org resolution

On every authenticated request, middleware resolves the active org in this order:

1. URL segment if the route is org-scoped: `/org/:slug/...` (preferred — bookmarkable, sharable, multi-tab safe).
2. `active_org` cookie (HttpOnly, signed) — fallback for legacy `/app/*` routes.
3. `users.last_active_organization` — fallback for first request.
4. First membership ordered by `joined_at` — defensive fallback.

Resolution is performed once per request and stored in `echo.Context` as `c.Set("org", orgRecord)` and `c.Set("membership", memberRecord)`. Helpers: `ctx.Org()`, `ctx.MustOrg()`, `ctx.Membership()`, `ctx.HasOrgRole("admin")`.

### Roles (per-org)

| Role | Capability |
|---|---|
| `owner` | Everything. Exactly one per org. Transferable via danger zone. Cannot be removed. Manages billing. |
| `admin` | Manage members & invitations (cannot remove owner), edit org settings, view billing, but cannot delete the org or change owner. |
| `member` | Full read/write on domain data within the org. No member/billing management. |
| `viewer` | Read-only. Cannot mutate domain data. Useful for clients / auditors. |

Encoded as a matrix in `internal/services/tenancy/permissions.go`:

```go
var matrix = map[Permission]map[Role]bool{
    PermInviteMember:    {Owner: true, Admin: true},
    PermRemoveMember:    {Owner: true, Admin: true},
    PermChangeRole:      {Owner: true, Admin: true},
    PermDeleteOrg:       {Owner: true},
    PermManageBilling:   {Owner: true, Admin: true},
    PermEditDomainData: {Owner: true, Admin: true, Member: true},
    PermReadDomainData: {Owner: true, Admin: true, Member: true, Viewer: true},
}
```

### Invitation flow

1. Admin/owner opens **Members** page, clicks **Invite**, enters email + role.
2. `tenancy.Invite(ctx, org, email, role)` creates an `invitations` row with a random 32-byte token, expires in 7 days, and triggers `email.SendInvite(...)`.
3. Recipient clicks the link → `GET /invite/:token` → if the email matches an existing user, sign them in and add membership; if not, redirect to signup pre-filled with the invite token cached in cookie.
4. After signup, hook detects the pending invite cookie and finalizes the membership.
5. Invite row is marked `accepted_at`. Audit log records `org.invite.accepted`.

Invitations can be **revoked** (sets `revoked_at`) and **resent** (regenerates token).

### Seat enforcement

Each plan has a `seats_limit`. Before creating an invite or accepting one, the service checks `org.seats_used + pending_invites < seats_limit`. If the limit is hit, the UI surfaces an upgrade CTA.

### Domain data: how org-scoping is enforced (defense in depth)

1. **Service layer (primary):** All repository methods accept `orgID` as the first non-context argument and add `organization = ?` to every query. There is no `FindAll()` — only `FindAllInOrg(orgID, ...)`.
2. **PocketBase rules (secondary):** The collection access rules shown in §7 reject any read/write where the requester isn't a member of the row's `organization`.
3. **HTTP middleware (tertiary):** `middleware.RequireOrg(role)` runs before any `/org/*` handler and bails with 403 if active membership doesn't satisfy the required role.

A request that bypassed any single layer would still be blocked by the next two.

### Org switching

The topbar's `org_switcher.templ` is an `@dropdown.Dropdown()` listing all memberships. Selecting an org issues `POST /app/switch-org` → server updates `users.last_active_organization`, sets the `active_org` cookie, and HTMX-redirects to `/org/:slug/`. The component is loaded with `hx-boost` so the swap is buttery-smooth.

### Onboarding flow

First-time login (no membership yet — shouldn't happen normally, but handled for migration cases):

1. Redirect to `/app/onboarding` → "Create your first workspace" form.
2. Submit creates the org + owner membership in a single PB transaction.
3. Redirect to `/org/:slug/`.

### Org deletion

Soft-delete first (move to `deleted_at`, hide from lists), then hard-delete after 30 days via a PB cron job. Deletion cascades: members, invitations, subscriptions, invoices, audit_log, and any future domain collections. Billing is canceled via Polar before the soft-delete row is written.

---

## 9. Authentication & Authorization

### Methods supported

- **Email + password** (PB built-in). Signup flow includes email verification.
- **Magic link** (email OTP). Uses PB's `RequestOTP` flow.
- **OAuth2:** Google, GitHub, Apple, Microsoft (configured per-environment).
- **Multi-factor (TOTP)**: Built on PocketBase v0.23+ MFA support.

### Session model

- PocketBase issues a JWT on successful auth.
- We store the JWT in an `HttpOnly`, `Secure`, `SameSite=Lax` cookie (`pb_auth`).
- Middleware `internal/server/middleware/auth.go` reads the cookie, validates via `app.NewAuthStore()`, attaches the user record to `echo.Context` as `c.Set("user", record)`.
- Helpers: `ctx.User()`, `ctx.MustUser()`, `ctx.HasRole("admin")`.

### Route guards

```go
// internal/server/routes.go

// User-level (no org context required)
app := r.Group("/app").BindFunc(middleware.RequireAuth(pb))
app.GET("/", handlers.Dashboard)
app.GET("/onboarding", handlers.Onboarding)
app.POST("/switch-org", handlers.SwitchOrg)

// Org-scoped: requires active org + minimum role
orgGroup := r.Group("/org/{slug}").BindFunc(
    middleware.RequireAuth(pb),
    middleware.ResolveOrg(pb),          // sets c.Set("org", ...) and c.Set("membership", ...)
    middleware.RequireOrgRole("viewer"), // any member at minimum
)
orgGroup.GET("/", orgHandlers.Overview)
orgGroup.GET("/billing", orgHandlers.Billing)

// Org admin-only
orgAdmin := orgGroup.Group("/members").BindFunc(middleware.RequireOrgRole("admin"))
orgAdmin.POST("/invite", orgHandlers.InviteMember)
orgAdmin.DELETE("/{userID}", orgHandlers.RemoveMember)

// Platform-staff only (Anthropic-style backoffice)
admin := r.Group("/admin").BindFunc(middleware.RequireAuth(pb), middleware.RequireSystemRole("admin"))
admin.GET("/users", handlers.AdminUsers)
```

### Authorization model — two dimensions

- **Platform role** (on `users.system_role`): `user` (default), `admin` (Anthropic-style staff). Gates `/admin/*` only.
- **Org role** (on `organization_members.role`): `owner | admin | member | viewer`. Gates `/org/:slug/*` and all domain operations.

The two are orthogonal: a user can be a platform admin without being an org owner anywhere, and vice versa.

### Password reset flow

1. User submits email on `/auth/forgot-password`.
2. Handler calls `services.Auth.RequestPasswordReset(email)`.
3. Service emits a PB password-reset token, renders the reset email with `services.Email.Templates.Reset(token)`.
4. Email delivered via PB mailer.
5. User clicks link `/auth/reset-password?token=...`, submits new password.
6. Handler validates token, calls PB `app.Save()` on the user with new password hash.

### Email verification flow

Identical pattern, with a `verified=true` flag flip. Unverified users land on `/auth/verify-email` until they confirm.

---

## 10. UI Component Library (templUI port)

### Strategy

We ship a **1:1 direct port** of every templUI component listed at [templui.io/docs/components](https://templui.io/docs/components). The port is performed once at scaffold time via templUI's CLI:

```bash
templui init
templui add "*"
```

After that, **the components live in your repo, under your ownership**. Updates are explicit:

```bash
templui --installed add        # Update all
templui add button --force      # Update one
```

### Components included (40+, in `components/ui/`)

The 40+ included are: Accordion, Alert, Aspect Ratio, Avatar, Badge, Breadcrumb, Button, Calendar, Card, Carousel, Charts, Checkbox, Collapsible, Copy Button, Date Picker, Dialog, Dropdown, Form, Icon, Input, Input OTP, Label, Pagination, Popover, Progress, Radio, Rating, Select Box, Separator, Sheet, Sidebar, Skeleton, Slider, Switch, Table, Tabs, Tags Input, Textarea, Time Picker, Toast, Tooltip.

Each lives in its own subfolder with a `*.templ` file (markup) and a `*.go` file (props/variants). Interactive components additionally ship a `*.min.js` and `*.js` file copied into `assets/js/`.

### Theme tokens

We use templUI's OKLCH-based theme (verbatim from the [How To Use](https://templui.io/docs/how-to-use) doc) in `assets/css/input.css`. Variables: `--background`, `--foreground`, `--card`, `--popover`, `--primary`, `--secondary`, `--muted`, `--accent`, `--destructive`, `--border`, `--input`, `--ring`, plus the full `--sidebar-*` family. Dark mode is a class variant: `.dark { ... }`, toggled by Alpine.

### Component composition example

```templ
package pages

import (
  "github.com/milzamsz/go-pocket/components/ui/button"
  "github.com/milzamsz/go-pocket/components/ui/card"
  "github.com/milzamsz/go-pocket/components/ui/input"
  "github.com/milzamsz/go-pocket/components/ui/label"
)

templ Login() {
  @card.Card(card.Props{Class: "w-full max-w-md"}) {
    @card.Header() {
      @card.Title() { Welcome back }
      @card.Description() { Sign in to your account }
    }
    @card.Content() {
      <form hx-post="/auth/login" hx-swap="outerHTML">
        @label.Label(label.Props{For: "email"}) { Email }
        @input.Input(input.Props{ID: "email", Name: "email", Type: "email", Required: true})
        @label.Label(label.Props{For: "password"}) { Password }
        @input.Input(input.Props{ID: "password", Name: "password", Type: "password", Required: true})
        @button.Button(button.Props{Type: "submit", Class: "w-full mt-4"}) { Sign in }
      </form>
    }
  }
}
```

### Marketing primitives (in `components/marketing/`)

Goilerplate-flavored components layered on top of templUI primitives:

- `hero.templ` — Hero with eyebrow, headline, subhead, dual CTAs, screenshot.
- `features.templ` — 3 / 4 / 6-column feature grid with Lucide icons.
- `pricing_table.templ` — Monthly/annual toggle, 3-tier table, "Most popular" badge.
- `testimonial.templ` — Card grid + carousel variants.
- `faq.templ` — Accordion-based FAQ.
- `cta.templ` — Section-wide call-to-action band.
- `logo_cloud.templ` — Customer logo strip.
- `footer.templ` — 4-column footer with newsletter subscribe.
- `navbar.templ` — Sticky nav with mobile sheet menu.

---

## 11. Page & Route Inventory

### Marketing (public)

| Path | Handler | Page |
|---|---|---|
| `GET /` | `marketing.Home` | `pages/marketing/home.templ` |
| `GET /pricing` | `marketing.Pricing` | `pages/marketing/pricing.templ` |
| `GET /about` | `marketing.About` | `pages/marketing/about.templ` |
| `GET /contact` | `marketing.Contact` | `pages/marketing/contact.templ` |
| `GET /blog` | `marketing.BlogIndex` | `pages/marketing/blog_index.templ` |
| `GET /blog/:slug` | `marketing.BlogPost` | `pages/marketing/blog_post.templ` |
| `GET /docs` | `marketing.DocsIndex` | `pages/marketing/docs_layout.templ` |
| `GET /docs/*path` | `marketing.DocsPage` | `pages/marketing/docs_page.templ` |
| `GET /help` | `marketing.HelpCenter` | `pages/marketing/help_center.templ` |
| `GET /sitemap.xml` | `marketing.Sitemap` | (XML) |
| `GET /robots.txt` | `marketing.Robots` | (text) |
| `GET /feed.xml` | `marketing.Feed` | (RSS) |

### Auth (public)

| Path | Handler |
|---|---|
| `GET /auth/login` | `auth.LoginPage` |
| `POST /auth/login` | `auth.Login` |
| `GET /auth/signup` | `auth.SignupPage` |
| `POST /auth/signup` | `auth.Signup` |
| `GET /auth/logout` | `auth.Logout` |
| `GET /auth/forgot-password` | `auth.ForgotPasswordPage` |
| `POST /auth/forgot-password` | `auth.ForgotPassword` |
| `GET /auth/reset-password` | `auth.ResetPasswordPage` |
| `POST /auth/reset-password` | `auth.ResetPassword` |
| `GET /auth/verify-email` | `auth.VerifyEmailPage` |
| `GET /auth/oauth/:provider` | `auth.OAuthRedirect` |
| `GET /auth/oauth/:provider/callback` | `auth.OAuthCallback` |

### App — user-level (authenticated, no active org required)

| Path | Handler |
|---|---|
| `GET /app` | `app.Dashboard` — redirects to `/org/:lastActiveSlug/` |
| `GET /app/onboarding` | `app.OnboardingPage` — create first org |
| `POST /app/onboarding` | `app.CreateFirstOrg` |
| `POST /app/switch-org` | `app.SwitchOrg` |
| `GET /app/settings/profile` | `app.SettingsProfile` |
| `POST /app/settings/profile` | `app.UpdateProfile` |
| `GET /app/settings/security` | `app.SettingsSecurity` |
| `POST /app/settings/security/password` | `app.ChangePassword` |
| `POST /app/settings/security/2fa` | `app.Toggle2FA` |
| `GET /app/settings/account` | `app.SettingsAccount` — email, locale, theme |

### Org-scoped (authenticated + member of `:slug`)

| Path | Min. role | Handler |
|---|---|---|
| `GET /org/:slug` | viewer | `org.Overview` |
| `GET /org/:slug/members` | viewer | `org.Members` |
| `GET /org/:slug/members/:userID` | viewer | `org.MemberProfile` |
| `GET /org/:slug/products/:id` | viewer | `org.ProductDetail` |
| `POST /org/:slug/members/invite` | admin | `org.InviteMember` |
| `DELETE /org/:slug/members/:userID` | admin | `org.RemoveMember` |
| `PATCH /org/:slug/members/:userID/role` | admin | `org.ChangeRole` |
| `GET /org/:slug/invitations` | admin | `org.PendingInvitations` |
| `POST /org/:slug/invitations/:id/resend` | admin | `org.ResendInvitation` |
| `POST /org/:slug/invitations/:id/revoke` | admin | `org.RevokeInvitation` |
| `GET /org/:slug/billing` | admin | `org.Billing` |
| `POST /org/:slug/billing/checkout` | admin | `org.StartCheckout` (Polar) |
| `POST /org/:slug/billing/portal` | admin | `org.OpenCustomerPortal` (Polar) |
| `GET /org/:slug/billing/invoices` | admin | `org.Invoices` |
| `GET /org/:slug/settings` | admin | `org.SettingsGeneral` |
| `POST /org/:slug/settings` | admin | `org.UpdateOrg` |
| `GET /org/:slug/settings/danger` | owner | `org.SettingsDanger` |
| `POST /org/:slug/settings/transfer` | owner | `org.TransferOwnership` |
| `POST /org/:slug/settings/delete` | owner | `org.DeleteOrg` |
| `GET /org/:slug/audit` | admin | `org.AuditLog` |

### Invitation acceptance (public — token-gated)

| Path | Handler |
|---|---|
| `GET /invite/:token` | `org.AcceptInvitationPage` |
| `POST /invite/:token` | `org.AcceptInvitation` |
| `POST /invite/:token/decline` | `org.DeclineInvitation` |

### Admin (platform-staff only, `users.system_role = admin`)

| Path | Handler |
|---|---|
| `GET /admin` | `admin.Dashboard` |
| `GET /admin/users` | `admin.Users` |
| `GET /admin/users/:id` | `admin.UserDetail` |
| `GET /admin/organizations` | `admin.Organizations` |
| `GET /admin/organizations/:id` | `admin.OrgDetail` |
| `GET /admin/analytics` | `admin.Analytics` |
| `GET /admin/audit` | `admin.Audit` |
| `GET /admin/settings` | `admin.Settings` |

### Webhooks

| Path | Handler |
|---|---|
| `POST /webhooks/polar` | `webhooks.Polar` (HMAC signature-verified) |

### PocketBase built-ins (kept untouched)

| Path | Purpose |
|---|---|
| `/_/` | PocketBase admin UI |
| `/api/*` | PocketBase REST API (lockdowned) |
| `/api/realtime` | SSE channel |

---

## 12. Feature Modules

### 12.1 Authentication module ✅

Already detailed in §9. Lives in `internal/services/auth/`.

### 12.2 Multi-tenancy (organizations) ★

Detailed in §8. Lives in `internal/services/tenancy/`. Key flows:

- Org CRUD (`Create`, `Update`, `Delete` with soft-delete + 30-day grace period).
- Membership management (add/remove, change role, transfer ownership).
- Invitations (create, send via Resend, accept, revoke, resend, expire via cron).
- Active-org resolution middleware.
- Seat enforcement gates.
- Hooks that fire `org.member.invited`, `org.member.joined`, `org.transferred`, etc. — surfaced to audit log, analytics, and webhooks.

### 12.3 Admin dashboard ✅

Custom admin UI (not PocketBase's `/_/`), at `/admin`, built with templUI components. Provides:

- User list with filtering (system_role, verified, # orgs, last_seen), search, bulk actions.
- User detail with edit, impersonate (audit-logged), reset password, list of orgs.
- Organization list: search by name/slug/owner, plan filter, MRR contribution.
- Org detail: members, billing, subscription history, audit log.
- Platform metrics: MRR, ARR, churn, signup funnel, plan distribution.
- Audit log viewer (cross-org).
- Settings: feature flags, maintenance mode toggle, email templates preview.

### 12.4 Landing + marketing pages ✅

The full marketing surface: home, pricing, about, contact, blog, docs. SEO meta tags via `layouts/marketing.templ` head block; OpenGraph images served from `/assets/img/og/*.png`. Server-rendered for instant FCP and crawlability.

### 12.5 Payments (Polar.sh) ★

We use [**Polar.sh**](https://polar.sh) — an open-source, developer-first MoR built on Stripe rails. Polar handles VAT/tax automatically (huge win for indie SaaS) and has clean Go SDK ergonomics.

**Wiring in `internal/services/billing/`:**

```go
import polargo "github.com/polarsource/polar-go"

// Constructed once at boot
client := polargo.New(
    polargo.WithServer("sandbox"),                          // or "production"
    polargo.WithSecurity(cfg.Polar.OrganizationAccessToken),
)
```

**The org IS the Polar customer.** When an org first upgrades, we call `Customers.Create(...)` with `external_id = org.id`, persist the returned `polar_customer_id` on the org, and use it for every subsequent Polar call.

**Flows:**

| Flow | Endpoint | What happens |
|---|---|---|
| Checkout | `POST /org/:slug/billing/checkout` | Service calls `Checkouts.Create` with `product_price_id`, `customer_id`, `success_url`, metadata `{org_id, user_id}`. Returns hosted checkout URL → 303 redirect. |
| Customer portal | `POST /org/:slug/billing/portal` | Service calls `CustomerSessions.Create` to mint a short-lived customer access token, then redirects to `https://polar.sh/customer-portal?token=...` (or your white-labeled portal). |
| Webhook ingest | `POST /webhooks/polar` | Verifies HMAC signature against `POLAR_WEBHOOK_SECRET`, dispatches by event type. |

**Webhook events handled:**

- `subscription.created` / `subscription.updated` / `subscription.canceled` → upsert `subscriptions` row, denormalize `plan` + `subscription_status` onto `organizations`.
- `order.created` / `order.paid` → upsert `invoices` row.
- `customer.created` / `customer.updated` → keep `polar_customer_id` in sync.
- All processing happens inside a single PB transaction to keep state coherent.

**Plan catalog** (`internal/services/billing/plans.go`):

```go
var Catalog = []Plan{
    {ID: "free",       Name: "Free",       Seats: 3,   PolarPriceID: ""},
    {ID: "pro",        Name: "Pro",        Seats: 10,  PolarPriceID: cfg.Polar.PriceProMonthly},
    {ID: "team",       Name: "Team",       Seats: 50,  PolarPriceID: cfg.Polar.PriceTeamMonthly},
    {ID: "enterprise", Name: "Enterprise", Seats: -1,  PolarPriceID: cfg.Polar.PriceEnterpriseMonthly},
}
```

**Sandbox vs production** is toggled by a single env var (`POLAR_SERVER=sandbox|production`). The SDK accepts it directly — no separate build tags or code paths.

**Why Polar over Stripe direct:**

- **MoR**: Polar handles VAT/sales tax remittance globally — Stripe Tax requires you to register in each jurisdiction yourself.
- **Open source**: Polar's dashboard, billing logic, and webhook system are all open source — you can self-host if you ever need to.
- **Simpler webhook surface**: ~6 events vs. Stripe's ~40. Less to handle, less to break.
- **Org-first model**: Polar's customer model maps cleanly to organizations (1 customer = 1 org).

### 12.6 Transactional email (Resend) ★

We use [**Resend**](https://resend.com) — purpose-built for transactional email with the cleanest Go SDK in the space.

**Setup:**

```go
import "github.com/resend/resend-go/v3"

client := resend.NewClient(cfg.Resend.APIKey)
```

**Send pattern** (`internal/services/email/service.go`):

```go
func (s *Service) SendInvite(ctx context.Context, inv *domain.Invitation, org *domain.Organization) error {
    html, err := s.render(templates.Invite, templates.InviteData{
        OrgName:   org.Name,
        InviterName: inv.InvitedBy.Name,
        AcceptURL: s.cfg.AppURL + "/invite/" + inv.Token,
        ExpiresAt: inv.ExpiresAt,
    })
    if err != nil { return err }

    _, err = s.resend.Emails.Send(&resend.SendEmailRequest{
        From:    s.cfg.Resend.From,                            // "Acme <hello@acme.dev>"
        To:      []string{inv.Email},
        Subject: fmt.Sprintf("You're invited to join %s", org.Name),
        Html:    html,
        Tags: []resend.Tag{
            {Name: "category", Value: "invite"},
            {Name: "org_id",   Value: org.ID},
        },
    })
    return err
}
```

**Templates** are written as `templ` components in `internal/services/email/templates/*.templ`. Templ renders type-safe HTML; we pipe through `vanng822/go-premailer` to inline CSS (Gmail's renderer ignores `<style>` blocks).

**Bulk + batch:** Resend's `Batch.Send` is used for weekly digests and broadcast campaigns.

**Domain setup:** Verify your sending domain in Resend, add the suggested DKIM/SPF/DMARC records. Document this in `/docs/email/setup`.

**Why Resend over SMTP/PocketBase mailer:**

- **Deliverability**: Built specifically to land in the inbox. SPF/DKIM/DMARC tooling built-in.
- **API ergonomics**: Cleaner than `net/smtp`, no MIME assembly, native Go SDK.
- **Tags + webhooks**: Tag every send with `category` + `org_id` for analytics and per-org bounce tracking via Resend webhooks.
- **No SMTP server to babysit**: Just an API key.

**Queue + retry:** Resend handles retries internally. For our domain-event-triggered sends (welcome, invite, password reset, billing notifications), we make the call inline. For broadcasts/digests, we enqueue rows into an `email_queue` collection and process them on a 30-second PB cron tick, calling `Batch.Send` for efficiency.

**Resend webhook ingestion** (`POST /webhooks/resend`):

- `email.sent`, `email.delivered`, `email.bounced`, `email.complained` → update per-recipient deliverability state on the user record (and stop sending to permanently bounced addresses).

### 12.7 File uploads + storage ✅

- PocketBase's `file` field type handles uploads natively; binary stored under `pb_data/storage/` or pushed to S3-compatible backend via PB's `FileSystem` config.
- Frontend uses `<input type="file">` posted via HTMX `hx-encoding="multipart/form-data"`.
- Thumbnails auto-generated for images (PB built-in `thumbs` field option).
- Signed URLs via PB's `/api/files/:collection/:id/:filename?token=...`.
- Per-org storage paths (e.g., `org_<id>/uploads/...`) for clean isolation and per-tenant quotas.

### 12.8 HTMX + Alpine.js interactivity ✅

- `htmx.min.js` loaded globally in `layouts/base.templ`.
- Alpine.js loaded globally for `x-data`, `x-show`, `x-transition`, `x-cloak`.
- templUI interactive components (datepicker, dropdown, dialog, popover, etc.) load their per-component scripts via `@datepicker.Script()` in the `<head>` of each page that uses them.
- Partial swaps: forms post with `hx-post`, server returns just the updated fragment (e.g., a re-rendered form with validation errors), swapped via `hx-swap="outerHTML"`.

### 12.9 Internationalization (i18n) ✅

- `internal/i18n` uses `go-i18n/v2`.
- Locales: `en` (default), `id`. Add more by dropping `*.toml` files.
- Middleware reads `Accept-Language`, cookie override (`locale=id`), and user preference; sets `c.Set("localizer", ...)`.
- Templ helper: `@i18n.T(ctx, "auth.signup.title")`.
- RTL support via `dir="rtl"` on `<html>` for relevant locales.

### 12.10 Dark mode ✅

- Three states: `light`, `dark`, `system`.
- Initial state set inline in `<head>` (no FOUC) via a 6-line script that reads `localStorage.theme` and applies `class="dark"` to `<html>`.
- Toggle in user menu (Alpine-controlled).
- All templUI components are already dark-mode aware via `.dark` variant.

### 12.11 SEO + sitemap ✅

- Per-page meta tags via `layouts/marketing.templ` accepting a `Meta` struct (title, description, image, type, canonical).
- OpenGraph & Twitter card tags.
- JSON-LD structured data for blog posts (`Article`), organization, breadcrumbs.
- `sitemap.xml` dynamically generated from registered marketing routes + blog posts.
- `robots.txt` allows public routes, disallows `/app/*`, `/admin/*`, `/api/*`, `/_/`.
- `rel="canonical"` on every marketing page.

---

## 13. Configuration & Environment

All config is loaded once at boot into `internal/config.Config` (a Go struct), populated from environment variables with `envconfig`-style parsing, with CLI flag overrides via `pflag`.

### `.env.example`

```bash
# App
APP_ENV=development               # development | staging | production
APP_NAME=go-pocket
APP_URL=http://localhost:8090
APP_PORT=8090
APP_SECRET=change-me-32-bytes-min  # Used for cookie signing, CSRF

# PocketBase
PB_DATA_DIR=./pb_data
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=change-me

# Resend (transactional email)
RESEND_API_KEY=re_xxxxxxxxxxxxxxxx
RESEND_FROM="Acme <hello@yourdomain.com>"
RESEND_REPLY_TO=support@yourdomain.com
RESEND_WEBHOOK_SECRET=whsec_xxx       # For bounce/complaint webhook verification

# Polar.sh (billing)
POLAR_SERVER=sandbox                  # sandbox | production
POLAR_ORGANIZATION_ACCESS_TOKEN=polar_oat_xxxxxxxxxxxxxxxxx
POLAR_WEBHOOK_SECRET=whsec_xxx        # For HMAC verification
POLAR_ORGANIZATION_ID=org_xxx         # Your Polar org ID
POLAR_PRICE_PRO_MONTHLY=price_xxx
POLAR_PRICE_PRO_YEARLY=price_xxx
POLAR_PRICE_TEAM_MONTHLY=price_xxx
POLAR_PRICE_TEAM_YEARLY=price_xxx
POLAR_PRICE_ENTERPRISE_MONTHLY=price_xxx

# OAuth
OAUTH_GOOGLE_CLIENT_ID=
OAUTH_GOOGLE_CLIENT_SECRET=
OAUTH_GITHUB_CLIENT_ID=
OAUTH_GITHUB_CLIENT_SECRET=

# Storage (optional; defaults to local filesystem)
S3_ENDPOINT=
S3_REGION=
S3_BUCKET=
S3_ACCESS_KEY=
S3_SECRET_KEY=

# Multi-tenancy
TENANCY_AUTO_CREATE_PERSONAL_ORG=true   # Create personal org on signup
TENANCY_INVITATION_TTL_DAYS=7
TENANCY_ORG_DELETE_GRACE_DAYS=30        # Soft-delete window before hard delete

# Observability
LOG_LEVEL=info                    # debug | info | warn | error
LOG_FORMAT=text                   # text | json
SENTRY_DSN=

# Feature flags
FEATURE_BLOG_ENABLED=true
FEATURE_OAUTH_ENABLED=true
FEATURE_BILLING_ENABLED=true
FEATURE_I18N_ENABLED=true
```

---

## 14. Development Workflow

### One-time setup

```bash
# Prerequisites
go version                                       # 1.26.3
go install github.com/a-h/templ/cmd/templ@v0.3.1020
go install github.com/go-task/task/v3/cmd/task@v3.49.1
# Tailwind: download standalone binary from tailwindcss/tailwindcss GitHub Releases

# Project bootstrap
git clone https://github.com/milzamsz/go-pocket myapp
cd myapp
cp .env.example .env
go mod tidy
task setup          # Runs migrations, seeds dev data, creates PB admin
```

### Daily loop

```bash
task dev            # Parallel: templ --watch + tailwindcss --watch + go run .
```

This is identical to templUI's workflow. `templ generate --watch` runs the Go server with `--proxy="http://localhost:8090"` and hot-reloads on `.templ` changes. `tailwindcss --watch` regenerates `output.css` on save.

### Taskfile commands

```yaml
version: "3"

tasks:
  setup:
    desc: First-time setup
    cmds:
      - go mod tidy
      - templui init
      - templui add "*"
      - task: migrate

  dev:
    desc: Start dev server with hot reload
    cmds:
      - task --parallel tailwind templ

  tailwind:
    cmds:
      - tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

  templ:
    cmds:
      - templ generate --watch --proxy="http://localhost:8090" --cmd="go run ./main.go" --open-browser=false

  build:
    desc: Production build (single static binary)
    cmds:
      - templ generate
      - tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify
      - CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version={{.VERSION}}" -o ./bin/go-pocket .
    vars:
      VERSION:
        sh: git describe --tags --always --dirty

  migrate:
    desc: Run pending migrations
    cmds:
      - go run . migrate up

  migrate:create:
    desc: Create a new migration
    cmds:
      - go run . migrate create {{.CLI_ARGS}}

  seed:
    desc: Seed dev data
    cmds:
      - go run ./cmd/tools/seed

  test:
    cmds:
      - go test ./... -race -cover

  lint:
    cmds:
      - go vet ./...
      - gofmt -l -s .

  templui:update:
    desc: Update all installed templUI components
    cmds:
      - templui --installed add

  release:
    desc: Build for linux/amd64 + linux/arm64
    cmds:
      - GOOS=linux GOARCH=amd64 task build
      - GOOS=linux GOARCH=arm64 task build
```

---

## 15. Build, Asset Pipeline & Embed

### What gets embedded into the binary

```go
// assets/embed.go
package assets

import "embed"

//go:embed all:css all:js all:img all:fonts
var Assets embed.FS
```

```go
// content/embed.go (blog + docs)
package content

import "embed"

//go:embed all:blog all:docs
var Files embed.FS
```

```go
// migrations/embed.go (via PocketBase's migrate package)
// PocketBase reads from this package directly via init() registration.
```

```go
// components/icons/embed.go (only if you bundle SVG icons separately)
```

### Build steps in order

1. `templ generate` — turns `*.templ` into `*_templ.go`.
2. `tailwindcss -i input.css -o output.css --minify` — produces ~30-50 KB CSS.
3. `go build -tags production -ldflags="-s -w -X main.Version=$(git describe)"`.

Result: **a single static binary**, typically 25-35 MB, that contains the entire app including assets, templates, migrations, blog, and docs.

### Cache-busting

CSS and JS asset URLs include a content hash query string in production: `/assets/css/output.css?v=<sha256-prefix>`. Computed once at boot from `embed.FS`.

---

## 16. Deployment (Docker / Dokploy)

go-pocket ships as a single Docker image and is designed to deploy in two equally supported ways: **Docker / docker-compose** (anywhere) and **[Dokploy](https://dokploy.com)** (recommended — open-source PaaS, self-hosted alternative to Heroku/Vercel/Netlify, runs on any VPS).

### Why Dokploy

Dokploy gives you the developer experience of Heroku/Vercel — git-push to deploy, environment variables UI, automatic Let's Encrypt via Traefik, rolling deploys, scheduled jobs, S3-backed backups, monitoring — but on **your** $5 VPS. It's an excellent fit for go-pocket because:

- Both run as Docker containers; no impedance mismatch.
- Dokploy handles Traefik + HTTPS + custom domains automatically — no Caddy config to maintain.
- Built-in **Schedule Jobs** can run `go-pocket` CLI subcommands (e.g., `migrate up`, `clean-expired-invitations`) on cron schedules from a UI.
- Built-in **Volume Backups** to S3 cover the `pb_data/` SQLite directory.
- Auto-deploy on git push from GitHub/Gitea.

### Dockerfile (multi-stage, single static binary)

```dockerfile
# syntax=docker/dockerfile:1.7

# ----- Build stage -----
FROM golang:1.26.3-alpine3.23 AS builder

RUN apk add --no-cache git nodejs npm curl \
    && curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss \
    && go install github.com/a-h/templ/cmd/templ@v0.3.1020

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate templ + tailwind
RUN templ generate \
    && tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify

# Build static binary
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /out/go-pocket .

# ----- Runtime stage -----
FROM alpine:3.23.3

RUN apk add --no-cache ca-certificates tzdata sqlite \
    && adduser -D -u 10001 gopocket

WORKDIR /app
COPY --from=builder --chown=gopocket:gopocket /out/go-pocket /app/go-pocket

USER gopocket

EXPOSE 8090
VOLUME ["/app/pb_data"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8090/healthz || exit 1

ENTRYPOINT ["/app/go-pocket"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
```

**Image size:** ~35 MB compressed (Alpine + static Go binary + assets embedded).

### `docker-compose.yml` (local + Dokploy-compatible)

```yaml
services:
  app:
    build:
      context: .
      args:
        VERSION: ${VERSION:-dev}
    image: go-pocket:latest
    restart: unless-stopped
    env_file: .env
    ports:
      - "8090:8090"
    volumes:
      - pb_data:/app/pb_data
    labels:
      # Traefik / Dokploy labels (set automatically by Dokploy UI)
      - "traefik.enable=true"
      - "traefik.http.routers.app.rule=Host(`${APP_HOST}`)"
      - "traefik.http.routers.app.tls.certresolver=letsencrypt"

volumes:
  pb_data:
    driver: local
```

### Deploy via Dokploy — step by step

1. **Install Dokploy** on a fresh VPS (Ubuntu 22.04+ recommended, 2 GB RAM minimum):
   ```bash
   curl -sSL https://dokploy.com/install.sh | sh
   ```
   This installs Docker, Traefik, and Dokploy itself in one shot. Visit `http://<vps-ip>:3000` to set up the admin account.

2. **Create a project** in Dokploy → "Create Application".

3. **Connect your Git provider** (GitHub, GitLab, Gitea, or Bitbucket) and select the `go-pocket` repository + branch.

4. **Choose build type:** "Dockerfile" → Dokploy auto-detects the `Dockerfile` at repo root.

5. **Set environment variables** in the Dokploy UI (copy from `.env.example`, fill in real values for `RESEND_API_KEY`, `POLAR_ORGANIZATION_ACCESS_TOKEN`, etc.). Dokploy encrypts these at rest.

6. **Mount a persistent volume:**
   - In the Volumes tab: mount path `/app/pb_data` ← named volume `go-pocket-data`.
   - This survives container restarts and redeployments.

7. **Configure domain:**
   - Domains tab → add `app.yourdomain.com`.
   - Toggle HTTPS → Dokploy provisions Let's Encrypt automatically via Traefik.
   - DNS: point an A record from `app.yourdomain.com` to your VPS IP.

8. **Set up auto-deploy** (optional but recommended):
   - Deployments tab → enable "Auto Deploy on Push".
   - Now every push to `main` triggers a rebuild + zero-downtime rolling restart.

9. **Schedule jobs** (in Dokploy's Schedule Jobs UI):
   - Daily: `./go-pocket cron expire-invitations`
   - Daily: `./go-pocket cron prune-analytics`
   - Weekly: `./go-pocket cron weekly-digest`
   - Monthly: `./go-pocket cron hard-delete-orgs` (purges orgs past the 30-day soft-delete grace period)

10. **Configure backups:**
    - S3 Destinations tab → add your S3 / Backblaze B2 / Cloudflare R2 credentials.
    - Volume Backups tab → schedule daily backups of the `go-pocket-data` volume.
    - Dokploy uses `sqlite3 .backup` against the live DB before archiving (no corrupt-snapshot risk).

### Alternative: plain Docker / docker-compose on any host

If you'd rather not use Dokploy, the same image works on **any** Docker host:

```bash
# Build
docker build -t go-pocket:latest .

# Run
docker run -d \
  --name go-pocket \
  --env-file .env \
  -p 8090:8090 \
  -v go-pocket-data:/app/pb_data \
  --restart unless-stopped \
  go-pocket:latest
```

Pair with **Caddy** or **Traefik** in front for HTTPS. The `docker-compose.yml` shown above works equally well outside Dokploy — just remove the Dokploy-specific labels if you're using your own reverse proxy.

### Zero-downtime deploys

Dokploy performs **rolling deploys** by default: a new container is spun up, healthchecks pass, traffic switches via Traefik, the old container drains and dies. Total observable downtime: **0 seconds**.

For plain Docker, use `docker compose up -d --no-deps --build app` — the brief container restart is sub-second.

### Backups recap

| Frequency | What | Where |
|---|---|---|
| Hourly | `sqlite3 .backup` of `data.db` | Local rotation, last 24 |
| Daily | Full volume snapshot | S3 / B2 / R2 via Dokploy |
| Weekly | Long-term archive | Cold storage |

Restore drill: documented in `/docs/deployment/disaster-recovery.md` — keep it tested.

### Scaling beyond a single VPS

PocketBase is single-writer SQLite, so vertical scaling is the first lever (Hetzner CCX13 → CCX23 → CCX33 covers most apps to ~100k MAU). Beyond that:

- **Dokploy Cluster** (Docker Swarm under the hood) for HA reverse proxy + scheduled jobs.
- **LiteFS** sidecar for read replicas + leader election.
- **Service-layer DB swap** → managed Postgres for write-heavy workloads (described in §5).

The Dockerfile + Dokploy deployment stays the same throughout — only the storage backend changes.

---

## 17. Observability, Security & Performance

### Logging

- Structured `slog` to stdout (text in dev, JSON in prod), captured by systemd-journal or Loki.
- PocketBase's own logs in `pb_data/auxiliary.db`, viewable in PB admin UI.
- Request ID middleware adds `X-Request-ID` for tracing across logs.

### Metrics

- Optional `/metrics` endpoint exposing Prometheus format (gated by admin auth or internal network).
- Counters: requests by route/status, signups, checkouts, errors.

### Error tracking

- Sentry SDK (`getsentry/sentry-go`) optional, configured via `SENTRY_DSN`.
- Panics caught by recover middleware → reported with stack trace + request context.

### Security checklist

- **CSRF** on all state-changing routes (custom middleware with double-submit cookie).
- **Rate limiting** on `/auth/*` (5 req/min/IP), `/invite/*` (10 req/min/IP), and `/webhooks/polar` + `/webhooks/resend` (signature-verified, no rate limit).
- **Tenant isolation**: triple-layer (service queries, PB rules, route middleware) — see §8.
- **Content Security Policy** strict: `script-src 'self'` (templUI is CSP-compliant), `style-src 'self'`, `img-src 'self' data:`, `connect-src 'self' https://api.polar.sh https://sandbox-api.polar.sh https://api.resend.com`. Plus `nonce` on the inline dark-mode init script.
- **HSTS** via Caddy.
- **bcrypt** cost 12 (PocketBase default).
- **Argon2id** for any custom token hashing (e.g., API keys).
- **SQL injection** N/A — all queries go through `dbx` parameterized builder.
- **XSS** — templ auto-escapes; never use `templ.Raw` on user input.
- **Open redirect** — `?return_to=...` is validated against an allowlist of internal paths.

### Performance budget

- TTFB < 100ms on /  (single SQLite query, in-memory plan catalog).
- `output.css` < 50 KB minified gzipped.
- Total JS on a typical authenticated page < 80 KB (Alpine + HTMX + ~3 templUI scripts).
- Lighthouse 95+ on marketing pages.

---

## 18. Testing Strategy

### Levels

1. **Unit** — service layer, pure functions. `go test ./internal/services/...`.
2. **Integration** — services + PocketBase via `tests.NewTestApp()` (in-memory SQLite). `go test ./internal/...`.
3. **HTTP** — handler-level tests using `httptest`, asserting status + HTML fragments via `goquery`.
4. **End-to-end** — optional Playwright suite under `e2e/` covering critical flows: signup → verify → upgrade → cancel.

### Coverage targets

- Services: 80%+
- Handlers: 60%+ (focus on critical paths)
- Total: 70%+

### CI

GitHub Actions workflow `.github/workflows/ci.yml`:

```yaml
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26.3' }
      - run: go install github.com/a-h/templ/cmd/templ@v0.3.1020
      - run: templ generate
      - run: go vet ./...
      - run: go test ./... -race -cover
```

---

## 19. Phase-by-Phase Implementation Roadmap

### Phase 0 — Scaffold (Day 1)

- [ ] `go mod init github.com/milzamsz/go-pocket`.
- [ ] Add PocketBase, templ, Echo (via PB) dependencies.
- [ ] Add `polargo` and `resend-go/v3` to go.mod.
- [ ] `templui init` + `templui add "*"`.
- [ ] Tailwind config + `input.css` with templUI theme tokens.
- [ ] `Taskfile.yml`, `.env.example`, `Dockerfile`, `docker-compose.yml`, `.gitignore`, `.dockerignore`.
- [ ] **`AGENTS.md` at repo root + `.agents/` directory** with rules, conventions, prompts (see §21).
- [ ] `main.go` boots PocketBase, registers a stub `OnServe` handler.
- [ ] `task dev` works; visiting `/` returns "hello".

**Deliverable:** Running app with hot reload, AI agent context wired up.

### Phase 1 — Foundations (Days 2-4)

- [ ] Folder structure carved out (`internal/`, `components/`, `migrations/`).
- [ ] Config loader (`internal/config`).
- [ ] Middleware chain (auth, logger, request ID, recover, CSRF, rate limit).
- [ ] Base layouts: `base.templ`, `marketing.templ`, `auth.templ`, `app.templ`.
- [ ] Asset serving (`/assets/*` + `utils.SetupScriptRoutes`).
- [ ] First page: `/` rendered via templUI's hero + features marketing primitives.
- [ ] Dark mode toggle works.
- [ ] 404 + 500 error pages.

**Deliverable:** Marketing skeleton ready, dark mode, asset pipeline embedded.

### Phase 2 — Auth (Days 5-7)

- [ ] Migration: `users` collection extended with `role`, `locale`, `theme`, etc.
- [ ] Login / signup / logout flows.
- [ ] Email verification.
- [ ] Forgot / reset password.
- [ ] OAuth (Google + GitHub) wired.
- [ ] Auth middleware + route guards.
- [ ] Session cookies (HttpOnly, Secure, SameSite=Lax).

**Deliverable:** Complete auth surface; user can sign up, verify, log in, log out.

### Phase 3 — User-level app shell + Settings (Days 8-9)

- [ ] App sidebar nav + topbar with user menu.
- [ ] `/app` dashboard placeholder (redirects to org).
- [ ] Profile settings (name, avatar upload).
- [ ] Security settings (change password, 2FA stub).
- [ ] Locale + theme preferences persist to user record.

**Deliverable:** Authenticated user shell with working personal settings.

### Phase 3.5 — Multi-tenancy ★ (Days 10-13)

- [ ] Migration: `organizations`, `organization_members`, `invitations` collections + composite uniqueness index.
- [ ] `internal/services/tenancy/` with service, repository, permissions matrix.
- [ ] Auto-create personal org hook on user signup.
- [ ] Org switcher component + `POST /app/switch-org`.
- [ ] `ResolveOrg` + `RequireOrgRole` middleware.
- [ ] Org-scoped routes: `/org/:slug/...` with overview, members, settings.
- [ ] Invitation flow: create, email, accept, decline, resend, revoke.
- [ ] Pending invitations list.
- [ ] Seat limits enforced at invite time.
- [ ] Transfer ownership flow.
- [ ] Org deletion with 30-day soft-delete + hard-delete cron.
- [ ] Audit log entries for every org mutation.

**Deliverable:** Full multi-tenant SaaS: multi-org users, role-gated routes, invitations, isolation tested.

### Phase 4 — Billing with Polar.sh (Days 14-16)

- [ ] Polar SDK wired (`polargo.New(WithServer, WithSecurity)`).
- [ ] Plan catalog + pricing page wired to env price IDs.
- [ ] Per-org checkout: `POST /org/:slug/billing/checkout` → Polar Checkout session.
- [ ] Customer portal: `POST /org/:slug/billing/portal` → Polar customer session redirect.
- [ ] `subscriptions` + `invoices` collections + webhook syncing.
- [ ] Webhook handler at `/webhooks/polar` with HMAC verification + 6-event dispatcher.
- [ ] Denormalized `plan` + `subscription_status` on org for fast UI checks.
- [ ] Org billing page (current plan, change plan, view invoices, cancel).
- [ ] Email notifications (via Resend): subscription started, payment failed, trial ending.

**Deliverable:** End-to-end per-org billing: invite team → upgrade org → cancel → see invoices, all via Polar.sh.

### Phase 5 — Email via Resend (Day 17)

- [ ] Resend SDK wired (`resend.NewClient(apiKey)`).
- [ ] Email service with templ-based templates + premailer inlining.
- [ ] Tags on every send (`category`, `org_id`) for analytics.
- [ ] Templates: welcome, verify, reset, **invite**, subscription_started, payment_failed, weekly_digest.
- [ ] Resend webhook handler for bounce/complaint tracking.
- [ ] Email queue collection + 30-second cron tick for batch sends.

**Deliverable:** All transactional emails sending via Resend, deliverability tracked.

### Phase 6 — Admin dashboard (Days 18-20)

- [ ] `/admin` shell with `system_role=admin` guard.
- [ ] User list with filters, search, pagination (templUI table + pagination).
- [ ] User detail view with edit, impersonate (audit-logged), list of orgs.
- [ ] Organization list with plan filter and MRR contribution.
- [ ] Org detail: members, billing, audit log.
- [ ] Platform-wide MRR / ARR / churn dashboards.
- [ ] Audit log viewer (cross-org).

**Deliverable:** Custom admin UI usable for day-to-day ops.

### Phase 7 — Content (blog + docs) (Days 21-23)

- [ ] `posts` collection + blog index/detail pages.
- [ ] File-based docs under `content/docs/*.md` rendered with markdown + frontmatter.
- [ ] Docs layout with sidebar, "On this page" TOC.
- [ ] RSS feed at `/feed.xml`.

**Deliverable:** Blog + docs sites live, matching goilerplate.com/docs layout.

### Phase 8 — SEO + i18n + polish (Days 24-25)

- [ ] Per-page meta tags, OG images, JSON-LD.
- [ ] Sitemap, robots.txt.
- [ ] i18n with `en` + `id` locales.
- [ ] Loading skeletons on slow handlers.
- [ ] Toast notifications wired site-wide.

**Deliverable:** Production-grade marketing surface.

### Phase 9 — Testing + observability (Days 26-27)

- [ ] Unit tests for services (esp. tenancy permission matrix).
- [ ] Integration tests via PB test app (multi-tenant isolation tests are critical).
- [ ] Handler tests for critical paths.
- [ ] slog + Sentry wiring.
- [ ] Health check at `/healthz`.

**Deliverable:** ≥70% coverage; metrics & errors observable; tenant isolation proven by tests.

### Phase 10 — Deploy + Docs + release (Days 28-30)

- [ ] `README.md` with quickstart.
- [ ] Repository `docs/` (ARCHITECTURE, DEPLOYMENT, CONTRIBUTING, ADRs).
- [ ] Polished `AGENTS.md` (final review for AI coding agent guidance).
- [ ] Production `Dockerfile` finalized + GitHub Actions workflow for image push.
- [ ] Dokploy deploy to staging, then production.
- [ ] Demo deploy at `demo.go-pocket.dev`.

**Deliverable:** Public v1.0 release; live demo running on Dokploy.

> **Total estimated effort:** ~30 focused days for a solo developer, ~12 days for a team of two. The multi-tenancy phase (3.5) adds ~3-4 days vs. the original single-tenant plan but pays for itself the moment your first customer asks "can my team join?".

---

## 20. Documentation Site (1:1 with goilerplate.com/docs)

The `/docs` route renders a documentation site whose structure mirrors goilerplate.com/docs. Content lives in `content/docs/*.md`, with frontmatter for ordering and grouping.

### Doc sections (mirroring Goilerplate's layout)

1. **Getting Started**
   - Introduction
   - Quick Start (`task setup` → `task dev` → visit localhost:8090)
   - Project Structure (links to §6 of this plan)
2. **Architecture**
   - Stack Overview
   - Folder Structure
   - PocketBase Integration
3. **Authentication**
   - Email & Password
   - OAuth Providers
   - Email Verification
   - Password Reset
   - Two-Factor Auth
4. **Components**
   - Overview
   - Theming (OKLCH tokens)
   - All templUI components (Accordion → Tooltip), one page each, with live examples + copy-button code blocks.
5. **Multi-Tenancy** ★
   - Organizations & Members
   - Roles & Permissions
   - Invitations
   - Org Switching
   - Tenant Isolation Model
6. **Pages**
   - Marketing
   - Auth
   - App / Dashboard
   - Org-Scoped (`/org/:slug/*`)
   - Admin
7. **Billing (Polar.sh)** ★
   - Polar Setup (sandbox & production)
   - Plan / Product Configuration
   - Webhooks
   - Customer Portal
   - Per-Org Subscriptions
8. **Email (Resend)** ★
   - Resend Setup & Domain Verification
   - Templates
   - Tags & Analytics
   - Bounce / Complaint Handling
9. **Storage**
   - File Uploads
   - Thumbnails
   - S3 Integration
   - Per-Org Storage Paths
10. **i18n**
    - Locales
    - Adding a Language
11. **SEO**
    - Meta Tags
    - Sitemap & Robots
    - OpenGraph Images
12. **Deployment** ★
    - Docker (any host)
    - Dokploy (recommended)
    - Auto-deploy on git push
    - Scheduled Jobs
    - Volume Backups
    - Disaster Recovery
13. **AI-Assisted Development** ★
    - AGENTS.md
    - Using `.agents/` rules
    - Cursor / Claude / Aider / Codex setup
14. **Reference**
    - Environment Variables
    - CLI Commands (`go-pocket serve`, `migrate`, `cron`, `admin`)
    - Migrations
    - Testing

### Doc page layout

Same as templUI's docs: left sidebar (sections + components), main content, right sidebar ("On This Page" anchor list). Code blocks use Shiki-style syntax highlighting (consider importing the `shiki` folder pattern from templUI, or use `gohtml`-based highlighter).

---

## 21. AI-Assisted Development (AGENTS.md & .agents/)

go-pocket is designed to be **co-built with AI coding agents** from commit #1. We follow the [AGENTS.md](https://agents.md) open standard (60k+ repos, backed by the Linux Foundation's Agentic AI Foundation) plus a project-specific `.agents/` directory for richer context.

### Why this matters

A Go SaaS boilerplate is large surface area: 40+ UI components, multi-tenant rules, hooks, migrations, and four service integrations. Without curated context, AI agents waste tokens re-discovering conventions on every task. `AGENTS.md` + `.agents/` lets Cursor, Claude Code, Aider, Codex, Gemini CLI, Copilot, Devin, and any future agent open the repo and immediately know:

- How to run dev/test/build commands.
- Where to put new code (handlers vs services vs templ components).
- That all DB access goes through service-layer methods filtered by `organization`.
- That `templ generate` must run after every `.templ` edit.
- That migrations are immutable once shipped.

### `AGENTS.md` at repo root (the canonical file)

Markdown, no fixed schema. The structure below is what we ship:

```markdown
# AGENTS.md — go-pocket

> Multi-tenant Go SaaS boilerplate on PocketBase + templUI.
> Read `.agents/architecture.md` and `.agents/conventions.md` before making non-trivial changes.

## Setup commands

- Install deps: `go mod download && templui add "*"`
- Generate templates: `templ generate`
- Run dev server (hot reload): `task dev`
- Run tests: `task test`
- Build production binary: `task build`
- Run migrations: `task migrate`

## Code style & conventions

- Go 1.26.3. `gofmt -s`. Errors wrapped with `fmt.Errorf("context: %w", err)`.
- Templ files end in `.templ`; ALWAYS run `templ generate` after editing one. The generated `*_templ.go` files must NOT be edited by hand.
- Tailwind utility classes only — no custom CSS in `.templ` files. Use theme tokens (`bg-background`, `text-foreground`, etc.).
- Handlers under `internal/server/handlers/` are thin: they parse request, call services, render templ. NO DB access in handlers.
- Services under `internal/services/` own all PocketBase interactions. Every repository method accepts `orgID` as a non-context argument when the resource is org-scoped.
- Migrations are append-only and immutable. NEVER modify a migration that has been merged to `main`. Add a new one instead.

## Multi-tenancy rules (critical)

- Every business collection has an `organization` relation. Service-layer queries MUST filter by it.
- New endpoints handling tenant data MUST live under `/org/:slug/*` and use the `RequireOrgRole` middleware.
- New collections require: (1) `organization` field, (2) PocketBase rules that join through `organization_members`, (3) service-layer methods that take `orgID`.

## Testing instructions

- Run all tests: `task test`
- Run a single package: `go test -race ./internal/services/tenancy/...`
- Multi-tenant isolation tests live in `internal/services/tenancy/*_test.go` — they MUST pass before any tenancy-touching PR is merged.
- Add or update tests for any code change.

## Pull request guidelines

- Title format: `<area>: <concise change>` (e.g., `billing: handle Polar trial_ends_at`).
- Run `task lint && task test` before committing.
- Include a one-line CHANGELOG.md entry for user-visible changes.
- Reference the migration number for any schema change.

## Things NOT to do

- Don't import `stripe-go`. Billing is **Polar.sh** — use the `polargo` package.
- Don't use `net/smtp` or PocketBase's mailer for outbound email. Use the `email` service which wraps `resend-go/v3`.
- Don't add new global state. Pass services via the `Deps` struct.
- Don't bypass the service layer. If you find yourself reaching for `app.Dao()` from a handler, write a service method instead.
- Don't introduce a JavaScript framework. We have Alpine + HTMX + per-component templUI scripts. That's the stack.

## See also

- `.agents/architecture.md` — full architecture summary
- `.agents/conventions.md` — file naming, error patterns, templ patterns
- `.agents/prompts/` — reusable prompt templates for common tasks
- `docs/` — long-form developer docs (ARCHITECTURE.md, ADRs)
```

### `.agents/` directory — richer context for sustained work

`.agents/` carries the longer-form context that would clutter `AGENTS.md`. Files an agent loads on demand:

#### `.agents/rules.md`

- Coding style deep-dive (e.g., "prefer `for range` over indexed loops").
- Test patterns (table-driven tests for services).
- Commit message style (Conventional Commits).
- Branching model.

#### `.agents/architecture.md`

A 200-300 line distillation of this PLAN.md focused on **what an agent needs to know to write code**:

- Layered architecture (handlers → services → repos → PB).
- Active-org resolution flow.
- Hook usage policy (thin glue only).
- Embedded asset model.
- Where to add a new feature end-to-end (collection → migration → service → handler → templ).

#### `.agents/conventions.md`

- File naming (`<noun>_<verb>.go`, `<noun>.templ`).
- Templ component structure (Props struct → templ func with `Props` arg → optional `Script()` func).
- Error patterns: domain errors in `internal/domain/errors.go`, wrapped with context at boundaries.
- HTMX patterns: `hx-swap="outerHTML"`, validation errors re-render the same fragment.

#### `.agents/prompts/`

Reusable Markdown prompts that an agent can be pointed at:

| File | Purpose |
|---|---|
| `add-feature.md` | "Add a new feature end-to-end" walkthrough: collection → migration → service → handler → templ → tests. |
| `add-migration.md` | How to write a reversible PB migration with `Register(up, down)`. |
| `add-templui-component.md` | "Add a new templUI component to this project" (run `templui add <name>`, regenerate Tailwind sources, load `Script()`). |
| `add-page.md` | "Add a new org-scoped page with auth + role guard + nav entry". |
| `debug.md` | Triage checklist when something breaks (logs, request ID, PB admin, replay webhook). |

#### `.agents/tools/`

Per-tool compatibility shims that point each agent at `AGENTS.md`. We use **symlinks** so we keep one source of truth:

```bash
# Created during setup
ln -s ../../AGENTS.md  .agents/tools/claude.md           # then CLAUDE.md → .agents/tools/claude.md
ln -s ../../AGENTS.md  .agents/tools/cursor.mdc          # then .cursorrules → .agents/tools/cursor.mdc
```

`.aider.conf.yml`:

```yaml
read: AGENTS.md
auto-commits: false
```

`.gemini/settings.json`:

```json
{
  "context": { "fileName": "AGENTS.md" }
}
```

`.github/copilot-instructions.md` (Copilot reads this automatically):

```markdown
This repo uses AGENTS.md as the canonical agent guide. Read AGENTS.md and .agents/architecture.md before suggesting changes.
```

### Maintenance discipline

`AGENTS.md` is **living documentation**. The rules to keep it useful:

1. **Every PR that changes a convention updates `AGENTS.md` or `.agents/`** in the same commit.
2. **One source of truth.** Don't mirror commands into `README.md`. README is for humans onboarding; AGENTS.md is for agents executing.
3. **No prose for prose's sake.** Bullet points, terse imperative voice ("Run X.", "Don't do Y."). Agents are not your audience for storytelling.
4. **Audit quarterly.** Skim the file; remove anything no longer true.

### How agents use this in practice

Example: an agent is asked "add a 'projects' feature scoped to organizations".

1. Agent reads `AGENTS.md` → sees the multi-tenancy rule + setup commands.
2. Agent reads `.agents/architecture.md` → understands layering.
3. Agent reads `.agents/prompts/add-feature.md` → gets the exact step-by-step.
4. Agent creates: migration (`migrations/170XXXXXXXX_init_projects.go`), service (`internal/services/projects/`), handlers (`internal/server/handlers/org/projects.go`), templ pages (`components/pages/org/projects_*.templ`), tests.
5. Agent runs `task test` and `templ generate` per the AGENTS.md instructions.
6. Submits PR with title `projects: introduce org-scoped projects feature`.

Result: the agent does in minutes what would otherwise take a human-led pairing session.

---

## 22. Appendix: Key Code Snippets

### `main.go`

```go
package main

import (
	"log"
	"os"

	"github.com/milzamsz/go-pocket/internal/app"
	"github.com/pocketbase/pocketbase"

	_ "github.com/milzamsz/go-pocket/migrations" // self-registering
)

func main() {
	pb := pocketbase.New()

	if err := app.Bootstrap(pb); err != nil {
		log.Fatal(err)
	}

	if err := pb.Start(); err != nil {
		log.Fatal(err)
	}

	_ = os.Args
}
```

### `internal/app/app.go`

```go
package app

import (
	"github.com/milzamsz/go-pocket/assets"
	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/server"
	"github.com/milzamsz/go-pocket/internal/services/billing"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func Bootstrap(pb *pocketbase.PocketBase) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Services — built once, wired everywhere
	emailSvc   := email.New(pb, cfg)         // wraps resend-go/v3
	tenancySvc := tenancy.New(pb, cfg, emailSvc)
	billingSvc := billing.New(pb, cfg, emailSvc, tenancySvc) // wraps polargo

	deps := &server.Deps{
		Config:  cfg,
		Email:   emailSvc,
		Tenancy: tenancySvc,
		Billing: billingSvc,
		Assets:  assets.Assets,
	}

	pb.OnServe().BindFunc(func(se *core.ServeEvent) error {
		server.RegisterRoutes(se.Router, pb, deps)
		return se.Next()
	})

	pb.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("system_role") == "" {
			e.Record.Set("system_role", "user")
		}
		return e.Next()
	})

	// Auto-create personal org on signup
	pb.OnRecordAfterCreateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
		if cfg.Tenancy.AutoCreatePersonalOrg {
			if _, err := tenancySvc.CreatePersonalOrg(e.Context(), e.Record); err != nil {
				return err
			}
		}
		return emailSvc.SendWelcome(e.Context(), e.Record)
	})

	// Audit invitations
	pb.OnRecordAfterCreateSuccess("invitations").BindFunc(func(e *core.RecordEvent) error {
		return emailSvc.SendInvite(e.Context(), e.Record)
	})

	return nil
}
```

### `internal/server/routes.go`

```go
package server

import (
	"github.com/milzamsz/go-pocket/internal/server/handlers/admin"
	apphandler "github.com/milzamsz/go-pocket/internal/server/handlers/app"
	"github.com/milzamsz/go-pocket/internal/server/handlers/auth"
	"github.com/milzamsz/go-pocket/internal/server/handlers/marketing"
	"github.com/milzamsz/go-pocket/internal/server/handlers/webhooks"
	"github.com/milzamsz/go-pocket/internal/server/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Deps struct {
	Config  any
	Email   any
	Tenancy any
	Billing any
	Assets  any
}

func RegisterRoutes(r *core.Router, pb *pocketbase.PocketBase, d *Deps) {
	// Global middleware
	r.BindFunc(middleware.RequestID)
	r.BindFunc(middleware.Logger)
	r.BindFunc(middleware.Recover)
	r.BindFunc(middleware.SecurityHeaders)

	// Assets
	registerAssets(r, d)

	// Marketing (public)
	r.GET("/", marketing.Home(d))
	r.GET("/pricing", marketing.Pricing(d))
	r.GET("/about", marketing.About(d))
	r.GET("/contact", marketing.Contact(d))
	r.GET("/blog", marketing.BlogIndex(d))
	r.GET("/blog/{slug}", marketing.BlogPost(d))
	r.GET("/docs", marketing.DocsIndex(d))
	r.GET("/docs/{path...}", marketing.DocsPage(d))
	r.GET("/sitemap.xml", marketing.Sitemap(d))
	r.GET("/robots.txt", marketing.Robots(d))
	r.GET("/feed.xml", marketing.Feed(d))

	// Auth
	authGroup := r.Group("/auth")
	authGroup.GET("/login", auth.LoginPage(d))
	authGroup.POST("/login", auth.Login(pb, d))
	authGroup.GET("/signup", auth.SignupPage(d))
	authGroup.POST("/signup", auth.Signup(pb, d))
	authGroup.GET("/logout", auth.Logout(pb))
	// ... reset, verify, oauth ...

	// App — user-level (no org context required)
	appGroup := r.Group("/app").BindFunc(middleware.RequireAuth(pb))
	appGroup.GET("/", apphandler.Dashboard(d))
	appGroup.GET("/onboarding", apphandler.OnboardingPage(d))
	appGroup.POST("/onboarding", apphandler.CreateFirstOrg(d))
	appGroup.POST("/switch-org", apphandler.SwitchOrg(d))
	appGroup.GET("/settings/profile", apphandler.SettingsProfile(d))
	appGroup.POST("/settings/profile", apphandler.UpdateProfile(pb, d))
	// ...

	// Org-scoped (active org + minimum role required)
	orgGroup := r.Group("/org/{slug}").BindFunc(
		middleware.RequireAuth(pb),
		middleware.ResolveOrg(pb, d),
		middleware.RequireOrgRole("viewer"),
	)
	orgGroup.GET("/", orghandler.Overview(d))
	orgGroup.GET("/members", orghandler.Members(d))
	orgGroup.GET("/billing", orghandler.Billing(d))
	orgGroup.POST("/billing/checkout", orghandler.StartCheckout(d))      // Polar
	orgGroup.POST("/billing/portal", orghandler.OpenCustomerPortal(d))    // Polar
	// ... admin-gated routes use .BindFunc(middleware.RequireOrgRole("admin")) ...

	// Invitation acceptance (public, token-gated)
	r.GET("/invite/{token}", orghandler.AcceptInvitationPage(d))
	r.POST("/invite/{token}", orghandler.AcceptInvitation(d))

	// Platform admin (system_role = admin)
	adminGroup := r.Group("/admin").BindFunc(middleware.RequireAuth(pb), middleware.RequireSystemRole("admin"))
	adminGroup.GET("/", admin.Dashboard(d))
	adminGroup.GET("/users", admin.Users(d))
	adminGroup.GET("/organizations", admin.Organizations(d))
	// ...

	// Webhooks (no auth — signature-verified)
	r.POST("/webhooks/polar", webhooks.Polar(d))
	r.POST("/webhooks/resend", webhooks.Resend(d))

	// Health
	r.GET("/healthz", func(e *core.RequestEvent) error {
		return e.String(200, "ok")
	})
}
```

### Sample migration (`migrations/1700000000_init_users.go`)

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Fields.Add(&core.TextField{Name: "name", Max: 100})
		users.Fields.Add(&core.FileField{Name: "avatar", MaxSelect: 1, MaxSize: 2 * 1024 * 1024, MimeTypes: []string{"image/jpeg", "image/png", "image/webp"}})
		users.Fields.Add(&core.SelectField{Name: "system_role", Values: []string{"user", "admin"}, MaxSelect: 1})
		users.Fields.Add(&core.TextField{Name: "locale", Max: 8})
		users.Fields.Add(&core.SelectField{Name: "theme", Values: []string{"light", "dark", "system"}, MaxSelect: 1})
		users.Fields.Add(&core.DateField{Name: "last_seen_at"})
		// last_active_organization is added in the organizations migration (forward reference avoided)

		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		for _, f := range []string{"name", "avatar", "system_role", "locale", "theme", "last_seen_at"} {
			users.Fields.RemoveByName(f)
		}
		return app.Save(users)
	})
}
```

### Sample multi-tenant migration (`migrations/1700000050_init_organizations.go`)

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// 1) organizations
		orgs := core.NewBaseCollection("organizations")
		orgs.Fields.Add(
			&core.TextField{Name: "slug", Required: true, Pattern: `^[a-z0-9-]+$`, Max: 64},
			&core.TextField{Name: "name", Required: true, Max: 100},
			&core.FileField{Name: "logo", MaxSelect: 1, MaxSize: 1 * 1024 * 1024, MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "image/svg+xml"}},
			&core.RelationField{Name: "owner", Required: true, CollectionId: "_pb_users_auth_", MaxSelect: 1},
			&core.TextField{Name: "polar_customer_id", Max: 100},
			&core.SelectField{Name: "plan", Values: []string{"free", "pro", "team", "enterprise"}, MaxSelect: 1},
			&core.SelectField{Name: "subscription_status", Values: []string{"trialing", "active", "past_due", "canceled", "none"}, MaxSelect: 1},
			&core.DateField{Name: "trial_ends_at"},
			&core.NumberField{Name: "seats_used"},
			&core.NumberField{Name: "seats_limit"},
			&core.JSONField{Name: "settings"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		orgs.AddIndex("idx_orgs_slug", true, "slug", "")
		if err := app.Save(orgs); err != nil { return err }

		// 2) organization_members
		members := core.NewBaseCollection("organization_members")
		members.Fields.Add(
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.RelationField{Name: "user", Required: true, CollectionId: "_pb_users_auth_", MaxSelect: 1, CascadeDelete: true},
			&core.SelectField{Name: "role", Required: true, Values: []string{"owner", "admin", "member", "viewer"}, MaxSelect: 1},
			&core.DateField{Name: "joined_at"},
			&core.AutodateField{Name: "created", OnCreate: true},
			&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
		)
		members.AddIndex("idx_member_org_user", true, "organization,user", "")
		// Tenant-aware rules
		members.ListRule  = pointer("organization.members_via_organization.user ?= @request.auth.id")
		members.ViewRule  = pointer("organization.members_via_organization.user ?= @request.auth.id")
		members.CreateRule = pointer("organization.members_via_organization.user ?= @request.auth.id && organization.members_via_organization.role ?~ \"owner|admin\"")
		members.UpdateRule = pointer("organization.members_via_organization.user ?= @request.auth.id && organization.members_via_organization.role ?~ \"owner|admin\"")
		members.DeleteRule = pointer("organization.owner = @request.auth.id || (organization.members_via_organization.user ?= @request.auth.id && organization.members_via_organization.role ?~ \"owner|admin\")")
		if err := app.Save(members); err != nil { return err }

		// 3) invitations
		invites := core.NewBaseCollection("invitations")
		invites.Fields.Add(
			&core.RelationField{Name: "organization", Required: true, CollectionId: orgs.Id, MaxSelect: 1, CascadeDelete: true},
			&core.EmailField{Name: "email", Required: true},
			&core.SelectField{Name: "role", Required: true, Values: []string{"admin", "member", "viewer"}, MaxSelect: 1},
			&core.TextField{Name: "token", Required: true, Max: 64},
			&core.RelationField{Name: "invited_by", Required: true, CollectionId: "_pb_users_auth_", MaxSelect: 1},
			&core.DateField{Name: "expires_at", Required: true},
			&core.DateField{Name: "accepted_at"},
			&core.DateField{Name: "revoked_at"},
			&core.AutodateField{Name: "created", OnCreate: true},
		)
		invites.AddIndex("idx_invite_token", true, "token", "")
		invites.AddIndex("idx_invite_email", false, "email", "")
		if err := app.Save(invites); err != nil { return err }

		// 4) Add users.last_active_organization (forward reference now resolvable)
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil { return err }
		users.Fields.Add(&core.RelationField{Name: "last_active_organization", CollectionId: orgs.Id, MaxSelect: 1})
		return app.Save(users)
	}, func(app core.App) error {
		for _, name := range []string{"invitations", "organization_members", "organizations"} {
			if c, err := app.FindCollectionByNameOrId(name); err == nil {
				_ = app.Delete(c)
			}
		}
		if users, err := app.FindCollectionByNameOrId("users"); err == nil {
			users.Fields.RemoveByName("last_active_organization")
			_ = app.Save(users)
		}
		return nil
	})
}

func pointer[T any](v T) *T { return &v }
```

### Base layout (`components/layouts/base.templ`)

```templ
package layouts

import (
	"github.com/milzamsz/go-pocket/components/ui/datepicker"
	"github.com/milzamsz/go-pocket/components/ui/dialog"
	"github.com/milzamsz/go-pocket/components/ui/dropdown"
	"github.com/milzamsz/go-pocket/components/ui/popover"
	"github.com/milzamsz/go-pocket/components/ui/toast"
)

type Meta struct {
	Title       string
	Description string
	Image       string
	Canonical   string
	Type        string // "website" | "article"
	Lang        string // "en" | "id"
}

templ Base(meta Meta) {
	<!DOCTYPE html>
	<html lang={ meta.Lang } class="scroll-smooth">
		<head>
			<meta charset="UTF-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
			<title>{ meta.Title }</title>
			<meta name="description" content={ meta.Description }/>
			<link rel="canonical" href={ meta.Canonical }/>
			<meta property="og:title" content={ meta.Title }/>
			<meta property="og:description" content={ meta.Description }/>
			<meta property="og:image" content={ meta.Image }/>
			<meta property="og:type" content={ meta.Type }/>
			<meta name="twitter:card" content="summary_large_image"/>
			<link rel="stylesheet" href="/assets/css/output.css"/>
			<link rel="icon" href="/assets/img/favicon.ico"/>
			<script>
				// No-FOUC dark mode init
				(function() {
					var t = localStorage.getItem('theme');
					if (t === 'dark' || (!t && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
						document.documentElement.classList.add('dark');
					}
				})();
			</script>
			<script src="/assets/js/alpine.min.js" defer></script>
			<script src="/assets/js/htmx.min.js"></script>
			@toast.Script()
			@dialog.Script()
			@dropdown.Script()
			@popover.Script()
			@datepicker.Script()
		</head>
		<body class="bg-background text-foreground antialiased">
			{ children... }
		</body>
	</html>
}
```

---

## 23. Admin Dashboard Starter Parity Plan (templUI)

This section defines a decision-complete implementation target for a **full admin dashboard starter** with capability parity to:

- `Kiranism/next-shadcn-dashboard-starter`

Parity means **feature and UX parity**, not framework parity. We keep Go + PocketBase + templ + templUI and map every major starter capability into server-rendered pages with HTMX/Alpine where needed.

### 23.1 Target outcome

By completion, `go-pocket` should ship a dashboard starter that includes:

- Polished dashboard shell with sidebar, header, breadcrumbs, command palette, responsive drawer.
- Analytics overview with stat cards, trend charts, and recent activity modules.
- Reusable data-table system with query-driven search/filter/sort/pagination.
- Org/team management equivalent to workspace/organization features.
- Billing, profile/settings, notification center, and plan-gated demo page.
- Optional productivity modules: kanban and chat.
- Theme system with light/dark/system plus multiple dashboard palettes.
- Error/empty/loading/skeleton states across all key screens.
- Cleanup tooling to remove optional modules when teams want a slimmer starter.

### 23.2 Architecture mapping (1:1 spirit)

- Next.js App Router pattern -> PocketBase route groups + templ page handlers.
- shadcn/ui component pattern -> templUI component primitives + local wrappers.
- React state/query tools -> server-driven query params + service-layer view models + HTMX partial swaps + Alpine local UI state.
- Clerk organizations/billing pattern -> PocketBase auth + `organizations` + `organization_members` + Polar billing routes.

No React/Next runtime is introduced. This starter remains single-binary Go.

### 23.3 Feature modules and route contracts

Add or finalize these route families:

- Dashboard core:
  - `GET /app/` (overview)
  - `GET /app/notifications`
  - `GET /app/profile`
  - `GET /app/settings/{profile|security|account}`
- Admin platform:
  - `GET /admin/`
  - `GET /admin/users`
  - `GET /admin/users/{id}`
  - `GET /admin/organizations`
  - `GET /admin/organizations/{id}`
  - `GET /admin/analytics`
  - `GET /admin/audit`
  - `GET /admin/settings`
- Org workspace:
  - `GET /org/:slug/` (workspace overview)
  - `GET /org/:slug/members`
  - `GET /org/:slug/invitations`
  - `GET /org/:slug/billing`
  - `GET /org/:slug/billing/invoices`
  - `GET /org/:slug/audit`
  - `GET /org/:slug/settings`
  - `GET /org/:slug/settings/danger`
- Optional parity modules:
  - `GET /org/:slug/kanban`
  - `GET /org/:slug/chat`
  - `GET /org/:slug/resources`
  - `GET /org/:slug/exclusive` (plan-gated demo)

All org-scoped routes remain under `/org/:slug/*` with existing role middleware and service-level org filters.

### 23.4 UI system requirements (templUI)

Standardize on templUI-based primitives and wrappers:

- App shell: sidebar, sheet/drawer, topbar, breadcrumb, command dialog, dropdown, avatar, tabs.
- Data surfaces: card, table, badge, pagination, skeleton, empty state, alert/error state.
- Input surfaces: input, textarea, select, switch, checkbox, tooltip/popover helpers.
- Feedback: toast, inline validation summary, success/error banners.

Rules:

- Admin and org views are operational, dense, and table-first.
- Marketing-style hero composition is not used inside authenticated dashboard pages.
- Mobile behavior is explicit: collapsible nav, horizontal table scrolling, stable action bars.

### 23.5 Data and collections for parity

Keep existing collections and add missing dashboard starter modules:

- Existing core:
  - `users`
  - `organizations`
  - `organization_members`
  - `invitations`
  - `subscriptions`
  - `webhook_events`
  - `audit_log`
- Add for parity modules:
  - `notifications` (user and optionally organization scoped)
  - `resources` (demo CRUD/table module)
  - `kanban_boards`, `kanban_columns`, `kanban_tasks`
  - `chat_conversations`, `chat_messages`
  - `user_preferences` (theme/layout/widget preferences)

For every org-scoped collection:

- include `organization` relation
- apply PB rules joined through `organization_members`
- enforce service-layer `organization = orgID` filters

### 23.6 Table and query contract

Unify table query semantics across admin/org list screens:

- `q`
- `page`
- `per_page`
- `sort`
- `direction`
- Optional feature filters: `status`, `role`, `plan`, `family`, `unread`

Handlers parse query params, call service methods, and return:

- full page render (normal request)
- table/list partial render (HTMX request)

This keeps behavior consistent while preserving server-first rendering.

### 23.7 Wave-based implementation plan

Wave 1: Shell + tokens + navigation foundation

- Finalize dashboard shell for app, org, and admin contexts.
- Normalize theme tokens for cards, borders, accents, success/warning/destructive, sidebar.
- Add responsive sidebar/drawer and command palette scaffold.

Wave 2: Reusable data-table system + admin parity pages

- Build one reusable table renderer pattern in templ for headers/rows/sort/pagination/empty/loading states.
- Apply to users, organizations, audit, analytics supporting lists.
- Add filter/search controls with query param roundtrip.

Wave 3: Org workspace parity pages

- Finalize members/invites/billing/settings/audit pages with robust interaction states.
- Add invoice table, plan badges, and subscription state mapping surfaces.
- Ensure role-gated actions render correctly by membership role.

Wave 4: Notifications, kanban, chat parity modules

- Notification center + page with read/unread and bulk mark-read.
- Kanban board UI and CRUD/reorder actions.
- Chat conversation list + thread + composer UI.

Wave 5: Polish and starter-hardening

- Error boundaries/pages and skeleton states.
- Accessibility and responsive sweep.
- Optional feature cleanup command spec and docs.
- Route/table docs alignment in `PLAN.md` and README.

### 23.8 Testing and acceptance gates

Service tests:

- Query parsing and filter/sort/pagination behavior.
- Notification read/unread transitions.
- Kanban/chat CRUD with org isolation.
- Plan-gated access checks.

Handler tests:

- Unauthorized/forbidden behavior for admin and org routes.
- Form validation and redirect semantics.
- HTMX partial responses for tables and interactive modules.

Integration tests:

- End-to-end flow: signup -> org -> invite -> billing -> admin visibility.
- Tenant isolation checks across all new org-scoped collections.

Release gates:

- `templ generate`
- `task css:build`
- `task lint`
- `go test ./...`
- `go build ./...`

### 23.9 Non-negotiable constraints

- Keep PocketBase as embedded backend and source of truth.
- Keep service-layer ownership of data access (no direct DB access in handlers).
- Keep `/api/*` reserved for PocketBase built-ins.
- Keep generated files out of manual edits.
- Keep multi-tenancy triple isolation intact for every new module.

---

## Closing notes

This blueprint is **opinionated on purpose**. Every choice — embedded PocketBase, templUI port, multi-tenant from commit #1, Polar.sh over Stripe, Resend over SMTP, Docker/Dokploy over bare-metal, AGENTS.md from the start — is selected so a solo founder can ship a real multi-tenant SaaS in weeks instead of months, without losing the ability to scale or migrate later.

When implementation begins, follow the roadmap in §19 phase-by-phase. Resist the urge to deviate from templUI's component patterns; they're already battle-tested. Treat the PocketBase service-layer abstraction as sacred — the day you outgrow SQLite, that abstraction is what saves you from a rewrite. Treat the tenant-isolation rules (§7 collection access rules + §8 service-layer filters + middleware) as **non-negotiable** — a single cross-tenant data leak will sink your reputation.

Use the AI agents. With `AGENTS.md` and `.agents/` in place from day one, Cursor / Claude / Aider / Codex become genuinely productive collaborators on this codebase, not novelties.

Ship something. Then iterate.
