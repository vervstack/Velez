---
id: "M9-T36"
title: "Fix sidebar active nav highlight + service page parent highlight"
status: "pending"
model: "stable-code"
created: "2026-05-04"
branch: "task/M9-T36-sidebar-active-nav"
---

# Task M9-T36 — Sidebar: Fix Active Nav Highlight

## Goal

Fix the sidebar nav item highlight so the active page is correctly indicated, and ensure sub-pages (e.g.
`/service/:key`) highlight their parent nav item.

## Context

`src/widgets/sidebar/Sidebar.tsx` — `NavItem` receives `isActive` computed as `activeNodeId == n.id` (line ~105).
`activeNodeId` is the selected node's UUID, never equal to a nav ID string like `'apps'`, so nav items are never
visually active regardless of the current route.

The correct prop to use is `activeNav` (also passed to `Sidebar` via props). Fix: replace `activeNodeId == n.id` with
`activeNav === n.id` in the `NAV_ITEMS.map(...)` call.

In `src/app/router/MainLayout.tsx`, `ROUTE_TO_NAV` maps exact route strings to nav IDs. Routes like `/service/:key` and
`/smerd/:name` are not in the map, so they fall back to `'apps'` via the `?? 'apps'` default — this is correct behavior.
No change needed in `MainLayout.tsx` unless the default fallback is wrong.

Verify by checking `Routes.ts` for any routes that should highlight a different nav item than `'apps'`:

- `/cp` → `'controlplane'` ✓
- `/vcn` → `'vcn'` ✓
- `/deployments` → `'deployments'` ✓
- `/apps`, `/service/:key`, `/smerd/:name` → `'apps'` (correct, service/smerd are sub-pages of Apps)
- `/search` → `'search'` ✓
- `/deploy`, `/new_verv_service` → arguably `'deployments'` but not critical

## Acceptance Criteria

- [ ] Navigating to `/apps` highlights the "Apps" nav item in the sidebar
- [ ] Navigating to `/service/:key` also highlights "Apps" (parent page)
- [ ] Navigating to `/cp` highlights "Control Plane"
- [ ] Navigating to `/deployments` highlights "Deployments"
- [ ] `yarn build:ui` passes with no TypeScript errors

## Files to Create / Modify

- `src/widgets/sidebar/Sidebar.tsx` — change `activeNodeId == n.id` to `activeNav === n.id` in `NAV_ITEMS.map`
- `src/app/router/MainLayout.tsx` — optionally add prefix-based matching for `/service` and `/smerd` routes if the
  `?? 'apps'` fallback is not sufficient

## Do NOT change

- Sidebar CSS — visual-only fix to the prop passed to `NavItem`
- Node selection logic (`activeNodeId`, `onNodeSelect`) — unrelated

## Notes

- The fix is a one-line change in `Sidebar.tsx`; the rest is verification.
- If `location.pathname` starts with `/service` or `/smerd` but the default fallback handles it, no change to
  `MainLayout.tsx` is needed — confirm before editing.