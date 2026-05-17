import cn from 'classnames';

import cls from '@/widgets/topbar/TopBar.module.css';

import {NodeBaseInfo, NodeStatus, VervPluginType} from "@/app/api/velez";
import {ListNodesQuery, ListPluginsQuery} from "@/processes/queries/control_plane.ts";
import {useNavigate} from "react-router-dom";
import {Routes} from "@/app/router/Routes.ts";

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'apps' | 'search';


interface LeftSideProps {
    collapsed: boolean;
    onCollapse: () => void;

    showAllNodes: boolean;
    activeNodeId?: string;
    onToggleAllNodes: () => void;
}

interface TopBarProps extends LeftSideProps {
    activeNav: NavId;
    onNavChange: (id: NavId) => void;
}

export default function TopBar(props: TopBarProps) {


    return (
        <div className={cls.TopBarContainer}>
            <LeftZone
                {...props}
            />
            <RightZone/>
        </div>
    );
}


function LeftZone(props: LeftSideProps) {
    return (
        <div className={cls.LeftZoneContainer}>
            <button
                className={cls.CollapseBtn}
                onClick={props.onCollapse}
                title={props.collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            >
                {props.collapsed ? '→' : '←'}
            </button>

            <div className={cls.Breadcrumbs}>
                <span className={cls.CrumbDim}>cluster</span>
                <span className={cls.CrumbSep}>/</span>
                {props.showAllNodes ? (
                    <span className={cls.CrumbActive}>all nodes</span>
                ) : (
                    <>
                        <span className={cls.CrumbNode}>{props.activeNodeId}</span>
                        {/*TODO*/}
                        {/*{activeNode?.status === NodeStatus.NodeStatus_Degraded && (*/}
                        {/*    <span className={cls.DegradedBadge}>degraded</span>*/}
                        {/*)}*/}
                    </>
                )}
                <button
                    className={cn(cls.AllNodesPill, {[cls.AllNodesActive]: props.showAllNodes})}
                    onClick={props.onToggleAllNodes}
                >
                    all nodes
                </button>
            </div>
        </div>)
}

function RightZone() {
    const navigate = useNavigate();

    const pluginsQuery = ListPluginsQuery();
    const nodesQuery = ListNodesQuery();

    function handleDeploy() {
        navigate(Routes.Deploy);
    }


    const isLoading = pluginsQuery.isLoading && nodesQuery.isLoading

    const isStateFullMode = !pluginsQuery.isLoading &&
        pluginsQuery.data
            ?.find(p => p.type == VervPluginType.statefull_pg) !== undefined;

    return (
        <div className={cls.RightZoneContainer}>

            {!isLoading && (isStateFullMode ? <SingleNodeStub/> : <NodesHealthStatus/>)}
            <button className={cls.DeployBtn} onClick={handleDeploy}>
                + Deploy
            </button>
        </div>
    )
}

function SingleNodeStub() {
    return (
        <div>
            <p>Single node mode</p>
            <button>Setup statefull</button>
            </div>
    )
}

function NodesHealthStatus() {
    const nodesQuery = ListNodesQuery();

    function countOnlineNodes(nodes: NodeBaseInfo[]) {
        return nodes.filter(n => n.status === NodeStatus.NodeStatus_Online).length;
    }

    return (
        <div className={cls.HealthCounter}>
            <span className={cls.HealthDot}/>
            {countOnlineNodes(nodesQuery.data?.nodes || [])}
            /
            {(nodesQuery.data?.nodes || []).length} nodes
        </div>
    )
}
