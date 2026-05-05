---
id: "M9-T33"
title: "ServiceInfoPage — full redesign with design system"
status: "pending"
model: "stable-code"
created: "2026-05-04"
branch: "task/M9-T33-service-page-redesign"
---

# Task M9-T33 — ServiceInfoPage: Full Redesign

## Goal

Rebuild `ServiceInfoPage` visual layout to match the dark design system established in M1–M8, replacing the old unthemed
CSS with proper design tokens and CSS Modules structure.

## Context

`src/pages/service/ServiceInfoPage.tsx` already has working data fetching (service, deployments, smerds via React Query)
and working actions (deploy, stop, restart, remove). The CSS (`ServiceInfoPage.module.css`) predates the design system
and uses hardcoded colors, no design tokens, and an ad-hoc layout.

Design tokens are in `src/index.module.css`. Reference components for style: `ServiceCard.tsx`, `NodeCard.tsx`,
`StatCard.tsx`, `Badge.tsx`, `StatusDot.tsx`.

The page structure should be rebuilt as a three-section layout:

1. **Header** — service name, `StatusDot`, action buttons (Deploy, Stop, Restart, Remove) — right-aligned
2. **Info grid** — two-column grid of labeled meta rows (ID, image, current deployment, status)
3. **Tabbed panels** — tabs: Overview | Deployments | Config — each renders the relevant sub-section

The existing `SmerdMetaSection` (ports, volumes, env vars) belongs in the Config tab. `DeploymentsSection` (paginated
list) belongs in the Deployments tab. Overview tab shows the info grid + placeholder slots for observability links and
runtime stats (to be filled by T34 and the BE task).

## Acceptance Criteria

- [ ] Page uses only design token CSS variables (`--color-*`, `--radius-*`, `--space-*`); no raw hex/px color values
- [ ] Header row renders service name, `StatusDot` for status, and action buttons in a single flex row
- [ ] Three tabs (Overview / Deployments / Config) switch content without a page reload
- [ ] Deployments tab renders the existing paginated deployment list, visually consistent with the new design
- [ ] Config tab renders ports, volumes, env vars from existing `SmerdMetaSection` logic
- [ ] Overview tab renders the meta grid; placeholder `<div>` slots present for observability links and runtime stats (
  so T34 can slot in without restructuring the page)
- [ ] `yarn build:ui` passes with no TypeScript errors

## Files to Create / Modify

- `src/pages/service/ServiceInfoPage.tsx` — rebuild layout, add tab state
- `src/pages/service/ServiceInfoPage.module.css` — full rewrite using design tokens
- `src/pages/service/parts/Header.tsx` — update to match new action bar design
- `src/pages/service/parts/Header.module.css` — update styling

## Do NOT change

- `src/processes/api/service.ts` — data fetching logic stays as-is
- `src/pages/service/parts/DeployMenu.tsx` — dialog content unchanged
- Any backend files

## Notes

- Use `function` declarations, not arrow functions for all components and handlers.
- The tab switcher is local `useState` — no router params needed.
- `StatusDot` component is at `src/components/base/StatusDot/StatusDot.tsx`.
- Coding rules from ROADMAP.md apply (rem units, CSS nesting, no `!important`).