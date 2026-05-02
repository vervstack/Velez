# T23 — NetworkTopologyMap Widget

**Files to create:**

- `src/widgets/vcn/NetworkTopologyMap.tsx`
- `src/widgets/vcn/NetworkTopologyMap.module.css`

## What it looks like

An SVG network diagram inside a panel. Background has a subtle dot grid. Nodes are circles with status colors. Edges are
lines (solid for active, dashed for degraded).

```
┌──────────────────────────────────────────────────────────┐
│ NETWORK TOPOLOGY                                          │
│  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·   │
│       ●node01 ──────── ●node02                           │
│                         │                                │
│               - - - - - ●node03                          │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

Node positions are fixed by `id` — callers must provide known node IDs.

## Data types

```ts
interface TopologyNode {
    id: string;
    host: string;
    status: 'online' | 'degraded' | 'offline';
}

interface TopologyEdge {
    from: string;  // node id
    to: string;    // node id
    status: 'active' | 'degraded';
}
```

## Props interface

```ts
interface NetworkTopologyMapProps {
    nodes: TopologyNode[];
    edges: TopologyEdge[];
}
```

## Component

The SVG viewBox is `0 0 520 220`. Node positions are hard-coded by id in a lookup. New nodes fall back to `[0, 0]` —
this is expected behavior for unknown nodes.

```tsx
// src/widgets/vcn/NetworkTopologyMap.tsx
import cls from '@/widgets/vcn/NetworkTopologyMap.module.css';
import SectionLabel from '@/components/base/SectionLabel';

interface TopologyNode {
    id: string;
    host: string;
    status: 'online' | 'degraded' | 'offline';
}

interface TopologyEdge {
    from: string;
    to: string;
    status: 'active' | 'degraded';
}

interface NetworkTopologyMapProps {
    nodes: TopologyNode[];
    edges: TopologyEdge[];
}

const NODE_POSITIONS: Record<string, [number, number]> = {
    node01: [130, 110],
    node02: [390, 60],
    node03: [390, 170],
};

const STATUS_COLOR: Record<string, string> = {
    online: 'var(--green)',
    degraded: 'var(--amber)',
    offline: 'var(--red)',
};

export default function NetworkTopologyMap({ nodes, edges }: NetworkTopologyMapProps) {
    return (
        <div className={cls.NetworkTopologyMapContainer}>
            <div className={cls.header}>
                <SectionLabel>Network topology</SectionLabel>
            </div>
            <svg
                className={cls.svg}
                viewBox="0 0 520 220"
                preserveAspectRatio="xMidYMid meet"
            >
                <defs>
                    <pattern id="vcn-grid" x="0" y="0" width="30" height="30" patternUnits="userSpaceOnUse">
                        <path d="M 30 0 L 0 0 0 30" fill="none" stroke="var(--border)" strokeWidth="0.5" />
                    </pattern>
                </defs>

                {/* Grid background */}
                <rect width="520" height="220" fill="url(#vcn-grid)" />

                {/* Edges */}
                {edges.map(function renderEdge(edge, i) {
                    const [x1, y1] = NODE_POSITIONS[edge.from] ?? [0, 0];
                    const [x2, y2] = NODE_POSITIONS[edge.to] ?? [0, 0];
                    const color = edge.status === 'degraded' ? 'var(--amber)' : 'var(--cyan)';
                    return (
                        <line
                            key={i}
                            x1={x1} y1={y1}
                            x2={x2} y2={y2}
                            stroke={color}
                            strokeWidth={edge.status === 'degraded' ? 1 : 1.5}
                            strokeOpacity={edge.status === 'degraded' ? 0.5 : 0.4}
                            strokeDasharray={edge.status === 'degraded' ? '5,4' : undefined}
                        />
                    );
                })}

                {/* Nodes */}
                {nodes.map(function renderNode(node) {
                    const [cx, cy] = NODE_POSITIONS[node.id] ?? [0, 0];
                    const color = STATUS_COLOR[node.status] ?? 'var(--fg-dim)';
                    return (
                        <g key={node.id}>
                            <circle cx={cx} cy={cy} r={20} fill="var(--bg2)" stroke={color} strokeWidth="1.5" strokeOpacity="0.7" />
                            <circle cx={cx} cy={cy} r={5} fill={color} fillOpacity="0.85" />
                            <text x={cx} y={cy + 34} textAnchor="middle" fontFamily="var(--font-mono)" fontSize="10" fill="var(--fg-dim)">{node.id}</text>
                            <text x={cx} y={cy + 46} textAnchor="middle" fontFamily="var(--font-mono)" fontSize="9" fill="var(--fg-faint)">{node.host}</text>
                        </g>
                    );
                })}

                {/* Legend */}
                <g transform="translate(400, 188)">
                    <line x1="0" y1="8" x2="18" y2="8" stroke="var(--cyan)" strokeWidth="1.5" strokeOpacity="0.5" />
                    <text x="22" y="11" fontFamily="var(--font-mono)" fontSize="9" fill="var(--fg-dim)">mesh</text>
                    <line x1="60" y1="8" x2="78" y2="8" stroke="var(--amber)" strokeWidth="1" strokeOpacity="0.5" strokeDasharray="4,3" />
                    <text x="82" y="11" fontFamily="var(--font-mono)" fontSize="9" fill="var(--fg-dim)">degraded</text>
                </g>
            </svg>
        </div>
    );
}
```

## CSS

```css
/* src/widgets/vcn/NetworkTopologyMap.module.css */
.NetworkTopologyMapContainer {
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1.5rem;
    overflow: hidden;
}

.header {
    margin-bottom: 1rem;
}

.svg {
    display: block;
    width: 100%;
    max-height: 13.75rem; /* 220px */
}
```

## Notes

- SVG `text` elements use `fontFamily="var(--font-mono)"` inline because SVG presentation attributes cannot reference
  CSS Modules classes, and CSS variables are supported by SVG in modern browsers.
- `NODE_POSITIONS` is static — to support dynamic layouts in a future task, replace it with a computed layout algorithm.
