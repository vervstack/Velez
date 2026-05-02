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
