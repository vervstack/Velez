# T14 — VCNPeerRow

**Files to create:**

- `src/components/vcn/VCNPeerRow.tsx`
- `src/components/vcn/VCNPeerRow.module.css`

## What it looks like

A single row in the VCN peer table. Grid with 6 columns.

```
node01 ↔ node02  │  mesh   │  1.2ms  │  [active]  │  2.4 GB  │  1.8 GB
node02 ↔ node03  │  mesh   │  19ms   │ [degraded] │  120 MB  │   88 MB
```

Column layout: `1fr 100px 80px 80px 80px 80px`
→ `name | type | latency | status | rx | tx`

## Data types

```ts
interface VCNPeerData {
    id: string;
    name: string;
    type: 'mesh' | 'gateway' | 'client';
    latency: string;   // formatted, e.g. "1.2ms" or "—"
    status: 'active' | 'degraded';
    rx: string;        // formatted, e.g. "2.4 GB"
    tx: string;
}
```

## Props interface

```ts
interface VCNPeerRowProps {
    peer: VCNPeerData;
    last?: boolean;   // omit bottom border on last row
}
```

## Component

```tsx
// src/components/vcn/VCNPeerRow.tsx
import cls from '@/components/vcn/VCNPeerRow.module.css';
import cn from 'classnames';
import Badge from '@/components/base/Badge';

export interface VCNPeerData {
    id: string;
    name: string;
    type: 'mesh' | 'gateway' | 'client';
    latency: string;
    status: 'active' | 'degraded';
    rx: string;
    tx: string;
}

interface VCNPeerRowProps {
    peer: VCNPeerData;
    last?: boolean;
}

export default function VCNPeerRow({ peer, last }: VCNPeerRowProps) {
    const statusColor = peer.status === 'active' ? 'var(--green)' : 'var(--amber)';

    return (
        <div className={cn(cls.VCNPeerRowContainer, { [cls.last]: last })}>
            <span className={cls.name}>{peer.name}</span>
            <span className={cls.type}>{peer.type}</span>
            <span className={cls.latency}>{peer.latency}</span>
            <div>
                <Badge label={peer.status} color={statusColor} />
            </div>
            <span className={cls.traffic}>{peer.rx}</span>
            <span className={cls.traffic}>{peer.tx}</span>
        </div>
    );
}
```

## CSS

```css
/* src/components/vcn/VCNPeerRow.module.css */
.VCNPeerRowContainer {
    display: grid;
    grid-template-columns: 1fr 6.25rem 5rem 5rem 5rem 5rem;
    padding: 0.8125rem 1.25rem;
    border-bottom: 1px solid var(--border);
    align-items: center;
}

.VCNPeerRowContainer.last {
    border-bottom: none;
}

.name {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--fg);
}

.type {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
}

.latency {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-mid);
}

.traffic {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
}
```
