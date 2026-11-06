# Jobs Engine (`internal/jobs/`)

Durable, checkpointed replacement for `internal/pipelines`: a `TaskHandler` per pipeline
(`Action()`, `NewContext()`, `BuildJobs()`) builds an ordered list of `NamedJob`s, each wrapped in a
`checkpointedJob` (DONE-skip, FAILED-short-circuit, resume-safe) and persisted as JSON to
`velez.tasks`/`velez.jobs` (Postgres in production, `local_storage` in single-node/dev — same
interface). `internal/app/custom.go` registers every `TaskHandler`; a single `taskWorker` polls and
claims work via `SELECT ... FOR UPDATE SKIP LOCKED`.

See `docs/jobs_migration.md` for per-pipeline migration status/checklist and
`docs/plans/testing.md` for the live-RPC cutover plan.

## Hard-won rules — read before touching a job scaffold or cutting one over to a live RPC

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
