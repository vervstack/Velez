---
id: "038"
title: "PluginMatrix: restore status display + separate Manage button with empty dialog"
status: "pending"
model: "qwen2.5-coder:3b"
created: "2026-05-05"
branch: "task/038-plugin-matrix-status-manage-dialog"
---

# Task 038 — PluginMatrix: restore status display + separate Manage button with empty dialog

## Goal

Replace the current enable/disable IconButton in each cell with a `StatusDot` showing the plugin's enabled/disabled
status, and add a separate per-row "Manage" button that opens an empty modal dialog.

## Context

`PluginMatrix.tsx` lives at `src/widgets/controlplane/PluginMatrix.module.css`. Currently each cell renders an
`IconButton` whose label is `✓` or `✕` and whose click handler calls `onEnable`/`onDisable`. The user wants:

1. **Status display** — each cell should show a `StatusDot` (from `@/components/base/StatusDot`) indicating `enabled` (
   green) or `disabled` (grey). No click action on the dot itself.
2. **Manage button** — a single "Manage" button per plugin row (not per node cell), placed at the right end of the row (
   or in the plugin name cell area). Clicking it opens a modal dialog.
3. **Dialog** — a simple modal overlay with a title ("Manage {pluginName}") and a close button. Body is empty for now.
   The dialog must be implemented as a separate component `PluginManageDialog` in
   `src/components/complex/PluginManageDialog/`.

Keep the `onEnable`/`onDisable` props on `PluginMatrixProps` — they are wired to real API calls elsewhere — but stop
calling them from cell clicks. The Manage button will wire them later once the dialog body is built.

## Acceptance Criteria

- [ ] Each node cell shows a `StatusDot` (green = enabled, grey/red = disabled) instead of an IconButton
- [ ] The `StatusDot` is not clickable — it is purely informational
- [ ] A "Manage" button appears once per plugin row (not per node)
- [ ] Clicking "Manage" opens `PluginManageDialog` for that plugin
- [ ] `PluginManageDialog` renders a modal overlay with title "Manage {pluginName}" and a close button; body is empty
- [ ] Closing the dialog (via close button or overlay click) hides it
- [ ] `onEnable` / `onDisable` props are preserved on the component but not called from any UI element yet
- [ ] `yarn build:ui` passes with no TypeScript errors

## Files to Create / Modify

- `src/widgets/controlplane/PluginMatrix.tsx` — replace cell IconButton with StatusDot, add Manage button per row
- `src/widgets/controlplane/PluginMatrix.module.css` — adjust cell layout for dot + new button column
- `src/components/complex/PluginManageDialog/PluginManageDialog.tsx` — new empty dialog component
- `src/components/complex/PluginManageDialog/PluginManageDialog.module.css` — new

## Do NOT change

- `src/components/base/StatusDot.tsx` — use as-is
- `src/components/base/IconButton.tsx` — may use for close button inside dialog
- `src/pages/controlplane/ControlPlanePage.tsx` — props interface is unchanged
- Any process / API layer files

## Notes

- Follow the CSS Modules coding rules: root class suffix `Container`, wrapper suffix `Wrapper`, `rem` units, no inline
  styles.
- The dialog overlay should use a CSS `position: fixed` overlay; no z-index hardcoding (use CSS custom property or layer
  order).
- `StatusDot` accepts a `status` prop — check its existing API before writing the call.
