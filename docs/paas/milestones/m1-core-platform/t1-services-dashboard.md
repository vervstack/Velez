# T1 — Services Dashboard

## Goal

Give the operator a clear at-a-glance view of every service registered in Velez and a drill-down into one service.

## Tasks

### 1.1 Services list (HomePage)
- [x] Fetch services via `ListServices`
- [x] Display name and image (image cross-referenced from smerds list by name match)
- [ ] Last-deployed timestamp — requires API change: `ServiceBaseInfo` must include `lastDeployedAt`
- [x] Link each row to `/service/:key`
- [x] Empty-state illustration when no services exist

### 1.2 Service detail (ServiceInfoPage)
- [x] Header: service name + overall health indicator (colored status badge)
- [x] Image tag shown in metadata section (from first matching smerd)
- [ ] Stop / Restart actions — requires new API endpoints (`StopService`, `RestartService`) — not present in gRPC API
- [x] Metadata section: image, environment variables, volumes, port bindings (fetched via `FetchSmerdsByServiceName`)
- [x] Navigation back to home

### 1.3 Smerds list (smerd containers)
- [x] On `HomePage`, show smerds alongside services
- [x] Status badge: running / exited / created
- [x] Link to `/smerd/:name`

### 1.4 Smerd detail (SmerdPage)
- [x] Container metadata: image, ports, status, created-at
- [x] Raw container inspect collapsible section

## Acceptance criteria

- [x] List refreshes on a short polling interval (React Query `refetchInterval`)
- [x] State changes are reflected without a page reload
- [x] No API calls made directly from components — all go through `src/processes/api/`
