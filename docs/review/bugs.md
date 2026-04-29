# Confirmed Bugs

These are verified issues in the current code — not design opinions, not speculative. Each has a specific file:line
reference.

---

## Critical — Wrong Behavior

### 1. Deploy watcher never tracks active deployments (`deploy_watcher.go:111`)

```go
// line 111
list.active = append(list.scheduled, dep) // BUG: appends to list.scheduled, not list.active
```

Every `RUNNING` deployment gets appended to `list.scheduled` instead of `list.active`. Result:

- `list.active` passed to `syncRunningBatch()` is always empty — health reconciliation never sees any running container.
- `list.scheduled` accumulates running deployments alongside genuinely scheduled ones, so `processScheduledBatch` tries
  to re-deploy already-running containers on every tick.

**Fix:** `list.active = append(list.active, dep)`

---

### 2. Healthcheck silently passes when container never starts (`healthcheck.go:52-68`)

```go
go func () {
defer close(errC)
for i := uint32(0); i < h.req.Healthcheck.Retries; i++ {
// ... inspects container ...
if cont.State.Status == dockerContainerStatusRunning {
errC <- nil
return
}
}
// loop exits without sending — channel just gets closed
}()

err := <-errC // receives nil from closed channel = "healthy"
```

If the container never reaches `running` state (crash loop, bad image, etc.), the goroutine exhausts retries and exits.
The deferred `close(errC)` causes `<-errC` to return `nil` — i.e., healthcheck passes. A container that failed to start
is reported as healthy.

**Fix:** Send an explicit error when the loop exhausts retries without success.

---

### 3. Nil pointer dereference in healthcheck when `containerId` is nil (`healthcheck.go:44`)

```go
if h.containerId == nil && *h.containerId == "" {
```

`&&` does not short-circuit on `true` — both sides are evaluated when the left is true. If `containerId` is nil,
`*h.containerId` panics. The intent is clearly `||`.

**Fix:** `if h.containerId == nil || *h.containerId == ""`

---

### 4. `SCHEDULED_UPGRADE` case is empty — upgrades silently do nothing (`deploy_watcher.go:159`)

```go
case deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE:
// empty
```

`listDeployments` puts both `SCHEDULEDDEPLOYMENT` and `SCHEDULEDUPGRADE` into `list.scheduled` (lines 107-109).
`processScheduledBatch` handles `SCHEDULEDDEPLOYMENT` but the `SCHEDULEDUPGRADE` branch is an empty case. Upgrade
deployments complete the tick with no action taken — status never updates, containers never change.

**Fix:** Implement the upgrade case (see `roadmap.md` Milestone 1).

---

### 5. `UpgradeDeploy()` is a stub that returns success (`verv_services/deploy.go:60`)

```go
func (v *VervService) UpgradeDeploy(ctx context.Context, request domain.UpgradeDeployReq) error {
return nil
}
```

Any call path that triggers an upgrade (e.g., `CreateDeploy` with an `Upgrade` spec) silently succeeds without doing
anything. No containers change, no error is returned, no status is recorded.

**Fix:** Wire the existing `UpgradeSmerd` pipeline.

---

## Data Integrity

### 6. Rollback uses `context.Background()` ignoring deadline (`runner.go:27`)

```go
rollbackErr := p.rollback(context.Background())
```

When a pipeline fails, rollback runs on a fresh context with no deadline. If the parent context was cancelled (timeout,
client disconnect), rollback can run indefinitely. In practice this means a timed-out deploy can leave its rollback
running in the background, deleting/restoring containers after the caller has already moved on.

**Fix:** Pass a separate context with an explicit rollback timeout rather than Background().

---

### 7. `wrapPgErr` is a no-op (`postgres/storage.go:53-55`)

```go
func wrapPgErr(err error) error {
return err
}
```

The function exists, is never called, and adds no value. Postgres errors propagate unwrapped — driver-specific error
codes (unique violation, foreign key, not found) are indistinguishable to callers.

**Fix:** At minimum, map `pq.Error` codes to sentinel errors. Call `wrapPgErr` at query sites.

---

## Wiring / Lifecycle

### 8. Worker `Stop()` is a no-op (`deploy_watcher.go:80-82`)

```go
func (d *deployWatcher) Stop() error {
return nil
}
```

`Stop()` is registered with `closer.Add` (custom.go:74), but it does nothing. The ticker keeps firing after shutdown
begins. The goroutine inside `starter.Do` has no ctx.Done() check, so it runs until the process dies. On a slow shutdown
with in-flight deployments, this can cause partial operations after the API layer is gone.

**Fix:** Stop the ticker and signal the goroutine to exit via a channel or context.

---

### 9. `ListDeployments.Response` is structurally empty (proto)

```proto
message ListDeployments {
  message Response {
    // empty
  }
}
```

The response message in `service_api.proto:121-123` has no fields. Any implementation of `ListDeployments` can only
return an empty response — there is no type-safe way to add deployment data without a proto change first.

**Fix:** Add fields (deployment list, total count) before implementing the handler.
