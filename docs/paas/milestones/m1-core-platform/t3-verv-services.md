# T3 — Verv Services Management

## Goal

Full create / read / delete lifecycle for Verv services registered in Velez, with enough context to understand what each service is and how to act on it.

## Tasks

### 3.1 Verv services list (HomePage or dedicated page)

- [x] Fetch via `ListServices`, display name with link to detail
- [ ] Columns: source repo, image, registered-at, action buttons — only name + status shown
- [ ] Filter / search bar for operators with many services

### 3.2 New service form (NewServicePage)

- [x] NewServicePage exists and calls `CreateService`
- [ ] Verify all fields: name, image, tag, env vars, volumes, port bindings (using `InitServiceWidget`)
- [ ] On success: redirect to the new service's detail page

### 3.3 Service delete

- [ ] "Remove" action on the service detail page — missing
- [ ] Confirmation dialog stating the service name
- [ ] Navigate back to list on success

### 3.4 Edit service (stretch goal for M1)

- [ ] Allow updating env vars or image tag in place

## Acceptance criteria

- [ ] Create round-trip is fully testable end-to-end against a running backend
- [ ] Delete requires explicit confirmation (no accidental removal)
- [ ] Form state is cleared on successful submit
