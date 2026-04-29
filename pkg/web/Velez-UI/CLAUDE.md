# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Exploration Rules

- Do NOT scan the full project
- Read only files directly relevant to the specific task
- Ask which file to edit if uncertain — don't explore to find out
- Never read more than 3 files before acting

## Commands

```bash
yarn dev           # Start dev server (Vite)
yarn build:ui      # Type-check + Vite build (UI only)
yarn build         # Build TypeScript client lib + link it + build UI
yarn lint          # ESLint

# Regenerate protobuf TypeScript bindings (requires moti CLI)
yarn gen-proto     # Runs moti g, then moves generated index.ts into place
```

The build has two parts that must be kept in sync:
1. `../@@vervstack/velez/` — TypeScript client library compiled from proto definitions
2. This UI app — consumes the library via `npm link @vervstack/velez`

When proto files change in `api/grpc/`, run `yarn gen-proto` to regenerate `src/app/api/velez/`, then rebuild the library via `yarn build:api && yarn link` before starting the dev server.

## Environment

`VITE_VELEZ_BACKEND_URL` — backend URL (grpc-gateway HTTP endpoint, default `http://0.0.0.0:53891`).  
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

### Proto regeneration

`moti.yaml` configures the `moti` tool to pull proto files from the Velez git repo and generate TypeScript via `grpc-gateway-ts` and `npm` plugins. The `replace` block redirects the module to the local checkout. After generation, `gen-proto` script moves the top-level `index.ts` out of the nested `@vervstack` directory.

## Coding Rules

- Components must be named function declarations — not `const Arrow = () => {}`
- All functions inside components (handlers, helpers) must also be named function declarations — never `const fn = () => {}`
- One file — one component
- `@/` resolves to `src/` — use it for all imports

## Styling Rules

- Always use CSS Modules — no inline styles for new code
- Use CSS nesting for child selectors inside a module
- Root style for a component must use the suffix `Container`; wrapper classes use suffix `Wrapper`
- Do not use `!important` or `z-index`
- Use `rem` units for font sizes and spacing; avoid hardcoded `px`/`em` in component CSS
- Animations: CSS `transition`/`animation`/`@keyframes` first — use `framer-motion` only when CSS cannot achieve the effect