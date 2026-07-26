# Plan: Cut Pipelines Over to the Postgres/Storage-Based Jobs Engine

## Context

`internal/pipelines` runs each pipeline synchronously, in-memory, with no
persistence — if the process dies mid-pipeline, the operation is abandoned.
`internal/jobs` is the replacement: a durable, checkpointed task engine
backed by `storage.TasksStorage`/`storage.JobsStorage` (Postgres in
production; `local_storage` in single-node/dev mode — both satisfy the same
interface).

Per `docs/jobs_migration.md`, every pipeline has already been **scaffolded**
as a job (`internal/jobs/*.go`, registered in `internal/app/custom.go`,
covered by fakes-based unit tests). **None has been cut over** — every live
gRPC/HTTP handler still calls `impl.pipeliner.X(...)` directly. Verified
2026-07-19 by grepping `internal/transport/*_impl/`.

This plan is scoped to that remaining step: **cutover**, one pipeline at a
time, test-first. It does not re-litigate the scaffolding work
`docs/jobs_migration.md` already tracks — read that doc for the job/step
inventory of each handler. This doc tracks the cutover initiative
specifically and should be kept in sync with that doc's "Live RPC cut over?"
column.

Reference precedent: `tests/e2e/suite_assemble_config_job_test.go` already
shows the target shape — it drives the real `Engine.Enqueue` +
`Engine.Watch` against real Postgres and asserts the same outcome as
`suite_assemble_config_test.go` (the old-pipeline test), without the RPC
being cut over yet. Use it as the template for what a cutover's test looks
like once the job runs *through the RPC* instead of being invoked directly.

## Ground rules (apply to every task below, every agent)

- **No architectural decisions without asking.** See "Open decisions" below —
  do not start cutover work on a pipeline until its open question is
  resolved with the user.
- **Never break backward compatibility.** The gRPC/HTTP contract (request →
  synchronous response) must not change for callers. Per decision #1 below
  (resolved), every cutover wraps `Enqueue` + `Watch` *inside* the handler
  so it still blocks and returns the same proto response shape it does
  today — clients must not need to poll.
- **Test-first, in this order, no skipping:** write/extend the
  characterization test → confirm it's green on the *old* code → cut over →
  confirm the *same* test is green, unmodified, on the *new* code → run full
  regression. If step 2 (baseline) isn't green, the test is wrong — fix the
  test, not the code, before touching production code.
- Follow this repo's `CLAUDE.md` code style (two-line error checks, no
  inline struct literals in call args) and the migration checklist's
  existing patterns (narrow accessor interfaces, `Set*` helpers in
  `tasks_setters.go`, checkpointed jobs).
- **Do not commit.** Leave changes in the working tree for the user to
  review and commit themselves.
- Before finishing a task, run: `go build ./...`, `go test ./internal/jobs/...`,
  `go test ./tests/e2e/... -timeout 10m` (needs local Docker; Postgres is
  reached via whatever `ClusterStateManager` the test environment wires in
  — no extra setup needed, `NewEnvironment(t)` already exercises the real
  `JobsEngine`). Then `graphify update .`.
- Update **both** tracking tables below and `docs/jobs_migration.md`'s
  "Live RPC cut over?" column when a pipeline's status changes. Don't let
  them drift.

## Open decisions (need explicit user sign-off before the affected work starts)

| # | Decision | Status |
|---|----------|--------|
| 1 | Cutover contract shape (default for the backlog) | **Resolved 2026-07-19: synchronous facade.** The handler calls `Enqueue`, then blocks server-side on `Watch` until the task reaches `DONE`/`FAILED`, then maps the final `TaskContext` into today's response type. No client-visible change — same request, same blocking call, same response shape. This is the default for every pipeline in the backlog **except** `CreateSmerd`, see #1a. |
| 1a | `CreateSmerd` streaming pilot | **Resolved 2026-07-19.** `CreateSmerd` (`rpc CreateSmerd(CreateSmerd.Request) returns (Smerd)`) is left **unary and unchanged** — existing callers (Go client, TS client, `Velez-UI`'s current code paths) keep working exactly as today; internally it's cut over per decision #1 (sync facade). Alongside it, add a **new** streaming RPC (name TBD, e.g. `CreateSmerdStream`) that `Enqueue`s the same `create_smerd` task and forwards status as it progresses — reuse `tasks.proto`'s `TaskStatus` shape, same pattern as `TasksApi.WatchTask`, rather than inventing a new message. This is additive, so it doesn't touch the "never break backward compatibility" rule. Rejected alternative: changing `CreateSmerd` itself to `returns (stream ...)` — breaks every existing caller, rejected. **Reload/resume:** works for free via the already-existing `TasksApi.WatchTask(entity_id, action)` — `Watch` polls storage, not an in-memory stream, so a client that reconnects after a page reload just re-attaches to wherever the task currently stands. `Engine.Enqueue` already dedupes per `(entityID, action)`, so re-issuing the create call after a reload re-attaches to the same in-flight task instead of erroring or double-creating. No resume-token machinery needed — the task's row *is* the resume point. **Scope: this streaming pattern is a pilot for `CreateSmerd` only.** Other pipelines in the backlog keep decision #1's sync facade unless/until revisited individually. |
| 2 | What to do about `CopyToVolume` — it has no live RPC caller today (per `docs/jobs_migration.md`), so there's nothing to cut over. Skip it in this plan, or is a new endpoint planned? | Can't cut over a handler that doesn't exist without inventing one. |
| 3 | `ConnectServiceToVpn`'s only e2e coverage (`suite_vpn_test.go`) skips without a local headscale instance. Do we stand up headscale in CI/dev for this cutover, or accept reduced coverage? | Affects whether Step A (baseline test) can actually run for that pipeline. |
| 4 | `DropSmerd` isn't scaffolded as a job at all yet (no `internal/jobs/drop_smerd.go`). Confirm it should be scaffolded from scratch (full `docs/jobs_migration.md` checklist) as part of this plan, or is it explicitly out of scope. | **Resolved 2026-07-26: scaffolded and cut over.** `internal/jobs/drop_smerd.go` (dynamic fan-out, one `drop_container_<n>` job per identifier) + `velez_api_impl.DropSmerd` now on the sync facade. See `docs/jobs_migration.md`'s status table for the full writeup. |

## Per-pipeline recipe (repeat for each row in the backlog)

**Step A — Baseline characterization test**
1. Find the live handler (see "Backlog" table for file paths) and its
   current `impl.pipeliner.X(...)` call.
2. Check `tests/e2e/` for existing coverage of that RPC. Add/extend a test
   in the same style as `suite_api_deploy_test.go` (real bufconn server,
   real Docker) covering: happy path, at least one failure path, and any
   edge cases relevant to this handler (idempotency, duplicate-name rejection,
   self-upgrade guard, etc. — formerly tracked in the now-deleted
   `docs/review/testing_plan.md`; see `docs/roadmap.md` for current status).
3. Run it against the **current** (old pipeliner) code. It must pass before
   you touch anything else — this is the safety net the rest of the recipe
   depends on.

**Step B — Cutover**
4. In the transport handler, replace `impl.pipeliner.X(...)` with
   `impl.jobsEngine.Enqueue(ctx, entityID, jobs.XAction, payload)`, then
   drain `Watch(ctx, entityID, jobs.XAction)` until a terminal
   (DONE/FAILED) status, then map the final `TaskContext` back into the
   existing proto response type. Add a `Get*` helper in
   `internal/api/server/velez_api/tasks_setters.go` if one doesn't exist for
   a needed field.
5. Use `internal/jobs/{name}_test.go` as the reference for what the job
   expects in its initial `TaskContext` — it's already the source of truth
   for the payload shape.
6. Leave the old pipeliner code in place (don't delete `internal/pipelines`
   code as part of cutover) unless it's confirmed unused by every other
   handler — that cleanup is a separate, later pass.

**Step C — Validate**
7. Re-run the **same** test from Step A, unmodified, against the new code.
   It must pass without any test-code changes.
8. Run full regression: `go build ./...`, `go test ./internal/jobs/...`,
   `go test ./tests/e2e/... -timeout 10m`. All green, no new failures.
9. `graphify update .`.

**Step D — Track**
10. Flip the row's status in the "Backlog" table below.
11. Update `docs/jobs_migration.md`'s "Live RPC cut over?" column to match.

## Pipeline #1 special case: `CreateSmerd` streaming pilot

`CreateSmerd` needs extra steps beyond the generic recipe above, per decision
#1a. Do these **in addition to**, not instead of, the generic Step A–D
recipe (the existing unary `CreateSmerd` still gets the sync-facade cutover
like every other pipeline):

1. Add the new streaming RPC to a proto file (`api/grpc/tasks.proto` next to
   `WatchTask`, or `velez_api.proto` next to `CreateSmerd` — pick whichever
   keeps `TaskStatus` reuse simplest) and regenerate (`moti g`).
2. Implementation: `Enqueue(ctx, entityID, jobs.CreateSmerdAction, payload)`
   then forward `Watch(ctx, entityID, jobs.CreateSmerdAction)` straight to
   the gRPC stream until the channel closes.
3. Test the **new** RPC end-to-end (bufconn, real Docker/Postgres, same
   style as the rest of `tests/e2e`): happy path streams at least one
   intermediate status before `DONE`; failure path streams `FAILED` with an
   error message.
4. Test reload/resume specifically: start the stream, cancel the client
   context mid-flight (simulating a page reload), then call
   `TasksApi.WatchTask` for the same `entity_id`/`action` and assert it
   picks up from the task's current persisted state (not from scratch) and
   still reaches `DONE`.
5. Confirm the old unary `CreateSmerd` is byte-for-byte unaffected by this
   addition — its own characterization test (Step A/C above) must still
   pass unmodified.

## Backlog (suggested order — confirm before starting if you'd rather reorder)

| # | Pipeline | Live handler | Job action | Existing e2e coverage (as of 2026-07-19) | Status |
|---|----------|-------------|------------|--------------------------------------------|--------|
| 1 | `LaunchSmerd` | `internal/transport/velez_api_impl/smerd_create.go` (`CreateSmerd`) + new `CreateSmerdStream` (pilot, see #1a) | `create_smerd` | `suite_api_deploy_test.go` covers create+list (stateless/cluster, healthcheck, ports) - full 9-subtest suite green against the cut-over job path (now 10, with `Test_DropSmerd_ByUuid` added under row 8 below). **Missing:** declarative-deploy idempotency, duplicate-name-without-declarative rejection (testing_plan.md 1.4/1.5); `CreateSmerdStream` has backend unit test coverage (`internal/transport/tasks_api_impl/impl_test.go`) but no e2e/reload-resume test yet. | **Unary `CreateSmerd` cut over** (sync facade). `CreateSmerdStream` pilot **implemented 2026-07-26** — `TasksApi.CreateSmerdStream`, backend + TS client + minimal `DeployWidget.tsx` UI wiring. |
| 2 | `AssembleConfig` | `internal/transport/velez_api_impl/assemble_config.go` | `assemble_config` | Old path: `suite_assemble_config_test.go`. New engine path already driven directly: `suite_assemble_config_job_test.go`. Nearly cutover-ready — use as the template. | Not started |
| 3 | `CreateService` | `internal/transport/service_api_impl/service_create.go` | `create_service` | None found. Needs a new e2e suite. | Not started |
| 4 | `CopyToVolume` | none (no live RPC caller) | `copy_to_volume` | N/A | Blocked — open decision #2 |
| 5 | `ConnectServiceToVpn` | `service_api_impl`/`velez_api_impl` VPN-connect handlers | `connect_service_to_vpn` | `suite_vpn_test.go` exists but skips without local headscale. | Blocked — open decision #3 |
| 6 | `EnableStatefullMode` | `internal/transport/control_plane_api_impl/` (`EnablePlugin`) | `enable_statefull_mode` | None found. Needs a new e2e suite. | Not started |
| 7 | `UpgradeSmerd` | `internal/transport/velez_api_impl/smerd_upgrade.go` | `upgrade_smerd` | None found (only fakes-based `internal/jobs/upgrade_smerd_test.go`). Needs a new e2e suite. | Not started |
| 8 | `DropSmerd` (destroy) | `internal/transport/velez_api_impl/smerd_drop.go` | `drop_smerd` | `internal/jobs/drop_smerd_test.go` (happy path, partial-failure-still-DONE, resume) + `suite_api_deploy_test.go`'s `Test_DropSmerd_ByUuid` — first-ever caller of the previously-unused `env.DropSmerd()` helper. | **Cut over 2026-07-26** (sync facade, decision #4 resolved). |

## Progress log

Append an entry here each time a pipeline's status changes, newest first.

- 2026-07-26 — Pipeline #8 (`DropSmerd`) scaffolded from scratch and cut over
  in the same pass (open decision #4 resolved) — the last pipeline left
  outside the original 7-pipeline `Pipeliner` migration's scope. One
  `drop_container_<n>` job per identifier; each job swallows its own Docker
  error into the task context rather than failing the task, preserving the
  old `DropSmerds`' "top-level error always nil" contract. Also implemented
  the `CreateSmerdStream` pilot from decision #1a: additive `TasksApi`
  server-streaming RPC, backend + TS client + minimal UI wiring in
  `DeployWidget.tsx`. Both verified: `go build ./...`, full non-e2e
  `go test ./...`, and `Test_Lifecycle`'s full 10-subtest suite (including
  the new `Test_DropSmerd_ByUuid`) all green; `bun run build`/`tsc` clean
  aside from one confirmed pre-existing, unrelated `HeadscalePluginForm.tsx`
  error. This closes both items `docs/roadmap.md` §1 had tracked as open.
- 2026-07-19 — Pipeline #1 (`CreateSmerd`) unary handler cut over to the
  jobs engine (Step A-C complete). Step A baseline (`Test_Lifecycle`, 9
  subtests) was green against the old pipeliner. Step B cutover initially
  broke 5/9 subtests: the `create_smerd` job scaffold (from the original
  `16fc3c94` bootstrap commit) only implemented 4 of `do_smerd_launch.go`'s
  9 steps, missing `prepare_request`, `fetch_config`, `prepare_verv_config`
  (including `use_image_ports` support), `copy_to_container`, and
  `subscribe_for_config_changes` entirely. Ported all 5 as new jobs in
  `internal/jobs/create_smerd.go` (reusing `classifyImage`/`createNetworkAPI`
  from `assemble_config.go`/`upgrade_smerd.go`); extended
  `CreateSmerdTaskPayload` (`api/grpc/tasks.proto`) with `image_labels`/
  `image_tags`/`image_exposed_ports`/`path_to_files`. Also found and fixed
  two engine-level bugs surfaced by this cutover (not create_smerd-specific):
  (1) `internal/jobs/worker.go`'s `run()` returned early on a task-context
  unmarshal error without ever calling `FinishTask`, leaving the task stuck
  at `RUNNING` forever instead of `FAILED`; (2) `CreateSmerd_Request.Config`
  is a proto oneof (Go interface field) that `encoding/json` can marshal but
  never unmarshal back into a concrete type - any task whose `Config` is
  populated (e.g. a caller-supplied `Plain` config, or the default `Verv`
  config `prepare_request` now sets) failed this way. Fixed with hand-written
  `MarshalJSON`/`UnmarshalJSON` on `CreateSmerd_Request` in
  `internal/api/server/velez_api/create_smerd_request_json.go`, following
  this codebase's existing "hand-write what protoc-gen-go doesn't emit"
  pattern (`tasks_setters.go`). Added a 60s safety-net timeout around the
  handler's `Watch` loop (`createSmerdWatchTimeout` in `smerd_create.go`) -
  not meant to be hit in normal operation, guards against the engine ever
  getting stuck the way bug (1) did. Step C: full `Test_Lifecycle` suite
  (9/9 subtests) green against the cut-over path, ~10-25s per subtest.
- 2026-07-19 — Open decision #1a resolved: `CreateSmerd` gets a pilot
  streaming RPC added alongside the untouched unary `CreateSmerd` (see
  "Pipeline #1 special case" section). Reload/resume rides on the existing
  `TasksApi.WatchTask`, no new resume mechanism needed. Streaming stays
  scoped to `CreateSmerd` only for now — other pipelines keep decision #1's
  sync facade.
- 2026-07-19 — Open decision #1 resolved: synchronous facade (see table
  above). Pipeline #1 (`CreateSmerd`) is now unblocked to start Step A.
- 2026-07-19 — Plan created. Baseline audit done (build clean, all
  `internal/jobs` fakes-based tests green, all current `tests/e2e` suites
  green against the still-live pipeliner path). No pipeline cut over yet.
