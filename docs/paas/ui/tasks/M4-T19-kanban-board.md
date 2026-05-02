# T19 — KanbanBoard Widget

**Files to create:**

- `src/widgets/deployments/KanbanBoard.tsx`
- `src/widgets/deployments/KanbanBoard.module.css`

## What it looks like

Three equal-width columns (Running / Degraded / Stopped), each scrollable vertically. Each column has a colored header
dot, label, and card count badge. Cards are `ServiceCard` components.

```
Running (5)       Degraded (2)      Stopped (2)
─────────────     ─────────────     ─────────────
[ServiceCard]     [ServiceCard]     [ServiceCard]
[ServiceCard]     [ServiceCard]     [ServiceCard]
[ServiceCard]                       — empty state —
```

When `statusFilters` hides a column, skip rendering it.

## Props interface

```ts
import { type ServiceCardData } from '@/components/service/ServiceCard';

interface KanbanBoardProps {
    services: ServiceCardData[];
    statusFilter?: Set<string>; // if non-empty, only show columns whose id is in the set
    onServiceAction: (serviceId: string, action: string) => void;
}
```

## Component

The component builds menu actions based on service status using a helper function.

```tsx
// src/widgets/deployments/KanbanBoard.tsx
import cls from '@/widgets/deployments/KanbanBoard.module.css';
import ServiceCard, { type ServiceCardData } from '@/components/service/ServiceCard';

interface KanbanBoardProps {
    services: ServiceCardData[];
    statusFilter?: Set<string>;
    onServiceAction: (serviceId: string, action: string) => void;
}

const COLUMNS = [
    { id: 'running',  label: 'Running',  color: 'var(--green)'   },
    { id: 'degraded', label: 'Degraded', color: 'var(--amber)'   },
    { id: 'stopped',  label: 'Stopped',  color: 'var(--fg-faint)'},
];

function buildActions(svc: ServiceCardData, onAction: (id: string, action: string) => void) {
    const actions: Array<{ label: string; icon: string; danger?: boolean; onClick: () => void }> = [];
    if (svc.status === 'running') {
        actions.push({ label: 'Restart',   icon: '↺', onClick: () => onAction(svc.name, 'restart') });
        actions.push({ label: 'Stop',      icon: '◼', danger: true, onClick: () => onAction(svc.name, 'stop') });
    }
    if (svc.status === 'stopped') {
        actions.push({ label: 'Start',     icon: '▶', onClick: () => onAction(svc.name, 'start') });
    }
    if (svc.status === 'degraded') {
        actions.push({ label: 'Restart',   icon: '↺', onClick: () => onAction(svc.name, 'restart') });
    }
    actions.push({ label: 'View logs', icon: '≈', onClick: () => onAction(svc.name, 'logs') });
    actions.push({ label: 'Shell',     icon: '$', onClick: () => onAction(svc.name, 'shell') });
    return actions;
}

export default function KanbanBoard({ services, statusFilter, onServiceAction }: KanbanBoardProps) {
    return (
        <div className={cls.KanbanBoardContainer}>
            {COLUMNS.map(function renderColumn(col) {
                if (statusFilter && statusFilter.size > 0 && !statusFilter.has(col.id)) return null;
                const colServices = services.filter(s => s.status === col.id);

                return (
                    <div key={col.id} className={cls.column}>
                        <div className={cls.columnHeader}>
                            <span className={cls.columnDot} style={{ background: col.color }} />
                            <span className={cls.columnLabel}>{col.label}</span>
                            <span className={cls.columnCount}>{colServices.length}</span>
                        </div>

                        <div className={cls.columnBody}>
                            {colServices.length === 0 && (
                                <div className={cls.empty}>
                                    <div className={cls.emptyDash}>—</div>
                                    <div className={cls.emptyText}>no services</div>
                                </div>
                            )}
                            {colServices.map(function renderCard(svc) {
                                return (
                                    <ServiceCard
                                        key={svc.name}
                                        service={svc}
                                        menuActions={buildActions(svc, onServiceAction)}
                                    />
                                );
                            })}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
```

## CSS

```css
/* src/widgets/deployments/KanbanBoard.module.css */
.KanbanBoardContainer {
    display: flex;
    gap: 1.125rem;
    flex: 1;
    padding: 1rem 1.5rem 1.25rem;
    overflow-x: auto;
    overflow-y: hidden;
    align-items: flex-start;
}

.column {
    width: 20.625rem; /* 330px */
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    height: 100%;
}

.columnHeader {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    padding: 0 0.125rem;
}

.columnDot {
    width: 0.4375rem;
    height: 0.4375rem;
    border-radius: 50%;
    display: block;
    flex-shrink: 0;
}

.columnLabel {
    font-family: var(--font-sans);
    font-size: 0.8125rem;
    font-weight: 700;
    color: var(--fg-mid);
}

.columnCount {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
    margin-left: auto;
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: 99px;
    padding: 0.0625rem 0.5rem;
}

.columnBody {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
    padding-bottom: 0.25rem;
    padding-right: 0.125rem;
}

.empty {
    border: 1px dashed var(--border);
    border-radius: var(--radius);
    padding: 2rem 1rem;
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
}

.emptyDash {
    font-family: var(--font-mono);
    font-size: 1.125rem;
    color: var(--fg-faint);
}

.emptyText {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-faint);
}
```
