"use strict";

// Minimal topojson.feature for world-atlas countries-110m.json (quantized arcs)
function topojsonFeature(topology, o) {
  const arcs = topology.arcs;
  const tr = topology.transform;
  const sx = tr.scale[0], sy = tr.scale[1];
  const tx = tr.translate[0], ty = tr.translate[1];
  const W = 1000, H = 500;
  function project(lon, lat) {
    return [(lon + 180) / 360 * W, (90 - lat) / 180 * H];
  }
  function decodeArc(idx) {
    const a = arcs[idx < 0 ? ~idx : idx];
    let x = 0, y = 0;
    const pts = a.map(p => {
      x += p[0]; y += p[1];
      return project(x * sx + tx, y * sy + ty);
    });
    return idx < 0 ? pts.reverse() : pts;
  }
  function ring(arcIdxList) {
    let pts = [];
    for (const i of arcIdxList) pts = pts.concat(decodeArc(i));
    // normalize winding to clockwise in screen coords (y grows downward)
    let area = 0;
    for (let i = 0, j = pts.length - 1; i < pts.length; j = i++) {
      area += (pts[j][0] + pts[i][0]) * (pts[j][1] - pts[i][1]);
    }
    return area < 0 ? pts : pts.reverse();
  }
  function rings(poly) {
    return poly.map(ring);
  }
  const feats = o.geometries.map(g => ({
    type: "Feature",
    properties: g.properties,
    geometry: g.type === "Polygon"
      ? { type: "Polygon", coordinates: rings(g.arcs) }
      : { type: "MultiPolygon", coordinates: g.arcs.map(rings) },
  }));
  return { type: "FeatureCollection", features: feats };
}
