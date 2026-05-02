# T24 — VCNPeerTable Widget

**Files to create:**

- `src/widgets/vcn/VCNPeerTable.tsx`
- `src/widgets/vcn/VCNPeerTable.module.css`

## What it looks like

A titled section with a bordered table. Header row + `VCNPeerRow` rows.

```
PEERS

┌─────────────────────────────────────────────────────────────┐
│ Name              │ Type    │ Latency │ Status    │ RX   │ TX │
│──────────────────────────────────────────────────────────────│
│ [VCNPeerRow]                                                  │
│ [VCNPeerRow]                                                  │
└─────────────────────────────────────────────────────────────┘
```

## Props interface

```ts
import { type VCNPeerData } from '@/components/vcn/VCNPeerRow';

interface VCNPeerTableProps {
    peers: VCNPeerData[];
}
```

## Component

```tsx
// src/widgets/vcn/VCNPeerTable.tsx
import cls from '@/widgets/vcn/VCNPeerTable.module.css';
import VCNPeerRow, { type VCNPeerData } from '@/components/vcn/VCNPeerRow';
import SectionLabel from '@/components/base/SectionLabel';

interface VCNPeerTableProps {
    peers: VCNPeerData[];
}

const TABLE_HEADERS = ['Name', 'Type', 'Latency', 'Status', 'RX', 'TX'];

export default function VCNPeerTable({ peers }: VCNPeerTableProps) {
    return (
        <div className={cls.VCNPeerTableContainer}>
            <div className={cls.sectionHeader}>
                <SectionLabel>Peers</SectionLabel>
            </div>
            <div className={cls.table}>
                <div className={cls.tableHeader}>
                    {TABLE_HEADERS.map(function renderHeader(h) {
                        return <span key={h} className={cls.headerCell}>{h}</span>;
                    })}
                </div>
                {peers.map(function renderRow(peer, i) {
                    return (
                        <VCNPeerRow
                            key={peer.id}
                            peer={peer}
                            last={i === peers.length - 1}
                        />
                    );
                })}
            </div>
        </div>
    );
}
```

## CSS

```css
/* src/widgets/vcn/VCNPeerTable.module.css */
.VCNPeerTableContainer { }

.sectionHeader {
    margin-bottom: 0.875rem;
}

.table {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

.tableHeader {
    display: grid;
    grid-template-columns: 1fr 6.25rem 5rem 5rem 5rem 5rem;
    padding: 0.625rem 1.25rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg3);
}

.headerCell {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}
```
