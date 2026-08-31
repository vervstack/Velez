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
| `Test_Environments` | partial | default-suffix backfill; two-env list scoping; drop-by-name / drop-by-uuid env scoping. `Test_SameNameInTwoEnvironments_AreDistinctContainers` **skipped** (jobs-engine dedup bug). |
| `Test_ServiceScoping` | yes | StopService / RestartService env-scoped to own container; cross-env no-op |
| `Test_ContainerRuntime_Matrix` + `_ListContainers_ScopesToEnvironment` + `_Network_PerEnvironmentIsolation` | partial | label-runtime naming/suffix; per-env docker network isolation. **`ListContainers` scoping is a RED test** (deliberate unfiltered stub). |
| `Test_UpgradeSmerd` | partial | happy-path upgrade (default env); non-existent fails. **`Test_UpgradeSmerd_InSuffixedEnvironment` is RED.** |
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
| Upgrade (default env) | covered | `Test_UpgradeSmerd_HappyPath` |
| Multi-environment scoping | covered | `Test_Environments`, `Test_ServiceScoping` |
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

| # | Commit | Notes |
|---|---|---|
| 0 | **Parallelize the e2e suite** | `t.Parallel()` on the top-level suites (the shared matreshka singleton + shared PortManager + per-test name suffixes + `-race` global fixes already make this safe), run with `-parallel N`, persistent docker image-cache volume across runs, tighten fixture healthcheck / `require.Eventually` poll intervals. Target: 2–4× wall-clock. Enables fast verification for every later phase. |
| 1 | **Verv-stack config assertions + fixture gaps + negative paths** | New `suite_verv_config_test.go`: deploy a verv image via `WithMatreshka()` **without** `IgnoreConfig`, config pre-seeded in matreshka; assert the rendered config file is in the container and derived env vars are set. Add restart-policy + `Plain` file-config + env-var assertions on a stable image (retire the intent from the dead Loki subtest). Add negative deploy tests to `Test_Lifecycle`: non-existent image, port collision, healthcheck-never-healthy, duplicate name. |
| 2 | **Service/deployment lifecycle e2e** | New `suite_service_lifecycle_test.go`: `CreateService` → `CreateDeploy` → deploy-watcher produces a running container → `ListDeployments` shows RUNNING → `UpgradeDeploy` → status transitions. Real Postgres (reuse the `suite_enable_statefull` disposable-PG pattern). |
| 3 | **EnableStatefull DSN via matreshka SD** | `buildRootDsnJob` (`internal/jobs/enable_statefull.go`, `!env.IsInContainer` branch) resolves `verv://<pg>` through service discovery instead of composing `localhost:<raw-port>`. Un-skip `Test_EnableStatefull`. Also `go test ./internal/jobs/...`. Trello: [#125](https://trello.com/c/KtAoIAhf). |
| 5 | **Thin RPC-gap checks** | `Version`, `SearchImages`, `GetHardware`, `GetServiceMetrics/Resources/Graph/Environments/Vervonomicon`. `MakeConnections`/`BreakConnections`: one round-trip test if live, else `reserved` them. |
| 6 | **Environment in jobs entity id** | Compose `entity_id = "<container-suffix>/<name>"` at all 4 Enqueue/Watch sites (`smerd_create.go`, `smerd_upgrade.go`, `smerd_drop.go`, `deploy_watcher.go`). Empty suffix = bare name, so PROD keys and existing `velez.tasks` rows are unchanged. Un-skip the 3 RED/skipped scoping tests. Also `go test ./internal/jobs/... ./internal/transport/...`. Trello: [#127](https://trello.com/c/8u706zaW). |
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
