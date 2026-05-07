import cls from '@/widgets/sidebar/Sidebar.module.css';
import cn from 'classnames';
import StatusDot from '@/components/base/StatusDot';
import SectionLabel from '@/components/base/SectionLabel';
import VelezIcon from '@/assets/icons/services/velez.svg';
import {NodeBaseInfo, NodeStatus} from "@/app/api/velez";
import SkeletonNodeRow from '@/components/node/SkeletonNodeRow';

function mapNodeStatus(status?: NodeStatus): 'online' | 'offline' | 'degraded' | 'stopped' {
    switch (status) {
        case NodeStatus.NodeStatus_Online: return 'online';
        case NodeStatus.NodeStatus_Degraded: return 'degraded';
        case NodeStatus.NodeStatus_Offline: return 'offline';
        default: return 'stopped';
    }
}

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'apps' | 'search';
type ToolId = 'secrets' | 'config' | 'logs' | 'settings';

interface SidebarProps {
    collapsed: boolean;

    nodes: NodeBaseInfo[];
    activeNodeId?: string;
    onNodeSelect: (id?: string) => void;
    isNodesLoading: boolean;

    activeNav: NavId;
    onNavChange: (id: NavId) => void;
    onToolNav?: (id: ToolId) => void;
}

const NAV_ITEMS: Array<{ id: NavId; label: string; icon: string }> = [
    {id: 'controlplane', label: 'Control Plane', icon: '⬡'},
    {id: 'vcn', label: 'VCN', icon: '◎'},
    {id: 'deployments', label: 'Deployments', icon: '⬕'},
    {id: 'apps', label: 'Apps', icon: '⬡'},
    {id: 'search', label: 'Search', icon: '⌕'},
];

const TOOL_ITEMS = [
    {id: 'secrets', label: 'Secrets', icon: '⊡'},
    {id: 'config', label: 'Config', icon: '≡'},
    {id: 'logs', label: 'Logs', icon: '≈'},
    {id: 'settings', label: 'Settings', icon: '◈'},
];

export default function Sidebar(
    {
        collapsed, nodes,
        activeNodeId, onNodeSelect,
        activeNav, onNavChange,
        isNodesLoading, onToolNav
    }: SidebarProps) {


    function renderDot(node: NodeBaseInfo) {
        function handleClick() {
            onNodeSelect(node.id);
        }

        return (
            <div key={node.id} className={cls.dotRow} title={node.id} onClick={handleClick}>
                <StatusDot
                    status={mapNodeStatus(node.status)}/>
            </div>
        );
    }

    return (
        <aside className={
            cn(cls.SidebarContainer, {
                [cls.collapsed]: collapsed,
            })}>

            <Logo collapsed={collapsed}/>
            <NodesList
                isNodesLoading={isNodesLoading}
                collapsed={collapsed}
                nodes={nodes}
                onNodeSelect={onNodeSelect}
                activeNodeId={activeNodeId}
                activeNav={activeNav}
                onNavChange={onNavChange}/>

            {collapsed && (
                <div className={cls.nodesCollapsed}>
                    {isNodesLoading ? (
                        <>
                            <div className={cls.skeletonDot}/>
                            <div className={cls.skeletonDot}/>
                            <div className={cls.skeletonDot}/>
                        </>
                    ) : (
                        nodes.map(renderDot)
                    )}
                </div>
            )}

            <div className={cls.divider}/>

            {/* Main nav */}
            <nav className={cls.nav}>
                {!collapsed && (
                    <div className={cn(cls.sectionHeader, cls.navSectionHeader)}>
                        <SectionLabel>Services</SectionLabel>
                    </div>
                )}
                {
                    NAV_ITEMS.map((n) =>
                        <NavItem
                            id={n.id}
                            label={n.label}
                            icon={n.icon}
                            isActive={activeNav === n.id}
                            onNavChange={onNavChange}
                            collapsed={collapsed}
                        />)}

                <div className={cls.divider}/>
                {!collapsed && (
                    <div className={cn(cls.sectionHeader, cls.navSectionHeader)}>
                        <SectionLabel>Tools</SectionLabel>
                    </div>
                )}

                {TOOL_ITEMS.map(function renderToolItem(item) {
                    function handleToolClick() {
                        if (onToolNav) {
                            onToolNav(item.id as ToolId);
                        }
                    }

                    return (
                        <div
                            key={item.id}
                            className={cn(cls.toolItem, {[cls.toolItemCollapsed]: collapsed})}
                            title={collapsed ? item.label : undefined}
                            onClick={handleToolClick}
                        >
                            <span className={cls.toolIcon}>{item.icon}</span>
                            {!collapsed && <span className={cls.toolLabel}>{item.label}</span>}
                        </div>
                    );
                })}
            </nav>

            <UserBar
                collapsed={collapsed}/>

        </aside>
    );
}


function Logo({collapsed}: { collapsed: boolean }) {
    return (
        <div className={cls.LogoContainer}>
            <img src={VelezIcon} alt="Velez" className={cls.logoIcon}/>
            {!collapsed && (
                <span className={cls.logoText}>
                        Velez
                        <span className={cls.logoSub}> / VervStack</span>
                    </span>
            )}
        </div>
    )
}

function NodesList(
    {
        collapsed, nodes,
        onNodeSelect, activeNodeId,
        isNodesLoading
    }: SidebarProps) {

    if (collapsed) return null;


    function renderNode(node: NodeBaseInfo) {
        function handleClick() {
            onNodeSelect(node.id);
        }

        return (
            <div
                key={node.id}
                className={
                    cn(cls.nodeRow, {
                        [cls.nodeActive]: activeNodeId === node.id,
                    })}
                onClick={handleClick}
            >
                <StatusDot
                    status={mapNodeStatus(node.status)}/>

                <div className={cls.NodeInfo}>
                    <div className={cn(
                        cls.nodeId, {
                            [cls.nodeIdActive]: activeNodeId === node.id,
                        })}>
                        {node.name}
                    </div>

                    <div className={cls.nodeHost}>{node.addr}</div>
                </div>

                {
                    node.status === NodeStatus.NodeStatus_Degraded && (
                        <span className={cls.degradedMark}>!</span>
                    )}
            </div>
        );
    }

    return (
        <div className={cls.NodesSection}>
            <div className={cls.sectionHeader}>
                <SectionLabel>Nodes</SectionLabel>
            </div>

            {!isNodesLoading ? nodes.map(renderNode) :
                <>
                    <SkeletonNodeRow/>
                    <SkeletonNodeRow/>
                    <SkeletonNodeRow/>
                </>
            }
        </div>
    )
}


interface NavItemProps {
    id: NavId;
    label: string;
    icon: string;
    isActive: boolean;
    onNavChange: (id: NavId) => void;
    collapsed: boolean;
}

function NavItem({
                     isActive, onNavChange,
                     id, collapsed,
                     label, icon
                 }: NavItemProps) {

    function handleClick() {
        onNavChange(id);
    }

    return (
        <div
            key={id}
            className=
                {cn(cls.navItem, {
                    [cls.navItemActive]: isActive,
                    [cls.navItemCollapsed]: collapsed
                })}
            onClick={handleClick}
            title={collapsed ? label : undefined}
        >
            <span className={cls.navIcon}>{icon}</span>
            {!collapsed && (
                <span className={
                    cn(cls.navLabel, {
                        [cls.navLabelActive]: isActive,
                    })}>
                                    {label}
                                </span>
            )}
        </div>
    );
}

function UserBar({collapsed}: { collapsed: boolean }) {
    return (
        <div className={cls.UserBarContainer}>
            <div className={cls.Avatar}>RS</div>

            {!collapsed && (
                <div>
                    <div className={cls.UserName}>RedSock</div>
                    <div className={cls.UserRole}>admin</div>
                </div>
            )}
        </div>)
}
