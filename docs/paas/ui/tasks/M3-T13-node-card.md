# T13 — NodeCard

**Files to create:**

- `src/components/node/NodeCard.tsx`
- `src/components/node/NodeCard.module.css`

## What it looks like

A horizontal card showing one cluster node with its health metrics and action buttons. Used in a vertical list on the
Control Plane page.

```
┌───────────────────────────────────────────────────────────────────────────────────┐
│ ● node01  [degraded]   │  CPU  38%  ████░░  │  Memory  52%  ████░░  │  5 services │  [shell] [drain] │
│   192.168.1.10         │                     │                        │             │                  │
│   eu-west              │                     │                        │             │                  │
└───────────────────────────────────────────────────────────────────────────────────┘
```

## Data types

```ts
interface NodeCardData {
    id: string;
    host: string;
    region: string;
    status: 'online' | 'degraded' | 'offline';
    cpu: number;       // percent 0-100
    mem: number;       // percent 0-100
    services: number;
}
```

## Props interface

```ts
interface NodeCardProps {
    node: NodeCardData;
    onShell?: () => void;
    onDrain?: () => void;
}
```

## Component

```tsx
// src/components/node/NodeCard.tsx
import cls from '@/components/node/NodeCard.module.css';
import cn from 'classnames';
import StatusDot from '@/components/base/StatusDot';
import Badge from '@/components/base/Badge';
import MiniBar from '@/components/base/MiniBar';
import IconButton from '@/components/base/IconButton';

export interface NodeCardData {
    id: string;
    host: string;
    region: string;
    status: 'online' | 'degraded' | 'offline';
    cpu: number;
    mem: number;
    services: number;
}

interface NodeCardProps {
    node: NodeCardData;
    onShell?: () => void;
    onDrain?: () => void;
}

export default function NodeCard({ node, onShell, onDrain }: NodeCardProps) {
    return (
        <div className={cn(cls.NodeCardContainer, { [cls.degraded]: node.status === 'degraded' })}>
            {/* Identity */}
            <div className={cls.identity}>
                <div className={cls.nameRow}>
                    <StatusDot status={node.status} pulse />
                    <span className={cls.nodeId}>{node.id}</span>
                    {node.status === 'degraded' && (
                        <Badge label="degraded" color="var(--amber)" />
                    )}
                </div>
                <div className={cls.host}>{node.host}</div>
                <div className={cls.region}>{node.region}</div>
            </div>

            {/* CPU */}
            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>CPU</span>
                    <span className={cn(cls.metricValue, {
                        [cls.valueRed]: node.cpu > 80,
                        [cls.valueAmber]: node.cpu > 60 && node.cpu <= 80,
                    })}>
                        {node.cpu}%
                    </span>
                </div>
                <MiniBar val={node.cpu} />
            </div>

            {/* Memory */}
            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>Memory</span>
                    <span className={cn(cls.metricValue, {
                        [cls.valueRed]: node.mem > 80,
                        [cls.valueAmber]: node.mem > 60 && node.mem <= 80,
                    })}>
                        {node.mem}%
                    </span>
                </div>
                <MiniBar val={node.mem} />
            </div>

            {/* Services count */}
            <div className={cls.servicesBlock}>
                <span className={cls.servicesCount}>{node.services}</span>
                <span className={cls.servicesLabel}>services</span>
            </div>

            {/* Actions */}
            <div className={cls.actions}>
                <IconButton label="shell" title="Open terminal" onClick={onShell} />
                <IconButton label="drain" title="Drain node" danger onClick={onDrain} />
            </div>
        </div>
    );
}
```

## CSS

```css
/* src/components/node/NodeCard.module.css */
.NodeCardContainer {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1.125rem 1.25rem;
    display: grid;
    grid-template-columns: 13.75rem 1fr 1fr 1fr auto;
    gap: 1.25rem;
    align-items: center;
    animation: fadeUp 0.3s ease both;
}

.NodeCardContainer.degraded {
    border-color: rgba(245, 166, 35, 0.44);
}

/* Identity */
.identity { }

.nameRow {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
}

.nodeId {
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    color: var(--fg);
    font-weight: 600;
}

.host {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
}

.region {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-faint);
}

/* Metric block */
.metric { }

.metricHeader {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.3125rem;
}

.metricLabel {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}

.metricValue {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-mid);
}

.valueRed   { color: var(--red); }
.valueAmber { color: var(--amber); }

/* Services */
.servicesBlock {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.servicesCount {
    font-family: var(--font-mono);
    font-size: 1.375rem;
    color: var(--fg);
    font-weight: 500;
}

.servicesLabel {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}

/* Actions */
.actions {
    display: flex;
    gap: 0.375rem;
}
```
