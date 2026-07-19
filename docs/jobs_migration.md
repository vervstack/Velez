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
| `AssembleConfig`       | `assemble_config` | Scaffolded    | `internal/jobs/assemble_config.go`    | 5          | No — `velez_api_impl.AssembleConfig` still calls the old pipeliner |
| `CopyToVolume`         | `copy_to_volume` | Scaffolded     | `internal/jobs/copy_to_volume.go`     | 3 + N (dynamic) | No — CopyToVolume has no live caller today (see docs/jobs_migrations/questions.md #1) |
| `ConnectServiceToVpn`  | `connect_service_to_vpn` | Scaffolded | `internal/jobs/connect_service_to_vpn.go` | 8 (9 pipeline steps, 1 folded — see questions.md) | No — `service_api_impl`/`velez_api_impl` still call the old pipeliner for VPN connect |
| `EnableStatefullMode`  | `enable_statefull_mode` | Scaffolded | `internal/jobs/enable_statefull.go` | 8 (7 pipeline steps + 1 new — see questions.md) | No — `control_plane_api_impl.EnablePlugin` still calls the old pipeliner |
| `UpgradeSmerd`         | `upgrade_smerd`  | Scaffolded     | `internal/jobs/upgrade_smerd.go`      | 15 (19 pipeline steps, 4 SingleFunc renames folded — see questions.md) | No — `velez_api_impl.UpgradeSmerd` still calls the old pipeliner |

**"Scaffolded" ≠ "cut over."** A scaffolded pipeline has a working, tested
`TaskHandler` registered in the worker (`internal/app/custom.go`), but the
gRPC handler that clients actually call still runs the old synchronous
`pipeliner`. Cutting a live RPC over to `Engine.Enqueue`/`Watch` changes its
contract from an immediate response to enqueue-then-poll — that's a separate,
bigger decision to make explicitly per pipeline, not a default next step.

## Suggested pick-up order (easiest first)

1. ~~**`AssembleConfig`**~~ — done. Same shape as `CreateService`/`create_smerd`:
   a short, fixed step list. The one wrinkle was the `getResult` post-processing
   (yaml/evon parsing) — the job version stashes that logic in a `parse_config`
   job since jobs don't have a `Runner.Result()` equivalent (their output is
   the persisted `TaskContext`, read back by the caller after the task is `DONE`).
2. ~~**`CopyToVolume`**~~ — done. Logic is simple (create loader container, copy
   files, drop container) but the step count is dynamic (one create/copy pair
   per mounted folder). `BuildJobs` builds this as a loop — no engine changes
   needed, just care in naming each job uniquely for checkpointing
   (`copy_file_<n>`).
3. ~~**`ConnectServiceToVpn`**~~ — done. 9 pipeline steps became 8 named jobs
   (the inline env-append closure was folded into `create_container`, same
   precedent as `CopyToVolume`'s combined mkdir+copy). `clientKey`,
   `loginServer`, `namespaceId` map directly to `TaskContext` fields.
4. ~~**`EnableStatefullMode`**~~ — done. 7 pipeline steps became 8 named jobs:
   the pipeline's inline password-generation setup (computed once per
   pipeline invocation) was promoted to its own `generate_credentials` job so
   a resumed task reuses the same passwords instead of regenerating ones
   that would mismatch an already-created container/user — same idempotency
   precedent as `ConnectServiceToVpn`'s `client_key`. The live
   `ClusterStateManager`/`StorageContainer` singleton swap was carried over
   as a job side effect unchanged. See questions.md for the SQL-testability
   gap this migration left open.
5. ~~**`UpgradeSmerd`**~~ — done, last as planned. 19 pipeline steps became 15
   named jobs: 4 pure-rename SingleFunc steps were folded into whichever
   real job immediately follows them, same fold precedent as
   ConnectServiceToVpn/CopyToVolume. The pipeline's single mutable
   `newLaunch`/`newContId` variables (reused across the scratch
   config-fetcher container and the final container) became a single
   `request`/`container_id` field pair reused the same way, replicating the
   original's rollback quirk on purpose rather than "fixing" it away. A
   likely-dead pipeline stage (the scratch container's config extraction,
   whose result is discarded and never read again) was preserved as-is
   rather than removed. See questions.md #18-21 for details and the
   `pauseAPI`/`renameAPI`/`copyFromAPI`/`createNetworkAPI` narrow-interface
   duplication this needed to stay unit-testable.

Every pipeliner method in the status table above is now scaffolded. No
pipeline has been cut over to serve its live gRPC endpoint from the jobs
engine - see the cross-cutting note below.

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
