# Pipelines → Jobs Migration

Goal: move business logic off the old in-memory, non-resumable `internal/pipelines`
runner onto the durable, checkpointed task engine in `internal/jobs` (see
`internal/jobs/job.go` for the core interfaces). This doc tracks what's moved,
what's left, and the recipe to move the next one — so each migration session can
start here instead of re-deriving the pattern from scratch.

## Status

| Pipeliner method     | Job action       | Status         | Handler                              | Steps/Jobs | Live RPC cut over? |
|-----------------------|------------------|----------------|---------------------------------------|------------|---------------------|
| `LaunchSmerd`          | `create_smerd`   | Cut over     | `internal/jobs/create_smerd.go`       | 9          | Yes — `velez_api_impl.CreateSmerd` now calls `Engine.Enqueue`+`Watch` (sync facade); scaffold grew from 4 to 9 jobs to reach parity with `do_smerd_launch.go` (prepare_request, fetch_config, prepare_verv_config, copy_to_container, subscribe_for_config_changes added) |
| `CreateService`        | `create_service` | Cut over     | `internal/jobs/create_service.go`     | 2          | Yes — `service_api_impl.CreateService` now calls `Engine.Enqueue`+`Watch` (sync facade); step parity was already 2-for-2 with `do_create_service.go`, no job changes needed |
| `AssembleConfig`       | `assemble_config` | Cut over     | `internal/jobs/assemble_config.go`    | 5          | Yes — `velez_api_impl.AssembleConfig` now calls `Engine.Enqueue`+`Watch` (sync facade); step parity was already 5-for-5 with `do_assemble_config.go` (4 pipeline steps + `getResult` folded into `parse_config`, precedent already noted below), no oneof fields in `AssembleConfigTaskPayload` so the round-trip test was skipped. The pipeline's post-`Result()` `cfgService.UpdateConfig(ctx, *cfg)` side effect (not modeled by any job) stays in the RPC handler, called after the `Watch` loop with a `domain.AppConfig` reconstructed from the final task's payload: `Meta.Name`/`Version`/`Format` map directly, `Meta.ConfType` converts the payload's plain string (`"verv"`/`"pg"`/`"plain"`) to `matreshka_api.ConfigTypePrefix` via `ConfigTypePrefix_value`, and `ContentRaw` comes from `content_raw`. `Meta.Content` (`*evon.Node`) is left nil rather than reconstructed from the payload's `content` field — the old pipeline's own `evon.Unmarshal(bytes, &appConfig.Content)` for the `ConfigFormat_env` case (the only format AssembleConfig ever produces) is a silent no-op (verified empirically: `reflect.Value.Elem()` on a nil `*evon.Node` yields an invalid `Value`, so no field mapping is built), so `Content` was already always nil in production before this cutover; reproduced as-is rather than fixed. `domain.AssembleConfig` (the pipeliner-only request struct) was removed alongside `do_assemble_config.go` since nothing else referenced it. |
| `CopyToVolume`         | `copy_to_volume` | Scaffolded     | `internal/jobs/copy_to_volume.go`     | 3 + N (dynamic) | No — CopyToVolume has no live caller today (see docs/jobs_migrations/questions.md #1) |
| `ConnectServiceToVpn`  | `connect_service_to_vpn` | Cut over     | `internal/jobs/connect_service_to_vpn.go` | 8 (9 pipeline steps, 1 folded — see questions.md) | Yes — `vcn_api_impl.ConnectService` now calls `Engine.Enqueue`+`Watch` (sync facade); step parity re-verified at cutover (8-for-8, matches scaffold). No oneof fields in `ConnectServiceToVpnTaskPayload`, so the round-trip test was skipped. **Removal shape differs from `CreateService`/`AssembleConfig`**: `do_connect_service_to_vpn.go` defines both the `Pipeliner.ConnectServiceToVpn` interface method/adapter (RPC-only, now dead, removed) *and* a package-level free function `ConnectServiceToVpn(req, nc, vpnClient, sdClient)` that two unrelated app-bootstrap call sites (`internal/cluster/configuration/service.go`, `internal/cluster/service_discovery/launch.go`) call directly, bypassing the `Pipeliner` interface, to connect the node's own matreshka/service-discovery sidecar to the VPN at startup. Those callers check `rerrors.Is` against the typed error the `Runner` returns — routing them through the jobs engine would flatten that into a string (`finalTask.Error.String`) and add persistence/checkpointing to synchronous in-process bootstrap logic, out of scope for this migration. Only the interface method and the `(p *pipeliner)` adapter were removed; the free function and both bootstrap callers are untouched and still compile/run exactly as before. |
| `EnableStatefullMode`  | `enable_statefull_mode` | Cut over     | `internal/jobs/enable_statefull.go` | 8 (7 pipeline steps + 1 new — see questions.md) | Yes — `control_plane_api_impl.EnablePlugin`'s `statefull_pg` case now calls `Engine.Enqueue`+`Watch` (sync facade); step parity re-verified at cutover (8-for-8, matches scaffold). `EnableStatefullTaskPayload.request` (`velez_api.EnableStatefullCluster`) is a plain non-oneof message (`optional bool`/`optional uint64`), unlike `CreateSmerd.Request`, so the round-trip test was skipped — confirmed by inspecting `api/grpc/control_plane_api.proto`. `EnablePlugin_Response` is always empty and the old pipeline's `getResult` was never read by the RPC handler, so no response reconstruction was needed (simpler than `AssembleConfig`, same shape as `CreateService`/`ConnectServiceToVpn`). The RPC's pre-existing quirk of returning a non-nil `&EnablePlugin_Response{}` even on error was preserved exactly. `domain.EnableStatefullClusterRequest`/`domain.StateClusterDefinition` (pipeliner-only request/result types) were removed alongside `do_enable_statefull.go` since nothing else referenced them. |
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

Every pipeliner method in the status table above is now scaffolded.
`LaunchSmerd`/`create_smerd`, `CreateService`/`create_service`,
`AssembleConfig`/`assemble_config`, `ConnectServiceToVpn`/
`connect_service_to_vpn`, and `EnableStatefullMode`/`enable_statefull_mode`
have been cut over to serve their live gRPC endpoints from the jobs engine;
the remaining two (`CopyToVolume`, `UpgradeSmerd`) still run the old
pipeliner - see the cross-cutting note below.

## Migration checklist (repeat per pipeline)

1. **Proto payload** — add a `{Name}TaskPayload` message to `api/grpc/tasks.proto`
   carrying only the business data jobs hand off to each other (mirror
   `CreateServiceTaskPayload`/`CreateSmerdTaskPayload`). Run `moti g` to
   regenerate. If a field needs to be mutated by a later job and isn't a plain
   proto3 scalar, add a hand-written `Set*` in
   `internal/api/server/velez_api/tasks_setters.go` (protoc-gen-go only emits
   getters).
   - **oneof check.** If the payload embeds a message with a `oneof` field
     (e.g. `CreateSmerd.Request.config`), write a throwaway test that
     marshals a populated instance via `encoding/json` and unmarshals it back
     into a fresh value, then asserts the oneof survived. A oneof is a Go
     interface field — `Marshal` serializes it fine, but `Unmarshal` can never
     allocate a concrete type back into an interface field, so this silently
     breaks with no compile-time warning. If the test fails, add hand-written
     `MarshalJSON`/`UnmarshalJSON` before writing any job code — pattern in
     `internal/api/server/velez_api/create_smerd_request_json.go`. Skipping
     this check is exactly what caused a multi-hour hang-debugging session
     during `CreateSmerd`'s cutover (see `docs/plans/testing.md`'s progress
     log, 2026-07-19).
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
   - **step-parity check.** List the original pipeline's `Steps: []steps.Step{...}`
     (in its `do_*.go` file) next to `BuildJobs()`'s returned list, one by one.
     Every original step must be accounted for: ported as a job, deliberately
     folded into an adjacent job (state the precedent, e.g. `docs/jobs_migrations/questions.md`'s
     mkdir+copy fold), or explicitly out of scope with a stated reason. An
     unexplained gap is a bug, not a simplification — `create_smerd.go`'s
     original scaffold silently dropped 5 of 9 steps (config-assembly,
     `use_image_ports`) this way, and it wasn't caught until cutover broke
     5 of 9 e2e subtests.
3. **Register** — add `registry.Register(jobs.NewXxxHandler(...))` next to the
   existing registrations in `internal/app/custom.go`.
4. **Tests** — `internal/jobs/{name}_test.go`. Reuse the fakes in
   `internal/jobs/fakes_test.go` (`fakeTasksStorage`, `fakeJobsStorage`); add a
   new fake there if the job needs a storage interface not yet faked. Cover at
   least: happy path end-to-end through `taskWorker.run`, and one failure path.
5. **Verify** — `go build ./...` and `go test ./...` clean. Run
   `graphify update .` so the code graph stays current.
6. **Update this table** — flip the row to Scaffolded and note the job count
   *and* the original pipeline's step count (e.g. "9 (4 originally, corrected)")
   so a future cutover session can see at a glance whether parity was checked.
7. **Cut-over is a separate, explicit decision** — ask before wiring a live RPC
   to `Engine.Enqueue`/`Watch`; don't do it as part of the scaffolding step.
   Before cutting over, re-run the step-parity and oneof checks above even if
   the pipeline was scaffolded long ago — a scaffold's own tests only cover
   what its (possibly incomplete) job list does, not what the original
   pipeline did.
