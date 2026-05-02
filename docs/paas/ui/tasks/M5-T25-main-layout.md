# T25 — MainLayout (rebuild)

**File to modify:** `src/app/router/MainLayout.tsx`

## What it does

Replaces the current top-navbar layout with a sidebar + topbar shell layout matching the design. The route-level content
is rendered inside the main content area via `<Outlet />`.

## Current layout (to replace)

Current `MainLayout` renders: `<PageHeader>` (top nav) → `<Outlet>` → `<Toaster>`.

## New layout structure

```
┌─────────┬──────────────────────────────────────────┐
│         │ TopBar                                    │
│ Sidebar ├──────────────────────────────────────────┤
│         │                                          │
│         │  <Outlet /> (page content)               │
│         │                                          │
└─────────┴──────────────────────────────────────────┘
```

State lives in MainLayout and is passed down to Sidebar and TopBar:

- `collapsed` — sidebar collapsed boolean
- `activeNodeId` — selected node id (default to first node or `''`)
- `showAllNodes` — boolean toggling cluster-wide view
- `activeNav` — current tab/nav id (synced with router)

## API data

Nodes come from the `ListSmerds` or a cluster nodes endpoint. For now, use static mock nodes until the API endpoint for
listing nodes is available. Use `useQuery` from `@tanstack/react-query`.

## Component

```tsx
// src/app/router/MainLayout.tsx
import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import cls from '@/app/router/MainLayout.module.css';
import Sidebar from '@/widgets/sidebar/Sidebar';
import TopBar from '@/widgets/topbar/TopBar';
import { Toaster } from '@/segments/Toaster';
import { Routes } from '@/app/router/Router';

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'search';

const MOCK_NODES = [
    { id: 'node01', host: '192.168.1.10', status: 'online'   as const },
    { id: 'node02', host: '192.168.1.42', status: 'online'   as const },
    { id: 'node03', host: '10.0.0.15',    status: 'degraded' as const },
];

const NAV_TO_ROUTE: Record<NavId, string> = {
    controlplane: Routes.ControlPlane,
    vcn:          Routes.VCN,
    deployments:  Routes.Deploy,
    search:       Routes.Home,
};

const ROUTE_TO_NAV: Record<string, NavId> = {
    [Routes.ControlPlane]: 'controlplane',
    [Routes.VCN]:          'vcn',
    [Routes.Deploy]:       'deployments',
    [Routes.Home]:         'search',
};

export default function MainLayout() {
    const navigate = useNavigate();
    const location = useLocation();

    const [collapsed, setCollapsed] = useState(false);
    const [activeNodeId, setActiveNodeId] = useState(MOCK_NODES[0].id);
    const [showAllNodes, setShowAllNodes] = useState(false);

    const activeNav: NavId = ROUTE_TO_NAV[location.pathname] ?? 'controlplane';

    function handleCollapse() {
        setCollapsed(prev => !prev);
    }

    function handleNavChange(id: NavId) {
        navigate(NAV_TO_ROUTE[id]);
    }

    function handleToggleAllNodes() {
        setShowAllNodes(prev => !prev);
    }

    function handleDeploy() {
        navigate(Routes.Deploy);
    }

    return (
        <div className={cls.MainLayoutContainer}>
            <Sidebar
                collapsed={collapsed}
                nodes={MOCK_NODES}
                activeNodeId={activeNodeId}
                onNodeSelect={setActiveNodeId}
                activeNav={activeNav}
                onNavChange={handleNavChange}
            />

            <div className={cls.content}>
                <TopBar
                    collapsed={collapsed}
                    onCollapse={handleCollapse}
                    nodes={MOCK_NODES}
                    activeNodeId={activeNodeId}
                    showAllNodes={showAllNodes}
                    onToggleAllNodes={handleToggleAllNodes}
                    activeNav={activeNav}
                    onNavChange={handleNavChange}
                    onDeploy={handleDeploy}
                />

                <main className={cls.main}>
                    <Outlet />
                </main>
            </div>

            <Toaster />
        </div>
    );
}
```

**Also create** `src/app/router/MainLayout.module.css`:

```css
/* src/app/router/MainLayout.module.css */
.MainLayoutContainer {
    display: flex;
    width: 100%;
    height: 100%;
    overflow: hidden;
    background: var(--bg);
}

.content {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    overflow: hidden;
}

.main {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
}
```

## Notes

- `MOCK_NODES` is temporary — replace with a React Query call to list cluster nodes once the API endpoint is available (
  tracked separately).
- `Routes` values must match the existing `src/app/router/Router.tsx` enum — do not change route paths.
- Remove the old `PageHeader` import and usage; `SettingsWidget` access should move to the sidebar Tools → Settings
  item (for now it can be omitted).
- The `Toaster` import path may need adjustment — check the existing export from `src/segments/Toaster.tsx`.
