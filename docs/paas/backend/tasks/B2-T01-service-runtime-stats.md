---
id: "B2-T01"
title: "Service runtime stats API (CPU%, memory)"
status: "done"
created: "2026-05-04"
branch: "task/B2-T01-service-runtime-stats"
---

# Task B2-T01 — Service Runtime Stats API

> **2026-07-26 update:** this shipped already, under the RPC name `GetServiceMetrics` rather than
> `GetServiceStats` (hence a literal grep for the latter turned up nothing and the roadmap listed
> this as not started). See `api/grpc/service_api.proto:73` (RPC) and `:241` (messages),
> `internal/service/service_manager/verv_services/metrics.go:14` (aggregation logic using
> `Docker.Stats()`), `internal/transport/service_api_impl/metrics.go:11` (handler), and
> `pkg/web/Velez-UI/src/widgets/service/ServiceHero/ServiceHero.tsx` (frontend consumer via
> `useGetServiceMetricsQuery`). The response is richer than this spec asked for — it also includes
> `replicas_running`/`replicas_desired`/`uptime_seconds`. Acceptance criteria below are left as
> originally written for reference; all are satisfied by the shipped implementation.

## Goal

Add a `GetServiceStats` RPC that returns current CPU usage (%) and memory usage (bytes + limit) for the running
container(s) of a given service, sourced from the Docker stats API.

## Context

The ServiceInfoPage (M9-T33) will display a runtime stats section (CPU%, RAM) in the Overview tab. Docker exposes
per-container stats via `ContainerStats` in the Go SDK (`github.com/docker/docker/client`). Velez already uses this
client in `internal/clients/docker/`.

The RPC should be added to `service_api.proto`. A service may have multiple running smerds; return aggregated stats (sum
of CPU%, sum of mem used, max of mem limit).

CPU% calculation from Docker stats:

```
cpuDelta = stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
systemDelta = stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
numCPUs = len(stats.CPUStats.CPUUsage.PercpuUsage)
cpuPercent = (cpuDelta / systemDelta) * numCPUs * 100.0
```

## Proto Changes

Add to `api/grpc/service_api.proto`:

```proto
message GetServiceStats_Request {
  string service_name = 1;
}

message GetServiceStats_Response {
  double cpu_percent    = 1;  // 0.0–100.0 per core, aggregated across containers
  uint64 memory_used    = 2;  // bytes currently used
  uint64 memory_limit   = 3;  // bytes available (container limit)
  repeated string smerd_ids = 4;  // IDs of containers included in stats
}

service ServiceAPI {
  // existing RPCs ...
  rpc GetServiceStats(GetServiceStats_Request) returns (GetServiceStats_Response) {}
}
```

Run `make codegen` after proto changes.

## Acceptance Criteria

- [ ] `GetServiceStats` RPC exists and responds without error for a running service
- [ ] Response includes non-zero `cpu_percent` and `memory_used` when at least one smerd is running
- [ ] Returns zeroed stats (no error) when no smerds are running for the service
- [ ] Docker stats call uses one-shot mode (`stream=false`) to avoid blocking
- [ ] Unit test covers the CPU% calculation formula and the no-smerd case
- [ ] `make lint` passes (or pre-existing lint failures are unchanged)

## Files to Create / Modify

- `api/grpc/service_api.proto` — add `GetServiceStats` RPC + messages
- `internal/transport/` — gRPC handler implementation for `GetServiceStats`
- `internal/service/` — business logic: fetch smerds by service name, call Docker stats, aggregate
- `internal/clients/docker/` — add `ContainerStats(ctx, containerID)` method if absent
- Tests alongside each new file

## Do NOT change

- Existing RPCs in `service_api.proto` — additive only
- `internal/pipelines/` — stats is a read-only query, no pipeline needed

## Notes

- Docker `ContainerStats` with `stream=false` returns a single snapshot — sufficient for a dashboard poll.
- Error handling: if Docker returns an error for one container, log and skip it; aggregate the rest.
- Two separate lines for error check — never `if err = ...; err != nil`.
- Struct literals must be assigned to named variables before passing to functions.