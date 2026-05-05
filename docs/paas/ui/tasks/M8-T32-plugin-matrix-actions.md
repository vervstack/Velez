---
id: "032"
title: "PluginMatrix — enable/disable actions and service page navigation"
status: "pending"
model: "qwen2.5-coder:3b"
created: "2026-05-04"
branch: "task/032-plugin-matrix-actions"
---

# Task 032 — PluginMatrix: enable/disable actions and service page navigation

## Goal

Extend `PluginMatrix` so each plugin row is clickable (navigates to `/service/:key`) and each status cell shows an
enable/disable toggle button instead of a static badge.

## Context

`PluginMatrix` lives at `src/widgets/controlplane/PluginMatrix.tsx`. It receives `nodes: NodeBaseInfo[]` and
`plugins: PluginStatus[]` and renders a grid of static `Badge` components showing `enabled`/`disabled` per node.

`PluginStatus` (defined inside the same file) currently has:

```ts
interface PluginStatus {
    pluginName: string;
    tag: string;
    nodeStatuses: Record<string, 'enabled' | 'disabled'>;
}
```

`ControlPlanePage` at `src/pages/controlplane/ControlPlanePage.tsx` constructs `PluginStatus` from
`vervServices: Service[]` — the service's `type` field is used as `pluginName`. The `key` needed for `/service/:key`
routing is not yet passed down.

The router already has `Routes.Service = '/service'` and `Arguments.Key = 'key'` in `src/app/router/Routes.ts`. The
destination page `ServiceInfoPage` at `src/pages/service/ServiceInfoPage.tsx` already exists.

Action callbacks (`onEnable`, `onDisable`) must be passed in as props from `ControlPlanePage` — do not call API stubs
directly from the widget. `ControlPlanePage` will wire stub handlers (e.g. `alert('not implemented')`) for now.

## Acceptance Criteria

- [x] `PluginStatus` interface gains an optional `serviceKey?: string` field used for navigation
- [x] Clicking a plugin name cell navigates to `/service/:serviceKey` using `react-router-dom` `useNavigate`; if
  `serviceKey` is absent, the click is a no-op
- [x] Each status cell renders an `IconButton` (enable or disable) instead of a static `Badge`, using the existing
  `src/components/base/IconButton.tsx`
- [x] `PluginMatrix` accepts `onEnable(pluginName: string, nodeId: string)` and
  `onDisable(pluginName: string, nodeId: string)` callback props
- [x] `ControlPlanePage` passes stub `onEnable`/`onDisable` handlers (shows `alert('not implemented')`) and populates
  `serviceKey` from `service.name` or `service.type`
- [ ] `yarn build:ui` passes with no type errors — BLOCKED by pre-existing error in `PageHeader.tsx` (Routes.Home
  missing)

## Files to Create / Modify

- `src/widgets/controlplane/PluginMatrix.tsx` — add `serviceKey` to `PluginStatus`, navigation on plugin name click,
  `onEnable`/`onDisable` props, replace `Badge` cells with `IconButton`
- `src/widgets/controlplane/PluginMatrix.module.css` — style clickable plugin name row and button cells
- `src/pages/controlplane/ControlPlanePage.tsx` — pass `serviceKey`, `onEnable`, `onDisable` to `PluginMatrix`

## Do NOT change

- `src/app/api/velez/` — auto-generated stubs, never edit manually
- `src/pages/service/ServiceInfoPage.tsx` — destination page, no changes needed
- `src/app/router/Routes.ts` — route constants already defined

## Notes

- Navigation must use `useNavigate` from `react-router-dom`, not `<a href>` or `window.location`
- Plugin name cell should visually indicate it is clickable (cursor pointer, subtle hover) only when `serviceKey` is
  present
- Enable/disable icons: use a simple circle-check / circle-x or similar — pick from whatever icon set is already
  imported in the project; do not add a new icon dependency