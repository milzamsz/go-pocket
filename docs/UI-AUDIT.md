# UI/UX Audit — go-pocket

Date: 2026-05-29
Scope: the application shell and design-token system (templ + Tailwind v4 + Alpine/HTMX).

This document records why the UI felt inconsistent ("confusing"), what was fixed
in this pass, and the recommended follow-ups that need a visual preview before
being applied.

## Root cause: three overlapping color systems

The UI mixed three different ways of expressing the same colors, so equivalent
surfaces and accents rendered differently depending on which one a given element
happened to use:

1. **Semantic theme tokens** — `bg-background`, `text-foreground`, `border-border`,
   `bg-card`, `text-primary`, `bg-accent`, `surface-container-*`. Defined in
   `assets/css/input.css` under `@theme` and `.dark`. This is the intended system.
2. **Hardcoded Tailwind literals** — `indigo-600/500/400`, `slate-200`,
   plus arbitrary hex like `#111319`, `#1b1f29`, `#181b24`. These never adapt to
   the theme and do not match the `primary`/border tokens.
3. **A second hand-maintained gray palette** inside the `.stitch-*` component
   classes (`#111319` panel, `#1b1f29` border) that diverged from the token
   palette (`#131315` surface, `#464554` border).

The practical symptom: a token-styled card next to a `.stitch-panel`, or the
indigo sidebar accent against the lavender `primary` token, showed subtly
mismatched grays, borders, and accent hues.

A secondary issue: the public/auth shell (`Surface` in `base.templ`) was
light-first with hardcoded `dark:` hex overrides, while the dashboard shell used
semantic tokens — two visual languages for one product.

## Fixed in this pass

All changes preserve both light and dark themes.

- **`assets/css/input.css`** — rewrote every `.stitch-*` component class to read
  the theme custom properties (`var(--color-card)`, `var(--color-border)`, etc.).
  Because those variables are redefined under `.dark`, each component now adapts
  automatically and stays in lockstep with the token system. Deleted the parallel
  hardcoded gray palette and all `.dark .stitch-*` overrides. Added a shared focus
  ring for `.stitch-input` and an `[x-cloak]` rule.
- **`components/layouts/base.templ`** (public/auth shell) — replaced
  `slate-*`/hex/`indigo-*` with `border-border`, `bg-card`, `text-primary`,
  `hover:bg-muted`. Wrapped the header links in a `<nav aria-label="Site">` and
  added visible focus rings. Added an `aria-label` to the theme toggle.
- **`components/layouts/dashboard.templ`** (app/org/admin shell):
  - Replaced all `indigo-*` literals (header badge, active nav, nav icons,
    avatar, breadcrumb) with `primary` tokens.
  - **Responsive:** the sidebar is now an off-canvas drawer on small screens
    (Alpine `navOpen` state, slide-in transform, click-away backdrop, close
    button) and static from `lg` up. Added a hamburger toggle in the top bar.
    Body padding scales `p-4 → sm:p-6 → lg:p-8`.
  - **Accessibility:** `aria-current="page"` on the active nav link;
    `aria-label`s on the bell, search (`<label class="sr-only">`), avatar, and
    drawer toggles; `:aria-expanded` on the hamburger; focus-visible rings on
    interactive elements; the avatar is now a real link to profile settings.
  - **Honest status:** removed the hardcoded "You are on the Free Trial. 7 days
    remaining" banner (it was shown to every org regardless of billing state and
    contradicted the hardcoded "Enterprise Plan" label). The sidebar subtitle now
    shows the `section` value that is already passed in, and the breadcrumb root
    uses `section` instead of a literal "Dashboard".
- **`components/layouts/icons.templ`** — added `menu` and `x` icons used by the
  new mobile drawer (the icon component has no fallback, so missing names render
  nothing).

> After editing `.templ` files you must run `task templ:generate` before
> building; the generated `*_templ.go` files are not hand-edited.

## Recommended follow-ups (need a visual preview)

### 1. Finish the token sweep in page templates

~87 hardcoded `indigo-*` / hex / `slate-*` occurrences remain in page-content
templates. They were left untouched because they affect content layout and some
are intentional marketing gradients — they should be migrated with a live
preview, not blind. Current counts:

| File | Occurrences |
|---|---|
| `components/pages/org/products.templ` | 26 |
| `components/pages/org/kanban.templ` | 21 |
| `components/pages/marketing/home.templ` | 14 |
| `components/pages/org/pages.templ` | 10 |
| `components/pages/marketing/docs.templ` | 7 |
| `components/pages/auth/pages.templ` | 4 |
| `components/pages/app/pages.templ` | 3 |
| `components/pages/marketing/blog.templ` | 2 |

Mapping to apply: `indigo-600 → primary`, `indigo-400/500 → primary` (with
`/10`–`/30` opacity for tints), `slate-200/#d8deea → border`,
`#111319/#0d0f14 → card`/`surface-container-low`, `text-slate-500 →
text-muted-foreground`.

### 2. Make the shell genuinely data-driven

`DashboardSurface` is called from ~13 sites across `app`, `org`, and `admin`
page templates, and none currently receive billing or user data. To wire real
state, introduce a `ShellState` value and thread it through:

```
type ShellState struct {
    OrgName   string
    PlanLabel string          // e.g. "Pro", "Free"
    User      struct{ Name, Email, AvatarURL string }
    Trial     struct{ Active bool; DaysRemaining int; UpgradeHref string }
}
```

- Handlers (`internal/server/handlers/scaffold.go` and the org handlers) read the
  org's `subscription_status` / `plan` denorm fields (already populated by the
  Polar webhook state store) and the current actor, then pass a `ShellState`
  into the page render.
- The trial banner renders only when `Trial.Active`; the plan label and avatar
  initials come from real data; the avatar can show `User.Name` initials instead
  of a generic icon.

This is a vertical slice (handler → page → layout) and should be built with the
app running so the rendered result can be verified.

### 3. Self-host fonts

`base.templ` loads Geist + Inter from Google Fonts, which conflicts with the
project's self-hosted, no-CDN posture (and required widening the new
Content-Security-Policy to allow `fonts.googleapis.com` / `fonts.gstatic.com`).
Self-hosting the woff2 files under `/assets/fonts/` and dropping the Google
origins from the CSP would tighten both privacy and the security policy.

### 4. Smaller polish

- Add a "skip to content" link in `Base` for keyboard users.
- Give the search field a real action (currently decorative).
- Consider `prefers-reduced-motion` handling for the drawer transition and the
  pulsing trial dot if the banner is reinstated.
