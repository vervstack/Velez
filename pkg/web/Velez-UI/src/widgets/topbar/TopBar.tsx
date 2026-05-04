import cn from 'classnames';

import cls from '@/widgets/topbar/TopBar.module.css';

import {NodeBaseInfo, NodeStatus} from "@/app/api/velez";

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'apps' | 'search';


interface TopBarProps {
    collapsed: boolean;
    onCollapse: () => void;

    nodes: NodeBaseInfo[];

    activeNodeId?: string;

    showAllNodes: boolean;
    onToggleAllNodes: () => void;
    activeNav: NavId;
    onNavChange: (id: NavId) => void;
    onDeploy: () => void;
}

export default function TopBar({
                                   collapsed,
                                   onCollapse,
                                   nodes,
                                   activeNodeId,
                                   showAllNodes,
                                   onToggleAllNodes,
                                   onDeploy,
                               }: TopBarProps) {
    const activeNode = nodes.find(n => n.id === activeNodeId);
    const onlineCount = nodes
        .filter(n => n.status === NodeStatus.NodeStatus_Online).length;

    return (
        <div className={cls.TopBarContainer}>
            <div className={cls.LeftZone}>
                <button
                    className={cls.CollapseBtn}
                    onClick={onCollapse}
                    title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                >
                    {collapsed ? '→' : '←'}
                </button>

                <div className={cls.Breadcrumbs}>
                    <span className={cls.CrumbDim}>cluster</span>
                    <span className={cls.CrumbSep}>/</span>
                    {showAllNodes ? (
                        <span className={cls.CrumbActive}>all nodes</span>
                    ) : (
                        <>
                            <span className={cls.CrumbNode}>{activeNodeId}</span>
                            {activeNode?.status === NodeStatus.NodeStatus_Degraded && (
                                <span className={cls.DegradedBadge}>degraded</span>
                            )}
                        </>
                    )}
                    <button
                        className={cn(cls.AllNodesPill, {[cls.AllNodesActive]: showAllNodes})}
                        onClick={onToggleAllNodes}
                    >
                        all nodes
                    </button>
                </div>
            </div>

            <div className={cls.RightZone}>
                <div className={cls.HealthCounter}>
                    <span className={cls.HealthDot}/>
                    {onlineCount}/{nodes.length} nodes
                </div>
                <button className={cls.DeployBtn} onClick={onDeploy}>
                    + Deploy
                </button>
            </div>
        </div>
    );
}
