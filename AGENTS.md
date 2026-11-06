# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Agent code of conduct

- do not make architectural decisions on your own, always suggest and ask before doing
- never break backward compatibility
- before handing out results run tests and validate that code changes didn't break things

## Project Overview

Velez is a lightweight node manager for the Vervstack ecosystem. It starts/stops services on machines using Docker,
acting as an alternative to manual docker-compose or Kubernetes. It integrates with:

- **matreshka** — external configuration service
- **makosh/headscale** — VPN/network management
- **rscli** — Go app building toolchain

The API is available at `<ip>:53890/api` (configurable).

## Commands

### Code Generation

```bash
make warmup    # Fetch proto dependencies
make codegen   # Generate Go + TypeScript from proto files (runs protopack + npm build)
```

### Linting

```bash
make lint      # golangci-lint (see .golangci.yaml for rules)
```

### Building

```bash
make build-local-container   # Build ARM64 Docker image tagged as velez:local
go build -o ./service ./cmd/service/main.go
```

### Frontend

```bash
# React UI
cd pkg/web/Velez-UI && bun install && bun run build
```

### Database Migrations

Migrations live in `migrations/` and use `pressly/goose`. SQL queries are compiled via `sqlc` into
`internal/storage/postgres/generated/`.

## Releases and E2E

Every master push produces an rc release; the minor release workflow gates on the full e2e suite.
Run e2e locally by group while iterating: `make e2e-group GROUP=<name>` (never the whole suite).
Full detail: `docs/ci_and_e2e.md`.

## Architecture

### Request Flow

```
gRPC/HTTP → transport/manager.go (cmux mux) → API impl → service layer → pipelines → docker/db
```

A single TCP port is multiplexed into gRPC and HTTP (grpc-gateway) using `cmux`.

### Layer Breakdown

| Layer     | Path                         | Responsibility                                      |
|-----------|------------------------------|-----------------------------------------------------|
| Entry     | `cmd/service/`               | Creates and starts App                              |
| App       | `internal/app/`              | Wires config, clients, services, and server         |
| Transport | `internal/transport/`        | gRPC + HTTP servers, API implementations            |
| Service   | `internal/service/`          | Business logic interfaces and implementations       |
| Pipelines | `internal/pipelines/`        | Multi-step orchestration (launch, upgrade, create)  |
| Jobs      | `internal/jobs/`             | Durable, resumable task/job engine replacing pipelines (see below) |
| Clients   | `internal/clients/`          | Docker, hardware, Matreshka, Makosh, Headscale      |
| Storage   | `internal/storage/postgres/` | PostgreSQL via sqlc-generated queries               |
| Domain    | `internal/domain/`           | Core data types (Service, Deployment, Volume, etc.) |

### Domain layer override

`internal/domain/` may reference generated proto (`velez_api`, `matreshka_api`) and sqlc types
directly — it does not follow the strict domain-purity rule. The full override text lives in
`docs/dev-profile-coverage.md` ("Project-level override"); keep both in sync.

### API (Proto definitions in `api/grpc/`)

- `velez_api.proto` — Container CRUD (CreateSmerd, ListSmerds, DropSmerd)
- `service_api.proto` — Service lifecycle (CreateService, CreateDeploy, ListDeployments)
- `control_plane_api.proto` — Cluster control
- `verv_closed_network.proto` — VPN operations
- Generated Go code lands in `internal/api/server/api/grpc/`

### Pipelines (`internal/pipelines/`)

Orchestrate multi-step Docker and cluster operations. Steps are organized into:

- `steps/smerd_steps/` — container create/start/drop
- `steps/service_steps/` — service validation/setup
- `steps/config_steps/` — config fetch/store from Matreshka
- `steps/network_steps/` — VPN setup
- `steps/cluster_steps/` — DB user creation, DSN setup

Key pipeline files: `do_smerd_launch.go`, `do_smerd_upgrade.go`, `do_create_service.go`.

### Jobs Engine (`internal/jobs/`)

Durable, checkpointed replacement for `internal/pipelines`: `TaskContext`s persisted as JSON to
`velez.tasks`/`velez.jobs` (Postgres in production, `local_storage` in single-node/dev), claimed
by a single `taskWorker` via `SELECT ... FOR UPDATE SKIP LOCKED`.

**Before touching a job scaffold or cutting one over to a live RPC, read the hard-won rules in
`docs/jobs_engine.md`** — proto `oneof` JSON round-trips, tasks that must always reach
`DONE`/`FAILED`, scaffold-vs-pipeline step equivalence, hung parallel e2e runs.

### E2E shared fixtures

`WithMatreshka()` (tests/e2e) is a single-process singleton — keep every caller in package
`tests/e2e`. Rule text: `docs/plans/e2e_flaky_lifecycle_matreshka.md`.

### Configuration

- `config/config.yaml` — production config
- `config/dev.yaml` — local dev overrides
- Environment variables use `VERV_NAME` prefix; parsed by the `matreshka` library

### Frontend (`pkg/web/`)

- `Velez-UI/` — React 18 + Vite application (Zustand state, React Query data fetching); the
  generated API client lives inside it, not as a separate published package.

## Velez is stateless — hard rule

**Never store anything long-term in Velez's process memory or on Velez's own disk.** A Velez restart
must lose nothing a user can see or use.

- **Single-node / dev mode:** state lives ON the Docker objects Velez manages — container env, labels,
  mounts, inspect data — and is read back from them. A secret Velez generates (admin token, password,
  registration token) is written to the container it belongs to at creation time, and `Get` reads it
  from that container. If the value is held by an external system that can return it (e.g. Garage's
  `showSecretKey`), ask that system instead of caching it.
- **Cluster mode:** Postgres is the only durable store.
- **In-memory is allowed only to bridge a short window** (the gap before a container exists, as
  `pendingPgaasPw` does), never as the source of truth. A new `secrets.NewStatic()` fallback, a
  `sync.Map`/map-with-mutex cache, or an on-disk JSON file holding user-visible state is a violation —
  derive it from the container instead.
- **Exceptions, both deliberate:**
  - In-flight operations in single-node mode (the task/job queue, transfer tokens and history, deployment
    history made by this process) live in memory only. A restart mid-operation can leave a half-created container
    with no task record; cluster mode is unaffected because tasks live in Postgres.
  - Velez's own identity (`local_state.Manager`: Velez/Matreshka/Headscale keys, cluster DSNs, sysbox settings)
    is stored in a JSON file at `local_state_path`. It describes Velez itself, not a managed service, and the path
    must be on a mounted volume.
- Before adding any new secret scope or stored value, answer: "after a Velez restart, where is this
  read back from?" If there is no answer, the design is wrong.

## Known pitfalls

Two incident-grade pitfalls with full writeups — read `docs/known_pitfalls.md` before working on
list/filter paths:

- Never mutate a value through a pointer field a caller owns (request structs are often built as
  `&callerOwnedField` — in-place normalization rewrites the caller's data).
- Docker's `name` filter is an unanchored substring match — always select a service's containers by
  the exact `VERV_SERVICE` label, never by name.

## Code Style

- **Error handling**: never assign and check an error in the same `if` statement. Always use two separate lines:
  ```go
  // correct
  err = doSomething()
  if err != nil { ... }

  // forbidden
  if err = doSomething(); err != nil { ... }
  ```

- **Struct literals**: never construct a struct literal inline inside a function call. Always assign to a named variable
  first:
  ```go
  // correct
  req := &velez_api.CreateSmerd_Request{
      Name:      name,
      ImageName: image,
  }
  result, err := client.CreateSmerd(ctx, req)

  // forbidden
  result, err := client.CreateSmerd(ctx, &velez_api.CreateSmerd_Request{
      Name:      name,
      ImageName: image,
  })
  ```

## Testing

- **No hand-written mocks/fakes — prefer real API calls.** This applies especially to container-runtime/Docker-facing
  code (real Docker daemon, not a fake `client.APIClient`), and also to Postgres (a real, disposable Postgres
  container — see `tests/e2e/suite_enable_statefull_test.go` for the existing pattern — not an in-memory storage
  fake). This is a project-wide, gradual migration away from `internal/jobs/fakes_test.go`'s hand-written fakes, not
  a one-shot rewrite: convert a fake's consumers to real API calls as you touch that code, don't leave new fakes
  behind, and don't feel obligated to convert everything at once. Prior fake-based tests are not force-reverted;
  they get replaced incrementally as their area of the code is worked on.
  - Real production constructors are usually enough on their own — `container_runtime.NewResolver`,
    `environments.NewStatic` (the real local_storage `EnvironmentsStorage` backend), and
    `container_manager.New` (the real `service.ContainerService`) all just need a real Docker client
    (`docker.NewClient`), no fake required.
  - Not in scope for this rule: pure in-process control-flow test doubles with no external dependency at all
    (e.g. a job-engine test double that just records call order), and anything that isn't really "calling an
    API" to begin with (e.g. `local_state.Manager`'s on-disk JSON persistence).
  - Expect these tests to run slower (real Docker/Postgres, not in-memory) and to need care around parallelism —
    default to `t.Parallel()` with a unique name/suffix per test (mirror `tests/e2e`'s `GetServiceName`/container
    suffix conventions) for every Docker-global-namespace resource a test touches; if a specific test flakes/
    collides under parallel execution, drop `t.Parallel()` for just that test, not the whole file.

## Key Dependencies

- **Docker**: `github.com/docker/docker`
- **gRPC + REST**: `google.golang.org/grpc`, `grpc-ecosystem/grpc-gateway`
- **Database**: `lib/pq` (Postgres), `sqlc` (query gen), `pressly/goose` (migrations)
- **Testing**: `stretchr/testify`, `gojuno/minimock`
- **Logging**: `rs/zerolog`
- **Internal framework**: `go.redsock.ru/*` utilities

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
