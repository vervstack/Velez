---
id: "M9-T35"
title: "ServiceInfoPage — environment tab switcher + tags strip"
status: "pending"
model: "stable-code"
created: "2026-05-04"
branch: "task/M9-T35-service-env-tags"
---

# Task M9-T35 — Environment Tab Switcher and Tags Strip

## Goal

Add an environment tab switcher and a tags strip to `ServiceInfoPage` so operators can see which environment a service
is running in and filter the deployment history by environment.

## Context

`ServiceInfoPage` (rebuilt in T33) fetches smerds via `FetchSmerdsByServiceName`. Smerds carry a Docker labels map (
`smerd.labels`) that contains metadata set at deploy time. Two conventions used in Velez deployments:

- `verv.env` label — one of `dev`, `staging`, `prod` (or absent)
- `verv.tag.*` labels — arbitrary tags, e.g. `verv.tag.team=backend`, `verv.tag.critical=true`

The environment tabs should be derived from the set of distinct `verv.env` values found across all fetched smerds. If
only one environment (or none) exists, tabs are hidden and not rendered.

Tags are read from `verv.tag.*` labels of the **currently active smerd** (the one whose ID matches
`service.currentDeploymentId` or the first running smerd). Each tag key-value pair renders as a `<Chip>` (
`src/components/base/chips/`).

## Acceptance Criteria

- [ ] Environment tabs appear below the header when more than one distinct `verv.env` value is detected across smerds;
  selecting a tab filters the Deployments tab to show only deployments from that env
- [ ] When `verv.env` is absent or uniform, the tab row is not rendered (not an empty strip)
- [ ] Tags strip renders below the env tabs showing `key: value` chips from `verv.tag.*` labels of the active smerd;
  absent if no tags
- [ ] Chips use the existing `EnvChip` or `Badge` component from `src/components/base/`
- [ ] Active env tab is visually distinct (accent underline using `--color-accent`)
- [ ] `yarn build:ui` passes with no TypeScript errors

## Files to Create / Modify

- `src/pages/service/ServiceInfoPage.tsx` — add env tab state and tags strip rendering
- `src/pages/service/ServiceInfoPage.module.css` — add tab and tags styles
- `src/processes/mappings/smerds.ts` — add helper to extract env and tags from smerd labels

## Do NOT change

- `src/processes/api/velez.ts` — no new API calls; uses existing smerd fetch
- Any backend files

## Notes

- Docker label key access: `smerd.labels?.["verv.env"]` (labels is `Record<string, string>`)
- Tags extraction: filter `Object.entries(smerd.labels ?? {})` for keys starting with `"verv.tag."`, strip prefix for
  display key.
- Coding rules from ROADMAP.md apply (function declarations, CSS Modules, rem units).
- T33 must be done first — this task slots into the layout it establishes.