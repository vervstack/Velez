# Port occupancy tracking

## Current state (as of the ContainerRuntime PullImage/ListOccupiedPorts pass)

`ports.Container` (`internal/clients/node_clients/ports/`) tracks which host ports are free/locked/held, seeded once
at boot with whatever's already occupied. That seed comes from `ContainerRuntime.ListOccupiedPorts`
(`internal/clients/node_clients/container_runtime/common.go`), which is `docker.Docker.ListOccupiedPorts`'s exact
query: list every container on the daemon, collect `PublicPort` off each one's port bindings. This only sees ports
Docker itself published - not ports a non-Docker process on the host happens to be listening on.

Boot sequencing (`internal/app/custom.go`): `node_clients.NewNodeClients` can't seed this itself anymore -
`ListOccupiedPorts` needs a resolved `ContainerRuntime`, and `RuntimeResolver` doesn't exist until `Custom.Init`
builds one a few lines later (needs `NodeClients.Docker().Client()`). So `NewNodeClients` constructs an unseeded
port manager, and `Custom.Init`'s `seedOccupiedPorts` helper reseeds it via `PortManagerContainer().Set(...)` once
the resolver exists. This is why the flow looks like two steps instead of one - see the doc comments on both
`node_clients.go`'s "Port manager" block and `custom.go`'s `seedOccupiedPorts`.

**Known gap:** a port bound by a non-Docker process on the host (or by Docker but not through this Velez
instance's tracked containers, e.g. another Docker Compose stack) is invisible to this seed. Velez could hand that
port out to a smerd, and `ContainerCreate`'s host bind would fail with an opaque Docker error instead of Velez's
own `ErrUnavailablePort`/pre-check message.

## Proposed improvement: host-level listening-port watcher via `/proc/net/tcp`

**Goal:** seed (and periodically refresh) occupied-port tracking from the actual host TCP listener table, not just
Docker's view of it - catching ports occupied by anything, not only Velez-managed containers.

**Why `/proc/net/tcp`/`/proc/net/tcp6` specifically, not a Go "list open ports" library:** Velez can run either as
a bare binary on the host or inside a container (`internal/cluster/env/containerinfo.IsInContainer()` already
distinguishes these). A library that inspects the *calling process's* network namespace (e.g. `gopsutil`'s
`net.Connections`) is correct for the bare-binary case but wrong for the containerized one - a container has its
own network namespace, so it would only see its own listening sockets, not the host's. `/proc/net/tcp` read
through a **host-mounted `/proc`** (bind-mounting the host's `/proc` into Velez's container - the same shape as the
existing `-v /var/run/docker.sock:/var/run/docker.sock` requirement noted in
`internal/clients/node_clients/node_clients.go`'s `dockerSocketHint`) sees the real host listener table regardless
of which mode Velez runs in, as long as the mount is present.

**Design sketch (not yet implemented):**

- New package, e.g. `internal/clients/node_clients/hostports/` (mirrors `internal/clients/node_clients/ports/`'s
  existing layering: a pure-parsing piece plus a thin wrapper).
- `ParseListeningPorts(procNetTCP io.Reader) ([]uint32, error)` - parses `/proc/net/tcp` and `/proc/net/tcp6`'s
  fixed-width format (hex local-address:port, `st` column `0A` = `TCP_LISTEN`), pure function, no filesystem
  access, so it's unit-testable against fixture strings without a real `/proc`.
- A thin `HostPortWatcher` wraps that, reading from `/proc/net/tcp`+`/proc/net/tcp6` at a configurable path
  (defaulting to `/proc/net/tcp{,6}`, overridable for tests and for a possible non-standard mount point) and
  merging with Docker's own `ListOccupiedPorts` view (union, not replacement - Docker's view still matters for
  attributing *which* container/environment holds a port, which `/proc/net/tcp` alone can't tell you).
  Bare-binary mode already sees the real host `/proc` for free; containerized mode needs the operator to mount
  `/proc` (host) -> `/proc` (container) or a dedicated path, similar to the Docker socket mount. Missing/unreadable
  `/proc/net/tcp` should degrade to today's Docker-only view with a logged warning, not fail startup - this is an
  enhancement to occupancy detection, not a hard dependency.
- Config: a new `environment.host_proc_path` (or similar) config field, empty by default (falls back to
  `/proc/net/tcp` directly, correct for the bare-binary/already-host-namespaced case); only needs setting when
  Velez runs containerized and the mount path differs from the default `/proc`.
- Wiring: `seedOccupiedPorts` (`internal/app/custom.go`) would union `ContainerRuntime.ListOccupiedPorts` with
  `HostPortWatcher.ListeningPorts()` instead of using the former alone. A periodic refresh (not just boot-time
  seeding) is a natural follow-up once the one-shot read works, so a port that becomes occupied by an external
  process *after* boot is caught too - `ports.Container.Set` already supports swapping in a freshly-seeded
  `PortManager` at any time, so this doesn't need new plumbing on that side.

**Explicitly out of scope for the first pass:** actually attributing an externally-occupied port to a specific
process (would need `/proc/<pid>/fd` inode matching, real added complexity) - for allocation-conflict purposes,
"is this port free" is the only question that matters; "who's holding it" stays a debugging nicety, not a
requirement.

**Status:** design only, not started. No code exists yet under `internal/clients/node_clients/hostports/`.
