# T20 — ServiceListView Widget

**Files to create:**

- `src/widgets/deployments/ServiceListView.tsx`
- `src/widgets/deployments/ServiceListView.module.css`

## What it looks like

A table-style list of services, grouped by status (Running → Degraded → Stopped). A sticky header row labels the
columns.

```
┌────────────────────────────────────────────────────────────────────┐
│       Service          Node     CPU     Memory   Uptime         │  ← sticky header
│───────────────────────────────────────────────────────────────────│
│ RUNNING (5)                                                       │
│ [ServiceListRow]                                                  │
│ ...                                                               │
│ DEGRADED (2)                                                      │
│ [ServiceListRow]                                                  │
│ STOPPED (2)                                                       │
└────────────────────────────────────────────────────────────────────┘
```

## Props interface

```ts
import { type ServiceCardData } from '@/components/service/ServiceCard';

interface ServiceListViewProps {
    services: ServiceCardData[];
    statusFilter?: Set<string>;
    onServiceAction: (serviceId: string, action: string) => void;
}
```

## Component

```tsx
// src/widgets/deployments/ServiceListView.tsx
import cls from '@/widgets/deployments/ServiceListView.module.css';
import ServiceListRow from '@/components/service/ServiceListRow';
import { type ServiceCardData } from '@/components/service/ServiceCard';

interface ServiceListViewProps {
    services: ServiceCardData[];
    statusFilter?: Set<string>;
    onServiceAction: (serviceId: string, action: string) => void;
}

const COLUMN_GROUPS = [
    { id: 'running',  label: 'Running',  color: 'var(--green)'   },
    { id: 'degraded', label: 'Degraded', color: 'var(--amber)'   },
    { id: 'stopped',  label: 'Stopped',  color: 'var(--fg-faint)'},
];

function buildActions(svc: ServiceCardData, onAction: (id: string, action: string) => void) {
    const actions: Array<{ label: string; icon: string; danger?: boolean; onClick: () => void }> = [];
    if (svc.status === 'running') {
        actions.push({ label: 'Restart', icon: '↺', onClick: () => onAction(svc.name, 'restart') });
        actions.push({ label: 'Stop',    icon: '◼', danger: true, onClick: () => onAction(svc.name, 'stop') });
    }
    if (svc.status === 'stopped') {
        actions.push({ label: 'Start',   icon: '▶', onClick: () => onAction(svc.name, 'start') });
    }
    if (svc.status === 'degraded') {
        actions.push({ label: 'Restart', icon: '↺', onClick: () => onAction(svc.name, 'restart') });
    }
    actions.push({ label: 'View logs', icon: '≈', onClick: () => onAction(svc.name, 'logs') });
    actions.push({ label: 'Shell',     icon: '$', onClick: () => onAction(svc.name, 'shell') });
    return actions;
}

export default function ServiceListView({ services, statusFilter, onServiceAction }: ServiceListViewProps) {
    const hasAny = services.length > 0;

    return (
        <div className={cls.ServiceListViewContainer}>
            {/* Sticky header */}
            <div className={cls.tableHeader}>
                {['', 'Service', 'Node', 'CPU', 'Memory', 'Uptime', ''].map(function renderHeader(h, i) {
                    return <span key={i} className={cls.headerCell}>{h}</span>;
                })}
            </div>

            {!hasAny && (
                <div className={cls.empty}>no services match current filters</div>
            )}

            {COLUMN_GROUPS.map(function renderGroup(group) {
                if (statusFilter && statusFilter.size > 0 && !statusFilter.has(group.id)) return null;
                const groupServices = services.filter(s => s.status === group.id);
                if (groupServices.length === 0) return null;

                return (
                    <div key={group.id}>
                        <div className={cls.groupLabel}>
                            <span className={cls.groupDot} style={{ background: group.color }} />
                            <span className={cls.groupName}>{group.label}</span>
                            <span className={cls.groupCount}>{groupServices.length}</span>
                        </div>
                        {groupServices.map(function renderRow(svc) {
                            return (
                                <ServiceListRow
                                    key={svc.name}
                                    service={svc}
                                    menuActions={buildActions(svc, onServiceAction)}
                                />
                            );
                        })}
                    </div>
                );
            })}
        </div>
    );
}
```

## CSS

```css
/* src/widgets/deployments/ServiceListView.module.css */
.ServiceListViewContainer {
    flex: 1;
    overflow-y: auto;
    margin: 1rem 1.5rem 1.25rem;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

/* Table header */
.tableHeader {
    display: grid;
    grid-template-columns: 1.75rem 1fr 7.5rem 5rem 5rem 5rem 2.25rem;
    gap: 0.75rem;
    padding: 0.625rem 1.25rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg3);
    position: sticky;
    top: 0;
}

.headerCell {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}

/* Empty state */
.empty {
    padding: 2.5rem 1.25rem;
    text-align: center;
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--fg-faint);
}

/* Group label row */
.groupLabel {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1.25rem 0.25rem;
}

.groupDot {
    width: 0.375rem;
    height: 0.375rem;
    border-radius: 50%;
    display: block;
    flex-shrink: 0;
}

.groupName {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
    letter-spacing: 0.06em;
    text-transform: uppercase;
}

.groupCount {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-faint);
}
```
