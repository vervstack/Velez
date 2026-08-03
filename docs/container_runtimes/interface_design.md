# Interface design

## Naming

`ContainerRuntime`, not `IContainer`. Go convention avoids the Hungarian `I`-prefix (interfaces are named by
capability), and `IContainer` reads as "one container instance" rather than "the engine that manages containers".
Concrete implementations: `labelBasedRuntime`, `dedicatedRuntime` (unexported — callers only ever hold the
interface or the resolver below).

## Wrapper types for Docker SDK parameters

`ContainerCreate` currently takes raw Docker SDK types directly as parameters. Rather than either (a) keeping bare
SDK types forever (blocks any future non-Docker backend from satisfying the interface) or (b) translating to fully
independent domain types now (speculative refactor for a capability — podman/containerd/Kata — that doesn't exist
yet), wrap them:

```go
// Thin pass-throughs today. Callers already code against the Velez-owned
// name, so when a future backend needs a genuinely different shape, these
// wrappers absorb that change without touching every call site.
type ContainerConfig struct {
    *container.Config
}
type HostConfig struct {
    *container.HostConfig
}
type NetworkingConfig struct {
    *network.NetworkingConfig
}

type ContainerCreateRequest struct {
    Config           *ContainerConfig
    HostConfig       *HostConfig
    NetworkingConfig *NetworkingConfig
    Platform         *v1.Platform
    ContainerName    string
}
```

## The interface

```go
// ContainerRuntime is what the service layer depends on instead of a
// concrete Docker client. Which concrete implementation serves a given
// environment (shared label-scoped daemon today; a dedicated daemon;
// podman/containerd/Kata someday) is resolved per-call and invisible here.
type ContainerRuntime interface {
    ContainerCreate(ctx context.Context, req ContainerCreateRequest) (container.CreateResponse, error)
    ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error)
    Remove(ctx context.Context, containerID string) error
    Stop(ctx context.Context, nameOrID string) error
    Restart(ctx context.Context, nameOrID string) error
    IsContainerRunning(ctx context.Context, nameOrID string) (running, exists bool, err error)
    Exec(ctx context.Context, containerID string, cfg container.ExecOptions) ([]byte, error)
    Stats(ctx context.Context, nameOrID string) (domain.ContainerStats, error)
    ListOccupiedPorts(ctx context.Context) ([]uint32, error)

    // Shared across every backend — implemented once on commonRuntime, embedded everywhere.
    PullImage(ctx context.Context, imageName string) (image.InspectResponse, error)
    CreateNetwork(ctx context.Context, name string) error
    ConnectToNetwork(ctx context.Context, req dockerutils.ConnectToNetworkRequest) error
    DisconnectFromNetworks(ctx context.Context, containerID string, networks []string) error
}
```

Note what's gone from today's signatures: no `suffix string` parameter, no `Client() client.APIClient` escape
hatch. Both are folded into the implementations below instead of being caller-supplied plumbing.

**Phase 1 implements only `ContainerCreate`** — see `roadmap.md`. The rest of the interface above is the target
shape; other methods stay on the existing `Docker` struct directly until their own phase.

## Common vs. divergent implementation (embedding)

```go
// commonRuntime backs operations that don't differ by backend. It only
// needs a raw client.APIClient — doesn't know or care whether that
// connection points at the shared daemon or a dedicated one.
type commonRuntime struct { cli client.APIClient }
func (c *commonRuntime) PullImage(...) {...}
func (c *commonRuntime) CreateNetwork(...) {...}
func (c *commonRuntime) ConnectToNetwork(...) {...}
func (c *commonRuntime) DisconnectFromNetworks(...) {...}

// labelBasedRuntime — tier 1, default. Stamps/filters by suffix label AND
// suffixes the actual container NAME (name_<suffix>) so two environments
// with the same logical smerd name don't collide on the shared daemon —
// see roadmap.md's "name conflict resolution" note. Suffix empty ("" —
// today's default/PROD) means "unsuffixed", preserving pre-multi-env
// container naming exactly.
type labelBasedRuntime struct {
    commonRuntime
    suffix string
}
func (r *labelBasedRuntime) ContainerCreate(...) {...}
func (r *labelBasedRuntime) ListContainers(...) {...}

// dedicatedRuntime — tier 2. Owns its own client.APIClient pointed at that
// environment's private daemon; no label/name-suffixing logic needed since
// the whole daemon is already exclusive to this environment.
type dedicatedRuntime struct {
    commonRuntime // cli here is a DIFFERENT connection than the shared one
}
func (r *dedicatedRuntime) ContainerCreate(...) {...}
func (r *dedicatedRuntime) ListContainers(...) {...}
```

## Resolution and hot-swap

```go
type RuntimeResolver interface {
    Runtime(ctx context.Context, environment string) (ContainerRuntime, error)
}
```

`Runtime()` reads the environment's row from whichever `EnvironmentsStorage` is currently active (static or pg —
the existing, already-solved state axis), fresh on every call: `docker_host == ""` → shared `labelBasedRuntime`
with that environment's suffix; `docker_host` set → a cached `dedicatedRuntime` for that host (connections
pooled/reused, not reopened per call). Because it re-reads storage on every call, flipping an environment between
tiers takes effect on the next call — **no explicit atomic-swap container needed**, unlike the state axis's
static→pg swap (which has to move a process-wide singleton, not a per-environment lookup).
