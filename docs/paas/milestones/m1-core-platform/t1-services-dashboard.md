# T1 — Services Dashboard

## Goal

Give the operator a clear at-a-glance view of every service registered in Velez and a drill-down into one service.

## Tasks

### 1.1 Services list (HomePage)
- Fetch services via `ListServices` (already exists in `src/pages/HomePage`)
- Display name, image, current state (running / stopped / unknown), and last-deployed timestamp
- Link each row to `/service/:key`
- Empty-state illustration when no services exist

### 1.2 Service detail (ServiceInfoPage)
- Header: service name, image tag, overall health indicator
- Quick-action bar: Stop / Restart / Open deploy flow
- Metadata section: environment variables, volumes, port bindings — read-only display
- Navigation back to home

### 1.3 Smerds list (smerd containers)
- On `HomePage`, show smerds alongside services (already fetched; needs layout cleanup)
- Status badge: running / exited / created
- Link to `/smerd/:name`

### 1.4 Smerd detail (SmerdPage)
- Container metadata: image, ports, status, started-at
- Raw container inspect collapsible section (for debugging)

## Acceptance criteria

- List refreshes on a short polling interval (React Query `refetchInterval`)
- State changes (e.g. container stops) are reflected without a page reload
- No API calls made directly from components — all go through `src/processes/api/`
