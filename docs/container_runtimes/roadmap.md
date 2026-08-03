# Roadmap

## Phase 1 (current pass): `ContainerCreate` only

Deliberately narrow scope, chosen because `ContainerCreate` is the one method that has to do real work in both
backends (label/name-conflict resolution for the shared daemon; nothing but routing for a dedicated one) — it
exercises the whole shape without requiring the full interface.

**In scope**
- `ContainerRuntime` interface — just the `ContainerCreate` method (see `interface_design.md`), plus the wrapper
  types (`ContainerConfig`/`HostConfig`/`NetworkingConfig`/`ContainerCreateRequest`).
- `RuntimeResolver` — resolves the label-based runtime only. The dedicated-instance branch is a stub (clear "not
  implemented yet" error) so the shape is right without building Phase 2 early.
- `labelBasedRuntime.ContainerCreate` — real name-conflict resolution, not just label-stamping: the actual Docker
  container **name** is `name` when suffix is empty (default/PROD, unchanged from today) or `name_<suffix>` when
  suffix is set. This revives the exact convention the old single-suffix pipeliner used
  (`internal/pipelines/do_smerd_launch.go`'s `req.Name + "_" + suffix`) before it was bypassed when environment
  threading replaced the static suffix field.
- One call-site change: `internal/jobs/create_smerd.go`'s `createContainerJob` calls
  `RuntimeResolver.Runtime(ctx, environment).ContainerCreate(...)` instead of
  `NodeClients.Docker().ContainerCreate(...)` directly.
- One RED→GREEN integration test (see "Test design" below).

**Explicitly deferred at the time of this phase (all since done — see the durable status note at the top of
`README.md` and the stage-by-stage log below):**
- `Inspect`/`Stop`/`Restart`/`Stats`/`Exec` — added on `labelBasedRuntime`, using the same
  `resolveOwnedContainer`/`resolveOwnedContainerInfo` ownership check as `Remove`/`Rename`.
  `container_manager.InspectSmerd`, `copy_to_volume.go`'s `copyFileJob`, `verv_services`'s
  `stop_restart.go`/`metrics.go` all resolve and call through `RuntimeResolver` now.
- `CreateNetwork`/`ConnectToNetwork`/`DisconnectFromNetworks` — added on `labelBasedRuntime`, suffixed exactly
  like container names (each environment gets its own `verv_<suffix>` network instead of every environment
  sharing one hardcoded `"verv"` network). `create_smerd.go`/`upgrade_smerd.go`'s network-join/pause/rollback
  logic goes through them instead of a raw `client.APIClient`.
- `PullImage`/`ListOccupiedPorts` — added on `commonRuntime` (node-wide, no suffix logic, embedded by every
  backend). Every caller (`assemble_config.go`, `create_smerd.go`, `upgrade_smerd.go`,
  `connect_service_to_vpn.go`, `enable_statefull.go`'s port-conflict check,
  `internal/app/custom.go`'s boot-time port-manager seeding) now resolves through `RuntimeResolver` instead of
  `node_clients.Docker` directly. See `docs/ports_management/roadmap.md` for why the boot-time seeding needed
  reordering (`RuntimeResolver` doesn't exist until after `NodeClients` does) and for a proposed future
  improvement (host-level `/proc/net/tcp`-based port watcher, to catch non-Docker port occupancy too).

(`ListContainers`, `Remove`, `Rename` and `IsContainerRunning` were originally deferred here too, but follow-up
passes in the same phase added all four — see "Names are always virtual at the interface boundary" and "Suffix
filtering has no 'unscoped' escape hatch" in `interface_design.md`, and the bug-fix entries below. `Rename`/
`IsContainerRunning` were added to unblock `internal/jobs/upgrade_smerd.go`'s suffix-aware container lookup/rename
steps — see the jobs-engine suffixed-environment upgrade fix. **Done:** the rename/drop/rollback jobs themselves
have since been cut over to call them — `renameContainerJob` and the new `dropOwnedContainerJob` (replacing two
`dropScratchContainerJob` call sites) now resolve a `ContainerRuntime` for the request's environment instead of
touching `node_clients.Docker`/raw `client.APIClient` directly, and `createContainerJob.Rollback`/
`renamingCreateContainerJob.Rollback` do the same for container removal. `IsContainerRunning` is still unused —
`pauseOldContainerJob`/`healthcheckJob` migrations were explicitly deferred, see the suffixed-environment upgrade
fix plan's "Deferred" note.)
- Dedicated-docker-instance real implementation — see Phase 2 below.
- pg-state matrix cells — need the cluster/`WithMatreshka` e2e fixture.
- A same-name-cross-environment test case — blocked on the task-dedup bug below.

## Test design

`tests/e2e/suite_container_runtime_test.go`, package `e2e`, following the precedent `suite_environments_test.go`
already set: single-node/`local_storage`, no `WithMatreshka()`, since Phase 1's only case needs no cluster/
Postgres.

```go
type containerRuntimeTestCase struct {
    name           string
    stateMode      string // "none" | "postgres"
    runtimeBackend string // "label" | "dedicated"
    environment    string
}

var containerRuntimeMatrix = []containerRuntimeTestCase{
    {name: "no-state/label-based (default)", stateMode: "none", runtimeBackend: "label", environment: "PROD"},
    // {stateMode: "postgres", runtimeBackend: "label"}      — next
    // {stateMode: "none",     runtimeBackend: "dedicated"}  — next
    // {stateMode: "postgres", runtimeBackend: "dedicated"}  — next
}

func Test_ContainerRuntime_Matrix(t *testing.T) {
    for _, tc := range containerRuntimeMatrix {
        t.Run(tc.name, func(t *testing.T) { runContainerRuntimeCase(t, tc) })
    }
}
```

Each subtest: stands up the fixture matching `stateMode`/`runtimeBackend` (only `none`/`label` actually wired —
others fail loudly with an explicit "add fixture support for X" until their turn comes, rather than silently
skipping), fires `CreateSmerd` through the real gRPC/transport path with `Environment: tc.environment`, polls the
task to `DONE` (reuse the existing e2e task-watch helper — see how `Test_Lifecycle` does it), then asserts against
the live Docker daemon that a container exists with the expected name/suffix label. Written RED first (interface
exists, resolver/runtime don't yet, or return "not implemented"), then implementation flips it green.

## Known bugs blocking later phases (found during the multi-environment pass, not yet fixed)

These don't block Phase 1 (which only exercises one environment, one case) but must be fixed before the matrix
cells that depend on them can go green:

1. **Jobs engine task dedup key has no environment component.** `internal/jobs`' smerd-create task enqueues on
   `req.GetName()` + action alone (see `smerd_create.go`). Creating the same-named smerd in two different
   environments dedups onto the same already-`DONE` task and returns the *first* environment's container UUID —
   confirmed live against real Docker. Must become environment-aware before a same-name-across-environments test
   case can be added to the matrix. **Still open** (unrelated to the name-virtualization/`Remove`-scoping fix
   below — `Test_SameNameInTwoEnvironments_AreDistinctContainers` stays `t.Skip`'d for this reason).
2. ~~**`DropSmerd` ignores environment scope entirely.**~~ **Fixed.** `labelBasedRuntime.Remove` now resolves
   the given identifier (bare name, already-suffixed name, or raw UUID) to a real container via
   `ContainerInspect`, then compares that container's `labels.SuffixLabel` exactly against `r.suffix` before
   removing — a mismatch (including a raw UUID that belongs to a different environment) is treated as "nothing to
   remove here" and reported as idempotent success, never an error and never an actual deletion. Both
   `tests/e2e/suite_environments_test.go`'s `Test_DropSmerd_ByBareName_SilentlyNoOpsInSuffixedEnvironment` and
   `Test_DropSmerd_ByUuid_CrossEnvironmentCollision` are green.

## Phase 2 (future): dedicated Docker instance

- `pkg/docker_setup` (public, self-contained, no `internal/...` imports) — config-authoring only:
  `GenerateDaemonConfig`, `GenerateSystemdUnit`, `WriteDaemonSetup`. No `os/exec`, no systemctl calls, no client
  construction — that orchestration lives in `internal/service/service_manager/verv_services/dedicated_docker.go`
  (new file).
- Orchestration: resolve host paths per binary/container mode (`containerinfo.IsInContainer()` already exists,
  `internal/cluster/env/containerinfo`) → **collision pre-flight** (no other environment already claims this
  `docker_host`; target socket path doesn't already exist; `systemctl status docker-<env>` doesn't already resolve
  to an existing unit; target data-root directory doesn't already exist non-empty) → call `pkg/docker_setup` →
  `systemctl daemon-reload && systemctl enable --now docker-<env>` → construct `*client.Client` via
  `client.WithHost(...)`, health-check it → persist `docker_host` on the environment row.
- Schema: `ALTER TABLE velez.environments ADD COLUMN IF NOT EXISTS docker_host TEXT NOT NULL DEFAULT ''`. Proto:
  `CreateEnvironment.Request.dedicated_docker bool`, `Environment.docker_host string`.
- Config: two new fields under `config.yaml`'s `environment:` schema list (`docker_setup_host_config_path`,
  `docker_setup_host_systemd_path`) — regenerated via `moti g`, same mechanism as `container_suffix`. Empty when
  running as a raw binary (real host paths used directly); must be set (mounted) when containerized, or
  `CreateEnvironment(dedicated_docker=true)` fails with a clear error.
- `DeleteEnvironment` cascade gains: `systemctl disable --now` + data-root cleanup for dedicated environments.
- **Security note (acknowledged, not blocking):** exec'ing `systemctl` against the host from inside a container
  needs the host's systemd/D-Bus socket reachable from the container — a real trust-boundary expansion beyond
  anything built so far. Coverage via AppArmor is planned as a separate future pass, not a Phase 2 blocker.

## Phase 3 (future)

- pg-state matrix cells (needs `WithMatreshka` cluster fixture).
- Client-side UX: `EnvironmentCreateDialog` gets a "dedicated Docker instance" checkbox plus inline instructions
  (what to mount / that binary mode needs no setup) so users know what enabling the toggle actually requires.
