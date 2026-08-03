# Container Runtimes

**Status: full `ContainerRuntime` interface implemented.** `ContainerCreate`, `ListContainers`, `Remove`, `Rename`,
`IsContainerRunning`, `Inspect`, `Stop`, `Restart`, `Stats`, `Exec`, `CreateNetwork`, `ConnectToNetwork` and
`DisconnectFromNetworks` are live on `labelBasedRuntime`; `PullImage` and `ListOccupiedPorts` are live on
`commonRuntime` (node-wide, no suffix logic - embedded by every backend). Every job/service call site that used to
go through `node_clients.Docker`/raw `client.APIClient` now resolves and calls through `RuntimeResolver` instead.
Remaining work is deleting the old pre-jobs-engine `internal/pipelines` package entirely (its last two callers,
`autoupgrade.go`/`deploy_watcher.go`, still need migrating) and Phase 2 (dedicated Docker instance, still a stub).
This directory is the durable record of the design session that produced this work, so future phases can continue
from here instead of re-deriving the shape from scratch. See [`interface_design.md`](interface_design.md) for the
concrete Go types/interfaces and [`roadmap.md`](roadmap.md) for what's in/out of scope per phase and the known
bugs that block later phases.

## Motivation

Velez's multi-environment feature (`velez.environments`, environment threaded through `CreateSmerd`/`ListSmerds`/
`DropSmerd`/`UpgradeSmerd`/`CreateService`/`CreateDeploy` — see `docs/roadmap.md` and the environments migrations)
today ties every container operation to a single global Docker connection (`internal/clients/node_clients/docker`,
constructed once via `client.FromEnv`). Isolation between environments is achieved purely by stamping/filtering a
`VELEZ_SUFFIX` label. That's the default and will stay the default — but some users want a genuinely separate
Docker daemon per environment (no shared kernel-level container namespace at all), and further out, other
container backends entirely (podman, containerd, Kata) are plausible. This needs a pluggable "how do we actually
talk to containers" layer between the service layer and the concrete backend, so the service layer never has to
know which one is active for a given environment.

## Two independent axes

This is a 2×2 (and growing) matrix, not one dimension:

| | **State** (where environment metadata lives) | **Runtime backend** (how containers are actually managed) |
|---|---|---|
| Already solved | `internal/storage/environments/{static,pg}.go` — Velez's existing local-vs-cluster storage switch, hot-swappable via the same atomic-pointer container pattern already used for `ClusterStateManagerContainer` | — |
| New (this doc set) | — | Label-based shared daemon (default, tier 1) vs. dedicated per-environment daemon (opt-in, tier 2) vs. future backends |

Do not conflate these. The "no state / state in pg" distinction the design session referred to is the **existing**
axis and needs no new interface — it's reused as-is. The **runtime-backend** axis is what `interface_design.md`
actually specifies.

## Current concrete Docker surface (reference, as of this design pass)

`internal/clients/node_clients/docker/client.go`'s `Docker` struct: `PullImage`, `Remove`, `Stop`, `Restart`,
`IsContainerRunning`, `ListOccupiedPorts`, `ListContainers`, `Exec`, `Client()` (raw `client.APIClient` escape
hatch), `ContainerCreate`, `Stats` — plus `dockerutils.CreateNetwork`/`ConnectToNetwork`/`DisconnectFromNetworks`,
which today bypass the wrapper entirely via `Client()`. `ContainerCreate` already takes raw Docker SDK types
directly (`*container.Config`, `*container.HostConfig`, `*network.NetworkingConfig`, `*v1.Platform`) — this is
what `interface_design.md`'s wrapper types are for.
