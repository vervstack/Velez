# Velez Roadmap

Velez is a lightweight node manager for the Vervstack ecosystem — see [`README.md`](../README.md) for what it does
day to day. This file is the single source of truth for what's **mid-implementation or not-started**, across both
the backend jobs-engine migration and the PaaS (backend + UI) effort. It replaces `docs/paas/roadmap.md` and
`docs/paas/ui/ROADMAP.md` (folded in below, then deleted) and `docs/review/` (a pre-jobs-engine snapshot; mined for
anything still true, then deleted — see §4).

Anything not mentioned here as open should be assumed done. Status claims below are backed by file:line references
verified against current code, not carried forward from older docs at face value.

---

## 2. PaaS platform

| #  | Milestone                                                        | Status                | Detail                                                                                        |
|----|------------------------------------------------------------------|-----------------------|-----------------------------------------------------------------------------------------------|
| M1 | [Core Platform](paas/milestones/m1-core-platform/overview.md)    | **In Progress**       | See breakdown below                                                                           |
| M2 | [Cluster & Networking](paas/milestones/m2-cluster-networking.md) | Backlog — not started | Control plane page, VCN page, multi-node targeting, node scheduling tags all still scope-only |
| M3 | [Observability](paas/milestones/m3-observability.md)             | Backlog — not started | Log streaming, metrics charts, event timeline, alerting, export — all scope-only              |
| M4 | [Access & Multi-tenancy](paas/milestones/m4-access-control.md)   | Backlog — not started | Auth, RBAC, namespaces, audit log, API tokens — all scope-only                                |
| M5 | [PaaS Automation](paas/milestones/m5-paas-automation.md)         | Backlog — not started | Auto-rollback, health-gated promotion, scaling policies, webhooks, secrets — all scope-only   |

M2-M5 are still draft scope documents with no completed work; nothing to verify against code yet.

### M1 — Core Platform, task-by-task

Verified against each task file's own checkboxes (`docs/paas/milestones/m1-core-platform/t*.md`) and spot-checked
against `pkg/web/Velez-UI`:

| Task                                                                                 | Status      | Open items                                                                                                                                                                                                                                                                                               |
|--------------------------------------------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [T1 — Services Dashboard](paas/milestones/m1-core-platform/t1-services-dashboard.md) | Done        | All items checked                                                                                                                                                                                                                                                                                        |
| [T2 — Deployments](paas/milestones/m1-core-platform/t2-deployments.md)               | Done (caveat) | 2026-07-26: Deploy button un-gated (was hardcoded disabled), "Upgrade" tab fully implemented, added `DeploymentStatusBadge` + short-lived post-submit polling. Progress display is still a coarse 7-value status, not true per-pipeline-step — no step-level data exists anywhere in `domain.Deployment` to back that, so this is a deliberate scope cut, not an oversight |
| [T3 — Verv Services](paas/milestones/m1-core-platform/t3-verv-services.md)           | Mostly done | 2026-07-26: added `repo` to `ServiceBaseInfo` (list columns done), real status badge on HomePage cards (was hardcoded "unknown"), per-card Stop/Restart/Remove actions. Built full-stack `RemoveService` (proto+backend+frontend) — the task file's "Remove action"/"confirmation dialog" checkboxes were previously marked done but nothing existed at any layer; now genuinely implemented, rejects deletion of a service with an attached smerd unless a "drop running instances" flag is set. New-service field verification resolved (verified `CreateService.Request` only defines `name` by design — image/tag/env/volumes/ports are `CreateDeploy`-time concerns). Remaining open: 3.4 edit-in-place (stretch goal, out of scope), create round-trip not e2e tested against a live backend |
| [T4 — Settings](paas/milestones/m1-core-platform/t4-settings.md)                     | Done | 2026-07-26: connection health indicator built — `StatusDot` in `ApiSettings.tsx` (via new `useConnectionHealth` hook, probes on mount + after save using the existing `ListServices` call) with a tooltip showing last-checked time and error                                                                                                                                                         |
| [T5 — UX Polish](paas/milestones/m1-core-platform/t5-ux-polish.md)                   | Mostly done | 2026-07-26: fixed a real bug where empty deployments rendered fabricated mock rows instead of an empty state; added skeleton loaders (ServiceInfoPage, SmerdPage — DeployPage/NewServicePage/VervClosedNetworkPage have no async load, N/A); added a header-wide loading indicator via `useIsFetching`; added inline retry-on-error for failed queries (HomePage, ServiceInfoPage, SmerdPage); added SmerdPage breadcrumb; added responsive hamburger + sidebar overlay below 1280px. Active-nav highlight was already done (`MainLayout.tsx`'s `ROUTE_TO_NAV`), just mis-tracked. Remaining open: no per-page loading indicator distinct from the global header one (arguably covered by the header indicator + skeletons) |

**Backend companion tasks** (`docs/paas/backend/tasks/`):

| Task                                                                                 | Status          | Evidence                                                                                                                                                                                                                                                                                                |
|--------------------------------------------------------------------------------------|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [B1-T01 — API enrichment](paas/backend/tasks/B1-T01-api-enrichment.md)               | Done            | `ListNodes` (`api/grpc/control_plane_api.proto:30`), `ListPeers` (`api/grpc/verv_closed_network.proto:50`), `ServiceBaseInfo` enrichment all present. One caveat noted in the task file itself: `make lint` was blocked by a pre-existing golangci-lint config version mismatch, unrelated to this task |
| [B2-T01 — Service runtime stats](paas/backend/tasks/B2-T01-service-runtime-stats.md) | Done | Shipped already, as `GetServiceMetrics` rather than `GetServiceStats` (a literal-string grep for the latter is why this was previously marked not started). Wired end-to-end: `internal/service/service_manager/verv_services/metrics.go`, `internal/transport/service_api_impl/metrics.go`, consumed by `ServiceHero.tsx` |
| [B3-T01 — Plugin service](paas/backend/tasks/B3-T01-plugin-service.md)               | Done            | `ListPlugins` RPC and `Plugin` message present (`api/grpc/control_plane_api.proto:37,148`)                                                                                                                                                                                                              |

---

## 3. UI redesign

**Component library (old `docs/paas/ui/ROADMAP.md` M1-M8, M10): Done, verified against code.** Spot-checked
components exist under `pkg/web/Velez-UI/src/{components,widgets}`: `PluginMatrix.tsx`, `NodeCard.tsx`,
`Sidebar.tsx`, `TopBar.tsx`, `StatCard`, `PluginManageDialog.tsx`, `StatefullPgPluginForm.tsx`,
`SimplePluginForm.tsx`, etc. — matching the task files' target paths.

**M9 — Service Page: Partial**, not fully done despite most of its tasks being struck through in the old roadmap:

| Task                                                                                    | Status                 | Evidence                                                                                                                                                                                                                           |
|-----------------------------------------------------------------------------------------|------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [T33 — ServiceInfoPage redesign](paas/ui/tasks/M9-T33-service-page-redesign.md)         | Done                   | `pkg/web/Velez-UI/src/pages/service/ServiceInfoPage.tsx` + `parts/Header.tsx` present; 5-tab strip implemented (`ServiceInfoPage.tsx:23-30`)                                                                                       |
| [T34 — ObservabilityLinksPanel widget](paas/ui/tasks/M9-T34-observability-links.md)     | Done                   | Implemented as `ObservabilityTools` (`pkg/web/Velez-UI/src/widgets/service/ObservabilityTools/ObservabilityTools.tsx`), wired into the Overview tab (`ServiceInfoPage.tsx:102`)                                                    |
| [T35 — Environment tab switcher + tags strip](paas/ui/tasks/M9-T35-service-env-tags.md) | Superseded, done differently | 2026-07-26: rather than a second `verv.env`-label-derived tab row, extended the pre-existing `EnvSwitcher` widget with real per-env data (`GetServiceEnvironments` RPC, previously generated but with zero call sites) and added a tags strip (`ServiceTagsStrip`/`TagChip`) from `verv.tag.*` smerd labels. "Filter deployments by env" dropped — `DeploymentInfo` has no env field to filter by. The 5-tab strip's `metrics`/`instances`/`history`/`access` placeholders are untouched, still "coming soon" (`ServiceInfoPage.tsx`) |
| T36-T38                                                                                 | Done                   | Struck through in the old roadmap and consistent with shipped code (sidebar active-nav fix, settings page, plugin matrix status + manage dialog)                                                                                   |

---

## 4. Genuinely open items carried forward from `docs/review/`

`docs/review/` (deleted as part of this cleanup) was a snapshot from a pre-jobs-engine code review describing Velez
as "early but functional." Every claim in `docs/review/bugs.md` (9 confirmed bugs) was re-verified against current
code and **all 9 are now fixed** — nothing from that file is carried forward. Most of `docs/review/incomplete.md` is
also fixed (`UpgradeDeploy` is fully implemented at `internal/service/service_manager/verv_services/deploy.go:68-143`;
`syncRunningBatch`/`deleteBatch` in `internal/workers/deploy_watcher.go` are no longer stubs; `ListDeployments.Response`
has real fields; `Custom.Stop()` performs real graceful shutdown; the `SCHEDULED_UPGRADE` case is implemented; a
self-upgrade guard exists at `internal/jobs/upgrade_smerd.go:107,239`). The items below are still genuinely true as of
2026-07-26 and have no other tracking home now that the file is gone:

- **Config subscription still commented out.** `internal/service/service_manager/services.go:75` —
  `//go handleConfigurationSubscription(configService, sm)` is still commented, `TODO VERV-128` still present there
  and in `internal/service/service_manager/configurator/configurator.go:27,37`. Running containers never receive
  live config updates after launch.
- **`UpgradeSmerd.Response` is still an empty message.** `api/grpc/velez_api.proto:201` — callers get no information
  about what changed (new container ID, image digest, etc.).
- **`CreateSmerd.Request` sidecar/resources fields still TODO.** `api/grpc/velez_api.proto:132-133`. (The `plain`
  config TODO from the same original review item has since been implemented — field 14, `PlainConfigSpec plain`.)
- **`SetupVcn` still silently disables itself on failure.** `internal/cluster/verv_closed_network/server.go:41` and
  `:53` both still return `DisabledVcnImpl{}` with no distinguishing signal between "not configured" and "configured
  but unreachable."
- **VCN features still unimplemented:** `DeleteNamespace` (`api/grpc/verv_closed_network.proto:43`) and
  `ConnectService.domain_name` (`:87`) — service-mesh DNS still doesn't work.
- **No deployment event log / history is still overwritten in place.** No migration in `migrations/*.sql` adds
  `replica_count`, `desired_count`, `rollback_target`, `owner_node`, or an append-only events table — status
  transitions still overwrite the current status field with no history.
- **Node ID still hardcoded to 1.** `internal/workers/deploy_watcher.go:50` — `nodeId: 1`. Blocks true multi-node
  operation; matches `docs/review/architecture.md`'s "Cluster State Is Node-Hardcoded" note and is directly relevant
  to M2 above.
- **Architecture notes relevant to future M2/podman/k8s work** (from `docs/review/architecture.md`, still valid,
  unactioned, no current owner): no `ContainerRuntime` abstraction (Docker client used directly everywhere, blocks a
  podman adapter); several proto types are Docker-shaped (`Port.exposed_to`, `RestartPolicyType`, `Hardware.cpu` as
  float) and will need a translation layer for k8s; no pod/sidecar-grouping concept; no matreshka circuit
  breaker/fallback, so a matreshka outage fails all new deploys even for containers that don't need dynamic config.
  These aren't scheduled against any milestone above — worth folding into M2's scope doc when that milestone is
  picked up.
