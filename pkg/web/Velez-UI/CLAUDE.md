# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
  # generate proto files
  bun gen
```

## Exploration Rules

- Do NOT scan the full project
- Read only files directly relevant to the specific task
- Ask which file to edit if uncertain — don't explore to find out
- Never read more than 3 files before acting


## Environment

`VITE_VELEZ_BACKEND_URL` — backend URL (grpc-gateway HTTP endpoint, default: the origin the UI is served from).  
`VITE_VELEZ_AUTH_HEADER` — optional auth header value.

Both can be overridden in `.env.local`. At runtime the user can also change the backend URL and auth header via the Settings widget; values are persisted to `localStorage` under the key `"settings"`.

## Architecture

### Layer structure

Layers are ordered: lower = more primitive. A layer may only import from layers below it — never from layers above.

| Layer | Path | Role |
|---|---|---|
| Pages | `src/pages/` | Route-level components, one directory per route. No reuse expected. |
| Widgets | `src/widgets/` | Self-contained feature blocks with business logic, composed into pages |
| Segments | `src/segments/` | Layout-level pieces shared across pages (PageHeader, Toaster) |
| Components | `src/components/` | Pure UI atoms — `base/` for primitives, `complex/` for composed elements. No business logic, no direct store access. |
| Processes | `src/processes/api/` | **All** gRPC calls must go here. Never call generated stubs directly from components or widgets. |
| Model | `src/model/` | App-level types (e.g. `Smerd`, `Service`) decoupled from proto types |
| App | `src/app/` | App wiring: routing, generated API clients (`src/app/api/`), layouts, settings |

`src/app/api/velez/` contains auto-generated grpc-gateway-ts stubs (`*.pb.ts`) — **do not edit manually**.

### Routing

Routes are defined in `src/app/router/Router.tsx` using `createBrowserRouter`. All routes are children of `MainLayout` which renders `<PageHeader>`, `<Outlet>`, and `<Toaster>`.

| Route | Page |
|---|---|
| `/` | HomePage — lists smerds and services |
| `/smerd/:name` | SmerdPage — single container detail |
| `/service/:key` | ServiceInfoPage — Verv service detail + deploy |
| `/new_verv_service` | NewServicePage |
| `/deploy` | DeployPage |
| `/cp` | ControlPlanePage |
| `/vcn` | VervClosedNetworkPage |

### API calls

All API calls go through `src/processes/api/` which calls the generated stubs in `src/app/api/velez/`. Each call receives an `InitReq` (`{ pathPrefix, headers }`) built from the settings hook. React Query is used for data fetching in pages/widgets.

### State management

- **Zustand** — global singleton stores (e.g. `useToaster` in `src/app/hooks/toaster/Toaster.ts`)
- **`useSettings` hook** — backend URL + auth header, persisted to localStorage
- **React Query** — server state / caching for API responses

### Toast / error handling

`useToaster` (Zustand store) exposes `bake(toast)`, `dismiss(title)`, and `catchGrpc(error)`. Call `catchGrpc` in `.catch()` blocks on API calls to surface gRPC errors as toasts. Toasts auto-dismiss after 5 seconds.

### Dialogs

One global dialog primitive — the `useDialog` Zustand store in `src/app/hooks/dialog/Dialog.tsx`, plus a single `<Dialog/>` host mounted once in `MainLayout`. Do not build bespoke modal state per feature.

- **Open:** `const { OpenDialog } = useDialog(); OpenDialog(<CreateAppDialog prop={x}/>)`. Trailing args stack as separate cards.
- **Close:** `const { CloseDialog } = useDialog()` — call it after a successful action and from any Cancel button.
- **Guard async work:** `LockClosing()` before an in-flight request, `UnlockClosing()` in `.finally()`. While locked, click-off and `CloseDialog()` are no-ops — call `UnlockClosing()` immediately before the success-path `CloseDialog()`.
- **One dialog per folder:** `src/dialogs/<Name>/<Name>.tsx` + `<Name>.module.css`, default-exported named function, root class `<Name>Container`. Sequential steps go in `src/dialogs/<Name>/screens/`, local pieces in `components/`.
- **Nested dialogs** (confirm, sub-steps) open via the same `OpenDialog` from inside a dialog — see `EnvironmentManageDialog` → `EnvironmentDeleteDialog`. A dialog never imports a page.
- **Opening from anywhere:** any layer may import `dialogs/` solely to call `OpenDialog(<X/>)` — the one sanctioned upward import.
- **Errors** inside a dialog route through `useToaster().catchGrpc`; never a native `alert`/`confirm`.
- **Chrome comes from `DialogShell`** (`src/components/DialogShell/`), never hand-rolled: the dialog root is
  `<div className={cls.<Name>Container}><DialogShell title=… onClose={CloseDialog}>…</DialogShell></div>`. The
  `<Name>Container` class sets width only (`width` + `max-width: var(--dialog-max-width)`) — no background,
  border, shadow, header or padding of its own.
- **Header is pinned, body scrolls.** The shell caps itself at `--dialog-max-height` and scrolls only its body,
  so the title and close button never scroll away. Never put `overflow-y: auto` on the dialog root.
- **Close button is always in the header** (`onClose`) — the shell renders the project's standard
  `<Button variant="ghost" sm>✕</Button>`. Omit it only while closing is deliberately blocked (an in-flight task
  screen).
- **Body padding and spacing belong to the shell** (`--dialog-padding`, flex column with gap). Children are the
  fields wrapper and the actions row directly — no extra `Content` div, no outer padding of their own. A screen
  that pads itself (`TaskProgressScreen`) renders with `isFlush`.

### Proto regeneration

`moti.yaml` configures the `moti` tool to pull proto files from the Velez git repo and generate TypeScript via `grpc-gateway-ts` and `npm` plugins. The `replace` block redirects the module to the local checkout. After generation, `gen-proto` script moves the top-level `index.ts` out of the nested `@vervstack` directory.

## Coding Rules

- **Form controls come from `@vervstack/chures`** (`Checkbox`, `Toggle`, `Dropdown`) everywhere, not only in
  dialogs. Before reaching for any `@/components/base/*` atom, check the kit's exports first.
  - **`Toggle` = a setting that applies immediately, or a view filter. `Checkbox` = a field that is submitted with a
    button (dialog/form).** A raw `<input type="checkbox">` is a lint error (`local/no-raw-checkbox`).
- Components must be named function declarations — not `const Arrow = () => {}`
- All functions inside components (handlers, helpers) must also be named function declarations — never `const fn = () => {}`
- One file — one component
- `@/` resolves to `src/` — use it for all imports
- Component can't have more that three levels of nesting. A root '*ComponentName*Container', Content elements and '*ComponentName*Wrapper's over them.
- If component can't use 3 levels of nesting model, it should be a separate into a private components below (unless components are reusable - then just ask and we feagure it out).
## Styling Rules

- Always use CSS Modules — no inline styles for new code
- Use CSS nesting for child selectors inside a module
- Root style for a component must use the suffix `Container`; wrapper classes use suffix `Wrapper`
- Do not use `!important` or `z-index`
- Use `rem` units for font sizes and spacing; avoid hardcoded `px`/`em` in component CSS
- Animations: CSS `transition`/`animation`/`@keyframes` first — use `framer-motion` only when CSS cannot achieve the effect
- Every appearance, disappearance (mount/unmount), and movement/reorder of a component or element must be animated — never an abrupt snap. Prefer a CSS `transition`/`animation`/`@keyframes` implementation; only reach for `framer-motion` when CSS genuinely cannot express the effect (e.g. list-reorder FLIP animations can still be done with CSS `transition: transform` driven by JS-measured offsets — that still counts as CSS-based).
- One component must have no more than three levels of enclosure: Root div, wrapper around content(s), content components. 
  Everything that doesn't fit - should be a separate component.
