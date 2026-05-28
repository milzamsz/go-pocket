---
name: Technical Elegance
colors:
  surface: '#131315'
  surface-dim: '#131315'
  surface-bright: '#39393b'
  surface-container-lowest: '#0e0e10'
  surface-container-low: '#1c1b1d'
  surface-container: '#201f22'
  surface-container-high: '#2a2a2c'
  surface-container-highest: '#353437'
  on-surface: '#e5e1e4'
  on-surface-variant: '#c7c4d7'
  inverse-surface: '#e5e1e4'
  inverse-on-surface: '#313032'
  outline: '#908fa0'
  outline-variant: '#464554'
  surface-tint: '#c0c1ff'
  primary: '#c0c1ff'
  on-primary: '#1000a9'
  primary-container: '#8083ff'
  on-primary-container: '#0d0096'
  inverse-primary: '#494bd6'
  secondary: '#44e2cd'
  on-secondary: '#003731'
  secondary-container: '#03c6b2'
  on-secondary-container: '#004d44'
  tertiary: '#ddb7ff'
  on-tertiary: '#490080'
  tertiary-container: '#b76dff'
  on-tertiary-container: '#400071'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#e1e0ff'
  primary-fixed-dim: '#c0c1ff'
  on-primary-fixed: '#07006c'
  on-primary-fixed-variant: '#2f2ebe'
  secondary-fixed: '#62fae3'
  secondary-fixed-dim: '#3cddc7'
  on-secondary-fixed: '#00201c'
  on-secondary-fixed-variant: '#005047'
  tertiary-fixed: '#f0dbff'
  tertiary-fixed-dim: '#ddb7ff'
  on-tertiary-fixed: '#2c0051'
  on-tertiary-fixed-variant: '#6900b3'
  background: '#131315'
  on-background: '#e5e1e4'
  surface-variant: '#353437'
typography:
  display:
    fontFamily: Geist
    fontSize: 48px
    fontWeight: '700'
    lineHeight: '1.1'
    letterSpacing: -0.04em
  headline-lg:
    fontFamily: Geist
    fontSize: 32px
    fontWeight: '600'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-lg-mobile:
    fontFamily: Geist
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Geist
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.3'
    letterSpacing: -0.01em
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: '1.6'
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.5'
  body-sm:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.5'
  label-md:
    fontFamily: Geist
    fontSize: 14px
    fontWeight: '500'
    lineHeight: '1'
    letterSpacing: 0.02em
  label-xs:
    fontFamily: Geist
    fontSize: 12px
    fontWeight: '600'
    lineHeight: '1'
    letterSpacing: 0.05em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 8px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  sidebar-width: 260px
  sidebar-collapsed: 80px
  container-max: 1440px
---

## Brand & Style

The design system is engineered for high-performance SaaS environments where precision meets premium aesthetics. It targets power users, developers, and data-driven decision-makers who value clarity without sacrificing visual sophistication.

The style is a synthesis of **Minimalism** and **Glassmorphism**, characterized by a "dark-first" philosophy. It utilizes ultra-thin 1px strokes, meaningful transparency, and vibrant light-source accents to create a sense of depth and technical mastery. The emotional response is one of calm control, high-tech reliability, and luxury-grade utility.

## Colors

The design system employs a sophisticated dark-palette hierarchy using deep Zinc and Slate tones. Color is used sparingly as a functional signal or a premium accent rather than a primary surface filler.

- **Backgrounds:** The base layer is a deep `#09090b`. Secondary surfaces (cards, sidebars) use a slightly lighter `#18181b` with subtle transparency.
- **Accents:** A core trio of Indigo (Primary), Teal (Success/Active), and Violet (Insight) provides a vibrant contrast. These should be applied using OKLCH color logic to ensure high-chroma "glow" effects against the dark background.
- **Borders:** Fixed at `#27272a` (Zinc-800) for standard containers and `#3f3f46` (Zinc-700) for interactive elements to maintain a sharp, technical look.

## Typography

This design system leverages **Geist** for headlines and UI labels to provide a precise, technical character, while **Inter** is used for body copy to ensure maximum legibility at smaller scales.

The typographic hierarchy follows a modular scale. Display and Headline styles use negative letter-spacing to feel tight and modern. Labels and technical data points should often use `label-xs` in uppercase to evoke a "instrumentation" feel. Monospaced numeric variants within Geist should be preferred for data tables and KPI values.

## Layout & Spacing

The layout is built on a strict **8px grid system**. This rhythm dictates padding, margins, and component heights to ensure mathematical harmony across the dashboard.

- **Structure:** A collapsible sidebar on the left controls primary navigation. The main content area sits within a fluid grid that caps at `container-max`.
- **Grid:** Use a 12-column grid for the main dashboard content. Gutters are fixed at `24px` (lg) to provide enough "air" between dense data points.
- **Top Bar:** A sticky top bar is required for all views, featuring a `backdrop-blur` (20px) and a semi-transparent background to allow content to scroll underneath while maintaining context.
- **Mobile:** On devices under 768px, the sidebar transitions to a bottom navigation bar or a full-screen overlay, and horizontal padding reduces to `16px`.

## Elevation & Depth

Hierarchy is established through **Tonal Layering** and **Glassmorphism** rather than traditional heavy shadows.

1.  **Level 0 (Base):** Deepest background (`#09090b`).
2.  **Level 1 (Cards/Sidebar):** `#18181b` with a 1px border.
3.  **Level 2 (Modals/Dropdowns):** Transparent backgrounds with `backdrop-blur: 12px` and a slightly more prominent border.
4.  **Accent Elevation:** KPI cards use a subtle "internal glow." This is achieved with a top-left inner shadow in the accent color (e.g., Indigo) at 10% opacity, simulating a light source hitting the edge of the glass.

Shadows, when used, are extremely diffused (e.g., `0 20px 40px rgba(0,0,0,0.4)`) and only appear on top-level floating elements like modals.

## Shapes

The shape language is "Soft-Modern." Using a base `roundedness: 2` (8px), the UI feels approachable but still structured and professional.

- **Buttons & Inputs:** Use the standard 8px radius.
- **Cards:** Use `rounded-lg` (16px) for main dashboard containers to create clear visual separation.
- **Interactive States:** On hover, certain elements may transition from a 1px border to a subtle "glow" border using a gradient stroke that incorporates the primary accent color.

## Components

### Buttons
Primary buttons use a solid Indigo background with a subtle top-light gradient. Secondary buttons are ghost-styled with 1px Zinc borders. All buttons feature a 150ms transition on hover.

### KPI Cards (Glowing)
These are the focal point of the dashboard. They feature a `1px` border that utilizes a linear gradient (Zinc-800 to Primary-500/20%). Inside, area charts use a high-vibrancy stroke with a fading gradient fill that terminates at 0% opacity near the card's bottom.

### Sidebar
The sidebar is dark-themed with a subtle right-border. Active links use a vertical "pill" indicator on the left and a low-opacity Indigo background tint.

### Inputs
Search and form fields are minimalist. They feature a Zinc-800 border that transforms into a Primary-500 border with a `0 0 0 2px` outer ring on focus. Use Inter-Medium for all input text.

### Area Charts
Charts must be frameless. Use OKLCH colors for lines to ensure they "pop" against the dark background. Gradients should be used under lines to provide a sense of volume and density.
