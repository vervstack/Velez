# T11 — ServiceCard (Kanban card)

**Files to modify/replace:**

- `src/components/service/ServiceCard.tsx`
- `src/components/service/ServiceCard.module.css`

## What it looks like

A card displayed in the kanban board columns (Running / Degraded / Stopped). Hover state raises the background and
brightens the border.

```
┌──────────────────────────────────┐
│ ● matreshka-be          [···]    │  ← status dot + name + 3-dot menu
│ [prod] [incident]                │  ← env + incident/freeze chips
│ godverv/matreshka:latest         │  ← image in mono dim
│ ● node01  192.168.1.10           │  ← node dot + id + host
│ CPU  12%  ████░░░░░░             │  ← MiniBar
│ MEM  180M ████░░░░░░             │
│ uptime  14d 3h                   │
└──────────────────────────────────┘
```

When `status === 'stopped'` skip the CPU/MEM/uptime block and show `container stopped` in dim italic.

## Data types

```ts
// Caller is responsible for building this from API types
interface ServiceCardData {
    name: string;
    image: string;
    status: 'running' | 'degraded' | 'stopped';
    cpu: number;        // percent 0-100
    mem: number;        // MiB
    uptime: string;
    restarts: number;
    env: string;        // 'prod' | 'stage'
    incident: boolean;
    releaseFrozen: boolean;
    node: {
        id: string;
        host: string;
        status: 'online' | 'degraded' | 'offline';
    };
}
```

## Props interface

```ts
interface ServiceCardProps {
    service: ServiceCardData;
    /** Actions shown in the 3-dot menu. Caller determines which actions are valid. */
    menuActions: import('@/components/complex/ThreeDotMenu/ThreeDotMenu').MenuAction[];
}
```

## Component

```tsx
// src/components/service/ServiceCard.tsx
import { useState } from 'react';
import cn from 'classnames';
import cls from '@/components/service/ServiceCard.module.css';
import StatusDot from '@/components/base/StatusDot';
import MiniBar from '@/components/base/MiniBar';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import FreezeChip from '@/components/base/chips/FreezeChip';
import ThreeDotMenu from '@/components/complex/ThreeDotMenu/ThreeDotMenu';

// types re-exported for callers
export interface ServiceCardData {
    name: string;
    image: string;
    status: 'running' | 'degraded' | 'stopped';
    cpu: number;
    mem: number;
    uptime: string;
    restarts: number;
    env: string;
    incident: boolean;
    releaseFrozen: boolean;
    node: { id: string; host: string; status: 'online' | 'degraded' | 'offline' };
}

interface ServiceCardProps {
    service: ServiceCardData;
    menuActions: Array<{ label: string; icon: string; danger?: boolean; onClick?: () => void }>;
}

export default function ServiceCard({ service, menuActions }: ServiceCardProps) {
    const [hovered, setHovered] = useState(false);

    function handleMouseEnter() { setHovered(true); }
    function handleMouseLeave() { setHovered(false); }

    function formatMem(mb: number) {
        return mb < 1000 ? mb + 'M' : (mb / 1000).toFixed(1) + 'G';
    }

    return (
        <div
            className={cn(cls.ServiceCardContainer, { [cls.hovered]: hovered })}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
        >
            {/* Header */}
            <div className={cls.header}>
                <div className={cls.nameRow}>
                    <StatusDot status={service.status} pulse />
                    <span className={cls.name}>{service.name}</span>
                    {service.restarts > 0 && (
                        <span className={cls.restartsBadge} title={`${service.restarts} restarts`}>
                            ↺{service.restarts}
                        </span>
                    )}
                </div>
                <ThreeDotMenu actions={menuActions} />
            </div>

            {/* Chips */}
            <div className={cls.chips}>
                <EnvChip env={service.env} />
                {service.incident && <IncidentChip />}
                {service.releaseFrozen && <FreezeChip />}
            </div>

            {/* Image */}
            <div className={cls.image}>{service.image}</div>

            {/* Node */}
            <div className={cls.node}>
                <StatusDot status={service.node.status} />
                <span className={cls.nodeId}>{service.node.id}</span>
                <span className={cls.nodeHost}>{service.node.host}</span>
            </div>

            {/* Stats */}
            {service.status !== 'stopped' ? (
                <div className={cls.stats}>
                    <div className={cls.statRow}>
                        <div className={cls.statItem}>
                            <div className={cls.statHeader}>
                                <span className={cls.statLabel}>CPU</span>
                                <span className={cn(cls.statValue, {
                                    [cls.valueRed]: service.cpu > 80,
                                    [cls.valueAmber]: service.cpu > 60 && service.cpu <= 80,
                                })}>
                                    {service.cpu}%
                                </span>
                            </div>
                            <MiniBar val={service.cpu} />
                        </div>
                        <div className={cls.statItem}>
                            <div className={cls.statHeader}>
                                <span className={cls.statLabel}>MEM</span>
                                <span className={cls.statValue}>{formatMem(service.mem)}</span>
                            </div>
                            <MiniBar val={service.mem} max={2048} />
                        </div>
                    </div>
                    <div className={cls.uptime}>
                        <span className={cls.statLabel}>uptime</span>
                        <span className={cls.statValue}>{service.uptime}</span>
                    </div>
                </div>
            ) : (
                <div className={cls.stopped}>container stopped</div>
            )}
        </div>
    );
}
```

## CSS

```css
/* src/components/service/ServiceCard.module.css */
.ServiceCardContainer {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.875rem 1rem; /* 14px 16px */
    transition: background 0.15s, border-color 0.15s;
    animation: fadeUp 0.3s ease both;
    cursor: default;
}

.ServiceCardContainer.hovered {
    background: var(--bg3);
    border-color: var(--border-m);
}

/* Header */
.header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 0.5rem;
    gap: 0.5rem;
}

.nameRow {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
}

.name {
    font-family: var(--font-mono);
    font-size: 0.78125rem; /* 12.5px */
    color: var(--fg);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.restartsBadge {
    font-family: var(--font-mono);
    font-size: 0.59375rem; /* 9.5px */
    color: var(--amber);
    border: 1px solid rgba(245, 166, 35, 0.3);
    border-radius: 0.25rem;
    padding: 0.0625rem 0.375rem;
    flex-shrink: 0;
}

/* Chips row */
.chips {
    display: flex;
    gap: 0.3125rem;
    flex-wrap: wrap;
    margin-bottom: 0.625rem; /* 10px */
}

/* Image */
.image {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
    margin-bottom: 0.625rem;
    word-break: break-all;
    line-height: 1.5;
}

/* Node row */
.node {
    display: flex;
    align-items: center;
    gap: 0.3125rem;
    margin-bottom: 0.625rem;
}

.nodeId {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}

.nodeHost {
    font-family: var(--font-mono);
    font-size: 0.59375rem;
    color: var(--fg-faint);
}

/* Stats */
.stats { }

.statRow {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem 0.75rem;
    margin-bottom: 0.25rem;
}

.statItem { }

.statHeader {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.1875rem;
}

.statLabel {
    font-family: var(--font-mono);
    font-size: 0.59375rem;
    color: var(--fg-dim);
}

.statValue {
    font-family: var(--font-mono);
    font-size: 0.59375rem;
    color: var(--fg-mid);
}

.valueRed   { color: var(--red); }
.valueAmber { color: var(--amber); }

.uptime {
    display: flex;
    justify-content: space-between;
}

.stopped {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-faint);
    font-style: italic;
}
```
