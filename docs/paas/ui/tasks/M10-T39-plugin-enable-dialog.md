---
id: "039"
title: "PluginManageDialog: enable action with per-plugin config forms"
status: "pending"
model: "qwen2.5-coder:3b"
created: "2026-05-06"
branch: "task/039-plugin-enable-dialog"
---

# Task 039 — PluginManageDialog: enable action with per-plugin config forms

## Goal

Fill `PluginManageDialog` with a status display and an "Enable" action — including config forms for plugins that require
extra parameters (`statefull_pg`, `headscale`) — and wire it to the `EnableService` API.

## Context

`PluginManageDialog` (`src/components/complex/PluginManageDialog/PluginManageDialog.tsx`) was created in T38 with an
empty body. Its props are `{ isOpen, pluginName, onClose }`. The parent `PluginMatrix` opens it when the user clicks
"Manage" on a row.

`ControlPlanePage` passes `onEnable` / `onDisable` to `PluginMatrix` but they currently just `alert()`.

The backend `EnableService` RPC (`POST /api/control_plane/service/enable`) accepts:

```ts
EnableServiceRequest {
  service: VervServiceType;   // which plugin to enable
  statefullCluster?: EnableStatefullCluster;   // only for statefull_pg
  headscaleServer?: EnableHeadscaleServer;     // only for headscale
}
```

The existing process function `EnableService(vervService, initReq)` in
`src/processes/api/control_plane.ts` already handles simple enable. There is also
`EnableStatefullPgCluster(payload, initReq)` for the stateful cluster case. There is no
`DisableService` API — the dialog should only surface the Enable action for now.

Plugin types and their payloads:

| `VervServiceType` | Extra config needed                                            |
|-------------------|----------------------------------------------------------------|
| `matreshka`       | none                                                           |
| `makosh`          | none                                                           |
| `webserver`       | none                                                           |
| `portainer`       | none                                                           |
| `statefull_pg`    | `EnableStatefullCluster` (`is_expose_port`, `expose_to_port`)  |
| `headscale`       | `EnableHeadscaleServer` (deploy config OR external connection) |

After a successful enable call the page must refresh the plugin list. Use
`queryClient.invalidateQueries({ queryKey: [CacheKey.Plugins] })` from
`@tanstack/react-query`.

## Acceptance Criteria

- [ ] Dialog body shows the plugin's current state (e.g. "Status: disabled")
- [ ] For simple plugins (`matreshka`, `makosh`, `webserver`, `portainer`): an "Enable" button appears when the plugin
  is
  disabled; clicking it calls `EnableService`, shows a success toast, closes the dialog, and invalidates the plugins
  query
- [ ] For `statefull_pg`: an optional "Expose port" checkbox and a port number input (visible only when checked) appear;
  confirming calls `EnableStatefullPgCluster` with the collected payload
- [ ] For `headscale`: a radio toggle between "Deploy new" (optional custom port + optional custom image) and "Connect
  to
  external" (URL + token inputs); confirming calls `EnableService` with the correct `headscaleServer` payload
- [ ] While the enable call is in-flight the button shows a loading state and is disabled
- [ ] On API error a toast is shown via `useToaster().catchGrpc`
- [ ] `ControlPlanePage` wires `onEnable` to open the dialog for the correct plugin (remove the `alert()` stub)
- [ ] `PluginManageDialog` receives the plugin's current `VervService.State` so it can render the status line
- [ ] `yarn build:ui` passes with no TypeScript errors

## Files to Create / Modify

- `src/components/complex/PluginManageDialog/PluginManageDialog.tsx` — add status display and enable action forms
- `src/components/complex/PluginManageDialog/PluginManageDialog.module.css` — style the dialog content area
- `src/widgets/controlplane/PluginMatrix.tsx` — pass plugin state into the dialog; pass `VervServiceType` not just name
- `src/pages/controlplane/ControlPlanePage.tsx` — replace `alert()` stubs with real `EnableService` call; invalidate
  query on success

## Do NOT change

- `src/processes/api/control_plane.ts` — process functions already exist; use them as-is
- `src/app/api/velez/control_plane_api.pb.ts` — auto-generated, never edit manually
- `src/components/base/IconButton.tsx`, `StatusDot.tsx` — use as-is

## Notes

- `PluginManageDialog` is a pure UI component (`src/components/complex/`); it must not call API functions directly. Pass
  an `onEnable` callback from the parent widget/page and call it from the dialog.
- Prefer a single `onEnable(type, payload?)` callback signature over separate per-plugin callbacks.
- Follow CSS Modules rules: root class `Container`, wrapper `Wrapper`, `rem` units, no inline styles.
- The dialog must not crash when opened for a plugin with `VervServiceType.unknown_service_type`.