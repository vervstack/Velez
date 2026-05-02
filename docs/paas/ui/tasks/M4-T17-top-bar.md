# T17 — TopBar Widget

**Files to create:**

- `src/widgets/topbar/TopBar.tsx`
- `src/widgets/topbar/TopBar.module.css`

## What it looks like

A 56px-tall horizontal bar at the top of the content area. Four regions left-to-right:

```
[←] cluster / node01 [degraded] [all nodes]  │  [Ctrl Plane] [VCN] [Deployments] [Search]  │  ● 2/3 nodes   [+ Deploy]
```

1. **Collapse button** (56×56) — toggles sidebar; shows `←` or `→`
2. **Breadcrumb** — `cluster / <nodeId>` or `cluster / all nodes` + optional degraded badge + "all nodes" pill toggle
3. **Tab switcher** — 4 tabs with active state (red background when active)
4. **Right zone** — node health counter (green dot + `2/3 nodes`) + Deploy button

## Props interface

```ts
type NavId = 'controlplane' | 'vcn' | 'deployments' | 'search';

interface TopBarNode {
    id: string;
    status: 'online' | 'degraded' | 'offline';
}

interface TopBarProps {
    collapsed: boolean;
    onCollapse: () => void;
    nodes: TopBarNode[];
    activeNodeId: string;
    showAllNodes: boolean;
    onToggleAllNodes: () => void;
    activeNav: NavId;
    onNavChange: (id: NavId) => void;
    onDeploy: () => void;
}
```

## Component

```tsx
// src/widgets/topbar/TopBar.tsx
import cls from '@/widgets/topbar/TopBar.module.css';
import cn from 'classnames';

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'search';

interface TopBarNode {
    id: string;
    status: 'online' | 'degraded' | 'offline';
}

interface TopBarProps {
    collapsed: boolean;
    onCollapse: () => void;
    nodes: TopBarNode[];
    activeNodeId: string;
    showAllNodes: boolean;
    onToggleAllNodes: () => void;
    activeNav: NavId;
    onNavChange: (id: NavId) => void;
    onDeploy: () => void;
}

const TABS: Array<{ id: NavId; label: string }> = [
    { id: 'controlplane', label: 'Control Plane' },
    { id: 'vcn',          label: 'VCN' },
    { id: 'deployments',  label: 'Deployments' },
    { id: 'search',       label: 'Search' },
];

export default function TopBar({
    collapsed,
    onCollapse,
    nodes,
    activeNodeId,
    showAllNodes,
    onToggleAllNodes,
    activeNav,
    onNavChange,
    onDeploy,
}: TopBarProps) {
    const activeNode = nodes.find(n => n.id === activeNodeId);
    const onlineCount = nodes.filter(n => n.status === 'online').length;

    return (
        <div className={cls.TopBarContainer}>
            {/* Collapse toggle */}
            <button
                className={cls.collapseBtn}
                onClick={onCollapse}
                title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            >
                {collapsed ? '→' : '←'}
            </button>

            {/* Breadcrumb */}
            <div className={cls.breadcrumb}>
                <span className={cls.crumbDim}>cluster</span>
                <span className={cls.crumbSep}>/</span>
                {showAllNodes ? (
                    <span className={cls.crumbActive}>all nodes</span>
                ) : (
                    <>
                        <span className={cls.crumbNode}>{activeNodeId}</span>
                        {activeNode?.status === 'degraded' && (
                            <span className={cls.degradedBadge}>degraded</span>
                        )}
                    </>
                )}
                <button
                    className={cn(cls.allNodesPill, { [cls.allNodesActive]: showAllNodes })}
                    onClick={onToggleAllNodes}
                >
                    all nodes
                </button>
            </div>

            {/* Tab switcher */}
            <div className={cls.tabs}>
                <div className={cls.tabGroup}>
                    {TABS.map(function renderTab(tab) {
                        function handleClick() { onNavChange(tab.id); }
                        return (
                            <button
                                key={tab.id}
                                className={cn(cls.tab, { [cls.tabActive]: activeNav === tab.id })}
                                onClick={handleClick}
                            >
                                {tab.label}
                            </button>
                        );
                    })}
                </div>
            </div>

            {/* Right zone */}
            <div className={cls.rightZone}>
                <div className={cls.healthCounter}>
                    <span className={cls.healthDot} />
                    {onlineCount}/{nodes.length} nodes
                </div>
                <button className={cls.deployBtn} onClick={onDeploy}>
                    + Deploy
                </button>
            </div>
        </div>
    );
}
```

## CSS

```css
/* src/widgets/topbar/TopBar.module.css */
.TopBarContainer {
    height: 3.5rem;
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--border);
    background: var(--bg2);
    flex-shrink: 0;
}

/* Collapse button */
.collapseBtn {
    width: 3.5rem;
    height: 3.5rem;
    flex-shrink: 0;
    background: transparent;
    border: none;
    border-right: 1px solid var(--border);
    color: var(--fg-dim);
    cursor: pointer;
    font-size: 0.875rem;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s;

    &:hover { color: var(--fg); }
}

/* Breadcrumb */
.breadcrumb {
    padding: 0 1.25rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    border-right: 1px solid var(--border);
    height: 100%;
    flex-shrink: 0;
}

.crumbDim {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-dim);
}

.crumbSep {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg-faint);
}

.crumbNode {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--fg);
    font-weight: 500;
}

.crumbActive {
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: var(--cyan);
    font-weight: 500;
}

.degradedBadge {
    font-family: var(--font-mono);
    font-size: 0.5625rem;
    color: var(--amber);
    border: 1px solid rgba(245, 166, 35, 0.44);
    border-radius: 0.25rem;
    padding: 0.0625rem 0.375rem;
}

.allNodesPill {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    font-weight: 500;
    padding: 0.1875rem 0.625rem;
    border-radius: 99px;
    cursor: pointer;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg-dim);
    margin-left: 0.25rem;
    transition: all 0.15s;

    &:hover {
        border-color: var(--border-m);
        color: var(--fg);
    }
}

.allNodesActive {
    border-color: rgba(10, 183, 238, 0.4);
    background: var(--cyan-dim);
    color: var(--cyan);
}

/* Tabs */
.tabs {
    padding: 0 1.25rem;
    display: flex;
    align-items: center;
    flex: 1;
}

.tabGroup {
    display: flex;
    gap: 0.125rem;
    background: var(--bg3);
    border-radius: 0.5625rem;
    padding: 0.1875rem;
    border: 1px solid var(--border);
}

.tab {
    font-family: var(--font-sans);
    font-size: 0.78125rem;
    font-weight: 400;
    padding: 0.375rem 1rem;
    border-radius: 0.4375rem;
    border: none;
    cursor: pointer;
    background: transparent;
    color: var(--fg-dim);
    transition: all 0.15s;

    &:hover:not(.tabActive) {
        color: var(--fg-mid);
    }
}

.tabActive {
    background: var(--red);
    color: #fff;
    font-weight: 700;
}

/* Right zone */
.rightZone {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0 1.25rem;
    border-left: 1px solid var(--border);
    height: 100%;
    flex-shrink: 0;
}

.healthCounter {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    font-family: var(--font-mono);
    font-size: 0.65625rem;
    color: var(--green);
}

.healthDot {
    width: 0.375rem;
    height: 0.375rem;
    border-radius: 50%;
    background: var(--green);
    display: block;
    animation: pulse 2s ease-in-out infinite;
}

.deployBtn {
    background: var(--red);
    border: none;
    border-radius: var(--radius-sm);
    color: #fff;
    font-family: var(--font-sans);
    font-weight: 700;
    font-size: 0.75rem;
    padding: 0.4375rem 1rem;
    cursor: pointer;
    transition: opacity 0.15s;
    white-space: nowrap;

    &:hover { opacity: 0.82; }
}
```
