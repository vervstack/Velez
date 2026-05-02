# T16 — Sidebar Widget

**Files to create:**

- `src/widgets/sidebar/Sidebar.tsx`
- `src/widgets/sidebar/Sidebar.module.css`

## What it looks like

A collapsible left sidebar. Expanded width: 220px. Collapsed width: 56px. Transition is animated.

Sections top to bottom:

1. **Logo area** — Velez icon + "Velez / VervStack" text (hidden when collapsed)
2. **Nodes list** — clickable rows with status dot, node id, host (collapsed: dots only)
3. **Divider**
4. **Main nav** — 4 items (Control Plane, VCN, Deployments, Search). Active = red tint.
5. **Divider**
6. **Tools nav** — 4 items (Secrets, Config, Logs, Settings). No active state.
7. **Bottom user row** — avatar initials circle + name/role (hidden when collapsed)

## Props interface

```ts
interface SidebarNode {
    id: string;
    host: string;
    status: 'online' | 'degraded' | 'offline';
}

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'search';

interface SidebarProps {
    collapsed: boolean;
    nodes: SidebarNode[];
    activeNodeId: string;
    onNodeSelect: (id: string) => void;
    activeNav: NavId;
    onNavChange: (id: NavId) => void;
}
```

## Component

```tsx
// src/widgets/sidebar/Sidebar.tsx
import cls from '@/widgets/sidebar/Sidebar.module.css';
import cn from 'classnames';
import StatusDot from '@/components/base/StatusDot';
import SectionLabel from '@/components/base/SectionLabel';
import VelezIcon from '@/assets/icons/services/velez.svg';

interface SidebarNode {
    id: string;
    host: string;
    status: 'online' | 'degraded' | 'offline';
}

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'search';

interface SidebarProps {
    collapsed: boolean;
    nodes: SidebarNode[];
    activeNodeId: string;
    onNodeSelect: (id: string) => void;
    activeNav: NavId;
    onNavChange: (id: NavId) => void;
}

const NAV_ITEMS: Array<{ id: NavId; label: string; icon: string }> = [
    { id: 'controlplane', label: 'Control Plane', icon: '⬡' },
    { id: 'vcn',          label: 'VCN',           icon: '◎' },
    { id: 'deployments',  label: 'Deployments',   icon: '⬕' },
    { id: 'search',       label: 'Search',        icon: '⌕' },
];

const TOOL_ITEMS = [
    { id: 'secrets',  label: 'Secrets',  icon: '⊡' },
    { id: 'config',   label: 'Config',   icon: '≡' },
    { id: 'logs',     label: 'Logs',     icon: '≈' },
    { id: 'settings', label: 'Settings', icon: '◈' },
];

export default function Sidebar({
    collapsed,
    nodes,
    activeNodeId,
    onNodeSelect,
    activeNav,
    onNavChange,
}: SidebarProps) {
    return (
        <aside className={cn(cls.SidebarContainer, { [cls.collapsed]: collapsed })}>
            {/* Logo */}
            <div className={cls.logo}>
                <img src={VelezIcon} alt="Velez" className={cls.logoIcon} />
                {!collapsed && (
                    <span className={cls.logoText}>
                        Velez
                        <span className={cls.logoSub}> / VervStack</span>
                    </span>
                )}
            </div>

            {/* Nodes */}
            {!collapsed && (
                <div className={cls.nodesSection}>
                    <div className={cls.sectionHeader}>
                        <SectionLabel>Nodes</SectionLabel>
                    </div>
                    {nodes.map(function renderNode(node) {
                        function handleClick() { onNodeSelect(node.id); }
                        return (
                            <div
                                key={node.id}
                                className={cn(cls.nodeRow, { [cls.nodeActive]: activeNodeId === node.id })}
                                onClick={handleClick}
                            >
                                <StatusDot status={node.status} />
                                <div className={cls.nodeInfo}>
                                    <div className={cn(cls.nodeId, { [cls.nodeIdActive]: activeNodeId === node.id })}>
                                        {node.id}
                                    </div>
                                    <div className={cls.nodeHost}>{node.host}</div>
                                </div>
                                {node.status === 'degraded' && (
                                    <span className={cls.degradedMark}>!</span>
                                )}
                            </div>
                        );
                    })}
                </div>
            )}

            {collapsed && (
                <div className={cls.nodesCollapsed}>
                    {nodes.map(function renderDot(node) {
                        function handleClick() { onNodeSelect(node.id); }
                        return (
                            <div key={node.id} className={cls.dotRow} title={node.id} onClick={handleClick}>
                                <StatusDot status={node.status} />
                            </div>
                        );
                    })}
                </div>
            )}

            <div className={cls.divider} />

            {/* Main nav */}
            <nav className={cls.nav}>
                {!collapsed && (
                    <div className={cn(cls.sectionHeader, cls.navSectionHeader)}>
                        <SectionLabel>Services</SectionLabel>
                    </div>
                )}
                {NAV_ITEMS.map(function renderNavItem(item) {
                    const active = activeNav === item.id;
                    function handleClick() { onNavChange(item.id); }
                    return (
                        <div
                            key={item.id}
                            className={cn(cls.navItem, { [cls.navItemActive]: active, [cls.navItemCollapsed]: collapsed })}
                            onClick={handleClick}
                            title={collapsed ? item.label : undefined}
                        >
                            <span className={cls.navIcon}>{item.icon}</span>
                            {!collapsed && (
                                <span className={cn(cls.navLabel, { [cls.navLabelActive]: active })}>
                                    {item.label}
                                </span>
                            )}
                        </div>
                    );
                })}

                <div className={cls.divider} />
                {!collapsed && (
                    <div className={cn(cls.sectionHeader, cls.navSectionHeader)}>
                        <SectionLabel>Tools</SectionLabel>
                    </div>
                )}

                {TOOL_ITEMS.map(function renderToolItem(item) {
                    return (
                        <div
                            key={item.id}
                            className={cn(cls.toolItem, { [cls.toolItemCollapsed]: collapsed })}
                            title={collapsed ? item.label : undefined}
                        >
                            <span className={cls.toolIcon}>{item.icon}</span>
                            {!collapsed && <span className={cls.toolLabel}>{item.label}</span>}
                        </div>
                    );
                })}
            </nav>

            {/* User */}
            <div className={cn(cls.user, { [cls.userCollapsed]: collapsed })}>
                <div className={cls.avatar}>RS</div>
                {!collapsed && (
                    <div className={cls.userInfo}>
                        <div className={cls.userName}>RedSock</div>
                        <div className={cls.userRole}>admin</div>
                    </div>
                )}
            </div>
        </aside>
    );
}
```

## CSS

```css
/* src/widgets/sidebar/Sidebar.module.css */
.SidebarContainer {
    width: 13.75rem; /* 220px */
    flex-shrink: 0;
    background: var(--bg2);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    transition: width 0.22s cubic-bezier(0.4, 0, 0.2, 1);
    overflow: hidden;
}

.SidebarContainer.collapsed {
    width: 3.5rem; /* 56px */
}

/* Logo */
.logo {
    height: 3.5rem;
    display: flex;
    align-items: center;
    padding: 0 1.125rem;
    border-bottom: 1px solid var(--border);
    gap: 0.625rem;
    flex-shrink: 0;
}

.collapsed .logo {
    padding: 0 0.875rem;
}

.logoIcon {
    width: 1.375rem;
    height: 1.375rem;
    flex-shrink: 0;
}

.logoText {
    font-family: var(--font-sans);
    font-weight: 700;
    font-size: 0.9375rem;
    color: var(--fg);
    white-space: nowrap;
}

.logoSub {
    color: var(--fg-dim);
    font-weight: 300;
    margin-left: 0.3125rem;
    font-size: 0.8125rem;
}

/* Nodes */
.nodesSection {
    padding: 0.875rem 0.75rem 0.5rem;
    flex-shrink: 0;
}

.sectionHeader {
    margin-bottom: 0.5rem;
    padding-left: 0.375rem;
}

.nodeRow {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4375rem 0.5rem;
    border-radius: 0.4375rem;
    cursor: pointer;
    background: transparent;
    border: 1px solid transparent;
    margin-bottom: 0.125rem;
    transition: background 0.12s, border-color 0.12s;

    &:hover {
        background: var(--bg3);
    }
}

.nodeRow.nodeActive {
    background: var(--bg3);
    border-color: var(--border);
}

.nodeInfo {
    flex: 1;
    min-width: 0;
}

.nodeId {
    font-family: var(--font-mono);
    font-size: 0.71875rem;
    color: var(--fg-mid);
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.nodeIdActive { color: var(--fg); }

.nodeHost {
    font-family: var(--font-mono);
    font-size: 0.59375rem;
    color: var(--fg-dim);
}

.degradedMark {
    font-size: 0.5625rem;
    color: var(--amber);
}

/* Collapsed node dots */
.nodesCollapsed {
    padding: 0.625rem 0;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
    align-items: center;
}

.dotRow {
    cursor: pointer;
    padding: 0.375rem;
}

/* Divider */
.divider {
    height: 1px;
    background: var(--border);
    margin: 0.25rem 0.75rem;
    flex-shrink: 0;
}

/* Nav */
.nav {
    padding: 0.5rem;
    flex: 1;
    display: flex;
    flex-direction: column;
}

.navSectionHeader {
    padding-left: 0.375rem;
    margin-bottom: 0.5rem;
    flex-shrink: 0;
}

.navItem {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.5625rem 0.625rem;
    border-radius: 0.5rem;
    cursor: pointer;
    margin-bottom: 0.125rem;
    background: transparent;
    border: 1px solid transparent;
    color: var(--fg-dim);
    transition: background 0.12s, color 0.12s;

    &:hover:not(.navItemActive) {
        background: var(--bg3);
    }
}

.navItem.navItemCollapsed {
    padding: 0.625rem 0;
    justify-content: center;
}

.navItem.navItemActive {
    background: var(--red-dim);
    border-color: rgba(237, 47, 50, 0.13);
    color: var(--red);
}

.navIcon { font-size: 0.875rem; flex-shrink: 0; }

.navLabel {
    font-family: var(--font-sans);
    font-size: 0.8125rem;
    font-weight: 400;
}

.navLabelActive { font-weight: 700; }

/* Tool items */
.toolItem {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.5rem 0.625rem;
    border-radius: 0.5rem;
    cursor: pointer;
    margin-bottom: 0.0625rem;
    color: var(--fg-dim);
    transition: background 0.12s, color 0.12s;

    &:hover {
        background: var(--bg3);
        color: var(--fg-mid);
    }
}

.toolItem.toolItemCollapsed {
    padding: 0.5rem 0;
    justify-content: center;
}

.toolIcon { font-size: 0.8125rem; flex-shrink: 0; }

.toolLabel {
    font-family: var(--font-sans);
    font-size: 0.78125rem;
}

/* User */
.user {
    border-top: 1px solid var(--border);
    padding: 0.75rem 0.875rem;
    display: flex;
    align-items: center;
    gap: 0.625rem;
    flex-shrink: 0;
}

.user.userCollapsed {
    justify-content: center;
    padding: 0.75rem 0;
}

.avatar {
    width: 1.75rem;
    height: 1.75rem;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--red), var(--cyan));
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: 0.6875rem;
    color: #fff;
    font-weight: 600;
}

.userInfo { }

.userName {
    font-family: var(--font-sans);
    font-size: 0.75rem;
    color: var(--fg);
    font-weight: 600;
}

.userRole {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
}
```
