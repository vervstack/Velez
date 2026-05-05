---
id: "M9-T37"
title: "Settings page — routed page for backend URL and auth token"
status: "pending"
model: "stable-code"
created: "2026-05-04"
branch: "task/M9-T37-settings-page"
---

# Task M9-T37 — Settings Page

## Goal

Create a dedicated `/settings` route that renders the connection settings form (backend URL and auth token) as a full
page, and wire the "Settings" tool item in the sidebar to navigate there.

## Context

`src/widgets/settings/SettingsWidget.tsx` currently renders as a modal overlay triggered from somewhere in the app. The
form stores values via `useCredentialsStore` from `src/app/settings/creds.ts`.

The sidebar's `TOOL_ITEMS` array in `src/widgets/sidebar/Sidebar.tsx` includes a
`{id: 'settings', label: 'Settings', icon: '◈'}` entry that renders as a static `<div>` with no click handler.

This task creates `src/pages/settings/SettingsPage.tsx` that contains the same form logic (backend URL, auth token with
show/hide, reset button, save button, build-time env display). The page is a standalone route — the modal overlay widget
is left untouched.

The sidebar "Settings" tool item should call `navigate('/settings')` when clicked. Since `TOOL_ITEMS` are currently
static divs, the sidebar needs to accept a `onToolNav` callback (or use `useNavigate` internally — prefer the callback
approach to keep Sidebar a controlled component).

`MainLayout` should handle the callback and call `navigate(Routes.Settings)`.

## Acceptance Criteria

- [ ] Navigating to `/settings` renders the settings form (URL input, token input with show/hide, reset, save)
- [ ] Save persists values to `useCredentialsStore` and shows a success toast or navigates back
- [ ] Clicking "Settings" in the sidebar tool section navigates to `/settings`
- [ ] The `/settings` route is registered in `src/app/router/Router.tsx`
- [ ] `Routes.Settings = '/settings'` is added to `src/app/router/Routes.ts`
- [ ] `yarn build:ui` passes with no TypeScript errors

## Files to Create / Modify

- `src/pages/settings/SettingsPage.tsx` — new page component with the credentials form
- `src/pages/settings/SettingsPage.module.css` — new, styled with design tokens
- `src/app/router/Routes.ts` — add `Settings = '/settings'`
- `src/app/router/Router.tsx` — add route for `/settings`
- `src/widgets/sidebar/Sidebar.tsx` — add `onToolNav?: (id: string) => void` prop; call it from the settings tool item's
  click handler
- `src/app/router/MainLayout.tsx` — pass `onToolNav` to `Sidebar`, handle `'settings'` → navigate to `Routes.Settings`

## Do NOT change

- `src/widgets/settings/SettingsWidget.tsx` — modal version is left as-is
- `src/app/settings/creds.ts` — credentials store is reused unchanged

## Notes

- The page layout should follow the same card/section structure as other M9 pages.
- Form logic can be a direct copy-paste from `SettingsWidget` adapted to page context (no `isOpen`/`onClose` needed).
- Use function declarations for all handlers; CSS Modules with design tokens; rem units.
- Coding rules from ROADMAP.md apply.