#!/usr/bin/env bash
# Recon scan for infra-map. Read-only over SSH. Regenerates data/infra-data.js (+ .json).
# Usage: ./scan.sh          — scans all nodes from nodes.json
#        ./scan.sh sc germ  — scan only listed node ids
set -o pipefail

cd "$(dirname "$0")"

NODES_FILE="nodes.json"
OUT_JSON="data/infra-data.json"
OUT_JS="data/infra-data.js"

only=("$@")
only_args=()
if [ ${#only[@]} -gt 0 ]; then only_args=("${only[@]}"); fi

# Remote gather — everything read-only. Emits simple key=value / JSON-lines sections.
remote_gather='
set -u
echo "HOST=$(hostname)"
echo "OS=$(. /etc/os-release 2>/dev/null; echo "${PRETTY_NAME:-unknown}")"
echo "KERNEL=$(uname -r)"
echo "ARCH=$(uname -m)"
echo "CPU_MODEL=$(grep -m1 "model name" /proc/cpuinfo 2>/dev/null | cut -d: -f2- | sed "s/^ //")"
echo "CPU_CORES=$(nproc 2>/dev/null || echo "?")"
awk "/MemTotal/ {print \"MEM_TOTAL_KB=\" \$2} /MemAvailable/ {print \"MEM_AVAIL_KB=\" \$2}" /proc/meminfo
df -k / 2>/dev/null | awk "NR==2 {print \"DISK_TOTAL_KB=\" \$2; print \"DISK_USED_KB=\" \$3}"
awk "{printf \"LOAD1=%s\nLOAD5=%s\nLOAD15=%s\n\", \$1, \$2, \$3}" /proc/loadavg
awk -F. "{print \"UPTIME_SEC=\" \$1}" /proc/uptime
docker version --format "DOCKER_SERVER={{.Server.Version}}" 2>/dev/null || echo "DOCKER_SERVER="
docker ps -a --format "{{json .}}" 2>/dev/null | sed "s/^/CONTAINER_JSON=/" || true
for net in $(docker network ls -q 2>/dev/null); do
  docker network inspect -f "{{.Name}}|{{.Driver}}|{{range \$k,\$v := .Containers}}{{\$v.Name}},{{end}}" "$net" 2>/dev/null | sed "s/^/NETWORK_LINE=/"
done
ip4=$(tailscale ip -4 2>/dev/null | head -1)
echo "TS_IP=${ip4:-}"
'

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

node_ids=$(python3 -c '
import json,sys
reg=json.load(open("nodes.json"))
wanted=set(sys.argv[1:])
for n in reg["nodes"]:
    if not wanted or n["id"] in wanted:
        print(n["id"])
' "${only_args[@]}")

for id in $node_ids; do
  ssh_alias=$(python3 -c "
import json
reg=json.load(open('nodes.json'))
print([n['ssh'] for n in reg['nodes'] if n['id']=='$id'][0])
")
  echo "scanning $id ($ssh_alias) ..."
  ssh -o BatchMode=yes -o ConnectTimeout=10 "$ssh_alias" "$remote_gather" > "$tmp_dir/$id.txt" 2>/dev/null \
    || echo "OFFLINE=true" > "$tmp_dir/$id.txt"
done

python3 - "$tmp_dir" "$NODES_FILE" "$OUT_JSON" <<'PYEOF'
import json, sys, os, time

tmp_dir, reg_path, out_path = sys.argv[1], sys.argv[2], sys.argv[3]
reg = json.load(open(reg_path))

def kb_to_mb(v):
    return round(int(v) / 1024) if v and v.isdigit() else None

nodes_out = []
for n in reg["nodes"]:
    raw = {}
    path = os.path.join(tmp_dir, n["id"] + ".txt")
    if os.path.exists(path):
        for line in open(path):
            line = line.rstrip("\n")
            if line.startswith("CONTAINER_JSON="):
                raw.setdefault("containers", []).append(json.loads(line[len("CONTAINER_JSON="):]))
            elif line.startswith("NETWORK_LINE="):
                raw.setdefault("networks", []).append(line[len("NETWORK_LINE="):])
            elif "=" in line:
                k, _, v = line.partition("=")
                raw[k] = v

    ts_status = None
    containers = []
    for c in raw.get("containers", []):
        labels = c.get("Labels") or ""
        label_map = {}
        for pair in labels.split(","):
            if "=" in pair:
                k, _, v = pair.partition("=")
                label_map[k] = v
        project = label_map.get("com.docker.compose.project") or label_map.get("VERV_SERVICE") or ""
        networks = [ns.split("/")[0] for ns in (c.get("Networks") or "").split(",") if ns]
        containers.append({
            "name": c.get("Names", ""),
            "image": c.get("Image", ""),
            "status": c.get("Status", ""),
            "state": c.get("State", ""),
            "project": project,
            "networks": networks,
            "sidecar": label_map.get("SIDECAR") == "true" or "-ts" in c.get("Names", ""),
            "velez_service": label_map.get("VERV_SERVICE", ""),
        })

    networks = []
    for line in raw.get("networks", []):
        parts = line.split("|")
        if len(parts) < 3:
            continue
        members = [m for m in parts[2].split(",") if m]
        networks.append({"name": parts[0], "driver": parts[1], "containers": members})

    running = sum(1 for c in containers if c["state"] == "running")
    node = dict(n)
    node.update({
        "offline": raw.get("OFFLINE") == "true",
        "hostname": raw.get("HOST", ""),
        "os": (raw.get("OS", "") + ", kernel " + raw.get("KERNEL", "") + " (" + raw.get("ARCH", "") + ")").strip(", "),
        "cpu_model": raw.get("CPU_MODEL", "") or "ARM Cortex-A76 (BCM2712)",
        "cpu_cores": int(raw["CPU_CORES"]) if str(raw.get("CPU_CORES", "")).isdigit() else None,
        "ram_total_mb": kb_to_mb(raw.get("MEM_TOTAL_KB")),
        "ram_avail_mb": kb_to_mb(raw.get("MEM_AVAIL_KB")),
        "disk_total_gb": round(int(raw["DISK_TOTAL_KB"]) / 1024 / 1024, 1) if str(raw.get("DISK_TOTAL_KB", "")).isdigit() else None,
        "disk_used_gb": round(int(raw["DISK_USED_KB"]) / 1024 / 1024, 1) if str(raw.get("DISK_USED_KB", "")).isdigit() else None,
        "load": [raw.get("LOAD1"), raw.get("LOAD5"), raw.get("LOAD15")],
        "uptime_days": round(int(raw["UPTIME_SEC"]) / 86400, 1) if str(raw.get("UPTIME_SEC", "")).isdigit() else None,
        "docker_version": (raw.get("DOCKER_SERVER") or "").replace("DOCKER_SERVER=", ""),
        "tailnet_ip": raw.get("TS_IP", ""),
        "containers_running": running,
        "containers_total": len(containers),
        "containers": containers,
        "networks": networks,
    })
    nodes_out.append(node)

# node-level links: tailnet + shared docker network names across nodes
links = []
ids = [n["id"] for n in nodes_out if not n["offline"]]
control = next((n["id"] for n in nodes_out if n.get("headscale_control")), None)
ts_nodes = [n["id"] for n in nodes_out if n.get("tailscale") and not n["offline"]]
if control:
    for t in ts_nodes:
        if t != control:
            links.append({"a": t, "b": control, "kind": "tailnet", "label": "headscale tailnet"})
for i, a in enumerate(ids):
    for b in ids[i + 1:]:
        na = next(n for n in nodes_out if n["id"] == a)
        nb = next(n for n in nodes_out if n["id"] == b)
        sa = {x["name"] for x in na["networks"]} - {"bridge", "host", "none"}
        sb = {x["name"] for x in nb["networks"]} - {"bridge", "host", "none"}
        # *_default networks are compose-local; same name on two nodes is a coincidence, not a link
        shared = sorted(n for n in sa & sb if not n.endswith("_default"))
        for net in shared:
            links.append({"a": a, "b": b, "kind": "docker-net", "label": net})

data = {"generated_at": time.strftime("%Y-%m-%d %H:%M:%S %z"), "nodes": nodes_out, "links": links}
os.makedirs(os.path.dirname(out_path), exist_ok=True)
json.dump(data, open(out_path, "w"), indent=1, ensure_ascii=False)
print("wrote", out_path)
PYEOF

# wrap json into a js file the static page can load without a server
printf 'window.INFRA_DATA = ' > "$OUT_JS"
cat "$OUT_JSON" >> "$OUT_JS"
printf ';\n' >> "$OUT_JS"
echo "wrote $OUT_JS"
