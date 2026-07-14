# Pipelines → Jobs Migration

Goal: move business logic off the old in-memory, non-resumable `internal/pipelines`
runner onto the durable, checkpointed task engine in `internal/jobs` (see
`internal/jobs/job.go` for the core interfaces). This doc tracks what's moved,
what's left, and the recipe to move the next one — so each migration session can
start here instead of re-deriving the pattern from scratch.

## Status

| Pipeliner method     | Job action       | Status         | Handler                              | Steps/Jobs | Live RPC cut over? |
|-----------------------|------------------|----------------|---------------------------------------|------------|---------------------|
| `LaunchSmerd`          | `create_smerd`   | Scaffolded     | `internal/jobs/create_smerd.go`       | 4          | No — `velez_api_impl.CreateSmerd` still calls the old pipeliner |
| `CreateService`        | `create_service` | Scaffolded     | `internal/jobs/create_service.go`     | 2          | No — `service_api_impl.CreateService` still calls the old pipeliner |
| `AssembleConfig`       | —                | Not started    | —                                      | 4 + result post-processing | — |
| `CopyToVolume`         | —                | Not started    | —                                      | variable (loops over input files) | — |
| `ConnectServiceToVpn`  | —                | Not started    | —                                      | 9          | — |
| `UpgradeSmerd`         | —                | Not started    | —                                      | ~18, multi-stage | — |
| `EnableStatefullMode`  | —                | Not started    | —                                      | ~7, bootstraps Postgres + mutates cluster state | — |

**"Scaffolded" ≠ "cut over."** A scaffolded pipeline has a working, tested
`TaskHandler` registered in the worker (`internal/app/custom.go`), but the
gRPC handler that clients actually call still runs the old synchronous
`pipeliner`. Cutting a live RPC over to `Engine.Enqueue`/`Watch` changes its
contract from an immediate response to enqueue-then-poll — that's a separate,
bigger decision to make explicitly per pipeline, not a default next step.

## Suggested pick-up order (easiest first)

1. **`AssembleConfig`** — same shape as `CreateService`/`create_smerd`: a short,
   fixed step list. The one wrinkle is the `getResult` post-processing (yaml/evon
   parsing) — the job version needs somewhere to stash that logic since jobs
   don't have a `Runner.Result()` equivalent (their output is the persisted
   `TaskContext`, read back by the caller after the task is `DONE`).
2. **`CopyToVolume`** — logic is simple (create loader container, copy files,
   drop container) but the step count is dynamic (one create/copy pair per
   mounted folder). `BuildJobs` can still build this as a loop, same as the
   pipeline does — no engine changes needed, just needs care in naming each
   job uniquely for checkpointing (e.g. `copy_file_<n>`).
3. **`ConnectServiceToVpn`** — 9 steps, moderate complexity, several steps share
   state via closures (`clientKey`, `loginServer`, `namespaceId`) which map
   naturally to `TaskContext` fields.
4. **`EnableStatefullMode`** — higher risk: bootstraps Postgres, swaps out the
   live `ClusterStateManager`/`StorageContainer` mid-pipeline. Migrate only
   after the pattern is well-proven on simpler cases.
5. **`UpgradeSmerd`** — most complex: ~18 steps, multiple in-place renames of
   the same container, several stages that reuse and mutate one shared
   `newLaunch` value. Do this last.

## Migration checklist (repeat per pipeline)

1. **Proto payload** — add a `{Name}TaskPayload` message to `api/grpc/tasks.proto`
   carrying only the business data jobs hand off to each other (mirror
   `CreateServiceTaskPayload`/`CreateSmerdTaskPayload`). Run `moti g` to
   regenerate. If a field needs to be mutated by a later job and isn't a plain
   proto3 scalar, add a hand-written `Set*` in
   `internal/api/server/velez_api/tasks_setters.go` (protoc-gen-go only emits
   getters).
2. **Handler + jobs** — new file `internal/jobs/{name}.go`:
   - `const {Name}Action = "{name}"`
   - narrow accessor interfaces (`GetX() T`, `SetX(T)`) instead of depending on
     the concrete payload type directly, so jobs stay reusable
   - a `TaskHandler` (`Action`, `NewContext`, `BuildJobs`)
   - one job type per step. If a step's logic is pure and has no pipeline-only
     dependency, it can often be reused as-is: `steps.Step` and `jobs.Job` are
     both just `Do(ctx context.Context) error`, so a `steps.XxxStep(...)` value
     satisfies `jobs.Job` with no wrapper needed (see `create_service.go`'s
     reuse of `service_steps.ValidateServiceName`). Steps with rollback need a
     `Rollback(ctx) error` method to satisfy `RollbackableJob`.
3. **Register** — add `registry.Register(jobs.NewXxxHandler(...))` next to the
   existing registrations in `internal/app/custom.go`.
4. **Tests** — `internal/jobs/{name}_test.go`. Reuse the fakes in
   `internal/jobs/fakes_test.go` (`fakeTasksStorage`, `fakeJobsStorage`); add a
   new fake there if the job needs a storage interface not yet faked. Cover at
   least: happy path end-to-end through `taskWorker.run`, and one failure path.
5. **Verify** — `go build ./...` and `go test ./...` clean. Run
   `graphify update .` so the code graph stays current.
6. **Update this table** — flip the row to Scaffolded and note the job count.
7. **Cut-over is a separate, explicit decision** — ask before wiring a live RPC
   to `Engine.Enqueue`/`Watch`; don't do it as part of the scaffolding step.
