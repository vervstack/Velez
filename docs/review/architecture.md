# Architecture & Design Concerns

These are not bugs — the code does what it's written to do. These are structural decisions that will need to be
revisited before adding podman, k8s, or meaningful multi-node support. Noted here so the cost is visible when planning
those milestones.

---

## Container Runtime Is Not Abstracted

The Docker client is used directly throughout the codebase — in pipelines, service managers, workers, cluster setup, and
transport implementations. There is no `ContainerRuntime` interface.

To add podman support today, you'd need to find every `docker.Client()` call site and add a branch. That's
unsustainable. The right structure is:

```
ContainerRuntime interface {
    CreateContainer(...)
    StartContainer(...)
    StopContainer(...)
    RemoveContainer(...)
    InspectContainer(...)
    PullImage(...)
    // ...
}
```

The Docker client wrapper in `internal/clients/node_clients/docker/` already has the shape of this — it's the right
place for the interface to live. Podman's API is compatible with Docker's, so the adapter would be thin.

---

## Proto Types Embed Docker Semantics

Several types in `velez_common.proto` and `velez_api.proto` map 1:1 to Docker concepts:

**`Port.exposed_to` (uint32)** — this is a host port number, Docker's port binding model. In k8s, ports are handled by
`Service` objects with selectors; there's no concept of "expose container port 80 as host port 8080". A higher-level
model would express _intent_ (expose this service on the network) rather than mechanism (bind to this host port).

**`RestartPolicyType` enum** — values are `unless_stopped`, `no`, `always`, `on_failure` — Docker's restart policy names
verbatim. k8s has `Always`, `OnFailure`, `Never` with different semantics. A podman system service maps differently
again.

**`Hardware.cpu` (float)** — Docker CPU shares work as fractional core allocation. k8s uses milliCPU (`100m` = 0.1
core). These aren't directly compatible; mapping float→string at the adapter layer is lossy.

**`Volume` (volume_name + container_path)** — only models named volumes. No concept of:

- Bind mounts (host path)
- k8s PersistentVolumeClaims
- Ephemeral volumes (emptyDir)

These don't need to change before podman support (podman uses the same volume model), but will need addressing for k8s.

---

## No Pod or Grouping Concept

Everything in Velez is a single container (`Smerd`). There's no concept of a group of containers that share a
lifecycle (Docker Compose's service, k8s Pod). This means:

- Sidecar containers (mentioned as TODO in the proto) have no natural home
- Logging agents, service mesh proxies, init containers have nowhere to attach
- A k8s Pod (multiple containers, shared network namespace) can't be expressed

The Headscale integration currently adds a tailscale sidecar via a separate container with shared networking — which
works for Docker but won't map to k8s where sidecars are first-class pod members.

---

## Network Model Assumes Docker Bridge

`NetworkBind` in the proto carries a `network_name` and `aliases`. This maps to Docker's overlay/bridge network model:
you create a named network, attach containers, they discover each other by alias.

For k8s:

- Pod-to-pod communication is flat and automatic; no named network required
- External exposure is via `Service` objects, not port bindings
- Network policies are separate resources

For multi-host Docker: named networks don't span hosts without Swarm/overlay configuration that Velez doesn't manage.

The current model works fine for single-node Docker. It becomes a mismatch at the abstraction boundary when adding
backends.

---

## Cluster State Is Node-Hardcoded

`deploy_watcher.go:44`:

```go
nodeId: 1,
```

The node ID used in all deployment queries is hardcoded to `1`. The deployments table has a `node_id` field; the schema
supports multi-node in principle. But the worker that drives deployments always assumes it's node 1. Adding a second
node means two watchers fighting over the same deployments.

Node identity should come from the local state manager (which does track a node ID) or from cluster registration.

---

## Matreshka Coupling

Config fetching from matreshka is woven into the container launch pipeline (`config_steps`). The pipeline step calls
matreshka synchronously; if matreshka is unreachable, the deploy fails.

There's no:

- Fallback (launch without config, use defaults)
- Caching (last-known-good config)
- Circuit breaker (stop hammering matreshka on repeated failures)

This means a matreshka outage causes all new deployments to fail, even for containers that don't need dynamic config.
The coupling should be optional — containers with `ignore_config: true` already skip this, but the infra around it
doesn't protect the rest.

---

## Deployment History Is Overwritten

Status updates (`UpdateDeploymentStatus`) replace the current status field. There is no events/transitions table. Once a
deployment reaches `RUNNING`, there's no record of how long it was `SCHEDULED`, when it transitioned, or what errors
occurred during `FAILED` states.

Debugging production issues will require log grepping. Any retry-on-failure logic, rollback detection, or SLA tracking
needs an append-only events table.
