# Roadmap

---

## Milestone 1 — Unfuck Current Setup

**Goal:** The basic lifecycle — deploy a container, upgrade it, drop it — works end-to-end without panics, silent
failures, or stale state.

### Complete the critical stubs

- **`syncRunningBatch()`** — reconcile running containers against DB: inspect each, mark FAILED if stopped/dead
- **`deleteBatch()`** — drop containers in SCHEDULED_DELETION, update status to DELETED

### Fix proto + storage blockers

- **`ListDeployments.Response`** — add fields: `repeated DeploymentInfo deployments = 1; uint64 total = 2;` with a
  `DeploymentInfo` message (id, status, spec_id, created_at)
- **`VervAppService`** — add `current_deployment_id` and `status` to `GetService` response
- Implement `ListDeployments` handler and storage query once proto is updated

### Fix wiring

- **`Custom.Stop()`** — implement graceful shutdown: stop workers, flush in-flight pipelines
- **`autoupgrade`** — run in a goroutine with context cancellation, not blocking the errgroup; add a startup timeout

### Add self-upgrade guard

- Before upgrading a container, check if it's the velez container itself (by label or name pattern); refuse or schedule
  a controlled restart

---

## Milestone 2 — Harden & Observe

**Goal:** Safe enough to run continuously. Failures are visible, recoverable, and don't cascade.

- Add deployment events table (append-only: deployment_id, status, error, timestamp)
- Config subscription (VERV-128): uncomment `handleConfigurationSubscription`, implement the restart action
- Circuit breaker on matreshka client: cache last-known-good config, stop failing deploys when matreshka is down
- Consistent gRPC error codes across all API handlers (some return raw errors, some use `codes.*`)
- Add schema fields to `deployments`: `replica_count`, `desired_count`, `rollback_target_spec_id`
- Move `local_state_path` default out of `/tmp` (lost on reboot) to a configurable persistent path
- Fix node ID: read from local state manager instead of hardcoding `1`

---

## Milestone 3 — Podman Support

**Goal:** Run the same service definitions on a podman socket with zero proto changes.

- Define `ContainerRuntime` interface in `internal/clients/node_clients/` covering the operations currently on the
  Docker wrapper
- Implement podman adapter (podman's API is Docker-compatible; adapter should be thin)
- Add `runtime` field to config (docker | podman), wire selection at app startup
- Test: create/list/drop a container through each backend

---

## Milestone 4 — K8s Mappers

**Goal:** Translate Velez service definitions into Kubernetes resources. No new runtime required — this is a
mapper/translator layer.

- Add `BackendType` to proto (`docker | podman | k8s`) — proto version bump required
- Abstract `Volume` type: add `volume_type` enum (named | host_path | pvc | emptydir)
- Abstract `Port` exposure model: separate host-bind from service-expose intent
- Implement k8s mapper: `domain.Service` → `k8s Deployment + Service + PVC` YAML/API call
- CPU/RAM: add a `milliCPU` alternative field alongside the float; k8s adapter uses milliCPU

---

## Milestone 5 — Multi-Node Clustering

**Goal:** Run services across multiple nodes with automatic placement and VPN mesh.

- Complete the three VPN setup paths in `verv_closed_network/server.go` (currently two paths return `DisabledVcnImpl{}`
  silently)
- Implement `ConnectSlave` node registration (proto message is empty — needs fields: node_id, address, capabilities)
- Service discovery: implement makosh integration for DNS/routing between services across nodes
- Cluster-aware scheduler: placement based on node capacity (hardware scanner already collects CPU/RAM)
- Heartbeat and failover: if a node goes silent, reschedule its RUNNING deployments on other nodes
