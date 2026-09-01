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
| `Test_EnableStatefull` | **SKIPPED** | statefull_pg plugin happy-path + ListPlugins/ListDeployments; unsupported-plugin fails. Blocked on host→DinD cluster-postgres DSN seam. |
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
| enable `statefull_pg` (cluster PG, migrations, DSN, ListPlugins, ListDeployments) | **gap** | test written, **skipped** |
| CreateService → CreateDeploy → deploy-watcher → running container | **gap** | fakes only |
| ListDeployments + status transitions | thin | only inside the skipped EnableStatefull test |
| UpgradeDeploy / scheduled upgrade / auto-rollback | **gap** | fakes only |
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
| C | **Cluster-PG-reachable-from-host seam + service/deployment lifecycle + EnableStatefull** *(merged old 2+3)* | Not via the `verv://` SD resolver (Phase 1 found it "produces zero addresses" from the in-process host app). Instead: expose the cluster Postgres on a host-reachable DinD port and inject a DSN override into the test app so `buildRootDsnJob`'s `!env.IsInContainer` branch dials a reachable address. Then: un-skip `Test_EnableStatefull`; new `suite_service_lifecycle_test.go` (`CreateService` → `CreateDeploy` → deploy-watcher → running container → `ListDeployments` RUNNING → `UpgradeDeploy` → status transitions). `go test ./internal/jobs/...`. Trello: [#125](https://trello.com/c/KtAoIAhf). |
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

### Phase C approach (resolved 2026-09-01)

The old Phase 3 ("resolve DSN via matreshka SD") leaned on the same `verv://` gRPC resolver
that Phase 1 found "produces zero addresses" from the in-process host test app. Confirmed:
do not fight the resolver. Phase C instead exposes cluster Postgres on a host-reachable DinD
port and injects a DSN override into the test app so `buildRootDsnJob`'s `!env.IsInContainer`
branch dials a reachable address directly. Old Phases 2 and 3 are done together as Phase C
because the service/deployment lifecycle is blocked on the same seam.

---

## 5. Progress log

Append-only. One line per landed phase so a fresh session can resume from here.

| # | Commit | Result |
|---|---|---|
| 0 | `253c1603` `[E2E] perf: parallelize the e2e suite` | `t.Parallel()` on 8 testify suites + 2 plain container-runtime tests; `Test_ContainerRuntime_Matrix` left serial (shared fixed suffix), `Test_EnableStatefull`/`Test_Vpn` still skipped. `Makefile` `-parallel 4`. `tests/dind/dind.go`: persistent `<name>-cache` volume at `/var/lib/docker`, survives Teardown, corrupt-cache retry guard, `Setup` split into `bringUp()`/`ensureCacheVolume()`. Postgres healthcheck poll 2s→500ms. Wall clock ~255s → ~145s (~1.75x); green ×2 under `-parallel 4`. Pre-commit gate passed (golangci-lint clean, `go test ./...` green). Test-only + Makefile + dind harness; no product code. The one earlier cold FAIL was a SIGTERM from a concurrent e2e run in the same repo, not a flake. |
| 1 | `[E2E] Tests: verv-stack config assertions + negative deploy paths` | New `tests/e2e/suite_verv_config_test.go` (`VervConfigSuite`, `t.Parallel()`): `Test_VervConfig_RenderedEnv` asserts verv classification (`MatreshkaConfigLabel=true`), `VERV_NAME` injection, and a `Plain` mount landing in the container; `Test_VervConfig_PlainFileMounted` asserts exact `Plain` bytes inside the running container; `Test_VervConfig_RestartPolicyApplied` asserts the container `HostConfig.RestartPolicy` (`always` currently maps to docker `on-failure`/retry 3 — asserted as-is, product gap). Real matreshka pre-seed did NOT land: `verv://matreshka` gRPC resolver "produces zero addresses" for both the raw configurator client and the `fetch_config` job under the e2e harness, so `!IgnoreConfig` + `Verv` + `WithMatreshka()` deploy fails at `fetch_config`. Fell back to reachable-only assertions with a `TODO(phase-1)` in the test. `tests/e2e/suite_api_deploy_test.go` `LifecycleSuite`: `Test_Negative_NonExistentImage`, `Test_Negative_PortCollision`, `Test_Negative_HealthcheckNeverHealthy`, `Test_Negative_DuplicateName` — each asserts `CreateSmerd` errors (or dedups) and leaves no running smerd; cleanup via `NewEnvironment`'s label-based `env.clean`. `tests/e2e/helper.go`: added `PostgresImage`/`NginxAlpineImage` consts (goconst). Product gaps surfaced, not fixed: (a) `always`→`on-failure` restart mapping in `parser.FromRestart`; (b) `healthcheckJob` never runs `Healthcheck.Command`, only checks `State.Status`; (c) `create_smerd` `copyToContainerJob` has no `mkdir -p` so a `Plain` path under a dir absent from the image fails the whole task. `make lint` clean; `make test-e2e` green (~79s). |
| 5 | `[E2E] Tests: thin RPC-gap e2e checks` | New `tests/e2e/suite_rpc_gaps_test.go` (`RpcGapsSuite`, one `NewEnvironment(t)` per method, single-node/local_storage — no matreshka, no cluster PG). Subtests: `Test_Version` (asserts non-empty version string); `Test_SearchImages` (`Name:"nginx"` — `image_list.go` always returns an empty `Images` slice, so only an error would surface: `NoError` + non-nil); `Test_GetHardware` (`NoError`, non-nil `Cpu`/`Ram`/`DiskMem` sub-messages only — values are host-dependent); `Test_GetServiceMetrics`/`Resources`/`Graph`/`Environments` (random unique name via `GetServiceName(t)`, all return empty-shape without error for a never-deployed service — assert `NoError` + non-nil + empty result slices); `Test_GetVervonomicon` (unconditional `&Response{}` stub — reachability only, noted in a comment); `Test_MakeAndBreakConnections` — live docker-network round-trip through `ApiGrpcImpl`: `CreateSmerd` (hello_world, `IgnoreConfig:true`) + a raw `bridge` network via the docker client, `MakeConnections{ServiceName: smerd UUID, TargetNetwork: net}` → `ContainerInspect` asserts the net is attached → `BreakConnections` → inspect asserts it is gone; `t.Cleanup` best-effort disconnect + `NetworkRemove`. Nothing reserved/skipped — all 9 subtests pass. The suite entrypoint intentionally omits `t.Parallel()`: the container+network churn in `Test_MakeAndBreakConnections`, run concurrently with the parallel deploy suites, added enough docker-daemon load to intermittently trip the pre-existing Phase-1 `Test_Negative_HealthcheckNeverHealthy` race (product gap G2) — seen once in an early full run, then not again after making this suite serial. No product code touched, no new helper const. `make lint` clean; `make test-e2e` green ×2 back-to-back after the change (~84s, ~87s wall). |
| 6 | `[Jobs] refactor: environment-scoped smerd entity ids` | Added `jobs.SmerdEntityID(suffix, name string) string` in `internal/jobs/entity_id.go` (+ `entity_id_test.go`: `SmerdEntityID("", "svc") == "svc"`, `SmerdEntityID("env", "svc") == "env/svc"`) — empty suffix returns the bare name so existing `velez.tasks` rows and single-environment / unconfigured-`ContainerSuffix` production keys are byte-identical. Call sites changed: `smerd_create.go` and `smerd_upgrade.go` now capture the suffix from `resolveEnvironment` (was discarded) and enqueue/watch on `SmerdEntityID(suffix, name)`; `tasks_api_impl` `CreateSmerdStream` gains a `service.VervServicesService` dep (wired in `custom.go`) and composes the same id; `deploy_watcher.go` `deploy()`/`upgrade()` compose with the spec's environment *name* (this worker deliberately never resolves suffixes — create_smerd's jobs do at run time — so name-scoping is used, with a `TODO(#127)`). Left alone: `smerd_drop.go` (keys on `uuid.New()`, no name dedup) and `tasks_api_impl.WatchTask` (pure reader, caller passes the composed id). Tests: un-skipped `Test_SameNameInTwoEnvironments_AreDistinctContainers` (`suite_environments_test.go`) — now green with the fix; `Test_UpgradeSmerd_InSuffixedEnvironment` and `Test_ContainerRuntime_ListContainers_ScopesToEnvironment` were verified already green on the pre-fix branch (fixed by earlier container_runtime phases, not the dedup bug — different names, no collision) so only their stale RED doc comments were updated. Updated `deploy_watcher_test.go` two entity-id assertions to the scoped form. `go build ./...` clean; `go test ./internal/jobs/... ./internal/transport/... ./internal/workers/...` green; `make lint` 0 issues; `make test-e2e` green (~88s wall). |
