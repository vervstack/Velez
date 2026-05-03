# M2 — Cluster & Networking

**Status:** Backlog — flesh out after M1 ships.

## Scope (draft)

- Control plane page (`/cp`): node list, cluster health, join/leave actions
- Verv Closed Network page (`/vcn`): peer list, VPN status, add/remove peers
- Multi-node deployment targeting — choose which node(s) a service runs on
- Node-level resource summary (CPU, RAM, disk) surfaced in the UI
- Node tags — see T-node-tags below

## T-node-tags — Node Scheduling Tags

### Goal

Allow operators to annotate nodes with scheduling intent tags so that the Velez scheduler can route deployments to the
most appropriate node automatically. Some tags are set manually; others are auto-assigned at node registration based on
detected hardware.

### Tags

| Tag                 | Set by             | Meaning                                                                                                                                                |
|---------------------|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `stateful-priority` | operator           | Node is preferred for stateful workloads (databases, services with persistent volumes). Scheduler bias: route stateful services here first.            |
| `short-living`      | operator           | Node is intended for short-lived / ephemeral containers (batch jobs, CI runners, one-shot tasks). Scheduler avoids placing long-running services here. |
| `test-only`         | operator           | Non-production node. Deployments with environment = `test` or `dev` may land here; production deploys are excluded.                                    |
| `cpu-heavy`         | auto (system info) | Node has high CPU core count / clock speed relative to the cluster. Auto-assigned at registration; scheduler routes CPU-intensive workloads here.      |
| `disk-heavy`        | auto (system info) | Node has large or fast local disk relative to the cluster. Auto-assigned at registration; scheduler routes disk-intensive workloads here.              |

### Tasks

- [ ] **Backend** — add `tags []string` field to `NodeInfo` proto and persist in DB
- [ ] **Auto-detection** — on node registration, inspect `runtime.NumCPU()` and disk stats; apply `cpu-heavy` /
  `disk-heavy` thresholds relative to cluster median
- [ ] **API** — `SetNodeTags(node_id, tags)` and `GetNodeTags(node_id)` endpoints in `control_plane_api.proto`
- [ ] **Scheduler** — when placing a deployment, score nodes by tag compatibility (stateful-priority, short-living,
  test-only exclusion, resource tags)
- [ ] **UI** — show tag chips on node cards in `/cp`; allow operator to add/remove manual tags via node detail panel

## Dependencies

- `control_plane_api.proto` and `verv_closed_network.proto` must be fully implemented in the backend
- Headscale integration stable in the service layer
