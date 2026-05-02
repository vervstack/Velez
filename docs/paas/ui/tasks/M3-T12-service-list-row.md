# T12 — ServiceListRow (table row)

**Files to create:**

- `src/components/service/ServiceListRow.tsx`
- `src/components/service/ServiceListRow.module.css`

## What it looks like

A single row in the list view of the deployments tab. Uses a CSS grid with 7 columns.

```
● matreshka-be   [prod]     node01   CPU 12% ████   MEM 180M ████   14d 3h   [···]
● postgres-main  [prod][!]  node03   CPU 89% ████   MEM 1.8G ████   3d 7h    [···]
```

Column layout: `28px 1fr 120px 80px 80px 80px 36px`
→ `dot | name+chips | node | cpu bar | mem bar | uptime | menu`

## Data types (same as ServiceCard, import from there)

```ts
import { type ServiceCardData } from '@/components/service/ServiceCard';
```

## Props interface

```ts
interface ServiceListRowProps {
    service: ServiceCardData;
    menuActions: Array<{ label: string; icon: string; danger?: boolean; onClick?: () => void }>;
}
```

## Component

```tsx
// src/components/service/ServiceListRow.tsx
import { useState } from 'react';
import cn from 'classnames';
import cls from '@/components/service/ServiceListRow.module.css';
import StatusDot from '@/components/base/StatusDot';
import MiniBar from '@/components/base/MiniBar';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import FreezeChip from '@/components/base/chips/FreezeChip';
import ThreeDotMenu from '@/components/complex/ThreeDotMenu/ThreeDotMenu';
import { type ServiceCardData } from '@/components/service/ServiceCard';

interface ServiceListRowProps {
    service: ServiceCardData;
    menuActions: Array<{ label: string; icon: string; danger?: boolean; onClick?: () => void }>;
}

export default function ServiceListRow({ service, menuActions }: ServiceListRowProps) {
    const [hovered, setHovered] = useState(false);

    function handleMouseEnter() { setHovered(true); }
    function handleMouseLeave() { setHovered(false); }

    function formatMem(mb: number) {
        return mb < 1000 ? mb + 'M' : (mb / 1000).toFixed(1) + 'G';
    }

    return (
        <div
            className={cn(cls.ServiceListRowContainer, { [cls.hovered]: hovered })}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
        >
            <StatusDot status={service.status} pulse />

            <div className={cls.nameCell}>
                <div className={cls.nameRow}>
                    <span className={cls.name}>{service.name}</span>
                    {service.restarts > 0 && (
                        <span className={cls.restarts}>↺{service.restarts}</span>
                    )}
                </div>
                <div className={cls.chips}>
                    <EnvChip env={service.env} />
                    {service.incident && <IncidentChip />}
                    {service.releaseFrozen && <FreezeChip />}
                </div>
            </div>

            <div className={cls.nodeCell}>
                <StatusDot status={service.node.status} />
                <span className={cls.nodeId}>{service.node.id}</span>
            </div>

            <div className={cls.barCell}>
                {service.status !== 'stopped' ? (
                    <>
                        <div className={cls.barHeader}>
                            <span className={cls.barLabel}>CPU</span>
                            <span className={cn(cls.barValue, {
                                [cls.valueRed]: service.cpu > 80,
                                [cls.valueAmber]: service.cpu > 60 && service.cpu <= 80,
                            })}>
                                {service.cpu}%
                            </span>
                        </div>
                        <MiniBar val={service.cpu} />
                    </>
                ) : <span className={cls.stopped}>—</span>}
            </div>

            <div className={cls.barCell}>
                {service.status !== 'stopped' ? (
                    <>
                        <div className={cls.barHeader}>
                            <span className={cls.barLabel}>MEM</span>
                            <span className={cls.barValue}>{formatMem(service.mem)}</span>
                        </div>
                        <MiniBar val={service.mem} max={2048} />
                    </>
                ) : <span className={cls.stopped}>—</span>}
            </div>

            <div className={cls.uptimeCell}>
                {service.status !== 'stopped'
                    ? service.uptime
                    : <span className={cls.stopped}>stopped</span>
                }
            </div>

            <ThreeDotMenu actions={menuActions} />
        </div>
    );
}
```

## CSS

```css
/* src/components/service/ServiceListRow.module.css */
.ServiceListRowContainer {
    display: grid;
    grid-template-columns: 1.75rem 1fr 7.5rem 5rem 5rem 5rem 2.25rem;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 1.25rem;
    border-bottom: 1px solid var(--border);
    background: transparent;
    transition: background 0.12s;
    cursor: default;
}

.ServiceListRowContainer.hovered {
    background: var(--bg3);
}

.nameCell {
    min-width: 0;
}

.nameRow {
    display: flex;
    align-items: center;
    gap: 0.4375rem;
    margin-bottom: 0.25rem;
}

.name {
    font-family: var(--font-mono);
    font-size: 0.78125rem;
    color: var(--fg);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.restarts {
    font-family: var(--font-mono);
    font-size: 0.59375rem;
    color: var(--amber);
    border: 1px solid rgba(245, 166, 35, 0.3);
    border-radius: 0.25rem;
    padding: 0.0625rem 0.3125rem;
    flex-shrink: 0;
}

.chips {
    display: flex;
    gap: 0.25rem;
    flex-wrap: wrap;
}

.nodeCell {
    display: flex;
    align-items: center;
    gap: 0.3125rem;
}

.nodeId {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
}

.barCell { }

.barHeader {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.1875rem;
}

.barLabel {
    font-family: var(--font-mono);
    font-size: 0.5625rem;
    color: var(--fg-dim);
}

.barValue {
    font-family: var(--font-mono);
    font-size: 0.5625rem;
    color: var(--fg-mid);
}

.valueRed   { color: var(--red); }
.valueAmber { color: var(--amber); }

.uptimeCell {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
}

.stopped {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-faint);
}
```
