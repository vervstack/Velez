# T3 — Verv Services Management

## Goal

Full create / read / delete lifecycle for Verv services registered in Velez, with enough context to understand what each service is and how to act on it.

## Tasks

### 3.1 Verv services list (HomePage or dedicated page)
- Fetch via `ListServices`
- Columns: name, source repo, image, registered-at, action buttons
- Filter / search bar for operators with many services

### 3.2 New service form (NewServicePage)
- Fields: service name, image name, tag, environment variables (key-value pairs, addable rows), volumes, port bindings
- Uses `CreateService` API call
- Inline validation: required fields, image name format
- On success: redirect to the new service's detail page

### 3.3 Service delete
- "Remove" action on the service detail page
- Confirmation dialog stating the service name
- Calls `DropSmerd` / appropriate API; navigates back to list on success

### 3.4 Edit service (stretch goal for M1)
- Allow updating env vars or image tag in place
- Only if the underlying API supports partial updates — skip if not

## Acceptance criteria

- Create round-trip is fully testable end-to-end against a running backend
- Delete requires explicit confirmation (no accidental removal)
- Form state is cleared on successful submit
