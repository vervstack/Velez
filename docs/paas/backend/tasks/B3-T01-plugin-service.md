---
id: "B3-T01"
title: "Plugin service with dual-mode storage and hot-switch"
status: "pending"
created: "2026-05-05"
branch: "task/B3-T01-plugin-service"
---

# Task B3-T01 — Plugin Service with Dual-Mode Storage and Hot-Switch

## Goal

Add a `ListPlugins` RPC to `ControlPlaneAPI` backed by a `PluginService` at the service layer that dynamically inspects
running containers when postgres is disabled, queries postgres when it is enabled, and switches between the two
implementations atomically using the same pattern as `ClusterStateManagerContainer`.

## Context

`ListVervServices` in `internal/transport/control_plane_api_impl/list_verv_services.go` currently resolves plugin state
inline by walking all smerds and matching container names. The new design extracts this into a proper service layer.

**Hot-switch pattern**: `internal/clients/cluster_clients/state/state.go` uses `atomic.Pointer[ClusterStateManager]` to
hold the active implementation. The `PluginServiceContainer` must replicate this — start with the docker-backed `noImpl`
-style implementation, then call `Set(pgImpl)` when postgres connects, and revert to `Set(dockerImpl)` when it
disconnects.

**Dynamic (no-pg) implementation**: call `ListSmerds` (already used in the handler today) and map known container
names (`makosh`, `matreshka`, `portainer`, `headscale`, `pg-statefull`) to `Plugin` entries with their running state —
same mapping logic already in `list_verv_services.go`. This is the default until pg is enabled.

**Postgres implementation**: executes a sqlc query that reads plugin state from a new or existing table. Define the
schema if the table does not exist (new migration).

The `ControlPlaneAPI` transport handler for `ListPlugins` must call `pluginService.ListPlugins(ctx, req)` — it must not
reproduce any discovery logic inline.

The `PluginServiceContainer` must be wired in `internal/app/` and its `Set` method called from the same place that
currently calls `ClusterStateManagerContainer.Set` (when pg comes up or goes down).

## Acceptance Criteria

- [ ] `ListPlugins` RPC exists in `control_plane_api.proto` with `Plugin` message containing at minimum `type` (
  VervServiceType) and `state` fields
- [ ] `PluginService` interface is defined in `internal/service/`
- [ ] `PluginServiceContainer` wraps the interface behind `atomic.Pointer` with a `Set(impl)` method
- [ ] Docker-backed implementation returns correct plugin states by querying running smerds (no postgres dependency)
- [ ] Postgres-backed implementation executes a sqlc query and returns the same shape
- [ ] `PluginServiceContainer.Set` is called when postgres connects/disconnects (matching the
  `ClusterStateManagerContainer` call sites)
- [ ] Transport handler `ListPlugins` delegates entirely to `pluginService.ListPlugins` — no inline container name
  matching
- [ ] `make codegen` succeeds after proto changes
- [ ] `go build ./...` passes
- [ ] `make lint` passes (or pre-existing lint failures are unchanged)

## Files to Create / Modify

- `api/grpc/control_plane_api.proto` — add `ListPlugins` RPC + `Plugin` + `ListPlugins.Request/Response` messages
- `internal/service/plugin_service.go` — `PluginService` interface + `PluginServiceContainer` type
- `internal/service/service_manager/plugins/docker_impl.go` — docker-backed implementation
- `internal/service/service_manager/plugins/docker_impl_test.go` — unit tests for docker impl
- `internal/service/service_manager/plugins/pg_impl.go` — postgres-backed implementation
- `internal/storage/plugin_storage.go` — `PluginsStorage` interface (if postgres impl needs its own storage type)
- `internal/storage/postgres/queries/plugins.sql` — sqlc query for plugin state
- `migrations/` — new goose migration if a plugins table is required
- `internal/transport/control_plane_api_impl/list_plugins.go` — gRPC handler
- `internal/app/` — wire `PluginServiceContainer`, call `Set` alongside existing `ClusterStateManagerContainer.Set`
  calls

## Do NOT change

- `internal/transport/control_plane_api_impl/list_verv_services.go` — existing RPC stays untouched
- Existing proto messages in `control_plane_api.proto` — additive only

## Notes

- Two separate lines for error check — never `if err = ...; err != nil`.
- Struct literals must be assigned to named variables before passing to functions.
- The docker impl's discovery logic is already proven in `list_verv_services.go` — reuse the same name → VervServiceType
  mapping rather than reimplementing it.
- If no plugins table exists in postgres yet, the pg impl may start as a thin read of the same container-name logic
  cached in pg; the schema can be minimal (plugin_type INT, state INT, updated_at TIMESTAMPTZ).
