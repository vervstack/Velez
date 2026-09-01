# E2E Test Coverage — Matrix & Phased Plan

Status snapshot as of 2026-08-31. Companion to `docs/plans/testing.md` (pipeline→jobs
cutover recipe) and `docs/plans/e2e_flaky_lifecycle_matreshka.md` (shared-fixture fix).
The pre-jobs-engine `docs/review/testing_plan.md` was deleted; its still-valid items were
folded into `docs/roadmap.md §4` and into this file.

All e2e lives in `tests/e2e/`, runs against real Docker via the per-binary DinD harness
(`tests/dind`), `make test-e2e`.

---

## 1. Suite inventory

| Suite | Runs? | Scope |
|---|---|---|
| `Test_Lifecycle` (`suite_api_deploy_test.go`) | yes | create→list lifecycle; stateless hello_world (+healthcheck, +default-config), nginx (image ports), postgres (healthcheck+ports); ClusterMode hw/nginx/pg **with `IgnoreConfig:true`**; DropSmerd by UUID. `Test_StatelessMode_Loki` skipped (stale config + moving tag). |
| `Test_HelloWorldCluster` | yes | pg + 2 verv hello_world apps on a shared bridge net; verv-service labels; cross-app API isolation / peer proxying |
| `Test_Environments` | yes | default-suffix backfill; two-env list scoping; drop-by-name / drop-by-uuid env scoping; `Test_SameNameInTwoEnvironments_AreDistinctContainers` (un-skipped Phase 6 — jobs-engine entity id now environment-scoped). |
| `Test_ServiceScoping` | yes | StopService / RestartService env-scoped to own container; cross-env no-op |
| `Test_ContainerRuntime_Matrix` + `_ListContainers_ScopesToEnvironment` + `_Network_PerEnvironmentIsolation` | yes | label-runtime naming/suffix; per-env docker network isolation; `ListContainers` scoping now green (label filter is unconditional). |
| `Test_UpgradeSmerd` | yes | happy-path upgrade (default env); non-existent fails; `Test_UpgradeSmerd_InSuffixedEnvironment` green (container_runtime suffix handling + Phase 6 entity-id scoping). |
| `Test_AssembleConfig` / `Test_AssembleConfigJob` | yes | config YAML *generation* for a verv image — via live pipeline and via jobs engine |
| `Test_ControlPlane` | yes | ListEnvironments with / without configured envs |
| `Test_EnableStatefull` | yes | statefull_pg plugin happy-path + ListPlugins/ListDeployments; unsupported-plugin fails. Un-skipped Phase C — cluster Postgres reached via the `ClusterPgDsn` host seam + `dindClusterPgPort`. |
| `Test_ServiceLifecycle` (`suite_service_lifecycle_test.go`) | yes | Phase C. Enable statefull_pg (shared `enableStatefullPgUnderDind`), then `CreateService` + `CreateDeploy(New)` through the real ServiceApi handlers; asserts the deployment + spec land in cluster Postgres, read back via `ListDeployments`' `ServiceName` join, status `SCHEDULED_DEPLOYMENT`. **Not** covered (see `TODO(#125)` + suite doc): watcher-driven `SCHEDULED_DEPLOYMENT→RUNNING` and `CreateDeploy(Upgrade)` — the deploy watcher and create_service handler bind their storage backend at startup, before the enable-statefull swap. |
| `Test_Vpn` | **SKIPPED** | whole suite `t.Skip("")` — headscale ConnectService + namespace CRUD |

Everything else is fakes-only unit tests (`internal/jobs/*_test.go`, `verv_services/*_test.go`,
`deploy_watcher_test.go`, transport `*_impl/*_test.go`).

---

## 2. Coverage vs. product surface

### 2.1 Basic — pure docker wrapper (no verv stack)

| Capability | Status | Note |
|---|---|---|
| Create / list / drop container, standard image | covered | `Test_Lifecycle`, `Test_DropSmerd_ByUuid` |
| Healthcheck config | covered | hw+healthcheck, postgres |
| Image-declared ports (`UseImagePorts`) | covered | nginx, postgres |
| Explicit port bindings + host reachability | thin | only as a side effect of the network-isolation test; no basic-suite reachability assertion |
| Env-var injection into container | thin | postgres passes `Env`; nothing asserts it landed in the container |
| Restart policy | **gap** | only in the skipped Loki subtest |
| `Plain` / `FileConfig` file injection | **gap** | only in the skipped Loki subtest |
| Volumes: create / mount / `CopyToVolume` | **gap** | fakes only (`internal/jobs/copy_to_volume_test.go`); no live RPC caller (open decision #2 in `testing.md`) |
| Upgrade (default + suffixed env) | covered | `Test_UpgradeSmerd_HappyPath`, `Test_UpgradeSmerd_InSuffixedEnvironment` |
| Multi-environment scoping | covered | `Test_Environments` (incl. same name in two envs → distinct containers), `Test_ServiceScoping` |
| `GetHardware` | unit only | `hardware_manager_test.go` |
| `Version`, `SearchImages` | **gap** | none |
| `MakeConnections` / `BreakConnections` | **gap** | none anywhere — confirm live vs dead |
| Failure paths (bad image, port clash, healthcheck timeout, dup name) | **gap** | no negative deploy test |

### 2.2 Basic + verv stack (matreshka reachable, single node)

| Capability | Status | Note |
|---|---|---|
| Node boots matreshka sidecar | covered | `WithMatreshka` shared singleton fixture |
| `AssembleConfig` for a verv image | covered | generation only |
| Container **actually receives** matreshka config (file + env) at launch | **gap** | every ClusterMode subtest sets `IgnoreConfig:true`; default-config subtest checks labels/`VERV_NAME` only |
| Config hot-reload / subscription | **gap** | product code commented out (VERV-128) |
| matreshka down → deploy behaviour | **gap** | no circuit-breaker (roadmap §4) |

### 2.3 Advanced — cluster mode + verv stack

| Capability | Status | Note |
|---|---|---|
| ClusterMode create hw/nginx/pg | shallow | only proves "works with a matreshka container present", config ignored |
| Multi-service verv cluster + dependency wiring | covered | `Test_HelloWorldCluster` |
| enable `statefull_pg` (cluster PG, migrations, DSN, ListPlugins, ListDeployments) | covered | `Test_EnableStatefull` (Phase C) — real cluster Postgres via the `ClusterPgDsn` host seam; migrations rolled, root/node DSN populated, `statefull_pg` == running |
| CreateService → CreateDeploy → deploy-watcher → running container | partial | Phase C `Test_ServiceLifecycle`: `CreateService` + `CreateDeploy(New)` RPCs green, deployment + spec persisted to cluster Postgres. Watcher-driven `SCHEDULED_DEPLOYMENT→RUNNING` container **not** covered — `TODO(#125)`; the deploy watcher binds storage at startup, pre-swap |
| ListDeployments + status transitions | partial | Phase C: persistence + `ServiceName` join asserted, status `SCHEDULED_DEPLOYMENT`. Live status transitions still **gap** (`TODO(#125)`) |
| UpgradeDeploy / scheduled upgrade / auto-rollback | **gap** | fakes only; `TODO(#125)` — blocked by the same pre-swap storage binding as the watcher path |
| Multi-node / `ConnectSlave` / scheduling tags | **gap** | `nodeId` hardcoded to 1 (roadmap §4) |
| `ConnectServiceToVpn` sidecar | **gap** | jobs fakes only; e2e skipped |

### 2.4 Networking

| Capability | Status | Note |
|---|---|---|
| Docker bridge net create + attach | covered | `Test_HelloWorldCluster`, `ContainerRuntime` |
| Per-env network isolation `verv_<suffix>` | covered | `Test_ContainerRuntime_Network_PerEnvironmentIsolation` |
| Aliases / inter-container DNS | covered (implicit) | `Test_HelloWorldCluster` (postgres alias, peer URL) |
| Default `verv` net at node boot | thin | exercised, not asserted |
| `MakeConnections` / `BreakConnections` RPC | **gap** | none |
| Headscale node join / `ConnectSlave` | **gap** | none |
| Headscale `ConnectService` (tailscale sidecar) | **gap** | `Test_Vpn` skipped |
| Headscale namespace CRUD (Create/List/Delete) | **gap** | `Test_Vpn` skipped |
| Headscale `ConnectUser` | **gap** | none |
| VCN `ListPeers` | unit only | `list_peers_test.go` |
| VCN service-mesh DNS (`domain_name`) | **gap** | unimplemented in product (roadmap §4) |

---

## 3. Phased plan

One branch (`verv/e2e-coverage-push`) off `origin/master`, sequential, one commit per phase,
single PR. Each phase verifies with `make test-e2e` unless noted.

One branch, sequential, one commit per phase. Order revised 2026-09-01 (see below): the
service/deployment lifecycle (old Phase 2) runs entirely through cluster Postgres
(`verv_services.CreateNewDeploy`/`UpgradeDeploy` → `dataStorage.TxManager().Execute` with a real
`*sql.Tx`; `local_storage.TxManager()` returns `nil`), so it is blocked on the same
host→DinD cluster-Postgres seam as `Test_EnableStatefull` (old Phase 3). Old Phases 2 and 3
are merged into one **Phase C**. Test-only, unblocked phases (5, 6) go first.

| # | Commit | Notes |
|---|---|---|
| 0 | **Parallelize the e2e suite** — DONE `253c1603` | `t.Parallel()` on the top-level suites, run with `-parallel N`, persistent docker image-cache volume across runs, tighter fixture poll intervals. Wall clock ~255s → ~145s. |
| 1 | **Verv-stack config assertions + fixture gaps + negative paths** — DONE `699f98d0` | `suite_verv_config_test.go` + negative deploy tests on `Test_Lifecycle`. Surfaced product gaps G1/G2/G3 (§6). |
| 5 | **Thin RPC-gap checks** *(next)* | New `suite_rpc_gaps_test.go`: `Version`, `SearchImages`, `GetHardware`, `GetServiceMetrics/Resources/Graph/Environments/Vervonomicon` — call each through the real impl, assert reachable shape (no cluster PG needed; service-read RPCs against a name with no deployment return empty, not error). `MakeConnections`/`BreakConnections`: one docker-network round-trip test if live, else document as `reserved`. |
| 6 | **Environment in jobs entity id** — DONE `[Jobs] refactor: environment-scoped smerd entity ids` | `jobs.SmerdEntityID(suffix, name)` folds the resolved environment suffix into the entity id at `smerd_create.go` / `smerd_upgrade.go` (resolved suffix) and `CreateSmerdStream` / `deploy_watcher.go` (environment name — those paths don't resolve suffixes). `smerd_drop.go` untouched (keys on `uuid.New()`). Empty suffix = bare name, so PROD keys and existing `velez.tasks` rows are unchanged. Un-skipped `Test_SameNameInTwoEnvironments_AreDistinctContainers`; `Test_UpgradeSmerd_InSuffixedEnvironment` and `_ListContainers_ScopesToEnvironment` were already green (fixed by earlier container_runtime phases, not the dedup bug) — stale RED comments updated. Trello: [#127](https://trello.com/c/8u706zaW). |
| C | **Cluster-PG-reachable-from-host seam + service/deployment lifecycle + EnableStatefull** *(merged old 2+3)* — DONE | Reused the existing unused `EnvironmentConfig.ClusterPgDsn` field as an advertise-address override (no proto/codegen/new field). `getRootDsnJob.applyBareBinaryHostPort` reads Host+Port from it on the `!env.IsInContainer` branch; empty (prod default) keeps byte-identical `localhost:<exposed port>`. Harness: `dindClusterPgPort=30020` (outside the PortManager band), `WithClusterPgDsn` opt, shared `enableStatefullPgUnderDind` helper (`helper_statefull_test.go`) with sidecar+volume pre-clean. Un-skipped `Test_EnableStatefull`; new `suite_service_lifecycle_test.go` covers `CreateService` + `CreateDeploy(New)` + cluster-PG persistence + `ListDeployments` `ServiceName` join. **Follow-up (C follow-up in §5): the storage-binding fix is now included** — deploy watcher & create_service handler resolve storage live per call, `TODO(#125)` cleared, `Test_ServiceDeploymentLifecycle` now drives the full lifecycle to `RUNNING` and through `CreateDeploy(Upgrade)` → `RUNNING`. Trello: [#125](https://trello.com/c/KtAoIAhf). |
| 4 | **Real headscale fixture** *(last)* | Shared per-binary headscale singleton in the DinD harness, un-skip `Test_Vpn` + namespace CRUD + `ConnectService`. Deferred to the end. Trello: [#126](https://trello.com/c/B7RAnMJa) — carries the user-provided image + config. |

---

## 4. Deferred — Phase 4: real headscale (Trello #126)

User's call: land every other phase first. Stand up a real headscale container as a shared
per-test-binary singleton in the DinD harness (same pattern as `tests/e2e/shared_matreshka.go`):
seed the image, bring one up in `TestMain`, publish its port on the bootstrap host, point
`WithStateVcnEnabled()` at that address instead of the hardcoded `http://localhost:8080`. Then
un-skip `tests/e2e/suite_vpn_test.go` and add assertions for namespace Create/List/Delete,
`ConnectService` producing a running tailscale sidecar, and `ConnectUser`.

`server_url: http://vcn.redsock.ru` + fixed `container_name: headscale` + fixed port 8080 will
need test-scoping (published ephemeral port, harness-set `server_url`) the same way the matreshka
fixture solved its name/port collision. Still open: which tailscale sidecar image `ConnectService`
uses and how auth keys get provisioned.

### docker-compose service (user-provided)

```yaml
headscale:
  image: headscale/headscale:0.27.2-rc.1
  container_name: headscale
  ports:
    - 8080:8080
  volumes:
    - ./data/config:/etc/headscale
    - ./data/lib:/var/lib/headscale
    - ./data/run:/var/run/headscale
  command: serve
  healthcheck:
    test: ["CMD", "headscale", "health"]
```

### config.yaml (user-provided)

```yaml
server_url: http://vcn.redsock.ru
listen_addr: 0.0.0.0:8080
metrics_listen_addr: 127.0.0.1:9090
grpc_listen_addr: 127.0.0.1:50443
grpc_allow_insecure: false
noise:
  private_key_path: /var/lib/headscale/noise_private.key
prefixes:
  v4: 100.64.0.0/10
  v6: fd7a:115c:a1e0::/48
  allocation: sequential
disable_check_updates: false
derp:
  server:
    enabled: true
    region_id: 1
    region_code: local
    region_name: "local-derp"
    stun_listen_addr: "0.0.0.0:3478"
    private_key_path: /var/lib/headscale/derp_private.key
  urls:
    - https://controlplane.tailscale.com/derpmap/default
ephemeral_node_inactivity_timeout: 30m
database:
  type: sqlite
  debug: false
  gorm:
    prepare_stmt: true
    parameterized_queries: true
    skip_err_record_not_found: true
    slow_threshold: 1000
  sqlite:
    path: /var/lib/headscale/db.sqlite
    write_ahead_log: true
    wal_autocheckpoint: 1000
log:
  level: info
  format: json
policy:
  mode: file
  path: "/etc/headscale/acl.json"
dns:
  magic_dns: false
  override_local_dns: false
unix_socket: /var/run/headscale/headscale.sock
unix_socket_permission: "0770"
logtail:
  enabled: false
randomize_client_port: false
```

---

## 6. Product gaps found mid-flight (need a decision — not fixed)

Surfaced by Phase 1 tests. All test-only phases assert current behaviour as-is; these
are candidates for Trello cards + product fixes, user's call.

| # | Gap | Where | Impact |
|---|---|---|---|
| G1 | `RestartPolicyType_always` (and `unless_stopped`) both collapse to docker `on-failure` with a capped retry count — never `always`/`unless-stopped`. | `internal/clients/node_clients/docker/dockerutils/parser/restart.go` `FromRestart` | "always restart" silently becomes "restart on failure, max 3". |
| G2 | `healthcheckJob` only inspects `State.Status == "running"`; it never executes `Healthcheck.Command`. | `internal/jobs/create_smerd.go` | A container that stays up with an always-failing healthcheck is reported healthy. `Healthcheck` config is close to inert. |
| G3 | `copyToContainerJob` writes via `dockerutils.WriteToContainer` with no `mkdir -p` of the parent dir (unlike `copy_to_volume.go`'s `copyFileJob`). | `internal/jobs/create_smerd.go:580` | A `Plain` file-config whose path is under a dir absent from the image fails the entire `create_smerd` task at `copy_to_container`. |

### Phase C approach (resolved 2026-09-01, landed 2026-09-01)

The old Phase 3 ("resolve DSN via matreshka SD") leaned on the same `verv://` gRPC resolver
that Phase 1 found "produces zero addresses" from the in-process host test app. Confirmed:
do not fight the resolver. Phase C instead exposed cluster Postgres on a host-reachable DinD
port (`dindClusterPgPort=30020`, deliberately outside the PortManager band) and reused the
already-present but unused `EnvironmentConfig.ClusterPgDsn` config field as an
advertise-address override — no proto change, no codegen, no new field. `getRootDsnJob`
(renamed helper `applyBareBinaryHostPort`) reads only Host+Port back out of it on the
`!env.IsInContainer()` branch; user/pwd/dbname still come from the sidecar's own env vars.
Empty `ClusterPgDsn` is production's default and keeps today's exact
`localhost:<exposed port>` behaviour. Old Phases 2 and 3 landed together because the
service/deployment lifecycle is blocked on the same seam.

**Limitation discovered mid-flight (not fixed — outside the Phase C seam, `TODO(#125)`):**
`internal/app/custom.go` constructs both the deploy watcher (`workers.NewDeployWatcher`, via
`clusterClients.StateManager().Deployments()`) and the create_service handler
(`jobs.NewCreateServiceHandler`, via `StateManager().Services()`) by resolving the storage
backend **once at startup** — the pre-swap `local_storage` backend. `enable_statefull` swaps
the backend under an atomic pointer, but those two already hold the old concrete storage, so
an in-process "enable statefull_pg, then drive a deployment" flow never reaches a RUNNING
container: the watcher is blind to the cluster-Postgres deployment, and `CreateDeploy` only
resolves the service because `Test_ServiceLifecycle` re-seeds it directly through the live
storage container. In production the node restarts with cluster storage already active
before the watcher starts, so this only bites the single-process test. The fix (make both
resolve storage live per call, as `VervService` already does) is a product change; a
follow-up card is recommended. Phase C therefore covers steps 1–4 of the lifecycle
(`enable` → `CreateService` → `CreateDeploy(New)` → persistence + `ListDeployments` join)
and marks the watcher-driven RUNNING transition and `CreateDeploy(Upgrade)` as `TODO(#125)`.

---

## 5. Progress log

Append-only. One line per landed phase so a fresh session can resume from here.

| # | Commit | Result |
|---|---|---|
| 0 | `253c1603` `[E2E] perf: parallelize the e2e suite` | `t.Parallel()` on 8 testify suites + 2 plain container-runtime tests; `Test_ContainerRuntime_Matrix` left serial (shared fixed suffix), `Test_EnableStatefull`/`Test_Vpn` still skipped. `Makefile` `-parallel 4`. `tests/dind/dind.go`: persistent `<name>-cache` volume at `/var/lib/docker`, survives Teardown, corrupt-cache retry guard, `Setup` split into `bringUp()`/`ensureCacheVolume()`. Postgres healthcheck poll 2s→500ms. Wall clock ~255s → ~145s (~1.75x); green ×2 under `-parallel 4`. Pre-commit gate passed (golangci-lint clean, `go test ./...` green). Test-only + Makefile + dind harness; no product code. The one earlier cold FAIL was a SIGTERM from a concurrent e2e run in the same repo, not a flake. |
| 1 | `[E2E] Tests: verv-stack config assertions + negative deploy paths` | New `tests/e2e/suite_verv_config_test.go` (`VervConfigSuite`, `t.Parallel()`): `Test_VervConfig_RenderedEnv` asserts verv classification (`MatreshkaConfigLabel=true`), `VERV_NAME` injection, and a `Plain` mount landing in the container; `Test_VervConfig_PlainFileMounted` asserts exact `Plain` bytes inside the running container; `Test_VervConfig_RestartPolicyApplied` asserts the container `HostConfig.RestartPolicy` (`always` currently maps to docker `on-failure`/retry 3 — asserted as-is, product gap). Real matreshka pre-seed did NOT land: `verv://matreshka` gRPC resolver "produces zero addresses" for both the raw configurator client and the `fetch_config` job under the e2e harness, so `!IgnoreConfig` + `Verv` + `WithMatreshka()` deploy fails at `fetch_config`. Fell back to reachable-only assertions with a `TODO(phase-1)` in the test. `tests/e2e/suite_api_deploy_test.go` `LifecycleSuite`: `Test_Negative_NonExistentImage`, `Test_Negative_PortCollision`, `Test_Negative_HealthcheckNeverHealthy`, `Test_Negative_DuplicateName` — each asserts `CreateSmerd` errors (or dedups) and leaves no running smerd; cleanup via `NewEnvironment`'s label-based `env.clean`. `tests/e2e/helper.go`: added `PostgresImage`/`NginxAlpineImage` consts (goconst). Product gaps surfaced, not fixed: (a) `always`→`on-failure` restart mapping in `parser.FromRestart`; (b) `healthcheckJob` never runs `Healthcheck.Command`, only checks `State.Status`; (c) `create_smerd` `copyToContainerJob` has no `mkdir -p` so a `Plain` path under a dir absent from the image fails the whole task. `make lint` clean; `make test-e2e` green (~79s). |
| 5 | `[E2E] Tests: thin RPC-gap e2e checks` | New `tests/e2e/suite_rpc_gaps_test.go` (`RpcGapsSuite`, one `NewEnvironment(t)` per method, single-node/local_storage — no matreshka, no cluster PG). Subtests: `Test_Version` (asserts non-empty version string); `Test_SearchImages` (`Name:"nginx"` — `image_list.go` always returns an empty `Images` slice, so only an error would surface: `NoError` + non-nil); `Test_GetHardware` (`NoError`, non-nil `Cpu`/`Ram`/`DiskMem` sub-messages only — values are host-dependent); `Test_GetServiceMetrics`/`Resources`/`Graph`/`Environments` (random unique name via `GetServiceName(t)`, all return empty-shape without error for a never-deployed service — assert `NoError` + non-nil + empty result slices); `Test_GetVervonomicon` (unconditional `&Response{}` stub — reachability only, noted in a comment); `Test_MakeAndBreakConnections` — live docker-network round-trip through `ApiGrpcImpl`: `CreateSmerd` (hello_world, `IgnoreConfig:true`) + a raw `bridge` network via the docker client, `MakeConnections{ServiceName: smerd UUID, TargetNetwork: net}` → `ContainerInspect` asserts the net is attached → `BreakConnections` → inspect asserts it is gone; `t.Cleanup` best-effort disconnect + `NetworkRemove`. Nothing reserved/skipped — all 9 subtests pass. The suite entrypoint intentionally omits `t.Parallel()`: the container+network churn in `Test_MakeAndBreakConnections`, run concurrently with the parallel deploy suites, added enough docker-daemon load to intermittently trip the pre-existing Phase-1 `Test_Negative_HealthcheckNeverHealthy` race (product gap G2) — seen once in an early full run, then not again after making this suite serial. No product code touched, no new helper const. `make lint` clean; `make test-e2e` green ×2 back-to-back after the change (~84s, ~87s wall). |
| C | `[E2E] Tests: cluster-pg host seam + service/deployment lifecycle` | **Seam (only product change):** reused the already-present, unused `EnvironmentConfig.ClusterPgDsn` field as an advertise-address override — no proto, no codegen, no new field. `internal/jobs/enable_statefull.go`: `getRootDsnJob` gains an `advertiseDsn string` field (sourced from `h.cfg.Environment.ClusterPgDsn` in `BuildJobs`); the `!env.IsInContainer()` block in `Do()` was extracted to `applyBareBinaryHostPort(pgCfg, cont)` (also clears a `nestif` lint hit) which, when `advertiseDsn != ""`, parses it via `resources.Postgres.ParseFromDsn` and takes **only** Host+Port (user/pwd/dbname still from the container's env vars); empty `advertiseDsn` is unchanged `localhost:<getExposedPgPort>`. `env.IsInContainer()` true ignores it entirely. Backward-compat: empty `ClusterPgDsn` (prod default in `config/config.yaml`) = byte-identical to before. **Unit tests** (`internal/jobs/enable_statefull_test.go`): `TestGetRootDsnJob_AdvertiseDsn_OverridesHostPort` (override wins, user/pwd still from env), `TestGetRootDsnJob_NoAdvertiseDsn_UsesLocalhostAndExposedPort` (kept), shared `newGetRootDsnInspectResp()` helper (kills goconst dupes). **Harness:** `tests/e2e/dind_ports.go` `dindClusterPgPort=30020` published by the DinD daemon but deliberately kept OUT of the PortManager band; `tests/e2e/helper_environment.go` `WithClusterPgDsn(dsn)` opt (post-config pass); new `tests/e2e/helper_statefull_test.go` (`_test.go`, not plain `.go`, because it needs `repoRoot` which lives in a test file) with `enableStatefullPgUnderDind(t, suffix)` — resolves the host-reachable sidecar addr, `t.Chdir`s to repo root (goose `./migrations`), brings up the env with the seam DSN, force-removes any stale sidecar container+volume left by the branch-named reused DinD (they carry a stale generated pg password → `password authentication failed`), runs `EnablePlugin(statefull_pg)` to `DONE`. **Suites:** un-skipped `Test_EnableStatefull` (`suite_enable_statefull_test.go`) — removed the `t.Skip` + stale `TODO(dind-harness)` block, happy path now wired through the shared helper, existing assertions + `Test_EnableStatefullMode_UnsupportedPlugin_Fails` kept, suite non-parallel; new `suite_service_lifecycle_test.go` (`ServiceLifecycleSuite`, `Test_ServiceLifecycle` entrypoint, not `t.Parallel()`, suffix `e2e-svc-lifecycle`, payload free of `.Config` per the oneof/json trap) covering `CreateService` + `CreateDeploy(New)` through the real `ServiceApiImpl`, cluster-PG persistence of the deployment + its spec, and `ListDeployments` `ServiceName`-join read-back with status `SCHEDULED_DEPLOYMENT`. **Not covered — `TODO(#125)` in the suite doc + a `t.Log`:** watcher-driven `SCHEDULED_DEPLOYMENT→RUNNING` container and `CreateDeploy(Upgrade)`. Root cause (see §6): `internal/app/custom.go` binds both `workers.NewDeployWatcher` and `jobs.NewCreateServiceHandler` to the pre-swap `local_storage` backend at startup; `enable_statefull`'s atomic-pointer swap doesn't reach them, so an in-process enable-then-deploy flow can't drive the watcher. `Test_ServiceLifecycle` re-seeds the service directly through the live storage container to get past `CreateDeploy`'s service lookup. Recommend a follow-up card to make both resolve storage live per call (as `VervService` already does). **Verify:** `go build ./...` clean; `go test ./internal/jobs/... ./internal/workers/... ./internal/transport/...` green; `make lint` 0 issues; `graphify update .` run (kept OUT of the commit); `make test-e2e` green — `ok go.vervstack.ru/Velez/tests/e2e 103.970s`, ~1:47 wall. The pre-existing `Test_Negative_HealthcheckNeverHealthy` flake did not recur in the final full run. |
| 6 | `[Jobs] refactor: environment-scoped smerd entity ids` | Added `jobs.SmerdEntityID(suffix, name string) string` in `internal/jobs/entity_id.go` (+ `entity_id_test.go`: `SmerdEntityID("", "svc") == "svc"`, `SmerdEntityID("env", "svc") == "env/svc"`) — empty suffix returns the bare name so existing `velez.tasks` rows and single-environment / unconfigured-`ContainerSuffix` production keys are byte-identical. Call sites changed: `smerd_create.go` and `smerd_upgrade.go` now capture the suffix from `resolveEnvironment` (was discarded) and enqueue/watch on `SmerdEntityID(suffix, name)`; `tasks_api_impl` `CreateSmerdStream` gains a `service.VervServicesService` dep (wired in `custom.go`) and composes the same id; `deploy_watcher.go` `deploy()`/`upgrade()` compose with the spec's environment *name* (this worker deliberately never resolves suffixes — create_smerd's jobs do at run time — so name-scoping is used, with a `TODO(#127)`). Left alone: `smerd_drop.go` (keys on `uuid.New()`, no name dedup) and `tasks_api_impl.WatchTask` (pure reader, caller passes the composed id). Tests: un-skipped `Test_SameNameInTwoEnvironments_AreDistinctContainers` (`suite_environments_test.go`) — now green with the fix; `Test_UpgradeSmerd_InSuffixedEnvironment` and `Test_ContainerRuntime_ListContainers_ScopesToEnvironment` were verified already green on the pre-fix branch (fixed by earlier container_runtime phases, not the dedup bug — different names, no collision) so only their stale RED doc comments were updated. Updated `deploy_watcher_test.go` two entity-id assertions to the scoped form. `go build ./...` clean; `go test ./internal/jobs/... ./internal/transport/... ./internal/workers/...` green; `make lint` 0 issues; `make test-e2e` green (~88s wall). |
| C follow-up | `[Jobs] Fix: resolve storage backend live in deploy watcher and create_service handler` + `[Tests] E2E: complete service deployment lifecycle (RUNNING + upgrade)` | **Storage-binding fix landed** (the `TODO(#125)` limitation from Phase C). `internal/jobs/create_service.go` and `internal/workers/deploy_watcher.go` no longer capture `.Services()` / `.Deployments()` at startup: each now holds the swappable `cluster_clients.ClusterStateManagerContainer` (via narrow consumer-side interfaces `ServicesStorageResolver` / `deploymentsStorageResolver`, mirroring `VervService.environments()` and the local `taskRunner` precedent) and re-resolves per call, so an `enable_statefull` atomic-pointer swap is observed in the same process. `custom.go` passes `StateManager()` instead of `StateManager().Services()`; `NewDeployWatcher` signature unchanged. Unit tests updated with tiny stub resolvers (`stubServicesResolver`, `stubStorageResolver`). **`Test_ServiceDeploymentLifecycle` now covers the full lifecycle end to end:** enable statefull_pg → `CreateService` (manual `UpsertService` seed workaround removed) → `CreateDeploy(New)` → deploy watcher drives `SCHEDULED_DEPLOYMENT` → `RUNNING` (real running container asserted via `ContainerInspect`) → `CreateDeploy(Upgrade)` onto `hello_world:v0.0.15` → watcher drives `SCHEDULED_UPGRADE` → `RUNNING`, container still up on the upgraded image. `TODO(#125)` scope note + `t.Log` removed. `go build ./...` clean; `go test ./internal/workers/... ./internal/jobs/... ./internal/transport/...` green; `make lint` 0 issues; `graphify update .` run (kept OUT of both commits); `make test-e2e` green. |
