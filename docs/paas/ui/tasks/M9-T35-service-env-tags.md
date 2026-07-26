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

**Superseded, 2026-07-26**: `ServiceInfoPage` already had a working `EnvSwitcher` widget (real `ListEnvironments`
RPC, but hardcoded per-env `status`/`health`) that this task's literal spec — a second, `verv.env`-label-derived
tab row — would have duplicated. Extended `EnvSwitcher` instead of building a parallel env-tabs concept:

- [x] ~~Environment tabs appear below the header when more than one distinct `verv.env` value is detected across
  smerds~~ — `EnvSwitcher` now wired to the real `GetServiceEnvironments` RPC (was already generated with a dead
  `useListServiceEnvsQuery` hook, zero call sites before this) instead of hardcoded per-env fields
- [ ] ~~selecting a tab filters the Deployments tab to show only deployments from that env~~ — dropped:
  `DeploymentInfo` carries no env field, so there's nothing to filter by without a separate proto change (not in
  scope this round)
- [x] Tags strip renders below `EnvSwitcher` showing `key: value` chips from `verv.tag.*` labels of the active
  smerd (`ServiceTagsStrip`, new `TagChip` component); absent if no tags
- [x] Chips use the existing chip pattern (`src/components/base/chips/`, new `TagChip.tsx` alongside `EnvChip.tsx`)
- [x] `bun run build` passes with no new TypeScript errors (one pre-existing, unrelated error in
  `HeadscalePluginForm.tsx` predates this change)

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