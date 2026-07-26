# Quickstart: Multi-Node

Velez's proto surface (`api/grpc/control_plane_api.proto`, `api/grpc/verv_closed_network.proto`)
defines a full multi-node story: a control plane, nodes joining a cluster, a Verv Closed Network
(VCN) built on Headscale, and services that span nodes. **Read this section before following any
steps below** — several of those RPCs are defined but not implemented yet, and this doc will not
paper over that.

## What's real today vs. what's a stub

Grounded directly in the current code (not the proto comments, which describe intent):

| Capability | Status | Evidence |
|---|---|---|
| A node self-hosts (or connects to) a Headscale VCN server on boot | **Works** | `internal/cluster/verv_closed_network/server.go` `SetupVcn` |
| `VcnApi.CreateNamespace` / `ListNamespaces` | **Works** (real headscale HTTP calls) | `internal/clients/cluster_clients/headscale/ns_create.go`, `ns_list.go` |
| `VcnApi.ConnectService` — attach a service to the VCN via a Tailscale sidecar | **Works** | `internal/jobs/connect_service_to_vpn.go`, `internal/transport/vcn_api_impl/service_connect.go` |
| `VcnApi.ConnectUser` | **Stub — no-op.** Returns success but does nothing. | `internal/clients/cluster_clients/headscale/node_register.go`: the entire body is commented out; `RegisterNode` just `return nil` |
| `VcnApi.ListPeers` | **Stub — hardcoded empty list.** Comment: "headscale integration is future work" | `internal/transport/vcn_api_impl/list_peers.go` |
| `ControlPlaneAPI.ConnectSlave` ("used by other Velez nodes to connect to cluster") | **Not implemented.** No handler exists — only `UnimplementedControlPlaneAPIServer` is embedded, so this RPC returns gRPC `Unimplemented`. | `internal/transport/control_plane_api_impl/` (no `connect_slave.go`); `impl.go` |
| `ControlPlaneAPI.EnablePlugin` with `headscale_server` payload | **Not implemented.** The handler's switch only handles `VervPluginType_statefull_pg`; every other plugin type (including `headscale`) falls through to `errUnsupportedService`. | `internal/transport/control_plane_api_impl/enable_plugin.go` |
| `ControlPlaneAPI.ListNodes` | Wired to storage, but **nothing in the codebase ever writes to it** — no node self-registers. Expect an empty list regardless of how many Velez processes are running. | `internal/service/service_manager/nodes_service/service.go` (read-only; no writer anywhere in `internal/`) |
| `CreateDeploy` targeting a specific node | **Not possible.** The request has no node field, and deployment creation hardcodes `NodeID: 1`. | `internal/service/service_manager/verv_services/deploy.go` |
| Config keys `vpn_server_url`, `master_node_address`, `cluster_pg_dsn` | **Declared but dead.** Present in `config/config.yaml`'s schema (`internal/config/environment.go`) but never read by any other code in the repo. | grep confirms zero non-definition references |

**Bottom line:** there is no automated "join this node to that cluster" flow today. What follows is
the closest working approximation, built from the pieces above, with the manual steps and their
caveats spelled out explicitly. Treat this as "how the VCN plumbing works today," not "how to run
a production multi-host cluster."

## Prerequisites

Same as `docs/quickstart/single-node.md`, on two (or more) separate Docker hosts:

- Docker + the same required bind mounts (`docker.sock`, `/dev/disk`, `/run`).
- Port `53890` free on each host for Velez's own API.
- Port `8080` free on the host that will run the Headscale server — that's `ApiPort` in
  `internal/patterns/headscale/headscale.go`, and it is **not** published by the docker run command
  in the root `README.md`; you must add `-p 8080:8080` yourself on that node.
- The `local_state_path` volume (`~/velez:/tmp/velez` in the README's docker run example) mounted
  and persisted on **every** node — this file is the only place VCN join information lives (see
  below), and losing it means re-doing the manual join.

## 1. Start the first node and let it self-host Headscale

Start Velez on node A exactly as in the single-node guide:

```shell
docker run \
   -p 53890:53890 \
   -p 8080:8080 \
   -d \
   --name velez \
   -v /var/run/docker.sock:/var/run/docker.sock \
   -v /dev/disk:/dev/disk \
   -v /run:/run \
   -v ~/velez:/tmp/velez \
   godverv/velez:latest
```

On first boot, `cluster.Setup` → `verv_closed_network.SetupVcn` checks node A's local state file
for `Network.Headscale.ServerUrl`. It's empty on a fresh node, so Velez deploys its own
`headscale/headscale:0.27.2-rc.1` container (named `headscale`, group label
`verv_closed_network`) and self-registers as that server's client
(`internal/clients/cluster_clients/headscale/service.go` `ConnectToContainer`), writing the
resulting URL and API key back into node A's own local state:

```shell
cat ~/velez/local_state.json
```

```json
{
  "Network": {
    "Headscale": {
      "ServerUrl": "http://localhost:8080",
      "Key": "<issued headscale API key>"
    }
  }
}
```

> **Caveat:** `ServerUrl` is derived from `getApiAddress` (`internal/clients/cluster_clients/headscale/service.go`),
> which returns `http://localhost:<port>` when Velez isn't itself containerized, or an internal
> Docker-network alias (`http://<alias>:8080`) when it is. **Neither is reachable from a second
> physical host.** If you want node B to reach node A's headscale server, replace the host part of
> `ServerUrl` with node A's actual reachable IP before using it in step 2 (the port, `8080` by
> default, is correct as long as you published it as shown above).

## 2. Point a second node at node A's Headscale server (manual — no join RPC exists)

Since `ConnectSlave` and the `EnableHeadscaleServer` plugin path aren't implemented, the only
mechanism the code actually supports is seeding node B's local state file with node A's Headscale
`ServerUrl`/`Key` **before Velez starts on node B for the first time** — `SetupVcn` only takes the
"connect as client" branch when both fields are already non-empty:

```go
// internal/cluster/verv_closed_network/server.go
if headscaleLocalConfig.Key != "" && headscaleLocalConfig.ServerUrl != "" {
    client, err := headscale.Connect(ctx, headscaleLocalConfig.ServerUrl, headscaleLocalConfig.Key)
    ...
}
```

On node B, before first start, create the local state file at the same `local_state_path` you plan
to mount (default `/tmp/velez/local_state.json`):

```shell
mkdir -p ~/velez
cat > ~/velez/local_state.json <<'EOF'
{
  "Network": {
    "Headscale": {
      "ServerUrl": "http://<node-A-reachable-ip>:8080",
      "Key": "<the Key you copied from node A's local_state.json>"
    }
  }
}
EOF
```

Then start Velez on node B the same way as node A (you can omit `-p 8080:8080` here since node B
isn't hosting its own Headscale server).

If this connects successfully, node B's `SetupVcn` calls `headscale.Connect`, which does a
`ListNamespaces` call against node A's server as a smoke test. If it fails, Velez logs the error and
**silently falls back to a disabled VCN** (`DisabledVcnImpl`) rather than erroring out the boot —
check node B's logs for `"couldn't connect to headscale server"` if VCN operations later fail with
`ErrServiceIsDisabled`.

There is currently no way to verify from the API that node B actually joined — see the gaps table
above (`ListNodes`, `ListPeers` don't reflect this).

## 3. Create a VCN namespace and attach a service to it (this part genuinely works)

Namespaces and service attachment go through the Headscale HTTP API for real, from whichever node
you call them on (it talks to whatever Headscale server that node's local state points at):

```shell
# Create a namespace ("user" in headscale terms)
curl -X POST "http://<node>:53890/api/vcn/namespaces/new" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <VelezKey>" \
  -d '{"name": "my-namespace"}'

# List namespaces
curl -X POST "http://<node>:53890/api/vcn/namespaces/list" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <VelezKey>" \
  -d '{}'

# Attach an existing service to the VCN (spins up a Tailscale sidecar container
# and joins it to the mesh via a headscale preauth key)
curl -X POST "http://<node>:53890/api/vcn/services/connect" \
  -H "Content-Type: application/json" \
  -H "Grpc-Metadata-Authorization: <VelezKey>" \
  -d '{"serviceName": "hello-world"}'
```

`ConnectService` (`api/grpc/verv_closed_network.proto`) enqueues the durable
`connect_service_to_vpn` job (`internal/jobs/connect_service_to_vpn.go`), which creates a
`<service>-<sidecar-suffix>` Tailscale container joined to the namespace, and blocks (up to 60s,
`connectServiceWatchTimeout`) until the job reaches `DONE`/`FAILED`. This is the real, working
per-service piece of "put something on the private network" — it just doesn't currently compose
into a cross-node deployment story (see next section).

Do **not** call `ConnectUser` (`/api/vcn/users/connect`) expecting it to do anything — it's a
documented no-op today (see the gaps table).

## 4. Deploying a service "across the cluster" — current limits

There is no way today to tell `CreateDeploy` (`api/grpc/service_api.proto`) which node to run on.
`internal/service/service_manager/verv_services/deploy.go`'s `CreateNewDeploy` hardcodes
`NodeID: 1` on every deployment it writes, and the reconciler (`internal/workers/deploy_watcher.go`)
only ever watches `nodeId: 1`. In practice, deployments created via the API run on whichever single
node's `deployWatcher` picks them up from its own storage — there's no cross-node scheduling.

What you *can* do today across nodes:

- Run independent `CreateService`/`CreateDeploy` calls against each node's own `:53890/api`
  directly — each node manages its own services/deployments independently.
- Use `ConnectService` (step 3) on each node so the resulting containers can reach each other over
  the shared Headscale VCN, if you completed the manual join in step 2.
- Point one service's config at another node's service by address/alias over the VCN mesh, the same
  way `internal/cluster/service_discovery/launch.go`'s Makosh setup connects the `makosh` service to
  the VCN via the same `ConnectServiceToVpn` pipeline.

Real cross-node scheduling (a control plane that places a deployment on a chosen node) is not
implemented — track `ControlPlaneAPI.ListNodes`/`ConnectSlave` in `api/grpc/control_plane_api.proto`
for where that would land once built.

## Troubleshooting / gotchas

- **`ErrServiceIsDisabled` from any `/api/vcn/*` or `ConnectService` call.** The calling node's VCN
  client is `DisabledVcnImpl` — either it never had a `ServerUrl` in local state and headscale
  self-deploy failed, or the connect-as-client smoke test in step 2 failed. Check that node's logs
  and `~/velez/local_state.json`.
- **Node B seemed to start fine but nothing about node A shows up anywhere.** Expected — see the
  gaps table. There's no node registry and no peer list populated by this flow; the only observable
  effect of a successful join is that VCN calls on node B succeed instead of returning
  `ErrServiceIsDisabled`.
- **`docker run` for node A didn't publish 8080.** `SetupVcn`/`LaunchHeadscale` binds headscale's
  API port to the host by default, but the README's docker run example only maps `53890`. Add
  `-p 8080:8080` (or whatever `EnableHeadscaleServer.DeployHeadscaleConfig.custom_port` would have
  set, had that plugin path been wired up) explicitly.
- **Local state file lost = VCN identity lost.** Exactly like the `VelezKey` in the single-node
  guide: if `local_state_path` isn't on a persisted volume, a node restart wipes
  `Network.Headscale`, and a self-hosting node will deploy a *second, unrelated* headscale server
  next boot (its old one is still running as an orphaned container) while a joined node falls back
  to `DisabledVcnImpl`.
- **`cluster_pg_dsn` / shared Postgres cluster state.** Same story as `vpn_server_url` — the config
  key isn't wired anywhere. `internal/cluster/cluster_state/master.go`/`worker.go` only activate
  Postgres-backed cluster state (shared `Nodes`/`Deployments`/`Tasks` storage across nodes) if
  `ClusterState.PgRootDsn`/`PgNodeDsn` are already present in local state — which today means
  seeding `local_state.json` by hand, the same way as the Headscale fields above. There is no
  documented, supported procedure for this yet; treat it as out of scope until the control-plane
  RPCs above exist.
