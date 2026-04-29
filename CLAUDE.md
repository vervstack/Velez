# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
cd pkg/web/Velez-UI && yarn install && yarn build
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

### Configuration

- `config/config.yaml` — production config
- `config/dev.yaml` — local dev overrides
- Environment variables use `VERV_NAME` prefix; parsed by the `matreshka` library

### Frontend (`pkg/web/`)

- `@vervstack/velez/` — TypeScript client library (compiled to `dist/`)
- `Velez-UI/` — React 18 + Vite application (Zustand state, React Query data fetching)

## Key Dependencies

- **Docker**: `github.com/docker/docker`
- **gRPC + REST**: `google.golang.org/grpc`, `grpc-ecosystem/grpc-gateway`
- **Database**: `lib/pq` (Postgres), `sqlc` (query gen), `pressly/goose` (migrations)
- **Testing**: `stretchr/testify`, `gojuno/minimock`
- **Logging**: `rs/zerolog`
- **Internal framework**: `go.redsock.ru/*` utilities