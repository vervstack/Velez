# Incomplete Features

Features that exist in structure — proto definitions, interfaces, wiring — but don't actually do anything yet. Organized
by layer.

---

## Proto Contracts

**`ListDeployments.Response` is empty** (`service_api.proto:121-123`)
No fields defined. The endpoint exists, the storage query exists, but there is no contract for what a deployment looks
like in a response. Cannot be implemented without a proto change.

**`CreateSmerd.Request` has two unimplemented fields** (`velez_api.proto:119-135`)

- `plain` config (line 119-120): marked `// TODO not implemented`
- `sidecar` and `resources` definitions: `// TODO` comments, no fields added yet — resources are needed for any
  meaningful scheduling

**`UpgradeSmerd.Response` is empty** (`velez_api.proto:202`)
The upgrade endpoint returns nothing. No way for callers to know what changed (new container ID, image digest, etc.).

---

## Service Layer

**`UpgradeDeploy()` returns nil** (`verv_services/deploy.go:60`)
The interface method exists; the implementation body is a single `return nil`. Upgrade deployments created via
`CreateDeploy` with an `Upgrade` spec silently do nothing.

**Config subscription is commented out** (`service_manager/services.go:54`)

```go
// TODO VERV-128
//go handleConfigurationSubscription(configService, sm)
```

`handleConfigurationSubscription` is fully implemented below (lines 76-99) — it watches for config patches from
matreshka and would restart affected containers. But it's commented out. Running containers never receive config updates
after launch.

The function itself also has an incomplete handler body at line 96 (`// TODO VERV-128`) — even when uncommented, the
restart/reconcile action is missing.

---

## Workers

**`syncRunningBatch()` is a stub** (`deploy_watcher.go:166-168`)

```go
func (d *deployWatcher) syncRunningBatch(ctx context.Context, active []domain.Deployment) error {
return nil
}
```

This is where running containers should be reconciled against their DB state (check health, detect crashes, trigger
restart). Currently does nothing.

**`deleteBatch()` is a stub** (`deploy_watcher.go:170-172`)

```go
func (d *deployWatcher) deleteBatch(ctx context.Context, deletion []domain.Deployment) error {
return nil
}
```

Containers scheduled for deletion are never actually removed. `SCHEDULED_DELETION` status is set, the deploy watcher
picks them up, and nothing happens.

---

## Transport / API

**`VervAppService` response is name+id only** (`service_api.proto:84-87`)
`GetService` returns a `VervAppService` with just `id` and `name` — no current deployment, no image, no status. Not
useful for anything beyond an existence check.

**`UpgradeSmerd` pipeline works but has no self-upgrade guard**
The `UpgradeSmerd` API implementation in `velez_api_impl` calls the upgrade pipeline, but there's no check to prevent
upgrading the velez container itself, which would kill the process mid-upgrade.

---

## Networking / VCN

**`SetupVcn` silently disables itself on any failure** (`verv_closed_network/server.go:40-65`)
Three paths:

1. No server URL → returns `DisabledVcnImpl{}` silently
2. Has key+URL but connection fails → logs error, returns `DisabledVcnImpl{}` silently
3. No key → tries to launch headscale locally

In cases 1 and 2, the caller (`cluster.Setup`) gets a disabled VCN implementation with no indication of whether this was
intentional (not configured) or a failure (configured but unreachable). Error handling is indistinguishable.

**`DeleteNamespace` not implemented** (`verv_closed_network.proto:42`)  
Commented as `// Not implemented` in proto. Namespace cleanup doesn't exist.

**`ConnectService.domain_name` not implemented** (`verv_closed_network.proto`)  
Field is defined but marked as not implemented — service mesh DNS won't work.

---

## Storage

**`deployments` table missing operational fields**

Current schema (`20251221075707_init_service_table.sql`) has no:

- `replica_count` / `desired_count` — no way to express "run 3 instances"
- `rollback_target` — no pointer to the previous good spec for rollback
- `owner_node` — no affinity tracking for multi-node

**No deployment event log**
Status transitions (`SCHEDULED → RUNNING → FAILED`) are overwritten in place. There's no history. When a deployment
fails, you can't see what the error was after the fact.

**`VervAppService` has no deployment link in storage**
The services table exists, deployments table exists, but `GetService` doesn't join them. A service record is essentially
just a name row.

---

## App Wiring

**`Custom.Stop()` does nothing** (`app/custom.go:98-101`)

```go
func (c *Custom) Stop() error {
return nil
}
```

Graceful shutdown of services, open DB connections, and in-flight pipeline state is left entirely to OS cleanup.

**`autoupgrade.Start()` blocks the errgroup** (`app/custom.go:82-88`)
`autoupgrade.New(...).Start()` is run inside an `errgroup.Go()` in `Custom.Start()`. If autoupgrade returns an error,
`errG.Wait()` propagates it and the app never finishes starting. If it hangs (e.g., Docker socket unavailable), the app
hangs at startup. There's no timeout or circuit breaker.
