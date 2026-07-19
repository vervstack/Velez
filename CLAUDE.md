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
# TypeScript library
cd pkg/web/@vervstack/velez && npm run build

# React UI
cd pkg/web/Velez-UI && bun install && bun run build
```

### Database Migrations

Migrations live in `migrations/` and use `pressly/goose`. SQL queries are compiled via `sqlc` into
`internal/storage/postgres/generated/`.

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

Durable, checkpointed replacement for `internal/pipelines`: a `TaskHandler` per pipeline
(`Action()`, `NewContext()`, `BuildJobs()`) builds an ordered list of `NamedJob`s, each wrapped in a
`checkpointedJob` (DONE-skip, FAILED-short-circuit, resume-safe) and persisted as JSON to
`velez.tasks`/`velez.jobs` (Postgres in production, `local_storage` in single-node/dev — same
interface). `internal/app/custom.go` registers every `TaskHandler`; a single `taskWorker` polls and
claims work via `SELECT ... FOR UPDATE SKIP LOCKED`. See `docs/jobs_migration.md` for
per-pipeline migration status/checklist and `docs/plans/testing.md` for the live-RPC cutover plan.

**Hard-won rules — read before touching a job scaffold or cutting one over to a live RPC:**

1. **Proto `oneof` fields silently break `encoding/json` round-trips.** A oneof (e.g.
   `CreateSmerd_Request.Config`) is a Go interface field. `encoding/json.Marshal` happily
   serializes whatever concrete type it holds, but `Unmarshal` can *never* allocate a concrete type
   back into an interface field — no error until the field is actually populated, and no linter
   catches it (Go's type system can't express the constraint). `internal/jobs` persists
   `TaskContext` via plain `encoding/json`, not `protojson`, so any payload that embeds a
   oneof-bearing message (directly, like `CreateSmerdTaskPayload.request`, or via a job that later
   sets the field) is at risk. **Before wiring a job whose payload embeds a message with a `oneof`
   field**, write a throwaway round-trip test (marshal → unmarshal → assert the oneof survived). If
   it fails, add hand-written `MarshalJSON`/`UnmarshalJSON` — pattern in
   `internal/api/server/velez_api/create_smerd_request_json.go`: shadow the promoted oneof field
   with a same-named, shallower dummy field so `encoding/json` never touches the real interface
   field, and flatten each oneof variant into its own explicit, tagged field instead. This is a
   *codebase* rule, not a create_smerd-specific one — `UpgradeSmerdTaskPayload` embeds the same
   `CreateSmerd.Request` type and sidesteps the bug today only by coincidence (its job never
   populates `.Config`).
2. **A claimed task must always reach `DONE`/`FAILED`.** Every code path in
   `taskWorker.run()` that returns before `runJobs` completes (no handler registered, context
   unmarshal failure, etc.) must go through `failTask()`, which calls `FinishTask`. A bare
   `return err` leaves the task stuck at `RUNNING` forever — reclaimable only after the 2-minute
   stale lease, and it will fail identically on every retry since nothing changed. This bit us once
   (worker.go's unmarshal-error path used to `return` directly) and is now fixed, but it's a trap
   for any new early-return added to `run()`.
3. **A "Scaffolded" job is not proven equivalent to the pipeline it replaces.** Nothing
   checks that a `TaskHandler.BuildJobs()` step count/coverage matches the original pipeline's step
   list. `create_smerd.go`'s scaffold (from the first jobs-engine commit) silently implemented only
   4 of `do_smerd_launch.go`'s 9 steps — missing config-assembly and `use_image_ports` entirely —
   and this went unnoticed until cutover, because the old pipeliner kept serving the live RPC the
   whole time the scaffold sat unused. Before cutting any pipeline over: read the pipeline's
   `Steps: []steps.Step{...}` list side-by-side with the job's `BuildJobs()` return, and account for
   every step (ported, deliberately folded per an existing precedent in
   `docs/jobs_migrations/questions.md`, or explicitly out of scope) — not just "the tests I thought
   to run passed."
4. **A hung parallel e2e run and a slow one look identical from the outside.** If
   `Test_Lifecycle` (or similar) hangs/times out only when run as a full parallel suite but passes
   when run alone, don't assume either "it's just slow" or "it's the same bug as before" —
   isolate the specific failing subtest, rerun it alone with a short timeout, and if it still hangs,
   force a live goroutine dump (`kill -QUIT <pid>` a few seconds in) or add temporary
   `println`-based tracing at each `checkpointedJob.Do()` call to see exactly which job step it's
   stuck in before changing any code or timeouts.

### Configuration

- `config/config.yaml` — production config
- `config/dev.yaml` — local dev overrides
- Environment variables use `VERV_NAME` prefix; parsed by the `matreshka` library

### Frontend (`pkg/web/`)

- `@vervstack/velez/` — TypeScript client library (compiled to `dist/`)
- `Velez-UI/` — React 18 + Vite application (Zustand state, React Query data fetching)

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
