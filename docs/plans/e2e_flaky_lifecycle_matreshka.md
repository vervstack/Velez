# Investigation: `Test_Lifecycle` flakiness (Test_ClusterMode_* subtests)

**Status: resolved.** Fixed via a shared, singleton matreshka fixture — see
"Resolution" at the bottom. The root-cause writeup and ruled-out theories
below are kept as-is for reference.

## Root cause

The three `Test_ClusterMode_*` subtests in `tests/e2e/suite_api_deploy_test.go`
(`Test_ClusterMode_HelloWorld`, `Test_ClusterMode_PlainNginx`,
`Test_ClusterMode_Postgres`) all call `NewEnvironment(t, WithMatreshka())`.
Every one of them runs `t.Parallel()`, and each gets its own in-process `App`,
but **all of them share one real Docker daemon** and try to stand up a
matreshka sidecar container using **hardcoded, non-unique identity**:

- Container name: `internal/cluster/configuration/deploy_local.go:17`
  — `Name = "matreshka"` (literal, used as `Config.Hostname` in
  `service.go:SetupMatreshka`, and as the container name Docker actually
  creates/looks up — confirmed via `initKey`'s
  `docker.ContainerInspect(ctx, Name)` at
  `internal/cluster/configuration/deploy_local.go:55`).
- Host port: `internal/cluster/configuration/service.go:60-75` —
  `grpcPort = "50049"` bound as a fixed `HostPort` unless
  `cfg.Environment.MatreshkaPort > 0` (nothing in
  `tests/config_mocks/velez_default_config.yaml` sets this, and no e2e test
  calls a port override for matreshka).

`cfg.Environment.ContainerSuffix` exists and is plumbed into
`docker.NewClient` (`internal/clients/node_clients/docker/client.go:32,137-138,197`)
and `tests/e2e/helper_environment.go:88`'s `WithContainerSuffix` option — but
it's only ever applied as a **label** (`labels.SuffixLabel`), never as part of
the actual container *name*. And no `Test_ClusterMode_*` subtest calls
`WithContainerSuffix` anyway, so it wouldn't help even if it did rename
things.

Each matreshka container is wrapped in `keep_alive.KeepAlive(...)`
(`service.go:97`), a background retry loop that keeps trying to (re)create the
container whenever it's not running, and tears it down on `App.Custom.Stop()`.
With 3 parallel envs all targeting the same name/port:

- Only one env's container create actually succeeds; the other two loop,
  logging `container name is taken;error creating container;error keeping
  service matreshka alive`.
- On teardown, whichever env's `Stop()` loses the race to remove the
  container first causes the others to log
  `Error response from daemon: removal of container matreshka is already in
  progress;error removing container;error dropping result;error killing
  service matreshka`.

This matches the project's own `CLAUDE.md` note (jobs-engine hard-won rule
#4) that `Test_Lifecycle` hangs/misbehaves specifically under full-parallel
execution and not standalone — except here it reproduced **solo**, running
only `Test_Lifecycle` by itself (`go test ./tests/e2e/ -run Test_Lifecycle -v
-timeout 10m`), because the 3 `ClusterMode` subtests are parallel *with each
other* regardless of what else is in the suite.

Because none of the three assertions in these subtests actually depend on
matreshka answering anything (`IgnoreConfig: true` is set, or the config path
doesn't block on it), the *test itself* often still passes despite the
background errors — which is exactly why this is flaky rather than a hard,
consistent failure: whether the run passes or not depends on scheduling luck
in the keep-alive retry loop and Docker's container-removal timing, not on
any assertion actually reading bad state.

## What was ruled out

Investigated and ruled out as the primary cause before finding the above
(kept here so a future session doesn't re-walk the same path):

- **Shared `PortManager`** (`tests/test_helper/ports.go`,
  `internal/clients/node_clients/ports/port_manager.go`) — is
  mutex-serialized correctly; `GetPort()` holds its lock for the entire
  check-and-reserve sequence, so two parallel envs can't be handed the same
  port from *this* manager. Real culprit is the separate,
  manager-bypassing hardcoded matreshka port above.
- **Shared Postgres `velez.tasks`/`velez.jobs` tables** across parallel
  `taskWorker`s — logically fine (`SELECT ... FOR UPDATE SKIP LOCKED`
  row-level locking), just adds DB load. Not implicated in the observed
  failure (which reproduced in under 1s-20s, no timeout/hang was hit).
- **Watch-loop timeouts sized for solo runs** (`createSmerdWatchTimeout` 60s,
  `upgradeSmerdWatchTimeout` 120s) — plausible theory before reproduction,
  but the actual failure has nothing to do with job/watch timing; it's a
  Docker-level container-identity collision.

## Reproduction

```
go test ./tests/e2e/ -run Test_Lifecycle -v -timeout 10m
```

Solo run 1: hard **FAIL** at 0.9s —
`Error response from daemon: removal of container matreshka is already in
progress;error removing container;error dropping result;error killing
service matreshka`.

Solo run 2 (immediately after): **PASS** in 19.3s, but with the same class of
errors logged mid-run (`container name is taken...`, `removal of container
matreshka is already in progress...`) — confirms the race exists regardless
of outcome; pass/fail is scheduling-dependent.

## Open decision (do not implement without sign-off — see repo `CLAUDE.md`:
"do not make architectural decisions on your own")

`configuration.Name`/`grpcPort` being fixed is a deliberate "one matreshka
per node" production assumption — not a bug in isolation. Fixing the e2e
flake means picking one of:

1. **Give the matreshka container name/port test-scoping**, e.g. derive the
   container name from `ContainerSuffix` (or a new test-only env knob)
   instead of the bare `Name` constant, and make `MatreshkaPort` default to
   an ephemeral/allocated port in tests. Touches production code
   (`deploy_local.go`, `service.go`) that's currently shared with real
   single-node deployments — needs care that a non-empty suffix/port still
   resolves correctly wherever something dials `"matreshka"` by DNS/service
   name (`matreshka.go`'s `NewClient` dials `verv://matreshka` via service
   discovery, not the container name directly, so this *might* be safe, but
   wasn't verified).
2. **Don't run `Test_ClusterMode_*` subtests in parallel with each other** —
   drop their `t.Parallel()` calls (or gate them behind a `sync.Mutex`/serial
   sub-suite) so only one matreshka container is ever being managed at a
   time. Minimal, test-only change, no production code touched, but slows
   the suite down and only band-aids the underlying "shared singleton name"
   assumption rather than fixing it.
3. **Stub/skip matreshka in these specific subtests** — since none of the
   three actually assert on matreshka behavior, consider whether they need
   `WithMatreshka()` (a real container) at all, versus a fake/mocked cluster
   client. Bigger test-design change, would need to confirm nothing else in
   `App.Custom.Init` depends on a real matreshka being reachable.

## Recommended next step

Get a decision on the above before writing code. Option 2 is the safest/
fastest if the goal is just "stop the flake," option 1 is the real fix if
matreshka-per-test isolation is wanted for future cluster-mode e2e coverage.

## Resolution

None of the three options above was actually chosen. The user's direction,
once we talked it through, was more accurate than any of them: **make the
sharing intentional instead of accidental** — the 3 `Test_ClusterMode_*`
subtests should keep running in parallel (that's the point: it simulates
multiple real nodes concurrently hitting one shared matreshka, matching
production, where there's always exactly one logical matreshka instance per
node regardless of how many nodes call it). The fix is a real, coordinated
singleton container, not per-test isolation (rejected: contradicts "only one
matreshka" prod semantics) and not de-parallelization (rejected: defeats the
purpose of testing concurrent-node behavior).

**Implementation** (full plan: `/Users/alexbukov/.claude/plans/hazy-bubbling-scott.md`):

- `internal/cluster/configuration/service.go` gained an additive,
  backward-compatible seam: `SharedInstance` / `StartSharedInstance` /
  `WithSharedInstance` / `sharedTaskFromContext`. `SetupMatreshka` now checks
  context for a pre-built shared task first; if present, it skips
  `startContainer` (the extracted container-create + `keep_alive.KeepAlive`
  block) entirely instead of racing to create a second one. With no context
  value — every existing production caller — behavior is unchanged.
- `tests/e2e/main_test.go` (new): a lazy `sync.Once`-guarded
  `getSharedMatreshka()` (same pattern as `tests/test_helper.GetSharedPortManager`)
  creates the one real container the first time any test calls
  `WithMatreshka()`; `TestMain` guarantees teardown exactly once after
  `m.Run()`.
- `tests/e2e/helper_environment.go`'s `WithMatreshka()` injects the shared
  instance into the env's context before `Custom.Init()` runs.
- `CLAUDE.md` and doc comments on `WithMatreshka()`/the `Test_ClusterMode_*`
  block both flag the hard constraint this implies: the fixture is a
  single-process (in-memory) singleton, so **every caller must stay in
  package `tests/e2e`** — a different package compiles to a separate test
  binary/OS process and would not observe it, silently reintroducing this
  exact collision.

**Two extra bugs found and fixed along the way, not in the original
diagnosis:**

1. `container_service_task.TaskV2.containerState` was written unguarded in
   `IsAlive()` and read unguarded in `GetPortBinding()`. Harmless when each
   `TaskV2` was short-lived and single-owner; a real, `-race`-provable race
   once one `TaskV2` is long-lived and shared across parallel callers
   (confirmed: removing the fix's `sync.RWMutex` and rerunning
   `Test_TaskV2_ConcurrentIsAliveAndGetPortBinding_NoRace -race` reliably
   reproduces the race; restoring the mutex closes it). Fixed with a
   `sync.RWMutex` guarding the field, no exported API change.
2. `keep_alive.AliveKeeper.Stop()` (external dep,
   `go.redsock.ru/toolbox@v0.0.13/keep_alive/start.go`) never returns:
   `Stop()` calls `a.ticker.Stop()` then blocks on `<-a.stopChan`, but
   `time.Ticker.Stop()` doesn't close the ticker's channel (documented Go
   behavior), so the background `for range a.ticker.C` goroutine blocks
   forever and `close(a.stopChan)` is never reached. `SharedInstance.Stop()`
   therefore only calls `task.Kill()`, never `ka.Stop()` — the ticker
   goroutine dies with the process at exit regardless. This looks like a
   latent bug in production shutdown too (wherever
   `cfg.Environment.ShutDownOnExit` wires `ka.Stop()` into `closer.Add()`),
   left as-is here since it's in an external, unmodifiable dependency and
   out of scope for this fix — worth a separate look someday.

**Verification:** `go build ./...`, `go vet ./...`,
`go test ./internal/cluster/... -race`, and repeated
`go test ./tests/e2e/ -run Test_Lifecycle -v -race -count=5 -timeout 10m`
runs with the three `Test_ClusterMode_*` subtests still visibly interleaved
(parallel), no more `container name is taken` / `removal of container
matreshka is already in progress` errors. Confirmed by A/B comparison
against the unmodified baseline (`git stash`): the baseline still logs the
container-collision errors under the same repeated-run command; this branch
never does, across 5 reruns plus one more clean solo run.

## Three more pre-existing bugs found while verifying — two fixed, one still open

None of these is caused by the matreshka change; all reproduced identically
on the unmodified baseline (confirmed by the same `git stash` A/B comparison
above, or — for bug #2, found later — by the fact that it's an unconditional
package-level global with no dependency on anything this branch touches).
Flagging so a future session doesn't have to rediscover them:

1. **FIXED. `internal/cluster/env/volume.go`'s `StartVolumes()` mutated
   package-level globals (`vervVolumeName`, `vervVolumePath`) with no
   synchronization.** `go test ./tests/e2e/ -run Test_Lifecycle -race`
   caught this independent of matreshka — it hit `Test_Stateless_HelloWorld`,
   `Test_Stateless_Nginx`, etc., none of which call `WithMatreshka()`. Any
   two `NewEnvironment()` calls running in parallel raced on these globals.
   Fixed by adding a package-level `sync.RWMutex` (`volumeMu`) guarding both
   vars — write-locked in `StartVolumes`/`createVervVolume` (the latter is
   only ever called with the lock already held, so it does not lock itself),
   read-locked in `GetVervVolumeName`/`GetVervVolumePath`. No signature
   changes (both getters have zero callers elsewhere in the repo, but they're
   exported, so this keeps the API identical). Proven fixed with a new
   `internal/cluster/env/volume_test.go` (`Test_StartVolumes_Concurrent_NoRace`,
   20 goroutines calling `StartVolumes`+both getters, run with `-race`) and by
   re-running the full e2e `-race` suite: the `volume.go`/`vervVolume*` race
   no longer appears anywhere in the output.

2. **FIXED (upstream, in the generator). `internal/config/load.go`'s
   `Load(configsPaths...)` funneled every call through a single
   **package-level global**, `var defaultConfig Config`, read/writing it in
   five separate unsynchronized steps and returning it by value:
   ```go
   func Load(configsPaths ...string) (Config, error) {
       var err error
       defaultConfig.MatreshkaConfig, err = matreshka.ReadConfigs(configsPaths...)
       ...
       defaultConfig.AppInfo = defaultConfig.MatreshkaConfig.AppInfo
       defaultConfig.Overrides = defaultConfig.MatreshkaConfig.ServiceDiscovery
       err = defaultConfig.MatreshkaConfig.Servers.ParseToStruct(&defaultConfig.Servers)
       ...
       err = defaultConfig.MatreshkaConfig.Environment.ParseToStruct(&defaultConfig.Environment)
       ...
       return defaultConfig, nil
   }
   ```
   Every parallel `NewEnvironment()` → `initConfig()` → `config.Load()` call
   (and `tests/e2e/shared_matreshka.go`'s `getSharedMatreshka`) reads/writes
   this one shared struct concurrently. `go test ./tests/e2e/ -run
   Test_Lifecycle -race -count=2`, run right after fixing bug #1 above so
   that race was silenced, surfaced ~80 separate `-race` reports here, hit by
   essentially every `NewEnvironment()` call including plain `Test_Stateless_*`
   ones — unrelated to matreshka-the-service despite the name overlap with
   the `go.vervstack.ru/matreshka` config library.

   `internal/config/load.go` carries a `// Code generated by RedSock CLI. DO
   NOT EDIT.` header — it's produced by the sibling `go.vervstack.ru/verv`
   CLI repo (`~/verv/verv`), from the template at
   `plugins/project/go_project/patterns/generators/config_generators/templates/autoload.go.pattern`.
   Fixed at the source: `Load()` now builds into a local `cfg Config` and
   returns that — no shared state at all, so concurrent calls (even with
   different `configsPaths`) can't race or interfere, full stop. `Init()`
   (the process-startup entry point, not used by the e2e tests, which call
   `Load()` directly) was separately hardened with `sync.Once` replacing its
   old non-atomic `if defaultConfig.AppInfo.Name != ""` guard — same
   `ErrAlreadyLoaded`-once-loaded contract, now thread-safe.
   Workflow: edited the template in `~/verv/verv`, `go build`/`go test
   ./plugins/project/go_project/patterns/generators/...` there (no snapshot
   test asserts the old template's exact text), `go install .`, then `verv
   tidy` inside this repo to regenerate `internal/config/load.go` (that
   command also touched several unrelated files — file-mode normalization,
   minor gofmt realignment, a new `.githooks/pre-commit`, `config/.env.example`,
   and a `config/config.yaml` reorder — kept per explicit decision, not part
   of this bug fix). Proven fixed with a new `internal/config/load_test.go`
   (`Test_Load_Concurrent_NoRace`, 20 goroutines calling `Load()` on the same
   path) and by re-running `go test ./tests/e2e/ -run Test_Lifecycle -race
   -count=3`: zero `DATA RACE` reports, down from ~80.

3. **STILL OPEN.** `go test ./tests/e2e/...` (either the full suite, or
   `Test_Lifecycle` run repeatedly via `-count=N` in the same process)
   intermittently fails with `helper_environment.go:152`'s
   `require.NoError(t, startServerMasterErr)` (or the `Custom.Init` error one
   line above it) reporting a bare `"closed;;"` error, sometimes as a `panic:
   Fail in goroutine after Test_X has completed` when the failing goroutine's
   subtest has already returned. Originally hypothesized to be a symptom of
   bug #2's `defaultConfig` corruption — **that hypothesis is now ruled
   out**: after fixing both bug #1 and bug #2, a `-race -count=3` run still
   hit this exact failure (`Test_ClusterMode_HelloWorld` and
   `Test_Stateless_Nginx`, on run 2 of 3) with **zero `DATA RACE` warnings
   anywhere in that run** — so it isn't a memory race `-race` can catch, and
   it isn't explained by either fixed bug. Genuinely a separate issue: some
   shared resource (a listener, a docker client, a fixed port, or similar)
   most likely gets legitimately contended or closed by one
   `TestEnvironment`'s teardown while another's `Custom.Init`/`Custom.Start`
   is still using it. Root cause not identified — per explicit instruction,
   not worth further resource investment this session; revisit if it starts
   blocking real work.
