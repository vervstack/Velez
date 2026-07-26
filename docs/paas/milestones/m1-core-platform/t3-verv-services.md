# T3 — Verv Services Management

## Goal

Full create / read / delete lifecycle for Verv services registered in Velez, with enough context to understand what each service is and how to act on it.

## Tasks

### 3.1 Verv services list (HomePage or dedicated page)

- [x] Fetch via `ListServices`, display name with link to detail
- [x] Columns: source repo, image, registered-at, action buttons (`HomePage.tsx` card grid — added real `status` badge (was hardcoded "unknown"), `repo` field (new `ServiceBaseInfo.repo`, sourced from the smerd's repo label), and per-card Stop/Restart/Remove actions)
- [x] Filter / search bar for operators with many services

### 3.2 New service form (NewServicePage)

- [x] NewServicePage exists and calls `CreateService`
- [x] Verify all fields: name, image, tag, env vars, volumes, port bindings (using `InitServiceWidget`) — verified: `CreateService.Request` only defines `name`, so `InitServiceWidget`'s single-field form is a correct match for the contract; image/tag/env/volumes/ports are configured later via `CreateDeploy`, not at service-creation time
- [x] On success: redirect to the new service's detail page

### 3.3 Service delete

- [x] "Remove" action on the service detail page (was previously not implemented at all despite being checked here — corrected; now backed by a new `RemoveService` RPC end-to-end)
- [x] Confirmation dialog stating the service name (`RemoveServiceDialog`, includes a "drop running instances" checkbox — delete is rejected server-side if a smerd/instance is attached unless that flag is set)
- [x] Navigate back to list on success (from ServiceInfoPage; HomePage's own card action refetches the list in place)

### 3.4 Edit service (stretch goal for M1)

- [ ] Allow updating env vars or image tag in place

## Acceptance criteria

- [ ] Create round-trip is fully testable end-to-end against a running backend — not verified this round (needs a live Postgres+Docker backend, not exercised headlessly)
- [x] Delete requires explicit confirmation (no accidental removal)
- [x] Form state is cleared on successful submit — N/A: `NewServicePage` navigates away immediately on success, nothing to visibly clear
