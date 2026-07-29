import cn from 'classnames';
import {useIsFetching} from "@tanstack/react-query";

import cls from '@/widgets/topbar/TopBar.module.css';

import {NodeBaseInfo, NodeStatus, VervPluginType} from "@/app/api/velez";
import {IsStatefullModeEnabled, ListNodesQuery, ListPluginsQuery} from "@/processes/queries/control_plane.ts";
import {useNavigate} from "react-router-dom";
import {Routes} from "@/app/router/Routes.ts";
import Button from "@/components/base/Button.tsx";
import IconButton from "@/components/base/IconButton.tsx";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import {openStatefullPgDialog} from "@/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx";

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'apps' | 'search';


interface LeftSideProps {
    collapsed: boolean;
    onCollapse: () => void;

    showAllNodes: boolean;
    activeNodeId?: string;
    onToggleAllNodes: () => void;

    onToggleMobileNav: () => void;
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
            <div className={cls.HamburgerWrapper}>
                <IconButton
                    label="☰"
                    title="Toggle navigation"
                    onClick={props.onToggleMobileNav}
                />
            </div>
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
    const isFetching = useIsFetching() > 0;

    function handleDeploy() {
        navigate(Routes.Deploy);
    }

    const isLoading = pluginsQuery.isLoading && nodesQuery.isLoading

    const isStateFullMode = IsStatefullModeEnabled()

    return (
        <div className={cls.RightZoneContainer}>
            {isFetching && <FetchingIndicator/>}
            {!isLoading && (isStateFullMode ? <NodesHealthStatus/> : <SingleNodeStub/>)}
            <Button
                variant={'primary'}
                onClick={handleDeploy}
            >Deploy</Button>
        </div>
    )
}

function FetchingIndicator() {
    return (
        <div className={cls.FetchingIndicatorContainer} title="Loading data...">
            <span className={cls.FetchingDot}/>
        </div>
    )
}

function SingleNodeStub() {
    const pluginsQuery = ListPluginsQuery();

    function onClick() {
        const plugin = pluginsQuery.data?.find(p => p.type == VervPluginType.statefull_pg)
            ?? new VervPlugin(VervPluginType.statefull_pg, "");

        openStatefullPgDialog(plugin);
    }

    return (
        <div className={cls.SingleNodeContainer}>
            <span className={cls.SingleNodeLabel}>
                Single node mode
            </span>
            <Button
                variant={'warn'}
                onClick={onClick}
            >
                Setup
            </Button>
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
