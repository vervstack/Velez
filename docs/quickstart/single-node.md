# Quickstart: Single Node

Stand up one Velez node with Docker, create a service, deploy it, and verify it's running. This
walkthrough uses the raw HTTP gateway (`<ip>:53890/api`) per the project's documented integration
point; the bundled Web UI (`Velez-UI`, served from `/` on the same port) and the
`@vervstack/velez` TypeScript client cover the same operations if you'd rather not hand-write
requests.

## Prerequisites

- Docker installed and running, with the socket reachable at `/var/run/docker.sock`.
- Port `53890` free on the host (Velez's API port — gRPC and HTTP are multiplexed on this single
  port via `cmux`). It's configurable but this is the shipped default (`config/config.yaml`).
- The port range `18501-19000` free-ish on the host — this is `available_ports` in
  `config/config.yaml`, the range Velez allocates from when a container asks to expose a port.
- Nothing else. Matreshka (`matreshka_is_enabled`) and Makosh (`makosh_is_enabled`) both default to
  `false` in `config/config.yaml`, and Postgres is **not required** for a single node: the jobs
  engine and the services/deployments storage both fall back to a local JSON-file store
  (`internal/storage/local_storage`) unless a cluster Postgres DSN has been wired up in local state
  (see `docs/quickstart/multi-node.md` and the Gotchas section below).

## 1. Start Velez

Reuses the setup from the repo root `README.md`.

```shell
docker run \
   -p 53890:53890 \
   -d \
   --name velez \
   -v /var/run/docker.sock:/var/run/docker.sock \
   -v /dev/disk:/dev/disk \
   -v /run:/run \
   -v ~/velez:/tmp/velez \
   godverv/velez:latest
```

The three bind mounts (`docker.sock`, `/dev/disk`, `/run`) are required — Velez uses them to manage
containers and read host hardware info. The `~/velez:/tmp/velez` mount is optional but strongly
recommended: it persists Velez's local state file (see below) across container restarts.

Alternatively, to build and run locally from source:

```shell
make build-local-container   # builds an ARM64 image tagged velez:local
# or, for a plain binary:
go build -o ./service ./cmd/service/main.go
./service   # reads config/config.yaml (or config/dev.yaml if you copy it over) from the cwd
```

## 2. Find your access key

Every write RPC (and most reads) requires an `Authorization` key, checked by
`internal/middleware/grpc_incoming_interceptor.go` against the `VelezKey` Velez generates on first
boot. It's persisted in the local state JSON file:

- Default path: `/tmp/velez/local_state.json` (`local_state_path` in `config/config.yaml`).
- With the docker run command above, that's `~/velez/local_state.json` on the host.

```shell
cat ~/velez/local_state.json | grep VelezKey
```

Pass it as the gRPC metadata key `Authorization`, which grpc-gateway exposes over HTTP as the
header `Grpc-Metadata-Authorization`:

```shell
curl -X GET "http://localhost:53890/api/version" \
  -H "Grpc-Metadata-Authorization: <your VelezKey>"
```

> **Note:** the repo root `README.md` shows this header as `Grpc-Metadata-Velez-Auth`. That's
> stale — the interceptor (`internal/middleware/grpc_incoming_interceptor.go`) checks the metadata
> key `Authorization`, not `Velez-Auth`. Use `Grpc-Metadata-Authorization`.

For local dev, you can skip auth entirely by setting `disable_api_security: true` in your config
(`config/dev.yaml` already has it off by default like `config.yaml` — flip it explicitly if you
want no-auth).

## 3. Create a service

A "service" is Velez's higher-level unit (name + deployment history); a "smerd" is the underlying
container. Creating a service just registers the name:

```shell
curl -X POST "http://localhost:53890/api/service/create" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <your VelezKey>" \
  -d '{"name": "hello-world"}'
```

(`ServiceApi.CreateService`, `api/grpc/service_api.proto`)

## 4. Create a deployment

`CreateDeploy` takes a `oneof specification`; for a first deployment use `new` with a
`CreateSmerd.Request` body (same shape as the standalone `CreateSmerd` RPC):

```shell
curl -X POST "http://localhost:53890/api/service/deploy/create" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <your VelezKey>" \
  -d '{
    "serviceName": "hello-world",
    "new": {
      "name": "hello-world",
      "imageName": "nginx:latest",
      "ignoreConfig": true,
      "settings": {
        "ports": [
          {"servicePortNumber": 80, "protocol": "tcp", "exposedTo": 18501}
        ]
      },
      "restart": {"type": "unless_stopped"}
    }
  }'
```

Field notes, grounded in `api/grpc/velez_api.proto` / `velez_common.proto`:

- `ignoreConfig: true` skips Matreshka config assembly — needed here since Matreshka is disabled
  by default.
- `settings.ports[].exposedTo` must land inside `available_ports` (`18501-19000` by default) or be
  omitted to let Velez pick a free port from that range.
- `useImagePorts: true` is an alternative to `settings.ports` — it publishes whatever the image
  itself exposes.
- Instead of `new`, you can pass `upgrade: {"deploymentId": ..., "image": "..."}` to roll a new
  image onto an existing deployment's spec.

**This call returns immediately, before the container exists.** `CreateDeploy` only writes a
`SCHEDULED_DEPLOYMENT` row (`internal/service/service_manager/verv_services/deploy.go`). A
background reconciler, `deployWatcher` (`internal/workers/deploy_watcher.go`), polls every 5
seconds, picks up scheduled deployments, and actually launches the container via the
`LaunchSmerd` pipeline. Expect a few seconds of lag between `CreateDeploy` returning and the
deployment showing `RUNNING`.

(If you want a synchronous create-and-wait call instead, use `VelezAPI.CreateSmerd`
(`POST /api/smerd/create`) directly — it enqueues a durable job and blocks until the container is
up, returning the full `Smerd` object. `CreateDeploy` is the service/deployment-history-aware path;
`CreateSmerd` is the low-level one.)

## 5. Verify it's running

Poll the deployment list until status flips to `RUNNING`:

```shell
curl -X POST "http://localhost:53890/api/service/deploy/list" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <your VelezKey>" \
  -d '{"serviceName": "hello-world"}'
```

`DeploymentStatus` values (`api/grpc/service_api.proto`): `SCHEDULED_DEPLOYMENT` → `RUNNING` (or
`FAILED`). Once running:

```shell
# Service-level view (name, current deployment id, status)
curl -X POST "http://localhost:53890/api/service/get" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <your VelezKey>" \
  -d '{"name": "hello-world"}'

# All services on this node
curl -X POST "http://localhost:53890/api/service/list" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <your VelezKey>" \
  -d '{}'

# Container-level view (ports, status, labels)
curl -X POST "http://localhost:53890/api/smerd/list" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <your VelezKey>" \
  -d '{"name": "hello-world"}'
```

Or just check Docker directly — a "smerd" is a plain container:

```shell
docker ps --filter "name=hello-world"
```

## Troubleshooting / gotchas

- **`CreateDeploy` "succeeds" but nothing ever runs.** Check the deployment status via
  `/api/service/deploy/list` — if it's stuck at `SCHEDULED_DEPLOYMENT`, the `deployWatcher` isn't
  making progress. Check the requested `settings.ports[].exposedTo` isn't already taken, and that
  `imageName` is pullable from wherever Docker is configured to pull from.
- **401/403 with a correct-looking key.** Metadata key is `Authorization`, header is
  `Grpc-Metadata-Authorization` (see step 2) — `Grpc-Metadata-Velez-Auth` from the root README does
  not work against the current interceptor.
- **Losing the access key on every restart.** If you don't mount `local_state_path`
  (`/tmp/velez/local_state.json` by default) to a host volume, Velez regenerates `VelezKey` on
  every restart and all previously-issued keys stop working.
- **`cluster_pg_dsn` in `config/config.yaml` does nothing.** It's declared in the config schema
  (`internal/config/environment.go`) but nothing in the codebase reads
  `cfg.Environment.ClusterPgDsn` — it isn't wired to anything. Postgres-backed cluster state is
  enabled a different way (see `docs/quickstart/multi-node.md`), not via this key.
- **Ports:** `available_ports` (`18501-19000` in prod config, `18501-18519` in `config/dev.yaml`)
  bounds what Velez will hand out via `exposedTo`/`useImagePorts`. Requesting a port outside that
  range, or one already bound, will fail the deployment.
