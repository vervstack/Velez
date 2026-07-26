# Velez Roadmap

Velez is a lightweight node manager for the Vervstack ecosystem — see [`README.md`](../README.md) for what it does
day to day. This file is the single source of truth for what's **mid-implementation or not-started**, across both
the backend jobs-engine migration and the PaaS (backend + UI) effort. It replaces `docs/paas/roadmap.md` and
`docs/paas/ui/ROADMAP.md` (folded in below, then deleted) and `docs/review/` (a pre-jobs-engine snapshot; mined for
anything still true, then deleted — see §4).

Anything not mentioned here as open should be assumed done. Status claims below are backed by file:line references
verified against current code, not carried forward from older docs at face value.

---

## 1. Backend jobs engine

**Status: functionally complete.** The original 7-pipeline `Pipeliner` migration is done, and both items that were
previously tracked as "what's left" below — `DropSmerd` and the `CreateSmerdStream` pilot — were completed
2026-07-26. See [`docs/jobs_migration.md`](jobs_migration.md) for the full per-pipeline history, methodology, and
the checklist to follow for any future pipeline migration. Summary: `LaunchSmerd`, `CreateService`,
`AssembleConfig`, `ConnectServiceToVpn`, `EnableStatefullMode`, `UpgradeSmerd`, and now `DropSmerd` are all
scaffolded as jobs *and* cut over to serve their live gRPC endpoints. Only `CopyToVolume` remains
scaffolded-but-not-cut-over, and it has no live caller to cut over to in the first place.

### What's left

- **`CopyToVolume`** (`internal/jobs/copy_to_volume.go`) stays scaffolded, no live caller — parked per
  `docs/jobs_migrations/questions.md` #1 unless a `CopyToVolume` RPC is ever added.

### Recently completed (2026-07-26)

- **`DropSmerd`** is now scaffolded (`internal/jobs/drop_smerd.go`) and cut over
  (`internal/transport/velez_api_impl/smerd_drop.go`). One `drop_container_<n>` job per identifier in
  `request.uuids ++ request.name`; each job swallows its own `Docker.Remove` error into the task context's
  `Failed`/`Successful` lists rather than failing the task, preserving the old `DropSmerds`' contract that the RPC's
  top-level error is always nil. `container_manager.DropSmerds` (the old logic) is not removed — `custom.go`'s
  `smerdsDropper` shutdown hook still calls it directly, bypassing the RPC layer, same carve-out shape as
  `ConnectServiceToVpn`/`UpgradeSmerd`. First-ever test coverage for this RPC, via the previously-unused
  `env.DropSmerd()` e2e helper (`Test_Lifecycle/Test_DropSmerd_ByUuid`). See `docs/jobs_migration.md`'s status
  table for details.
- **`CreateSmerdStream`** now exists: an additive server-streaming RPC on `TasksApi` (not `VelezAPI` — avoids a
  circular proto import, since `tasks.proto` already imports `velez_api.proto`) that enqueues the same
  `create_smerd` task as the untouched unary `CreateSmerd` and forwards `TaskStatus` updates via the same
  `Engine.Watch` mechanism `WatchTask` already uses. Same `(name, create_smerd)` entity/action key as unary
  `CreateSmerd`, so a client on either path attaches to the same in-flight task. TS client regenerated for the
  first time (`pkg/web/Velez-UI/src/app/api/velez/tasks.pb.ts` didn't exist before this — neither did one for the
  pre-existing `WatchTask`), and wired into `DeployWidget.tsx`'s smerd-creation flow to show live status instead of
  a static spinner.

> `docs/plans/testing.md`'s own "Backlog" status table (its lines ~140-149) is stale — it still shows most pipelines
> as "Not started" even though they were subsequently cut over. Trust `docs/jobs_migration.md`'s status table (kept
> in sync at each cutover) over that table.

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
| [T2 — Deployments](paas/milestones/m1-core-platform/t2-deployments.md)               | Mostly done | No per-step progress display during a deploy; "Upgrade" tab in `DeployMenu` explicitly not implemented                                                                                                                                                                                                   |
| [T3 — Verv Services](paas/milestones/m1-core-platform/t3-verv-services.md)           | Partial     | List columns (source repo/image/registered-at) not added; new-service field verification incomplete; no navigate-to-list-on-delete-success; edit-in-place not started; create round-trip not e2e tested                                                                                                  |
| [T4 — Settings](paas/milestones/m1-core-platform/t4-settings.md)                     | Mostly done | Connection health indicator (status dot + health-check call + tooltip) not built                                                                                                                                                                                                                         |
| [T5 — UX Polish](paas/milestones/m1-core-platform/t5-ux-polish.md)                   | Partial     | No skeleton loaders (plain "Loading..." text only); no retry button on failed queries; no empty states for deployments/smerds lists; no breadcrumb trail; no active-route sidebar highlight noted here (though M9-T36 elsewhere fixed one active-nav bug); responsive layout down to 1280px not verified |

**Backend companion tasks** (`docs/paas/backend/tasks/`):

| Task                                                                                 | Status          | Evidence                                                                                                                                                                                                                                                                                                |
|--------------------------------------------------------------------------------------|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [B1-T01 — API enrichment](paas/backend/tasks/B1-T01-api-enrichment.md)               | Done            | `ListNodes` (`api/grpc/control_plane_api.proto:30`), `ListPeers` (`api/grpc/verv_closed_network.proto:50`), `ServiceBaseInfo` enrichment all present. One caveat noted in the task file itself: `make lint` was blocked by a pre-existing golangci-lint config version mismatch, unrelated to this task |
| [B2-T01 — Service runtime stats](paas/backend/tasks/B2-T01-service-runtime-stats.md) | **Not started** | `GetServiceStats` RPC absent from `api/grpc/service_api.proto` (grepped, no match). Blocks the M9-T33 Overview tab's stats display, which currently has no runtime-stats section                                                                                                                        |
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
| [T35 — Environment tab switcher + tags strip](paas/ui/tasks/M9-T35-service-env-tags.md) | **Partial — not done** | The 5-tab strip exists, but only `overview` renders real content; `metrics`/`instances`/`history`/`access` all render a `"{activeTab} — coming soon"` placeholder (`ServiceInfoPage.tsx:114-118`). No tags strip found in the file |
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
