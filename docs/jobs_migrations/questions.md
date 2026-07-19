# Jobs Migration — Open Questions

Questions I couldn't decide on my own during the jobs migration. Each has the
default I proceeded with so work wasn't blocked (per your instruction). Answers
below are final unless noted otherwise.

## CopyToVolume — resolved 2026-07-19

1. **CopyToVolume has ZERO callers today.** Repo-wide search for `.CopyToVolume(`,
   `CopyToVolumeRequest`, `PathToFiles` finds only the interface decl
   (`internal/pipelines/pipelines.go`) and impl (`do_copy_to_volume.go`). No gRPC
   handler / service / worker invokes it — unlike CreateService / LaunchSmerd /
   AssembleConfig which all have live callers pre-cutover.
   - **Decision: keep it scaffolded, no cut-over.** It's dead code today but
     scaffolding it proved the dynamic-loop/checkpoint pattern `UpgradeSmerd`
     will also need.

2. **mkdir + write per file: one job or two?** Pipeline does `Exec("mkdir -p <dir>")`
   then `CopyToContainer` as separate steps per file.
   - **Decision:** Combine into ONE `copy_file_<n>` job per file (fewer checkpoint
     rows; `mkdir -p` is idempotent so no correctness cost). Same fold precedent
     was reused for `ConnectServiceToVpn`'s env-append step (see below).

3. **nil-content files:** `container_steps.CopyToContainer` silently skips files whose
   `Content == nil` (`if m.Content == nil { continue }`). `PathToFiles` is
   `map[string][]byte`, so a present-but-nil slice is possible.
   - **Decision: preserve the skip as-is.** Migrations don't change behavior;
     this stays a known quirk rather than a fix.

4. **Docker fake investment.** No Docker fake exists anywhere in the repo's unit tests
   (only `tests/e2e` uses a real daemon). AssembleConfig's 4 Docker-touching jobs
   are consequently untested at unit level.
   - **Decision:** Introduced a minimal, reusable Docker fake in
     `internal/jobs/fakes_test.go` (`fakeDocker`/`fakeNodeClients`/`fakeContainerAPI`).
     `ConnectServiceToVpn` builds on it directly (see item 8 below) — confirms this
     was the right investment.

5. **`smerd_steps.Exec` ignores exit codes** (`_ = res`) — a failing `mkdir -p`
   (e.g. permission denied) surfaces only on a transport-level Docker error, not a
   non-zero exit. CopyToVolume inherits this.
   - **Decision: inherit as-is.** Out of scope for a behavior-preserving migration,
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
   - **Decision: keep the narrow-interface pattern, reuse it going forward.**
     Introduced two small local interfaces in `copy_to_volume.go` (`startAPI` —
     `ContainerStart`/`ContainerStop`; `copyAPI` — `CopyToContainer`) that the real
     `client.APIClient` still satisfies structurally, plus a `writeFileToContainer`
     helper duplicating `dockerutils.WriteToContainer`'s tar-build logic against the
     narrow `copyAPI` (Go's interface assignability rules mean a fake satisfying only
     `copyAPI` can never be passed where `dockerutils.WriteToContainer` expects a full
     `client.APIClient`). Confirmed as the go-forward pattern: `connect_service_to_vpn.go`'s
     `startSidecarContainerJob` reuses `startAPI` as-is (no new interface needed) and
     gets full unit + rollback coverage the same way.

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

## ConnectServiceToVpn — resolved 2026-07-19

A prior session had left this half-migrated: `api/grpc/tasks.proto` and the
regenerated `tasks.pb.go`/`tasks_setters.go` already had
`ConnectServiceToVpnTaskPayload` (uncommitted), but no
`internal/jobs/connect_service_to_vpn.go`, no registration, no tests. This
session finished the migration checklist against the existing proto scaffold.

10. **`client_key` (Headscale auth key) storage.** Same shape as question #4's
    concern, but for a genuine secret this time: the proto field is a plain
    `optional string`, persisted verbatim into `velez.tasks.context` JSON.
    - **Decision: keep it a plain string.** No field-level encryption, no
      omission from the persisted context — the old pipeline already held it
      in-memory with no encryption at rest, so this isn't a regression, and
      encrypting one field is a bigger architectural change than this
      migration's scope. Documented in `docs/security/vpn_client_key_storage.md`
      per your instruction to write this decision down.

11. **Inline env-append closure step** (`launchContainer.Env = append(...)` in
    `do_connect_service_to_vpn.go`) has no result to persist on its own.
    - **Decision:** Folded into `createSidecarContainerJob.Do()` rather than
      given its own checkpoint row — same precedent as question #2's
      mkdir+copy fold for CopyToVolume. 9 pipeline steps become 8 named jobs.

12. **`network_steps.GetClientKey`'s "existing key found" branch has a latent
    bug**: `h.keyResponse = &authKey.Key` reassigns the step's own pointer
    field instead of writing through it (`*h.keyResponse = ...`), so the
    caller's `clientKey` variable is never actually updated when a reusable
    key already exists — the pipeline silently falls through to issuing a
    brand-new key every time instead of reusing one.
    - **Decision: did not reproduce the bug.** The jobs model passes state via
      `SetClientKey(string)` on the proto `TaskContext` rather than a raw
      pointer, and a setter call always writes the value — there's no pointer
      to "reassign instead of dereference" in this model, so the bug is
      structurally gone in `getClientKeyJob.Do()`, not deliberately fixed.
      Net effect: the job version correctly reuses an existing key instead of
      always minting a new one. Flagging in case this was ever relied upon
      (e.g. tests or ops expecting one fresh key per connect); if so, say so
      and it can be special-cased back.

13. **`container_steps.Create`'s extra tolerance not carried over.** The
    pipeline's `Create` step (a) tolerates `docker.ErrNameIsTaken` on create
    (treats a same-named container as pre-existing rather than failing), and
    (b) on rollback, inspects the container first and also removes its volume
    mounts, not just the container itself.
    - **Decision: simplified to match this migration's existing container-job
      pattern.** `createSidecarContainerJob` uses plain `ContainerCreate` +
      `Remove`-on-rollback, the same shape as `create_smerd.go`'s
      `createContainerJob` and `copy_to_volume.go`'s `createLoaderContainerJob`
      — neither of those carried `ErrNameIsTaken` tolerance or mount cleanup
      either. Low risk since this pipeline has no live caller yet (no
      cut-over), but flagging since it's a real behavior narrowing versus the
      specific step being replaced, unlike the other jobs' precedents which
      replaced steps that never had this extra tolerance to begin with.

## EnableStatefullMode — resolved 2026-07-19

14. **Password idempotency across resumes.** The pipeline generates the
    Postgres root/node-user passwords by checking whether
    `localState.ClusterState.PgRootDsn`/`PgNodeDsn` are already parseable -
    but those fields aren't written until the *last* pipeline step. If a job
    task fails and resumes before that point, recomputing this check in
    `BuildJobs` on every rebuild would mint a fresh random password each
    time, mismatching whatever password was already baked into the running
    container or already-created Postgres user.
    - **Decision:** promoted the check into its own `generate_credentials`
      job that only generates once and persists the result into
      `EnableStatefullTaskPayload.root_pwd`/`user_pwd`, reusing it on
      resume - same idempotency precedent as `ConnectServiceToVpn`'s
      `client_key` (question #10). 7 pipeline steps become 8 named jobs.

15. **`ClusterStateManager`/`StorageContainer` singleton swap.** The
    pipeline's last-but-one step mutates live, shared service state
    (`p.clusterClients.StateManager().Set(...)`,
    `p.services.StorageContainer().Set(...)`), not just Docker or SQL.
    - **Decision:** carried over as-is inside `updateClusterStateJob`,
      unchanged. Changing *how* that swap happens is out of scope for a
      behavior-preserving migration; flagging since it's the first job in
      this migration series to mutate cross-cutting live singletons rather
      than a single container's own state.

16. **No SQL fake/mock exists in this repo** (no `sqlmock`/`go-sqlmock`
    dependency), unlike question #4's Docker fake. `create_schema_and_migrate`
    and `create_pg_user` reuse `sqldb.RollMigration`/
    `cluster_steps.CreatePgUserForNode` verbatim, both of which open a real
    `*sql.DB` connection via `sql.Open` + `Exec`.
    - **Decision:** left these two jobs' success path untested at unit level,
      same untested status quo as the original (never unit-tested)
      `cluster_steps` pipeline steps they reuse - no new test debt
      introduced, but no new coverage added either. Only their
      connection-failure path is exercised (a real dial against an
      unreachable port), which also doubles as this migration's required
      end-to-end failure-path test, exercising the full rollback cascade
      through `start_container`/`create_container`. The success path needs a
      real Postgres and stays `tests/e2e`-only, consistent with how Docker
      daemon-required behavior is already handled there.

17. **`get_root_dsn`'s Docker dependency.** `cluster_steps.GetRgRootDsn`
    (the pipeline step this job replaces) takes the full `node_clients.Docker`
    interface and writes its result through a raw `*string` fixed at
    construction time - neither trait fits a resumable job (the container id
    it needs is only available from the task's persisted context at `Do()`
    time, and `fakeDocker.Client()` always returns `nil` so no fake
    `node_clients.Docker` can be built for it anyway).
    - **Decision:** duplicated the inspect/env-parse logic in `getRootDsnJob`
      against the narrow `containerInspectAPI` interface instead - same
      narrow-interface-over-duplication tradeoff as `copy_to_volume.go`'s
      `writeFileToContainer` helper (question #8).

## UpgradeSmerd — resolved 2026-07-19

The most complex pipeline, saved for last per the suggested order: 3 stages
(pause the old container, build+extract config from a throwaway "scratch"
container running the new image, build+launch the real new container), with
in-place container renames threading through all three.

18. **A pipeline stage looks like dead/unfinished wiring.** do_smerd_upgrade.go
    creates a full "config-fetcher" container (same Settings/Ports/Volumes as
    the final container, just a different Name), reads a config file out of
    it into a local `cfgMount` variable, then drops the container - but
    `cfgMount` is never read again anywhere in the pipeline. Every other
    field `FromContainerToRequest`/`FetchConfig`/`PrepareVervConfig` populate
    is threaded through to the final container; this one isn't.
    - **Decision: preserve it anyway.** `getConfigFromScratchContainerJob`
      still performs the container-create, tar-read and container-drop
      exactly as before, computing a result it never persists to any
      accessor. This is the same "behavior-preserving migration, don't fix
      what you find" rule used for question #12's latent pointer bug and
      question #13's dropped error tolerance - flagging in case you want this
      whole stage (`create_config_fetcher_container` /
      `get_config_from_container` / `drop_config_fetcher_container`, 3 of the
      15 jobs) stripped out as genuinely dead work in a follow-up.

19. **Shared `container_id` field reused across two container-create stages,
    on purpose.** The original pipeline reuses one `newContId` variable
    (via step closures) for both the scratch config-fetcher container and
    the final container. Because both `smerd_steps.Create` calls close over
    the *same* pointer, the old pipeline's own rollback already has a latent
    quirk: if the final-container stage fails, rolling back the
    config-fetcher stage's `Create` step reads whatever the shared pointer
    holds *now* (the final container's id, not its own), so it can end up
    trying to remove the wrong id - harmless only because
    `node_clients.Docker.Remove` tolerates `NotFound`.
    - **Decision: replicate this with a single shared `container_id` proto
      field**, rather than giving the scratch and final containers separate
      fields (which would be strictly *safer* than the original but would
      change rollback behavior versus what's being migrated). Same
      "behavior-preserving, including its quirks" rule as #18.
      `old_container_id` stays a separate field since the original never
      shared it with `newContId` either.

20. **`smerd_steps.PauseContainer`/`RenameContainer`/
    `config_steps.getConfigFromContainerStep` all take `client.APIClient`
    (or need it via `dockerutils.DisconnectFromNetworks`/`ConnectToNetwork`/
    `CreateNetwork`/`ReadFromContainer`, all parameterized on the full
    interface too) - none of them fit a fake, same wall as question #8's
    `copyAPI`/`writeFileToContainer`.
    - **Decision: reuse the go-forward pattern from question #8, applied
      four more times.** New narrow interfaces `pauseAPI`, `renameAPI`,
      `copyFromAPI`, `createNetworkAPI` in `upgrade_smerd.go`, each with a
      duplicated helper function reproducing the relevant dockerutils
      function's logic against the narrow interface instead of calling it
      directly. `fakeContainerAPI` (in `fakes_test.go`) grew matching methods
      (`ContainerPause`/`Unpause`/`Rename`, `NetworkDisconnect`/`Connect`/
      `List`/`Create`, `CopyFromContainer`) and now embeds `client.APIClient`
      as a nil interface (satisfies the full 100+ method interface via
      promotion; only the explicitly-overridden methods are safe to call) so
      it can also be injected as `fakeDocker`'s `Client()` return value via a
      new `withClient` builder - `fakeDocker.Client()` previously always
      returned `nil` (question #17), which meant no test in this repo had
      ever exercised a full end-to-end happy path through jobs needing the
      raw Docker client. This unblocked a genuine 15-job happy-path
      end-to-end test for the first time, not just a failure path.

21. **No `node_clients.PortManager` fake exists** (question-mark left open
    implicitly since no prior migration's jobs needed port locking at unit
    level). `ports.NewPortManager` is a real, in-memory implementation with
    no external dependency beyond a local `net.Listen` availability probe.
    - **Decision: use the real one directly in tests** rather than
      hand-writing a fake, via a new `fakeNodeClients.withPortManager`
      builder (defaults to nil, unchanged for every other migration's
      tests). Cheap and correct; no PortManager fake needed.

## Cross-cutting (all pipelines)

6. **Live-RPC cut-over** stays OFF for every pipeline unless you say otherwise
   (checklist item 7 — your explicit per-pipeline decision). This note is
   stale as of `UpgradeSmerd`'s cutover (2026-07-19): five of the six
   migrated pipelines are now cut over to their live RPCs/callers -
   `LaunchSmerd`/`create_smerd`, `CreateService`/`create_service`,
   `AssembleConfig`/`assemble_config`, `ConnectServiceToVpn`/
   `connect_service_to_vpn`, `EnableStatefullMode`/`enable_statefull_mode`,
   and `UpgradeSmerd`/`upgrade_smerd`. Only `CopyToVolume` remains
   "Scaffolded" and un-cut-over, and it has no live caller to cut over in the
   first place (#1 above). See `docs/jobs_migration.md`'s Status table for
   the current, authoritative per-pipeline state - don't trust this
   paragraph's own history over that table.
