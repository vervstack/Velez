# M1 — Core Platform

**Goal:** a working browser UI where an operator can see what is running, inspect deployments, trigger new deploys, manage Verv services, and configure the connection to the backend.

## Exit criteria

- [x] Services list loads and reflects live backend state
- [ ] Each service shows its deployments and current status (deployments listed; service status always "unknown", no
  env/volumes/ports display)
- [x] User can trigger a new deployment from the UI
- [ ] Verv services can be created, viewed, and removed (create + view + delete UI done; backend delete API not yet wired)
- [x] Settings (backend URL, auth header) are editable and persisted

## Task groups

| File | Scope |
|------|-------|
| [t1-services-dashboard.md](t1-services-dashboard.md) | Services list page and service detail |
| [t2-deployments.md](t2-deployments.md) | Deployments list, status display, and deploy trigger |
| [t3-verv-services.md](t3-verv-services.md) | Verv service CRUD and lifecycle actions |
| [t4-settings.md](t4-settings.md) | Settings panel — connection config and preferences |
| [t5-ux-polish.md](t5-ux-polish.md) | Loading states, error handling, toasts, empty states |

## Approach

Increment on what already exists in `src/pages/` rather than rewriting. Each task group is a self-contained slice — UI, API process calls, and model types — so groups can be worked in parallel.
