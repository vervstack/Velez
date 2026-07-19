# Task/Job Engine — Velez vs. Artel Comparison & Merge Direction

Research comparing Velez's `internal/jobs` durable task engine against Artel's
"Tract" workflow engine, with the goal of extracting a shared Go+Postgres
job/task execution library. No code was written as part of this research —
this is analysis only.

## Sources

- Velez: `internal/jobs/` (`job.go`, `engine.go`, `worker.go`, `checkpoint.go`,
  `registry.go`, `create_smerd.go`, `create_service.go`, `assemble_config.go`),
  `internal/storage/postgres/` (tasks/jobs queries), `internal/storage/local_storage/`,
  `internal/transport/tasks_api_impl/`, `api/grpc/tasks.proto`,
  `docs/jobs_migration.md`.
- Artel (`~/ruf/artel`): `internal/domain/tract.go`, `internal/service/v1/tract/`,
  `internal/repository/pg/repos/tracts/`, `migrations/033_tracts.sql`,
  `migrations/034_gitlab_tract_tools.sql`, `api/grpc/tracts.proto`,
  `internal/transport/tracts_api/`, `internal/transport/tract_webhook/`,
  `docs/tract-risks.md`.

## Per-system deep dive

### Velez `internal/jobs`

1. **Data model** — two Postgres tables (`velez.tasks`, `velez.jobs`).
   `tasks`: one row per `(entity_id, action)`, `status` enum
   (`PENDING|RUNNING|DONE|FAILED`), `context JSON` payload, `claimed_at`/`claimed_by`
   lease fields, `UNIQUE(entity_id, action)`. `jobs`: one row per `(task_id, job_name)`
   step, `status` (`RUNNING|DONE|FAILED`), `error`. No retry-count column on either
   table. Step payload isn't stored separately — only the task-level `context` is.
2. **Definition of work** — 100% Go, static per action. `TaskHandler.BuildJobs`
   returns a fixed `[]NamedJob` slice built from a live `TaskContext`. No
   loops/conditionals in the step graph itself; jobs share data via narrow Go
   accessor interfaces over the concrete proto payload.
3. **Execution/scheduling** — single in-process polling loop
   (`time.Ticker`, 2s interval) claiming work via
   `UPDATE ... WHERE id = (SELECT ... FOR UPDATE SKIP LOCKED LIMIT 1)`. A
   `claimed_at` lease (2 min) reclaims orphaned tasks from crashed workers.
   Multi-replica-safe and restart-survivable; no LISTEN/NOTIFY.
4. **Checkpointing/resumability** — `checkpointedJob` wraps every step: skip if
   already `DONE`, fail-fast if already `FAILED`, re-run if `RUNNING` (crash
   mid-step) or absent. On resume, the whole step list is rebuilt and
   re-walked top-to-bottom; already-done steps are skipped in O(1).
5. **Retry & failure** — no automatic retry/backoff. A `FAILED` step is
   terminal until something external clears its checkpoint row. Rollback
   *is* supported: `RollbackableJob` — on first failure, `runJobs` walks
   completed steps backward and calls `Rollback` on each.
6. **Concurrency control** — enqueue-time dedup via
   `INSERT ... ON CONFLICT (entity_id, action) DO NOTHING` (concurrent
   `Enqueue` calls converge on one row); execution-time exclusivity via the
   `SKIP LOCKED` claim.
7. **Data passing** — a shared mutable `TaskContext` (proto pointer) threaded
   between steps by Go closures in-process, and persisted as JSON to
   `tasks.context` after every successful step for resume.
8. **API/observability** — single `WatchTask` server-streaming RPC
   (`api/grpc/tasks.proto`), implemented as a 1s poll loop over
   `GetTaskByEntityAction` internally, pushed to the client as a stream.
   No public `EnqueueTask` RPC — enqueuing happens from other API handlers
   via `Engine.Enqueue`.
9. **Extensibility** — new task type = new proto payload + new
   `internal/jobs/{name}.go` implementing `TaskHandler` + registration in
   `internal/app/custom.go`. Always Go code + rebuild + redeploy; no
   data-driven registration path.

### Artel "Tract"

1. **Data model** — `tracts` (definition: `id, user_id, definition JSONB, enabled, ...`),
   `tract_runs` (one row per execution: `status`, `trigger_output JSONB`, `error`),
   `tract_run_steps` (one row per executed step instance: `step_id, step_type,
   input/output JSONB, status, error`), plus `triggers`/`tract_trigger_links`.
   Definition and run state are cleanly separated; no retry-count column.
   No FK from the JSONB definition into tool/connection tables — flagged as a
   known risk in `docs/tract-risks.md`.
2. **Definition of work** — hybrid. `TractStep.Type` is a closed enum
   (`action|condition|parallel|group|script`) forming a JSONB tree. `action`
   steps reference a tool by name from the `mcp_tools` catalog table (fully
   data-driven HTTP-call templates, executed generically by `HttpExecutor`) —
   adding a new HTTP-backed action is a pure DB insert, no Go. Builtins and
   `script` (embedded JS via `goja`) are the exceptions requiring Go/code.
   Adding a new step *kind* (not instance) still requires a Go switch-case.
3. **Execution/scheduling** — no engine loop, no locking, no LISTEN/NOTIFY.
   A run is a bare goroutine spawned inline (`go t.runManual(...)`, webhook
   dispatch, MCP tool) against the server's lifecycle context. No multi-replica
   coordination — each server instance just runs whatever it triggered.
4. **Checkpointing/resumability** — persist-before-apply per step (row
   inserted as `running` before the call executes), but **no actual resume
   logic** — killing the process kills the run. A boot-time `SweepStaleRuns`
   just marks orphaned `running` rows as `failed` for UI hygiene. Explicitly
   called out in `docs/tract-risks.md` as a deliberate v1 simplification.
5. **Retry & failure** — manual, run-level only. `RetryRun` starts a brand-new
   full run reusing the original trigger payload; it does not skip completed
   steps. No automatic retry/backoff. No rollback/compensation concept.
6. **Concurrency control** — none. No lock, no "already running" guard;
   repeated triggers spawn independent concurrent runs against the same tract.
7. **Data passing** — explicit template references (`{{step_id.path}}`,
   `{{trigger.path}}`) resolved against a per-run in-memory JSON outputs map,
   snapshotted before each step to avoid races between parallel branches.
8. **API/observability** — `GetRun` (one-shot) plus `WatchRun`, a
   server-streaming RPC that is itself a 700ms poll loop wrapping `GetRun` —
   no pub/sub in the engine.
9. **Extensibility** — new HTTP-backed action = pure data (`mcp_tools` insert).
   New step kind, condition operator, or script language = Go change in
   `executeStep`'s switch and every validator. The step-tree model is
   explicitly *not* an arbitrary DAG — sequence/branch/parallel-group only,
   a deliberate ceiling per `docs/tract-risks.md`.

## Common ground (both systems converged independently — kept in the merge)

| Aspect | Shared approach |
|---|---|
| Definition vs. run state | Keep separate regardless of how definitions are authored. |
| Per-step durable audit trail | One row per step-attempt (status/error/timestamps) joined to the parent run. |
| Watch/observe API | Server-streaming RPC that's a poll loop over the DB row under the hood — not real pub/sub. Both landed here independently; keep it. |
| JSON payload on the run row | Step input/output as JSON, not typed columns. |
| Whole-context persistence | Persist accumulated context after every step for resumability. |

## Divergent decisions — resolved

Presented to the user as a 4-question decision; answers below.

| Decision | Chosen |
|---|---|
| Trigger/pickup mechanism | **Poll + `SELECT ... FOR UPDATE SKIP LOCKED`** (Velez's current approach) — proven, multi-replica-safe, restart-survivable. LISTEN/NOTIFY and Artel's bare-goroutine model were rejected. |
| Step definition model | **Static Go step lists per action** (Velez's current model) — max type safety and rollback support. The data-driven catalog / hybrid catalog options were declined. |
| Execution graph shape | **Tree: sequence + branch + parallel-group** (Artel's current model) — no arbitrary DAG. |
| Retry/backoff | **Add automatic per-step retry with backoff** — new capability neither system has today. |

## Synthesized target architecture

The combination is coherent: fully compiled/typed, tree-shaped, self-healing.

1. **Data model (3 tables)** — merge Velez's simplicity with Artel's
   definition/run split:
   - `tasks`/`runs`: `id`, `entity_id`, `action`, `status`, `context JSONB`,
     `claimed_at`, `claimed_by`, `attempt`/`retry_count`, `error`, timestamps.
     Keeps Velez's `UNIQUE(entity_id, action)`.
   - `steps`: one row per step *attempt*: `task_id`, `step_path` (must now
     encode tree position, not just a flat name — e.g.
     `"parallel_group_2/branch_a/create_container"`), `status`, `attempt`,
     `error`, `input/output JSONB`, timestamps.
   - No separate definition table — since step lists are static Go, the
     definition lives in code via `Registry`, as in Velez today.

2. **Definition mechanism** — static Go, but tree-capable. `BuildJobs` needs
   to return a tree, not a flat slice: combinators like `jobs.Sequence(...)`,
   `jobs.Parallel(...)`, `jobs.Branch(cond, then, else)` alongside
   `NamedJob`. The flat `[]NamedJob` becomes one case of a more general tree
   node. Rollback and retry both need to be defined against this tree.

3. **Trigger/scheduling** — unchanged from Velez: ticker poll,
   `FOR UPDATE SKIP LOCKED`, lease-based reclaim. Agnostic to what's inside a
   task, so no change needed to support the tree/retry additions.

4. **Checkpointing & resume** — extends Velez's DONE-skip/FAILED-short-circuit
   gate to walk the tree: sequence nodes checkpoint in order; parallel nodes
   checkpoint each branch independently, so a crash mid-parallel-group
   resumes only the unfinished branches.

5. **Retry & backoff** — add `attempt`/`max_attempts`/`next_retry_at` to the
   `steps` row. Checkpoint gate becomes: DONE→skip, FAILED with attempts
   remaining→re-run after backoff, FAILED and exhausted→terminal (then
   rollback fires as today). Backoff policy (fixed/exponential) is a property
   of the step-kind registration, not per-instance config.

6. **Rollback/compensation** — `RollbackableJob` carries forward unchanged
   for sequence nodes. For a `parallel` node, rollback needs a policy
   decision (e.g. roll back all completed branches concurrently before
   continuing backward through the outer sequence) — to be decided at
   implementation time.

7. **Concurrency/dedup** — unchanged from Velez: `ON CONFLICT DO NOTHING` at
   enqueue, claim-lease at execution. Independent of tree shape.

8. **Data passing** — typed context, unchanged from Velez: shared mutable
   `TaskContext` threaded via closures, persisted as JSON after each step.
   No template-string/JSON-ref layer (that was Artel's data-driven-composition
   mechanism, not chosen).

9. **Watch/API** — unchanged: poll-loop-under-a-streaming-RPC, as both
   systems already independently do.

## Open trade-off to confirm before implementation

Choosing static-Go-only step definition means the merged **library** won't
give Artel's product the "non-engineer composes a workflow from the canvas"
capability directly — that requires data-driven step composition, which
wasn't chosen. If Artel's UI still needs to let users build tracts visually,
that composition layer would have to stay an *application-level* concern on
top of the shared library (e.g. one specific step-kind that's an
"interpreter" walking a stored JSON tree at runtime) rather than something
the core engine provides generically. Worth confirming this is acceptable
for Artel's product, and worth scoping how much of Artel's canvas-builder UX
would survive that split, before treating this direction as final.
