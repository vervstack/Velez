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

## Cross-cutting (all pipelines)

6. **Live-RPC cut-over** stays OFF for every pipeline unless you say otherwise
   (checklist item 7 — your explicit per-pipeline decision). All migrations remain
   "Scaffolded": handler registered + tested, old pipeliner still serves live RPCs.
