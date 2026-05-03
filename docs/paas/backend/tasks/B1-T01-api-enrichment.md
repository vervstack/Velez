# B1-T01 — Extend API for UI wiring (backend)

**Status:** pending

## Goal

Close the API gaps that block full mock replacement in the UI (tracked in T31). Three
additions: richer service listing, a nodes listing endpoint, and VCN peer details.

---

## 1. Enrich `ServiceBaseInfo` in `ListServices`

**Proto file:** `api/grpc/service_api.proto`

`ServiceBaseInfo` currently only carries `name`. Extend it with the fields the UI cards need:

```proto
message ServiceBaseInfo {
  string name        = 1;
  string image_name  = 2;  // primary container image (imageName from latest smerd)
  string status      = 3;  // running | degraded | stopped — derived from smerd status
  string env         = 4;  // value of the 'env' label if set, else empty
}
```

**Backend change:** in `ListServices` handler, for each service join the latest associated
smerd to populate `image_name` and `status`. If no smerd exists, leave them empty.

Run `make codegen` after proto changes to regenerate Go + TypeScript stubs.

---

## 2. Add `ListNodes` endpoint

**Proto file:** `api/grpc/control_plane_api.proto` (or new `node_api.proto`)

Add a simple node listing RPC:

```proto
message NodeInfo {
  string id     = 1;
  string host   = 2;
  string status = 3;  // online | degraded | offline
}

message ListNodesRequest  {}
message ListNodesResponse { repeated NodeInfo nodes = 1; }

service ControlPlaneAPI {
  // existing RPCs ...
  rpc ListNodes(ListNodesRequest) returns (ListNodesResponse) {}
}
```

**Backend change:** implement the handler. For single-node setups, return the local machine's
IP/hostname with `status = "online"`. Multi-node support can be added later.

---

## 3. Add VCN peer details to `ListNamespaces`

**Proto file:** `api/grpc/verv_closed_network.proto`

Extend `Namespace` or add a `GetNamespace` / `ListPeers` RPC that returns connection metrics:

```proto
message VcnPeer {
  string id      = 1;
  string host    = 2;
  string type    = 3;  // gateway | mesh | client
  string status  = 4;  // online | offline
  string latency = 5;  // e.g. "12ms"
  string rx      = 6;  // e.g. "1.2 MiB"
  string tx      = 7;
}

message ListPeersRequest  { string namespace_key = 1; }
message ListPeersResponse { repeated VcnPeer peers = 1; }
```

**Backend change:** implement via headscale client or local WireGuard stats.

---

## Acceptance Criteria

- [ ] `ListServices` response includes `imageName` and `status` for services that have a running smerd
- [ ] `GET /api/control/nodes` (or equivalent) returns at least the local node
- [ ] Go unit tests cover the enrichment join and the single-node case
- [ ] `make codegen` runs clean after proto changes
- [ ] `make lint` passes
- [ ] VCN peer listing endpoint exists and returns at least empty list without error

## Files to create / modify

- `api/grpc/service_api.proto` — extend `ServiceBaseInfo`
- `api/grpc/control_plane_api.proto` — add `ListNodes` RPC
- `api/grpc/verv_closed_network.proto` — add `ListPeers` RPC
- `internal/transport/` — handler implementations for new RPCs
- `internal/service/` — business logic for node listing and smerd join
- Tests alongside each handler

## Notes

- Run `make codegen` before writing any Go — proto changes must be reflected first
- The smerd join for `ServiceBaseInfo` should be best-effort (no error if smerd is absent)
- `ListNodes` single-node implementation can read hostname from `os.Hostname()` and detect
  IP via `net.InterfaceAddrs()`