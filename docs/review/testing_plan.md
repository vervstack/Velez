# E2E Testing Plan — Milestone 1

This document covers the test scenarios for Milestone 1 ("Unfuck Current Setup").
Tests use real Docker, optionally real PostgreSQL, and a real in-process Velez server
via `bufconn`. No Matreshka, Makosh, Headscale, or other Vervstack services.

---

## Bugs to fix before tests can run

These are blockers discovered during codebase review. Tests cannot be written or run until
these are resolved.

| # | Location                                            | Issue                                                                                                                                                                              |
|---|-----------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1 | `tests/helper_environment.go:38`                    | `local_state` package not imported — tests don't compile                                                                                                                           |
| 2 | `api/grpc/service_api.proto`                        | `ListDeployments.Response` is empty — proto needs `DeploymentInfo` message, `repeated DeploymentInfo deployments` and `uint64 total` fields                                        |
| 3 | `api/grpc/service_api.proto`                        | `VervAppService` missing `current_deployment_id` and `status` fields                                                                                                               |
| 4 | `internal/transport/service_api_impl/`              | `ListDeployments` handler not implemented (falls to `Unimplemented`)                                                                                                               |
| 5 | `internal/storage/postgres/queries/deployments.sql` | `deployment_specifications.name` has a `UNIQUE` constraint; current code generates `"service_id=X"` as the name, causing duplicate-key error on second deploy for the same service |
| 6 | `internal/app/custom.go` `Custom.Stop()`            | Returns `nil` immediately — no graceful shutdown; in-flight pipelines are abandoned                                                                                                |
| 7 | `internal/cluster/autoupgrade/autoupgrade.go`       | `<-au.stopC` in the select falls through to the next loop iteration instead of returning — goroutine never actually stops                                                          |
| 8 | `internal/app/custom.go` `Custom.Start()`           | `autoupgrade` is created fresh inside `Start()` and not reachable via `Stop()` — context cancellation cannot reach it                                                              |

---

## Environment modes

### Mode A — Docker only (no PG)

- `NewEnvironment(t)` with no options
- State manager uses `noImpl` — deployment scheduling is silently disabled (`ErrServiceIsDisabled`)
- Tests cover direct container operations only

### Mode B — Docker + PostgreSQL

- `NewEnvironment(t, WithPostgres())` — new `TestEnvOpt` that starts a `postgres:16` container (or connects to a local
  PG), runs migrations, and wires in `pgState`
- Tests cover the full service/deployment lifecycle via the `ServiceApi`
- Each test gets an isolated PG schema or a per-test database to avoid cross-contamination

---

## Suite 1 — Container lifecycle (Mode A)

Basic smerd (container) CRUD. These replace and extend the existing `Test_Deploy_Generic_Api`
and `Test_Deploy_Verv_Api` tests.

### 1.1 Create and list a simple container

- Call `CreateSmerd` with `hello_world` image, `ignore_config: true`
- Assert response: name matches, `status = running`, UUID non-empty, `created_at` set
- Call `ListSmerds` by name
- Assert exactly one result returned, fields match `CreateSmerd` response

### 1.2 Create container with healthcheck

- Call `CreateSmerd` with postgres image, custom healthcheck (`pg_isready -U postgres`, interval=2s, retries=3)
- Assert container reaches `running` status within timeout
- Healthcheck command should be visible in the returned `Smerd`

### 1.3 Drop a running container

- Create a container
- Call `DropSmerd` with UUID
- Assert `successful` list contains UUID, `failed` is empty
- Call `ListSmerds` by UUID, assert empty

### 1.4 Declarative deploy idempotency

- Create same container twice with `is_declarative_deploy: true`
- Assert both calls return success (second call sees healthy container and skips)
- Assert only one container exists in `ListSmerds`

### 1.5 Duplicate name rejected without declarative flag

- Create container with name "foo"
- Create again with same name, `is_declarative_deploy: false` (default)
- Assert second call returns an error

### 1.6 Upgrade container image

- Create a container with `hello_world:v0.0.14`
- Call `UpgradeSmerd` with a different tag (e.g. `hello_world:v0.0.13`)
- Assert old container is gone, new container with new image tag is running under the same name
- Assert `ListSmerds` shows one container with the new image

### 1.7 Self-upgrade guard

- Set up so that the Velez container itself is running (simulate by labeling a test container
  with the velez self-identification label or by name pattern)
- Call `UpgradeSmerd` targeting that container
- Assert the call returns an error (refused) rather than proceeding

---

## Suite 2 — Volume scenarios (Mode A)

### 2.1 Named volume persists across container recreation

- Create a postgres container with a named volume mounted at `/var/lib/postgresql/data`
- Write data: exec a SQL `CREATE TABLE` and `INSERT`
- Drop the container (without removing the Docker volume)
- Re-create a new postgres container with the same named volume
- Assert the data is still accessible (run `SELECT` against it)

### 2.2 Volume present in ListSmerds response

- Create a container with an explicit volume binding
- Call `ListSmerds`, inspect the returned `Smerd.volumes`
- Assert `volume_name` and `container_path` match what was requested

### 2.3 Transfer volume to a different container

- Create container A with a named volume at `/data`
- Write a file into the volume (via `CopyToVolume` or exec)
- Drop container A
- Create container B (different image, same volume name at same path)
- Assert the file written by container A is visible inside container B

### 2.4 CopyToVolume API

- Create a named volume
- Call `CopyToVolume` (if exposed via pipeline) with a map of `path → bytes`
- Create a container mounting that volume
- Assert the files are visible inside the container at the expected paths

---

## Suite 3 — Service and deployment lifecycle (Mode B — requires PG)

This suite exercises `ServiceApi`: `CreateService`, `CreateDeploy`, `GetService`,
`ListDeployments`, `ListServices`.

### 3.1 Create a service and verify it is listed

- Call `CreateService` with name "my-service"
- Call `ListServices`, assert "my-service" appears in response, `total >= 1`
- Call `GetService` by name, assert `id` and `name` are correct
- Assert `GetService` response includes `current_deployment_id` (nil at this point) and `status`

### 3.2 Deploy a service — new specification

- Create a service
- Call `CreateDeploy` with a `new` spec (hello_world image, ignore_config: true)
- Poll `ListDeployments` until status = `RUNNING` (or up to 30s timeout)
- Assert `ListDeployments.Response.total = 1`, `deployments[0]` has `id`, `status`, `spec_id`, `created_at`
- Assert the container is actually running in Docker (`ListSmerds` by name)
- Call `GetService`, assert `current_deployment_id` matches the running deployment's id

### 3.3 Deploy a service — failed deployment is recorded

- Create a service
- Call `CreateDeploy` with a non-existent image (`no-such-image:v0.0.1`)
- Poll `ListDeployments` until status = `FAILED` (or up to 30s timeout)
- Assert deployment status is `FAILED` in response
- Assert no container was left running

### 3.4 Upgrade deployment — new image

- Create a service and deploy it with `hello_world:v0.0.14`
- Wait for status `RUNNING`
- Call `CreateDeploy` with an `upgrade` spec pointing to the running deployment, new image = `hello_world:v0.0.13`
- Poll `ListDeployments` until the new deployment reaches `RUNNING`
- Assert: old deployment moves to `DELETED`, new deployment is `RUNNING`
- Assert: only one container with this service name is running in Docker
- Assert: container is using the new image tag
- Call `GetService`, assert `current_deployment_id` reflects the new deployment

### 3.5 Multiple services do not interfere

- Create two services "svc-alpha" and "svc-beta"
- Deploy both concurrently
- Assert `ListDeployments?service_id=<alpha>` returns only alpha's deployments
- Assert `ListDeployments?service_id=<beta>` returns only beta's deployments

### 3.6 ListDeployments pagination

- Create a service
- Deploy it 5 times sequentially (each fails fast — use bad images — to accumulate records quickly)
- Assert `ListDeployments` with `limit=2, offset=0` returns 2 records and `total=5`
- Assert `ListDeployments` with `limit=2, offset=4` returns 1 record

---

## Suite 4 — Wiring and shutdown (Mode A)

### 4.1 Graceful shutdown — in-flight containers cleaned up

- Start env
- Begin creating a large container (slow image pull)
- Call `env.Custom.Stop()` while pull is in progress
- Assert `Stop()` returns without panic within a reasonable deadline
- Assert no zombie containers remain after shutdown

### 4.2 Autoupgrade — goroutine stops on context cancel

- Start env with a short autoupgrade interval (inject a mock or short-circuit)
- Cancel the root context
- Assert autoupgrade goroutine exits within 2× check period (no goroutine leak)

### 4.3 DeployWatcher — stops cleanly

- Start env
- Stop deploy watcher via `worker.Stop()`
- Assert subsequent ticks do not process anything (second `Stop()` is idempotent)

---

## Suite 5 — Multi-node with PG (Mode B, single-machine simulation)

Simulate two Velez nodes against the same PG instance. Node IDs are distinct; each node
only picks up deployments assigned to it.

### 5.1 Deployment assigned to node 1 does not run on node 2

- Start two `TestEnvironment` instances sharing the same PG DSN but different node IDs
- Create a deployment scheduled to node 1
- Assert only node 1's deploy watcher processes it; node 2's watcher ignores it

### 5.2 Node heartbeat updates `last_online`

- Start env in Mode B
- Wait 2× heartbeat interval
- Query PG `velez.nodes` directly
- Assert `last_online` was updated within the last heartbeat interval

---

## Test infrastructure requirements

### Docker helper utilities needed

- `ExecInContainer(ctx, contID, cmd) (stdout string, err error)` — run a command inside a container
- `VolumeExists(ctx, name) bool` — check a Docker volume exists
- `CreateVolume(ctx, name) error` — create a named volume without a container
- `RemoveVolume(ctx, name) error` — delete a Docker volume in cleanup

### PG helper needed

- `WithPostgres() TestEnvOpt` — spin up postgres container (or reuse local), run goose migrations,
  inject DSN as `ClusterStateManager`
- Each test gets an isolated `search_path` or schema to avoid cross-test state

### Polling helper needed

```go
// WaitFor polls cond every interval until it returns true or timeout elapses.
func WaitFor(t *testing.T, timeout, interval time.Duration, cond func () bool)
```

### Assertions to add to `helper.go`

- `AssertDeploymentStatus(t, env, deploymentId, expected)` — polls and asserts a deployment reaches a target status
- `AssertContainerRunning(t, env, name)` — asserts Docker container is in running state
- `AssertContainerGone(t, env, name)` — asserts no container with given name exists

---

## Files to create

| File                             | Purpose                                                                              |
|----------------------------------|--------------------------------------------------------------------------------------|
| `tests/suite_containers_test.go` | Suite 1 — container lifecycle                                                        |
| `tests/suite_volumes_test.go`    | Suite 2 — volume scenarios                                                           |
| `tests/suite_services_test.go`   | Suite 3 — service + deployment lifecycle (PG)                                        |
| `tests/suite_wiring_test.go`     | Suite 4 — wiring and shutdown                                                        |
| `tests/suite_multi_node_test.go` | Suite 5 — multi-node simulation                                                      |
| `tests/helper_pg.go`             | `WithPostgres()` opt, migrations, teardown                                           |
| `tests/helper_docker.go`         | `ExecInContainer`, `VolumeExists`, `CreateVolume`, `RemoveVolume`                    |
| `tests/helper_wait.go`           | `WaitFor`, `AssertDeploymentStatus`, `AssertContainerRunning`, `AssertContainerGone` |

---

## Order of implementation

1. Fix compile blocker (missing import in `helper_environment.go`)
2. Fix proto + codegen for `ListDeployments.Response` and `VervAppService`
3. Implement `ListDeployments` handler
4. Fix `deployment_specifications.name` uniqueness bug
5. Implement `Custom.Stop()` graceful shutdown
6. Fix `autoupgrade` context cancellation
7. Add self-upgrade guard
8. Write test infrastructure (`helper_pg.go`, `helper_docker.go`, `helper_wait.go`)
9. Write test suites 1–5 in order