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
