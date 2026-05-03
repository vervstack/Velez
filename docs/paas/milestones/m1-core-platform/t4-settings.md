# T4 — Settings Panel

## Goal

Operators can configure the connection to the Velez backend and any per-session preferences without editing files.

## Tasks

### 4.1 Settings widget (already partially exists)

- [x] Backend URL input — validated as a valid HTTP/HTTPS URL
- [x] Auth header input — masked by default, reveal-on-click toggle
- [x] Save button — persists to `localStorage`
- [x] Reset-to-defaults button

### 4.2 Connection health indicator

- [ ] Small status indicator in the page header (green / yellow / red)
- [ ] Health-check call on mount and after settings save
- [ ] Tooltip showing last-checked time and error message if unhealthy

### 4.3 Settings validation

- [x] Block save if the URL is empty or malformed
- [x] Warn if the URL is HTTP instead of HTTPS
- [x] Warn if auth header is empty

### 4.4 Environment variable display (informational)

- [x] Read-only section showing `VITE_VELEZ_BACKEND_URL` and `VITE_VELEZ_AUTH_HEADER` baked at build time

## Acceptance criteria

- [x] Changing backend URL immediately affects subsequent API calls without a page reload
- [x] Settings survive a full page refresh
- [x] Auth header is never logged to the browser console
