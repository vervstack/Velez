# T30 — AppsPage + AppCard (logical services view)

**Files to create:**

- `src/pages/apps/AppsPage.tsx`
- `src/pages/apps/AppsPage.module.css`
- `src/components/apps/AppCard.tsx`
- `src/components/apps/AppCard.module.css`

**Files to modify:**

- `src/app/router/Routes.ts` — add `Apps = '/apps'`
- `src/app/router/Router.tsx` — add `/apps` route → `AppsPage`
- `src/widgets/sidebar/Sidebar.tsx` — add "Apps" to `NavId` and `NAV_ITEMS`
- `src/app/router/MainLayout.tsx` — add `apps` ↔ `Routes.Apps` mapping

## What it does

Adds an **Apps** entry to the sidebar's "Services" section. The Apps page lists logical services — one card per
service, not per container. Cards share the visual skeleton of `ServiceCard` but expand on hover to show deployment
metadata, and click-navigate to `/service/:key`.

## Sidebar change

Extend `NavId` in `Sidebar.tsx` and `MainLayout.tsx`:

```ts
type NavId = 'controlplane' | 'vcn' | 'deployments' | 'apps' | 'search';
```

Add to `NAV_ITEMS` (insert between `deployments` and `search`):

```ts
{
    id: 'apps', label
:
    'Apps', icon
:
    '⬡'
}
,
```

Add to `NAV_TO_ROUTE` / `ROUTE_TO_NAV` in `MainLayout.tsx`:

```ts
apps: Routes.Apps,          // NAV_TO_ROUTE
    [Routes.Apps]
:
'apps',      // ROUTE_TO_NAV
```

## AppCard data type

```ts
export interface AppData {
    name: string;           // service name / key
    image: string;          // primary container image
    status: 'running' | 'degraded' | 'stopped';
    env: string;            // 'prod' | 'stage'
    incident: boolean;
    releaseFrozen: boolean;
    node: { id: string; host: string; status: 'online' | 'degraded' | 'offline' };
    // extra fields shown only on hover:
    deployments: number;    // count of active deployments
    lastDeployed: string;   // human-readable, e.g. '2h ago'
    configSource: string;   // e.g. 'matreshka' | 'none'
    version: string;        // image tag or semver, e.g. 'v2.1.4'
}
```

## AppCard layout

Base state (no hover) — identical layout to `ServiceCard` minus the CPU/MEM mini-bars and uptime row:

```
┌──────────────────────────────────┐
│ ● matreshka-be          [···]    │  ← status dot + name + 3-dot menu
│ [prod] [incident]                │  ← chips
│ godverv/matreshka:latest         │  ← image dim mono
│ ● node01  192.168.1.10           │  ← node row
└──────────────────────────────────┘
```

Hover state — the card height expands with a smooth CSS transition, revealing an extra block below the node row:

```
│ ─────────────────────────────── │  ← thin divider
│ deployments  3                  │
│ last deploy  2h ago             │
│ config       matreshka          │
│ version      v2.1.4             │
│ [↗ Open] [▶ Deploy]             │  ← action buttons
```

Use `max-height` transition (`max-height: 0` → `max-height: 10rem`) with `overflow: hidden` on the extra block.
The card itself should have `cursor: pointer` and fire `onClick` to navigate to the service page.

## AppCard props

```ts
interface AppCardProps {
    app: AppData;
    onOpen: (name: string) => void;
    onDeploy: (name: string) => void;
}
```

No 3-dot menu — replace with the two action buttons in the hover block.

## AppsPage component

```tsx
// src/pages/apps/AppsPage.tsx
import {useMemo, useState} from 'react';
import {useNavigate} from 'react-router-dom';
import cls from '@/pages/apps/AppsPage.module.css';
import AppCard, {type AppData} from '@/components/apps/AppCard';
import {Routes, Arguments} from '@/app/router/Routes';

const MOCK_APPS: AppData[] = [
    {
        name: 'matreshka-be',
        image: 'godverv/matreshka:latest',
        status: 'running',
        env: 'prod',
        incident: false,
        releaseFrozen: false,
        node: {id: 'node01', host: '192.168.1.10', status: 'online'},
        deployments: 2,
        lastDeployed: '14d ago',
        configSource: 'matreshka',
        version: 'latest'
    },
    {
        name: 'api-gateway',
        image: 'internal/gateway:v2.1',
        status: 'running',
        env: 'prod',
        incident: true,
        releaseFrozen: false,
        node: {id: 'node02', host: '192.168.1.42', status: 'online'},
        deployments: 1,
        lastDeployed: '6d ago',
        configSource: 'matreshka',
        version: 'v2.1'
    },
    {
        name: 'postgres-main',
        image: 'postgres:16-alpine',
        status: 'degraded',
        env: 'prod',
        incident: true,
        releaseFrozen: false,
        node: {id: 'node03', host: '10.0.0.15', status: 'degraded'},
        deployments: 1,
        lastDeployed: '3d ago',
        configSource: 'none',
        version: '16-alpine'
    },
    {
        name: 'redis-cache',
        image: 'redis:7-alpine',
        status: 'degraded',
        env: 'prod',
        incident: false,
        releaseFrozen: false,
        node: {id: 'node02', host: '192.168.1.42', status: 'online'},
        deployments: 1,
        lastDeployed: '1d ago',
        configSource: 'none',
        version: '7-alpine'
    },
    {
        name: 'prometheus',
        image: 'prom/prometheus:latest',
        status: 'stopped',
        env: 'stage',
        incident: false,
        releaseFrozen: false,
        node: {id: 'node01', host: '192.168.1.10', status: 'online'},
        deployments: 0,
        lastDeployed: 'never',
        configSource: 'none',
        version: 'latest'
    },
];

export default function AppsPage() {
    const navigate = useNavigate();
    const [search, setSearch] = useState('');

    const filtered = useMemo(function computeFiltered() {
        const q = search.trim().toLowerCase();
        if (!q) return MOCK_APPS;
        return MOCK_APPS.filter(a => a.name.toLowerCase().includes(q) || a.image.toLowerCase().includes(q));
    }, [search]);

    function handleOpen(name: string) {
        navigate(Routes.Service + '/' + name);
    }

    function handleDeploy(name: string) {
        console.log('deploy', name);
    }

    return (
        <div className={cls.AppsPageContainer}>
            <div className={cls.toolbar}>
                <input
                    className={cls.search}
                    placeholder="Filter apps…"
                    value={search}
                    onChange={e => setSearch(e.target.value)}
                />
                <span className={cls.count}>{filtered.length} apps</span>
            </div>
            <div className={cls.grid}>
                {filtered.map(function renderCard(app) {
                    return (
                        <AppCard
                            key={app.name}
                            app={app}
                            onOpen={handleOpen}
                            onDeploy={handleDeploy}
                        />
                    );
                })}
            </div>
        </div>
    );
}
```

## AppsPage CSS

```css
/* src/pages/apps/AppsPage.module.css */
.AppsPageContainer {
    padding: 1.25rem 1.5rem;
    overflow-y: auto;
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.toolbar {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex-shrink: 0;
}

.search {
    height: 2rem;
    padding: 0 0.75rem;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--fg);
    font-family: var(--font-mono);
    font-size: 0.78125rem;
    outline: none;
    width: 18rem;

    &:focus {
        border-color: var(--border-m);
    }

    &::placeholder {
        color: var(--fg-faint);
    }
}

.count {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
}

.grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(17rem, 1fr));
    gap: 0.75rem;
    align-content: start;
}
```

## AppCard CSS highlights

```css
/* src/components/apps/AppCard.module.css */
.AppCardContainer {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.875rem 1rem;
    cursor: pointer;
    transition: background 0.15s, border-color 0.15s;
}

.AppCardContainer:hover {
    background: var(--bg3);
    border-color: var(--border-m);
}

.extra {
    max-height: 0;
    overflow: hidden;
    transition: max-height 0.22s ease;
}

.AppCardContainer:hover .extra {
    max-height: 10rem;
}
```

Reuse all existing base classes (`nameRow`, `chips`, `image`, `node`) from `ServiceCard.module.css` — either import
the shared selectors or duplicate only what differs. Prefer importing `ServiceCard.module.css` classes where identical
to avoid drift.

## Routes update

```ts
// src/app/router/Routes.ts — add:
Apps = '/apps'
```

```tsx
// src/app/router/Router.tsx — add inside children:
{
    path: Routes.Apps,
        element
:
    <AppsPage/>,
}
,
```

## Notes

- Replace `MOCK_APPS` with a React Query `useQuery` call to `ListServices` from `service_api.proto` once the API
  endpoint maps to `src/processes/api/velez.ts`. Map response fields to `AppData` in `src/processes/mappings/`.
- The "Deploy" button in the hover block can navigate to `Routes.Deploy` or open a deploy modal — `console.log` is
  sufficient for this task.
- Do NOT change `ServiceCard.tsx` or the Deployments page.
- Run `yarn build` in `pkg/web/Velez-UI` after changes — task is complete only if the build passes with no TypeScript
  errors.
