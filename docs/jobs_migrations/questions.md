# Jobs Migration — Open Questions

Questions I couldn't decide on my own during the jobs migration. Each has the
default I proceeded with so work wasn't blocked (per your instruction). Override
any of these and I'll adjust on the next pass.

## CopyToVolume (migrating now)

1. **CopyToVolume has ZERO callers today.** Repo-wide search for `.CopyToVolume(`,
   `CopyToVolumeRequest`, `PathToFiles` finds only the interface decl
   (`internal/pipelines/pipelines.go`) and impl (`do_copy_to_volume.go`). No gRPC
   handler / service / worker invokes it — unlike CreateService / LaunchSmerd /
   AssembleConfig which all have live callers pre-cutover.
   - **Question:** Is this dead code (delete instead of migrate?), or is there a
     planned caller not yet wired up? Affects what `entity_id` a future
     `Engine.Enqueue(..., "copy_to_volume", ...)` would use — no convention to mirror.
   - **My default:** Migrate it anyway (scaffold only, no cut-over — there's nothing
     to cut over). Low risk, and it proves the dynamic looped-jobs + checkpoint
     pattern that UpgradeSmerd will also need. If you'd rather delete it, say so.

2. **mkdir + write per file: one job or two?** Pipeline does `Exec("mkdir -p <dir>")`
   then `CopyToContainer` as separate steps per file.
   - **My default:** Combine into ONE `copy_file_<n>` job per file (fewer checkpoint
     rows; `mkdir -p` is idempotent so no correctness cost). Alternative is a literal
     1:1 mirror `mkdir_file_<n>` + `copy_file_<n>`.

3. **nil-content files:** `container_steps.CopyToContainer` silently skips files whose
   `Content == nil` (`if m.Content == nil { continue }`). `PathToFiles` is
   `map[string][]byte`, so a present-but-nil slice is possible.
   - **My default:** Preserve the skip (inherit existing behavior; migrations don't
     change behavior). Flag: could be treated as a bug now that it's unit-testable.

4. **Docker fake investment.** No Docker fake exists anywhere in the repo's unit tests
   (only `tests/e2e` uses a real daemon). AssembleConfig's 4 Docker-touching jobs
   are consequently untested at unit level.
   - **My default:** Introduce a MINIMAL, reusable Docker fake in
     `internal/jobs/fakes_test.go` covering only the methods CopyToVolume needs
     (container create/inspect/remove, exec, write), so ConnectServiceToVpn /
     UpgradeSmerd can build on it. If the raw `client.APIClient` surface makes this
     impractical, I'll fall back to unit-testing pure helpers + a create-container
     failure path only, matching AssembleConfig's coverage level.

5. **`smerd_steps.Exec` ignores exit codes** (`_ = res`) — a failing `mkdir -p`
   (e.g. permission denied) surfaces only on a transport-level Docker error, not a
   non-zero exit. CopyToVolume inherits this.
   - **My default:** Inherit as-is (out of scope for a behavior-preserving migration),
     same as the `EnvFormat` quirk pinned in `assemble_config_test.go`.

## CopyToVolume (added during implementation)

7. **Skip `ContainerInspect` after `ContainerCreate`.** `smerd_steps.Create` (and
   `create_smerd.go`'s `createContainerJob`) re-inspect the just-created container
   and use the inspected ID rather than `ContainerCreate`'s own response ID.
   - **My default:** `createLoaderContainerJob` uses `ContainerCreate`'s response ID
     directly and skips the inspect call, following `assemble_config.go`'s
     `createScratchContainerJob` precedent instead. This keeps the job on the narrow
     `node_clients.Docker` interface only (no `client.APIClient` needed), so it's
     fully unit-testable with the existing `fakeDocker`. Functionally identical (the
     inspected ID is always the same ID Docker just returned); flag if you want the
     inspect call kept for parity with `create_smerd.go`.

8. **`startLoaderContainerJob` and `copyFileJob` still need `client.APIClient`**
   (`ContainerStart`/`ContainerStop` aren't on `node_clients.Docker` at all, and
   `dockerutils.WriteToContainer` takes `client.APIClient` by name). Question #4's
   default said I'd fall back to pure-helper + create-container-failure tests only
   if this wall was hit.
   - **My default:** Went one step further than the documented fallback: introduced
     two small local interfaces in `copy_to_volume.go` (`startAPI` — `ContainerStart`/
     `ContainerStop`; `copyAPI` — `CopyToContainer`) that the real `client.APIClient`
     still satisfies structurally, and a `writeFileToContainer` helper that
     duplicates `dockerutils.WriteToContainer`'s tar-build logic against the narrow
     `copyAPI` instead of calling it directly (Go's interface assignability rules
     mean a fake satisfying only `copyAPI` can never be passed where
     `dockerutils.WriteToContainer` expects a full `client.APIClient`). This bought
     full unit coverage (happy path + every required failure path, including the
     cascading-rollback scenario) at the cost of ~40 lines of duplicated tar logic
     and two fields per job now typed as project-local interfaces instead of
     `client.APIClient` directly, unlike `create_smerd.go`'s equivalent jobs. Flag if
     you'd rather keep `client.APIClient` untouched everywhere and accept the
     narrower AssembleConfig-level coverage.

9. **Regenerating via `moti g` re-broke `pkg/docs.swagger_ui.go`.** That file's
   `//go:embed all:swaggers` directive doesn't match its own directory (it's
   generated at `pkg/`, but `swaggers/` only exists under `pkg/docs/`), so `go build`
   fails whenever it's present. Git had it staged as added but already deleted from
   disk before this session started, implying it was already removed once. Running
   `moti g` (needed to add `CopyToVolumeTaskPayload`) regenerated the whole
   `api/grpc/*.proto` set including this file's plugin output.
   - **My default:** Deleted it again after `moti g`, restoring the pre-existing
     (deleted-from-disk) state rather than leaving the build broken. Not a
     CopyToVolume-specific change; flagging in case there's a moti.yaml fix needed
     so this stops recurring on every `moti g` run.

## Cross-cutting (all pipelines)

6. **Live-RPC cut-over** stays OFF for every pipeline unless you say otherwise
   (checklist item 7 — your explicit per-pipeline decision). All migrations remain
   "Scaffolded": handler registered + tested, old pipeliner still serves live RPCs.
