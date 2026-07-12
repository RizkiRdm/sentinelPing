# DESIGN.md

Project: SentinelPing
Scope: MVP UI/UX System Specification

---

## Design Principles

- Simplicity over decoration.
- Consistency over novelty.
- Accessibility over aesthetics.
- Status legibility over visual flourish (this is a monitoring tool —
  state MUST be understood in <1 second of glance).
- Implementation simplicity over custom components.

---

## Visual Identity

- Density: Medium-high information density. Dashboard is a monitoring
  console, not a marketing page — MUST prioritize scanability over
  whitespace-heavy layouts.
- Atmosphere: Technical, calm, utilitarian. MUST NOT use playful
  illustration, decorative gradients, or marketing-style hero sections
  inside the authenticated dashboard.
- Visual Consistency: MUST use a single design system (shadcn/ui component
  primitives) across all screens — no ad-hoc one-off component styling.
- Motion Intensity: LOW. Motion exists only to communicate state change
  (e.g. status color transition), NOT for decorative purposes.
- Brand Characteristics: Monospace accents for technical data (tokens,
  timestamps, ping IDs) to reinforce "developer tool" identity.

---

## Color System

- Semantic Color Roles:
  - `healthy` → green
  - `running` → blue
  - `late` → amber/yellow
  - `failed` → red
  - `pending` → gray/neutral
- Color Tokens: MUST be defined as CSS variables (`--color-healthy`,
  `--color-running`, `--color-late`, `--color-failed`, `--color-pending`),
  NOT hardcoded hex values inline in components.
- Accessibility Requirements: State colors MUST NOT be the ONLY indicator
  of state — MUST be paired with a text label and/or icon (STRICT:
  color-only status indicators are PROHIBITED, per colorblind accessibility
  requirement).
- Contrast Requirements: MUST maintain minimum 4.5:1 text contrast ratio
  for all body text; MUST maintain minimum 3:1 for large text (18px+ bold
  or 24px+ regular).
- State Colors: apply consistently across badge, table row indicator, and
  chart line color for the same state — MUST NOT use different color
  mappings across components for the same semantic state.
- Dark Mode Behavior: MUST be supported at MVP (developer tool audience
  strongly prefers dark mode as default). Dark mode MUST use adjusted
  (desaturated) variants of state colors to preserve contrast ratio.
- Light Mode Behavior: MUST be available as toggle; MUST NOT be
  the only mode.

---

## Typography

- Typography Scale: MUST use a constrained scale (e.g. 12/14/16/20/24/32px)
  — MUST NOT introduce arbitrary font sizes outside this scale.
- Font Roles: Sans-serif for UI text/labels; monospace for tokens,
  timestamps, ping URLs, and any copy-pasteable technical value.
- Spacing Rules: MUST use a consistent spacing scale (4px base unit:
  4/8/12/16/24/32/48).
- Line Height Rules: body text MUST use 1.5 line-height minimum for
  readability; monospace technical values MAY use 1.2.
- Readability Constraints: MUST NOT set body text below 14px.
- Responsive Typography: MUST NOT scale typography down below the defined
  minimum on mobile breakpoints — MUST reduce layout density instead of
  shrinking text below readable size.

---

## Component System

- Buttons: primary (main action, e.g. "Create Monitor"), secondary
  (cancel/dismiss), destructive (delete monitor — MUST use red + require
  confirmation). States: default, hover, focus, disabled, loading.
- Forms/Inputs: MUST show inline validation errors adjacent to the field,
  NOT only in a toast/summary. MUST support keyboard-only completion.
- Cards: used for individual monitor summary in dashboard grid/list view.
  MUST display: name, current state badge, last ping time, next expected
  ping time.
- Dialogs: used for monitor creation/edit and destructive-action
  confirmation. MUST trap focus while open. MUST close on Escape key.
- Navigation: MUST be a persistent sidebar or top bar (single-level, MVP
  has no nested navigation — only Dashboard, Monitor Detail, Settings).
- Tables: used for ping history and state transition history. MUST support
  basic pagination (STRICT: no infinite scroll in MVP — simpler to
  implement, sufficient for expected data volume).
- Loading States: MUST use skeleton placeholders matching the shape of the
  eventual content, NOT generic spinners, for list/table views. Spinners
  ARE acceptable for button-level async actions (e.g. form submit).
- Skeleton States: MUST match final layout dimensions to prevent layout
  shift on content load.
- Empty States: MUST include a clear call-to-action (e.g. "No monitors yet
  — create your first monitor" with button), NOT a blank page.
- Error States: MUST display a human-readable message plus a retry action
  where applicable (e.g. failed data fetch MUST show "Retry" button, not
  just an error message).

---

## Layout Rules

- Grid System: 12-column grid for dashboard layout at desktop breakpoint.
- Spacing System: 4px base unit (see Typography section).
- Responsive Breakpoints: mobile <640px, tablet 640-1024px, desktop
  >1024px.
- Container Widths: MUST cap main content max-width at 1280px on large
  screens to preserve readability (MUST NOT stretch tables/cards
  full-width on ultra-wide monitors).
- Z-index Hierarchy: base content (0) < sticky nav (10) < dropdown (20) <
  dialog/modal (30) < toast notification (40). MUST NOT use arbitrary
  z-index values outside this defined hierarchy.
- Content Density Rules: monitor list/table view MUST prioritize density
  (more rows visible) over card-based decorative spacing.

---

## Interaction Rules

- Hover Behavior: MUST provide visible hover state on all interactive
  elements (buttons, table rows, links).
- Focus Behavior: MUST provide visible focus outline on all interactive
  elements — STRICT: MUST NOT remove default focus outline without
  providing an equivalent custom visible replacement.
- Active States: MUST provide pressed/active visual feedback on buttons.
- Disabled States: MUST reduce opacity and MUST prevent pointer
  interaction (`cursor: not-allowed`) on disabled elements.
- Keyboard Navigation: all primary flows (create monitor, view detail,
  logout) MUST be completable via keyboard alone.
- Touch Interactions: MUST provide minimum 44x44px touch target size for
  all interactive elements on mobile/tablet breakpoints.

---

## Motion Rules

- Transitions: state color changes MUST use a short transition (150-200ms)
  — MUST NOT be instant (jarring) or slow (>300ms, feels laggy).
- Animation Timing: MUST use ease-out for entrances, ease-in for exits.
- Duration Limits: MAX_LIMIT 300ms for any UI transition in MVP scope.
- Motion Accessibility: MUST respect `prefers-reduced-motion` media query
  — MUST disable non-essential transitions when set.
- Performance Constraints: MUST NOT animate layout-affecting properties
  (width/height/top/left) — MUST use `transform`/`opacity` only.

---

## Accessibility Rules

- WCAG Target Level: AA.
- Keyboard Support: REQUIRED for all interactive flows (see Interaction
  Rules).
- Screen Reader Support: all icon-only buttons MUST have `aria-label`.
  Status badges MUST have accessible text equivalent, not conveyed by
  color/icon alone.
- Focus Visibility: REQUIRED, see Interaction Rules.
- Color Accessibility: see Color System contrast requirements.
- Touch Target Requirements: minimum 44x44px, see Interaction Rules.

---

## Responsive Rules

- Mobile Behavior: navigation MUST collapse to a hamburger/drawer pattern.
  Table views MUST convert to stacked card format below 640px (STRICT: no
  horizontal scroll tables on mobile).
- Tablet Behavior: sidebar MAY collapse to icon-only; table views remain
  tabular.
- Desktop Behavior: full sidebar + tabular views.
- Content Adaptation Rules: monitor detail page MUST reflow chart from
  full-width to stacked-below-summary on mobile.
- Navigation Adaptation Rules: primary nav MUST remain reachable within 1
  tap/click at all breakpoints.

---

## Design Tokens

```json
{
  "colors": {
    "healthy": "#22c55e",
    "running": "#3b82f6",
    "late": "#eab308",
    "failed": "#ef4444",
    "pending": "#9ca3af"
  },
  "spacing": [4, 8, 12, 16, 24, 32, 48],
  "typography": {
    "scale": [12, 14, 16, 20, 24, 32],
    "sans": "Inter, system-ui, sans-serif",
    "mono": "JetBrains Mono, monospace"
  },
  "radius": { "sm": 4, "md": 8, "lg": 12 },
  "shadows": {
    "sm": "0 1px 2px rgba(0,0,0,0.05)",
    "md": "0 4px 6px rgba(0,0,0,0.1)"
  },
  "zIndex": { "nav": 10, "dropdown": 20, "modal": 30, "toast": 40 },
  "animation": { "fast": "150ms", "base": "200ms", "max": "300ms" }
}
```

---

## Forbidden Patterns

- Hidden navigation (nav MUST always be discoverable, no gesture-only
  access).
- Inaccessible contrast (below 4.5:1 for body text).
- Layout shifts on load (MUST use skeleton states, see Component System).
- Autoplay media (N/A for MVP scope but explicitly PROHIBITED if ever
  introduced).
- Excessive animations (see Motion Rules duration limits).
- Inconsistent component behavior (same component MUST behave identically
  across all screens it appears on).
- Modal stacking (MUST NOT open a dialog on top of another dialog).
- Destructive actions without confirmation (delete monitor MUST require
  explicit confirm dialog).

---

## Implementation Constraints

- Component Reuse Requirements: MUST reuse shadcn/ui primitives before
  creating any custom component from scratch.
- Styling Constraints: MUST use TailwindCSS utility classes exclusively —
  MUST NOT introduce separate CSS-in-JS libraries or additional styling
  frameworks.
- Framework Compatibility Requirements: MUST target React 18+ function
  components with hooks only — MUST NOT introduce class components.
- Performance Limits: dashboard initial load MUST NOT exceed 2s on 100
  monitors (per PRD.md Success Metrics) — chart rendering (Recharts) MUST
  be lazy-loaded on the monitor detail route only, MUST NOT be bundled
  into the main dashboard list bundle.

---

## Priority Order (Conflict Resolution)

1. Accessibility Rules
2. Usability Rules
3. Implementation Constraints
4. Consistency Rules
5. Visual Identity
6. Motion Rules

Higher-priority rules override lower-priority rules. Accessibility
overrides aesthetics. Usability overrides visual preference. Consistency
overrides individual component customization. When uncertain, prefer the
simpler implementation.
