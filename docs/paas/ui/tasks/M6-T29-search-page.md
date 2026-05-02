# T29 — SearchPage

**Files to create:**

- `src/pages/search/SearchPage.tsx`
- `src/pages/search/SearchPage.module.css`

## What it looks like

A centered search interface. The input is autofocused. Results appear below grouped by type.

```
┌──────────────────────────────────┐
│  ⌕ Search services, nodes…      │  ← large, autofocused input
└──────────────────────────────────┘

SERVICES (3)
● matreshka-be     [prod]  node01
● api-gateway      [prod] [incident]  node02
● postgres-main    [prod] [incident]  node03

NODES (1)
● node03  degraded  10.0.0.15
```

## Routing

Add `Routes.Search = '/search'` to `src/app/router/Router.tsx` and register the route under `MainLayout`. The Sidebar "
Search" nav item should point to this route.

## Props

No props — self-contained page. Search state is local.

## Component

```tsx
// src/pages/search/SearchPage.tsx
import { useState, useEffect, useRef, useMemo } from 'react';
import cls from '@/pages/search/SearchPage.module.css';
import StatusDot from '@/components/base/StatusDot';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import { type ServiceCardData } from '@/components/service/ServiceCard';

// Mock data — replace with React Query calls
const MOCK_SERVICES: ServiceCardData[] = [
    { name: 'matreshka-be', image: 'godverv/matreshka:latest', status: 'running',  cpu: 12, mem: 180,  uptime: '14d 3h', restarts: 0,  env: 'prod',  incident: false, releaseFrozen: false, node: { id: 'node01', host: '192.168.1.10', status: 'online'   } },
    { name: 'api-gateway',  image: 'internal/gateway:v2.1',    status: 'running',  cpu: 33, mem: 512,  uptime: '6d 4h',  restarts: 0,  env: 'prod',  incident: true,  releaseFrozen: false, node: { id: 'node02', host: '192.168.1.42', status: 'online'   } },
    { name: 'postgres-main',image: 'postgres:16-alpine',        status: 'degraded', cpu: 89, mem: 1800, uptime: '3d 7h',  restarts: 12, env: 'prod',  incident: true,  releaseFrozen: false, node: { id: 'node03', host: '10.0.0.15',    status: 'degraded' } },
    { name: 'prometheus',   image: 'prom/prometheus:latest',    status: 'stopped',  cpu: 0,  mem: 0,    uptime: 'stopped',restarts: 0,  env: 'stage', incident: false, releaseFrozen: false, node: { id: 'node01', host: '192.168.1.10', status: 'online'   } },
];

const MOCK_NODES = [
    { id: 'node01', host: '192.168.1.10', status: 'online'   as const },
    { id: 'node02', host: '192.168.1.42', status: 'online'   as const },
    { id: 'node03', host: '10.0.0.15',    status: 'degraded' as const },
];

export default function SearchPage() {
    const [query, setQuery] = useState('');
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(function autoFocus() {
        inputRef.current?.focus();
    }, []);

    function handleQueryChange(e: React.ChangeEvent<HTMLInputElement>) {
        setQuery(e.target.value);
    }

    const { matchedServices, matchedNodes } = useMemo(function computeResults() {
        const q = query.trim().toLowerCase();
        if (!q) return { matchedServices: [], matchedNodes: [] };

        const matchedServices = MOCK_SERVICES.filter(s =>
            s.name.toLowerCase().includes(q) ||
            s.image.toLowerCase().includes(q) ||
            s.node.id.toLowerCase().includes(q)
        );

        const matchedNodes = MOCK_NODES.filter(n =>
            n.id.toLowerCase().includes(q) ||
            n.host.toLowerCase().includes(q)
        );

        return { matchedServices, matchedNodes };
    }, [query]);

    const hasResults = matchedServices.length > 0 || matchedNodes.length > 0;

    return (
        <div className={cls.SearchPageContainer}>
            <div className={cls.searchWrapper}>
                <span className={cls.searchIcon}>⌕</span>
                <input
                    ref={inputRef}
                    className={cls.searchInput}
                    value={query}
                    onChange={handleQueryChange}
                    placeholder="Search services, nodes…"
                />
            </div>

            {query && !hasResults && (
                <div className={cls.empty}>no results for "{query}"</div>
            )}

            {matchedServices.length > 0 && (
                <section className={cls.section}>
                    <div className={cls.sectionHeader}>
                        <span className={cls.sectionTitle}>Services</span>
                        <span className={cls.sectionCount}>{matchedServices.length}</span>
                    </div>
                    {matchedServices.map(function renderService(svc) {
                        return (
                            <div key={svc.name} className={cls.resultRow}>
                                <StatusDot status={svc.status} />
                                <span className={cls.resultName}>{svc.name}</span>
                                <div className={cls.resultChips}>
                                    <EnvChip env={svc.env} />
                                    {svc.incident && <IncidentChip />}
                                </div>
                                <span className={cls.resultMeta}>{svc.node.id}</span>
                            </div>
                        );
                    })}
                </section>
            )}

            {matchedNodes.length > 0 && (
                <section className={cls.section}>
                    <div className={cls.sectionHeader}>
                        <span className={cls.sectionTitle}>Nodes</span>
                        <span className={cls.sectionCount}>{matchedNodes.length}</span>
                    </div>
                    {matchedNodes.map(function renderNode(node) {
                        return (
                            <div key={node.id} className={cls.resultRow}>
                                <StatusDot status={node.status} />
                                <span className={cls.resultName}>{node.id}</span>
                                {node.status === 'degraded' && (
                                    <span className={cls.degradedLabel}>degraded</span>
                                )}
                                <span className={cls.resultMeta}>{node.host}</span>
                            </div>
                        );
                    })}
                </section>
            )}
        </div>
    );
}
```

## CSS

```css
/* src/pages/search/SearchPage.module.css */
.SearchPageContainer {
    padding: 2rem 1.5rem;
    overflow-y: auto;
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    max-width: 48rem;
}

/* Search input */
.searchWrapper {
    position: relative;
}

.searchIcon {
    position: absolute;
    left: 1rem;
    top: 50%;
    transform: translateY(-50%);
    font-family: var(--font-mono);
    font-size: 1.125rem;
    color: var(--fg-faint);
    pointer-events: none;
}

.searchInput {
    width: 100%;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.875rem 1rem 0.875rem 2.75rem;
    font-family: var(--font-mono);
    font-size: 0.875rem;
    color: var(--fg);
    outline: none;
    transition: border-color 0.15s;

    &::placeholder { color: var(--fg-faint); }
    &:focus { border-color: var(--border-h); }
}

/* Empty */
.empty {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--fg-faint);
}

/* Result sections */
.section {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
}

.sectionHeader {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
}

.sectionTitle {
    font-family: var(--font-mono);
    font-size: 0.5625rem;
    color: var(--fg-dim);
    letter-spacing: 0.1em;
    text-transform: uppercase;
}

.sectionCount {
    font-family: var(--font-mono);
    font-size: 0.5625rem;
    color: var(--fg-faint);
}

/* Result row */
.resultRow {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.625rem 0.75rem;
    border-radius: var(--radius-sm);
    transition: background 0.12s;
    cursor: pointer;

    &:hover {
        background: var(--bg3);
    }
}

.resultName {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--fg);
    font-weight: 500;
}

.resultChips {
    display: flex;
    gap: 0.25rem;
}

.resultMeta {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
    margin-left: auto;
}

.degradedLabel {
    font-family: var(--font-mono);
    font-size: 0.59375rem;
    color: var(--amber);
    border: 1px solid rgba(245, 166, 35, 0.44);
    border-radius: 0.25rem;
    padding: 0.0625rem 0.375rem;
}
```

## Notes

- Add `Routes.Search = '/search'` to the Routes enum in `src/app/router/Router.tsx`.
- Update the Sidebar widget (T16) `NAV_TO_ROUTE` to map `'search'` → `/search`.
- Replace `MOCK_SERVICES` and `MOCK_NODES` with React Query hooks when the page is hooked up to the API.
