# infra-map — vervstack infrastructure map

Static demo of the vervstack node/service topology. No build step, no dependencies, works offline.

## Run

```bash
./scan.sh            # collect fresh data from all nodes (read-only over SSH)
open index.html      # or python3 -m http.server inside this folder
```

`scan.sh` writes two files:

- `data/infra-data.json` — same data, pretty-printed, for other tools
- `data/infra-data.js` — wrapped as `window.INFRA_DATA`, loaded by `index.html`

Open `index.html` directly from disk — no server needed. `data/world-110m.json` is bundled,
so the map works offline.

Rescan one node: `./scan.sh sc`

## What is collected (all read-only)

Per node, over SSH (`BatchMode`, 10s connect timeout; a dead node is marked `offline`):

- hostname, OS, kernel, arch
- CPU model / cores, RAM total + available, disk usage of `/`
- load average, uptime, docker server version
- `docker ps -a` — name, image, status, compose project / `VERV_SERVICE` labels, networks
- `docker network inspect` — network name, driver, member containers
- `tailscale ip -4` if tailscale is installed
- links between nodes:
  - headscale tailnet (from the `tailscale` / `headscale_control` flags in `nodes.json`)
  - docker networks with the same name present on two+ nodes (e.g. `verv`, `metrics`)

## nodes.json — static registry

Coordinates, display names and roles are not discoverable over SSH, so they live here.
Edit `lat`/`lon` to move a node on the map (home `sc` currently uses a placeholder
position — set your real coordinates there).

Flags:

- `tailscale: true` — node runs a tailscale client, gets a tailnet link to the control plane
- `headscale_control: true` — node hosts the headscale control plane

## Views

- **Nodes** — world map (equirectangular), nodes as dots at their coordinates.
  Nodes closer than ~44 px on screen collapse into a cluster dot with a count;
  click it to zoom in and split. Curved blue line = headscale tailnet;
  dashed grey lines = same-named docker network on both ends. Click a node for a
  detail panel (specs, load, RAM/disk bars, IPs, containers, networks).
- **Services** — no map. One column per node, each service (compose project /
  `VERV_SERVICE`) is a circle, containers are squares inside it (grey-red = exited,
  hover for image/status). Solid lines = services sharing a docker network on one
  node; dashed purple = the same network name spanning nodes. Click a service for
  its panel.

## Velez-UI port

Plain JS + SVG, no framework, one data contract (`window.INFRA_DATA` / `infra-data.json`).
To port: reuse `groupServices()` + the panel renderers as-is; replace the SVG pan/zoom
layer with the UI-kit canvas if desired. The data contract is the integration point —
a future Velez RPC can emit the same JSON and the views won't change.

## Suggested improvements for the fleet (found during recon)

1. **Log rotation.** All three nodes run docker's `json-file` logging driver with no
   rotation configured (default = unlimited). Add to `/etc/docker/daemon.json`:
   ```json
   { "log-driver": "json-file", "log-opts": { "max-size": "10m", "max-file": "3" } }
   ```
   then `systemctl reload docker` (does not restart containers).
2. **sc: reclaim ~40 GB.** `docker system df` shows 27.7 GB reclaimable images,
   12.3 GB build cache, 14.5 GB dangling volumes. `docker system prune` (add `-a`
   carefully) and `docker volume prune`.
3. **nl: RAM at ~90 %, no swap.** 1.6 of 1.8 GB used. Either add a small swapfile or
   move something off the box.
4. **Velez version drift.** sc runs `v0.1.120-rc`, germ `v0.1.119-rc`, nl `v0.1.111`.
5. **Stale containers.** `s3_tst_web_ui`, `victoriametrics`, `victoriametrics-ts`,
   `socks5`, `web-certbot-1` (exited 1 — check cert renewal on germ) are long-dead.
6. **germ: 319 days uptime** — pending kernel updates won't apply until reboot.
