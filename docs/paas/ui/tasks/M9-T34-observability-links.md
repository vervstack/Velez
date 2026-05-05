---
id: "M9-T34"
title: "ObservabilityLinks widget for ServiceInfoPage"
status: "pending"
model: "stable-code"
created: "2026-05-04"
branch: "task/M9-T34-observability-links"
---

# Task M9-T34 — ObservabilityLinksPanel Widget

## Goal

Create an `ObservabilityLinksPanel` widget that renders a row of icon+label links to external observability tools (
Grafana, Loki, OpenSearch, Jaeger/Tempo), with URLs assembled from a configurable base URL and the current service name.

## Context

The `ServiceInfoPage` Overview tab (from T33) has a placeholder slot for this panel. This widget is purely frontend — no
new API needed. Tool URLs are configurable via the existing Settings mechanism (
`src/app/hooks/settings/useSettings.ts`). The settings store currently tracks `backendUrl` and `authHeader`; this task
extends it with optional observability base URLs.

Each link URL is constructed as `{baseUrl}/{serviceName}` (or a query param pattern) — the exact pattern per tool:

- Grafana: `{grafanaUrl}/d/service-overview?var-service={serviceName}`
- Loki (Grafana datasource): `{grafanaUrl}/explore?left=...&var-service={serviceName}`
- OpenSearch: `{opensearchUrl}/app/discover#{serviceName}`
- Jaeger: `{jaegerUrl}/search?service={serviceName}`

If a base URL is not configured, that link is hidden (not rendered as a broken link).

## Acceptance Criteria

- [ ] Widget renders as a horizontal strip of pill-shaped links, each with a tool icon (SVG or emoji fallback) and label
- [ ] Only links with a configured base URL are rendered; empty config shows "No observability links configured"
  placeholder text
- [ ] Clicking a link opens the URL in a new tab (`target="_blank" rel="noopener noreferrer"`)
- [ ] Settings panel (accessible from the existing Settings widget) exposes four new optional text inputs: Grafana URL,
  OpenSearch URL, Jaeger URL, Loki URL; values persisted to localStorage under `"settings"` key alongside existing
  fields
- [ ] `ObservabilityLinksPanel` is placed in the Overview tab placeholder slot from T33 — import it into
  `ServiceInfoPage.tsx`
- [ ] `yarn build:ui` passes with no TypeScript errors

## Files to Create / Modify

- `src/widgets/service/ObservabilityLinksPanel.tsx` — new widget
- `src/widgets/service/ObservabilityLinksPanel.module.css` — new
- `src/app/hooks/settings/useSettings.ts` — add four optional URL fields to settings shape
- `src/widgets/settings/SettingsWidget.tsx` — add four URL inputs (or wherever settings form lives)
- `src/pages/service/ServiceInfoPage.tsx` — import and place panel in Overview tab

## Do NOT change

- `src/processes/api/` — no API calls in this widget
- Any backend files

## Notes

- Use function declarations only; one file = one component.
- Settings shape extension must be backwards-compatible: missing keys default to `""`.
- The widget file goes in `src/widgets/service/` (create the directory if absent).
- Coding rules from ROADMAP.md apply.