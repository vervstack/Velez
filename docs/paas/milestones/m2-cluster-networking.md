# M2 — Cluster & Networking

**Status:** Backlog — flesh out after M1 ships.

## Scope (draft)

- Control plane page (`/cp`): node list, cluster health, join/leave actions
- Verv Closed Network page (`/vcn`): peer list, VPN status, add/remove peers
- Multi-node deployment targeting — choose which node(s) a service runs on
- Node-level resource summary (CPU, RAM, disk) surfaced in the UI

## Dependencies

- `control_plane_api.proto` and `verv_closed_network.proto` must be fully implemented in the backend
- Headscale integration stable in the service layer
