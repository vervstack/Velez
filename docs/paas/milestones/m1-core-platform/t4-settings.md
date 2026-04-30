# T4 — Settings Panel

## Goal

Operators can configure the connection to the Velez backend and any per-session preferences without editing files.

## Tasks

### 4.1 Settings widget (already partially exists)
- Backend URL input — validated as a valid HTTP/HTTPS URL
- Auth header input — masked by default, reveal-on-click toggle
- Save button — persists to `localStorage` under key `"settings"`
- Reset-to-defaults button

### 4.2 Connection health indicator
- Small status indicator in the page header (green / yellow / red)
- On mount and after settings save, fire a lightweight health-check call (e.g. `ListSmerds` with a short timeout)
- Tooltip showing last-checked time and error message if unhealthy

### 4.3 Settings validation
- Block save if the URL is empty or malformed
- Warn (but don't block) if the URL is HTTP instead of HTTPS
- Warn if auth header looks empty while the backend likely requires one

### 4.4 Environment variable display (informational)
- Read-only section showing which `VITE_*` env vars were baked in at build time
- Helps operators debug mismatches between build-time and runtime config

## Acceptance criteria

- Changing backend URL immediately affects subsequent API calls without a page reload
- Settings survive a full page refresh
- Auth header is never logged to the browser console
