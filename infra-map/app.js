"use strict";

const W = 1000, H = 500;
const svg = document.getElementById("map");
const panel = document.getElementById("panel");
const NS = "http://www.w3.org/2000/svg";

const state = {
  mode: "map",
  t: { x: 0, y: 0, k: 1 },
  selected: null,
  world: null,
  data: window.INFRA_DATA || null,
};

function el(tag, attrs, parent) {
  const e = document.createElementNS(NS, tag);
  for (const k in attrs) e.setAttribute(k, attrs[k]);
  if (parent) parent.appendChild(e);
  return e;
}

function project(lon, lat) {
  return [(lon + 180) / 360 * W, (90 - lat) / 180 * H];
}

function applyT() {
  const { x, y, k } = state.t;
  for (const g of svg.querySelectorAll(".zoomable")) {
    g.setAttribute("transform", `translate(${x},${y}) scale(${k})`);
  }
}

function clampK(k) { return Math.min(14, Math.max(0.6, k)); }

// ---- pan / zoom ----
(function bindPanZoom() {
  let drag = null;
  svg.addEventListener("pointerdown", ev => {
    if (ev.target.closest(".node,.cluster,.svc,.svcbox")) return;
    drag = { x: ev.clientX, y: ev.clientY };
    svg.classList.add("dragging");
    svg.setPointerCapture(ev.pointerId);
  });
  svg.addEventListener("pointermove", ev => {
    if (!drag) return;
    state.t.x += ev.clientX - drag.x;
    state.t.y += ev.clientY - drag.y;
    drag = { x: ev.clientX, y: ev.clientY };
    applyT();
  });
  svg.addEventListener("pointerup", () => { drag = null; svg.classList.remove("dragging"); });
  svg.addEventListener("wheel", ev => {
    ev.preventDefault();
    const r = svg.getBoundingClientRect();
    const mx = ev.clientX - r.left, my = ev.clientY - r.top;
    const k2 = clampK(state.t.k * Math.exp(-ev.deltaY * 0.0012));
    const f = k2 / state.t.k;
    state.t.x = mx - (mx - state.t.x) * f;
    state.t.y = my - (my - state.t.y) * f;
    state.t.k = k2;
    applyT();
    if (state.mode === "map") { renderMapNodes(); renderMapLinks(); } // re-cluster
  }, { passive: false });
})();

// ---- layers ----
let linksG, landG, nodesG, svcG;
function makeLayers() {
  const root = el("g", { class: "zoomable" }, svg);
  linksG = el("g", {}, root);
  landG = el("g", {}, root);
  nodesG = el("g", {}, root);
  svcG = el("g", { style: "display:none" }, root);
}

function drawLand() {
  landG.textContent = "";
  if (!state.world) return;
  const d = [];
  for (const f of state.world.features) {
    const polys = f.geometry.type === "Polygon" ? [f.geometry.coordinates] : f.geometry.coordinates;
    for (const poly of polys) {
      for (const ring of poly) {
        d.push("M" + ring.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join("L") + "Z");
      }
    }
  }
  el("path", {
    d: d.join(""),
    fill: "#151e28",
    stroke: "#26303d",
    "stroke-width": 0.6,
    "stroke-linejoin": "round",
  }, landG);
}

const NODE_COLORS = { sc: "#3fb96f", germ: "#4da3ff", nl: "#d9a3ff" };
function nodeColor(n) { return NODE_COLORS[n.id] || "#4da3ff"; }

function worldToScreen(wx, wy) {
  return [wx * state.t.k + state.t.x, wy * state.t.k + state.t.y];
}

// ---- map view: links ----
function renderMapLinks() {
  linksG.textContent = "";
  if (state.mode !== "map" || !state.data) return;
  const pos = {};
  for (const n of state.data.nodes) {
    pos[n.id] = project(n.lon, n.lat);
  }
  for (const l of state.data.links) {
    if (!pos[l.a] || !pos[l.b]) continue;
    const [x1, y1] = pos[l.a], [x2, y2] = pos[l.b];
    const mx = (x1 + x2) / 2, my = (y1 + y2) / 2;
    const dx = x2 - x1, dy = y2 - y1;
    const len = Math.hypot(dx, dy) || 1;
    const cx = mx - dy / len * len * 0.18, cy = my + dx / len * len * 0.18;
    const isTail = l.kind === "tailnet";
    el("path", {
      d: `M${x1},${y1} Q${cx},${cy} ${x2},${y2}`,
      fill: "none",
      stroke: isTail ? "#4da3ff" : "#5b6b7d",
      "stroke-width": isTail ? 1.6 / state.t.k : 1 / state.t.k,
      "stroke-dasharray": isTail ? "none" : `${5 / state.t.k},${4 / state.t.k}`,
      opacity: 0.8,
    }, linksG);
  }
}

// ---- map view: nodes with overlap clustering ----
function renderMapNodes() {
  nodesG.textContent = "";
  if (!state.data) return;
  const rScreen = 15;
  const thresh = 44;

  const items = state.data.nodes.map(n => {
    const [wx, wy] = project(n.lon, n.lat);
    const [sx, sy] = worldToScreen(wx, wy);
    return { n, wx, wy, sx, sy, group: null };
  });

  // union-find on screen-distance
  const parent = items.map((_, i) => i);
  const find = i => parent[i] === i ? i : (parent[i] = find(parent[i]));
  for (let i = 0; i < items.length; i++) {
    for (let j = i + 1; j < items.length; j++) {
      if (Math.hypot(items[i].sx - items[j].sx, items[i].sy - items[j].sy) < thresh) {
        parent[find(i)] = find(j);
      }
    }
  }
  const groups = new Map();
  items.forEach((it, i) => {
    const r = find(i);
    if (!groups.has(r)) groups.set(r, []);
    groups.get(r).push(it);
  });

  for (const g of groups.values()) {
    if (g.length === 1) {
      drawNode(g[0], rScreen);
    } else {
      const cx = g.reduce((s, it) => s + it.wx, 0) / g.length;
      const cy = g.reduce((s, it) => s + it.wy, 0) / g.length;
      const r = (rScreen + 7) / state.t.k;
      const c = el("g", { class: "cluster", cursor: "pointer" }, nodesG);
      el("circle", { cx, cy, r, fill: "#233246", stroke: "#4da3ff", "stroke-width": 1.5 / state.t.k, opacity: 0.95 }, c);
      const label = el("text", {
        x: cx, y: cy, "text-anchor": "middle", "dominant-baseline": "central",
        "font-size": 13 / state.t.k, fill: "#d6dee8", "pointer-events": "none",
      }, c);
      label.textContent = g.length;
      const title = el("text", {
        x: cx, y: cy + r + 12 / state.t.k, "text-anchor": "middle",
        "font-size": 11 / state.t.k, fill: "#7c8a9b", "pointer-events": "none",
      }, c);
      title.textContent = `${g.length} nodes — click to zoom`;
      c.addEventListener("click", () => zoomTo(cx, cy, state.t.k * 2.4));
    }
  }
}

function drawNode(it, rScreen) {
  const { n, wx, wy } = it;
  const r = rScreen / state.t.k;
  const g = el("g", { class: "node", cursor: "pointer" }, nodesG);
  const down = n.offline;
  el("circle", {
    cx: wx, cy: wy, r: r + 4 / state.t.k,
    fill: "none", stroke: nodeColor(n), "stroke-width": 1 / state.t.k, opacity: down ? 0.3 : 0.55,
  }, g);
  el("circle", {
    cx: wx, cy: wy, r,
    fill: nodeColor(n), opacity: down ? 0.3 : 0.9,
    stroke: "#0b0f14", "stroke-width": 1.5 / state.t.k,
  }, g);
  const lbl = el("text", {
    x: wx, y: wy + r + 13 / state.t.k, "text-anchor": "middle",
    "font-size": 12 / state.t.k, fill: down ? "#4a5563" : "#d6dee8",
    "font-weight": 600, "pointer-events": "none",
  }, g);
  lbl.textContent = n.name;
  const sub = el("text", {
    x: wx, y: wy + r + 25 / state.t.k, "text-anchor": "middle",
    "font-size": 10 / state.t.k, fill: "#7c8a9b", "pointer-events": "none",
  }, g);
  sub.textContent = n.containers_running != null ? `${n.containers_running} containers` : "";
  g.addEventListener("click", ev => { ev.stopPropagation(); showNodePanel(n); });
}

function zoomTo(wx, wy, k2) {
  const r = svg.getBoundingClientRect();
  const mx = r.width / 2, my = r.height / 2;
  state.t.x = mx - wx * k2;
  state.t.y = my - wy * k2;
  state.t.k = clampK(k2);
  applyT();
  renderMapNodes();
}

// ---- info panel ----
function esc(s) { return String(s == null ? "" : s).replace(/[&<>"]/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c])); }

function bar(pct, cls) {
  return `<div class="bar ${cls}"><div style="width:${Math.min(100, Math.max(0, pct)).toFixed(0)}%"></div></div>`;
}

function showNodePanel(n) {
  state.selected = { type: "node", id: n.id };
  const ramPct = n.ram_total_mb ? (1 - (n.ram_avail_mb || 0) / n.ram_total_mb) * 100 : null;
  const diskPct = n.disk_total_gb ? n.disk_used_gb / n.disk_total_gb * 100 : null;
  panel.innerHTML = `
    <button class="close" onclick="closePanel()">×</button>
    <h2>${esc(n.name)} <span style="color:${nodeColor(n)}">●</span></h2>
    <div class="sub">${esc(n.title || "")} — ${esc(n.location || "")}${n.offline ? " — <b style='color:var(--bad)'>offline</b>" : ""}</div>
    <dl class="kv">
      <dt>role</dt><dd>${esc(n.role || "")}</dd>
      <dt>hostname</dt><dd>${esc(n.hostname)}</dd>
      <dt>os</dt><dd>${esc(n.os)}</dd>
      <dt>public ip</dt><dd>${esc(n.public_ip || "—")}</dd>
      <dt>tailnet ip</dt><dd>${esc(n.tailnet_ip || "—")}</dd>
      <dt>cpu</dt><dd>${esc(n.cpu_model)} (${esc(n.cpu_cores)} cores)</dd>
      <dt>load</dt><dd>${esc((n.load || []).join(" / "))}</dd>
      <dt>uptime</dt><dd>${esc(n.uptime_days)} days</dd>
      <dt>docker</dt><dd>${esc(n.docker_version || "—")}</dd>
      <dt>containers</dt><dd>${esc(n.containers_running)} running / ${esc(n.containers_total)} total</dd>
    </dl>
    ${ramPct != null ? `<h3>RAM ${esc(n.ram_total_mb)} MB</h3>${bar(ramPct, "ram")}` : ""}
    ${diskPct != null ? `<h3>Disk ${esc(n.disk_used_gb)} / ${esc(n.disk_total_gb)} GB</h3>${bar(diskPct, "disk")}` : ""}
    ${n.note ? `<h3>note</h3><div class="sub">${esc(n.note)}</div>` : ""}
    <h3>containers</h3>
    <div>${(n.containers || []).map(c => `<span class="chip ${esc(c.state)}" title="${esc(c.image)} — ${esc(c.status)}">${esc(c.name)}</span>`).join("")}</div>
    <h3>docker networks</h3>
    <div>${(n.networks || []).map(x => `<span class="chip net">${esc(x.name)} <span style="color:var(--dim)">${x.containers.length}</span></span>`).join("")}</div>
  `;
  panel.classList.add("open");
}

function showServicePanel(nodeId, svc) {
  const node = state.data.nodes.find(x => x.id === nodeId);
  panel.innerHTML = `
    <button class="close" onclick="closePanel()">×</button>
    <h2>${esc(svc.name)}</h2>
    <div class="sub">${esc(node ? node.name : "")} — ${svc.containers.length} containers</div>
    <h3>containers</h3>
    ${svc.containers.map(c => `
      <div><span class="chip ${esc(c.state)}">${esc(c.name)}</span>
      <span class="sub">${esc(c.image)} — ${esc(c.status)}</span></div>`).join("")}
    <h3>networks</h3>
    <div>${svc.networks.map(x => `<span class="chip net">${esc(x)}</span>`).join("")}</div>
  `;
  panel.classList.add("open");
}

window.closePanel = function () {
  panel.classList.remove("open");
  state.selected = null;
};

// ---- service view ----
function svcColor(name) {
  let h = 0;
  for (const ch of name) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
  return `hsl(${h % 360} 45% 55%)`;
}

function groupServices(node) {
  const map = new Map();
  for (const c of node.containers || []) {
    const key = c.project || c.name;
    if (!map.has(key)) map.set(key, { name: key, containers: [], networks: new Set(), node: node.id });
    const s = map.get(key);
    s.containers.push(c);
    c.networks.forEach(x => s.networks.add(x));
  }
  for (const s of map.values()) {
    s.networks = [...s.networks].filter(x => x && x !== "bridge" && x !== "host" && x !== "none");
    s.containers.sort((a, b) => a.name.localeCompare(b.name));
  }
  return [...map.values()].sort((a, b) => b.containers.length - a.containers.length);
}

function renderServiceView() {
  svcG.textContent = "";
  linksG.textContent = "";
  nodesG.textContent = "";
  if (!state.data) return;

  const perNode = state.data.nodes.map(n => ({ node: n, services: groupServices(n) }));
  const colW = 340, rowH = 130, padX = 60, padY = 90;

  const svcPos = new Map();
  perNode.forEach((pn, ci) => {
    const x0 = padX + ci * colW;
    // node label
    const t = el("text", { x: x0, y: 30, "font-size": 15, "font-weight": 700, fill: nodeColor(pn.node) }, svcG);
    t.textContent = `${pn.node.name} — ${pn.node.hostname || ""}`;
    pn.services.forEach((s, ri) => {
      const n = s.containers.length;
      const r = Math.max(30, Math.min(64, 24 + n * 6));
      const cx = x0 + colW / 2 - 40;
      const cy = padY + ri * rowH + r;
      s.cx = cx; s.cy = cy; s.r = r;
      svcPos.set(pn.node.id + "/" + s.name, s);
    });
  });

  // links: services sharing a docker network (same node solid, cross node dashed)
  const byNet = new Map();
  for (const pn of perNode) {
    for (const s of pn.services) {
      for (const net of s.networks) {
        if (!byNet.has(net)) byNet.set(net, []);
        byNet.get(net).push(s);
      }
    }
  }
  for (const [net, list] of byNet) {
    if (list.length < 2) continue;
    for (let i = 0; i < list.length - 1; i++) {
      const a = list[i], b = list[i + 1];
      const cross = a.node !== b.node;
      el("path", {
        d: `M${a.cx},${a.cy} L${b.cx},${b.cy}`,
        stroke: cross ? "#d9a3ff" : "#445468",
        "stroke-width": cross ? 1.4 : 1,
        "stroke-dasharray": cross ? "6,5" : "none",
        opacity: 0.7, fill: "none",
      }, svcG).setAttribute("data-net", net);
    }
  }

  // service circles + container squares
  for (const pn of perNode) {
    for (const s of pn.services) {
      const g = el("g", { class: "svc", cursor: "pointer" }, svcG);
      const anyRunning = s.containers.some(c => c.state === "running");
      el("circle", {
        cx: s.cx, cy: s.cy, r: s.r,
        fill: "#151e28", opacity: anyRunning ? 1 : 0.45,
        stroke: svcColor(s.name), "stroke-width": 1.6,
      }, g);
      // squares in grid inside circle
      const size = 13, gap = 3;
      const cols = Math.max(1, Math.floor((s.r * 1.5) / (size + gap)));
      s.containers.forEach((c, i) => {
        const col = i % cols, row = Math.floor(i / cols);
        const x = s.cx - (cols * (size + gap) - gap) / 2 + col * (size + gap);
        const y = s.cy - 8 + row * (size + gap) - Math.floor(s.containers.length / cols) * 0;
        const rect = el("rect", {
          x, y, width: size, height: size, rx: 2,
          fill: c.state === "running" ? svcColor(s.name) : "#5c2b29",
          opacity: c.state === "running" ? 0.95 : 0.8,
        }, g);
        rect.setAttribute("class", "svcbox");
        const title = el("title", {}, rect);
        title.textContent = `${c.name}\n${c.image}\n${c.status}`;
      });
      const lbl = el("text", {
        x: s.cx, y: s.cy + s.r + 14, "text-anchor": "middle",
        "font-size": 11, fill: "#d6dee8", "pointer-events": "none",
      }, g);
      lbl.textContent = s.name.length > 26 ? s.name.slice(0, 25) + "…" : s.name;
      g.addEventListener("click", ev => { ev.stopPropagation(); showServicePanel(pn.node.id, s); });
    }
  }
}

// ---- mode toggle ----
function setMode(mode) {
  state.mode = mode;
  document.querySelectorAll("#topbar .seg button").forEach(b => {
    b.classList.toggle("active", b.dataset.mode === mode);
  });
  landG.style.display = mode === "map" ? "" : "none";
  svcG.style.display = mode === "services" ? "" : "none";
  if (mode === "map") {
    state.t = { x: 0, y: 0, k: 1 };
    applyT();
    renderMapLinks();
    renderMapNodes();
  } else {
    linksG.textContent = "";
    renderServiceView();
  }
  window.closePanel();
}

// ---- boot ----
async function boot() {
  makeLayers();
  try {
    const res = await fetch("data/world-110m.json");
    const topo = await res.json();
    state.world = topojsonFeature(topo, topo.objects.countries);
  } catch (e) { /* land stays blank */ }
  drawLand();

  document.querySelectorAll("#topbar .seg button").forEach(b => {
    b.addEventListener("click", () => setMode(b.dataset.mode));
  });
  svg.addEventListener("click", ev => {
    if (!ev.target.closest(".node,.cluster,.svc")) window.closePanel();
  });

  if (!state.data) {
    document.getElementById("stamp").textContent = "data/infra-data.js missing — run ./scan.sh";
    return;
  }
  document.getElementById("stamp").textContent = "scanned " + state.data.generated_at;
  setMode("map");
}

boot();
