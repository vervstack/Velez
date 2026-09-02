import cls from '@/widgets/topbar/TopBar.module.css';

import {NodeBaseInfo, NodeStatus, VervPluginType} from "@/app/api/velez";
import {IsStatefullModeEnabled, ListNodesQuery, ListPluginsQuery} from "@/processes/queries/control_plane.ts";
import Button from "@/components/base/Button.tsx";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import {openStatefullPgDialog} from "@/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx";
import {useBreadcrumbs} from "@/app/hooks/breadcrumbs/Breadcrumbs.ts";
import BreadcrumbsBar from "@/components/complex/BreadcrumbsBar/BreadcrumbsBar.tsx";

export default function TopBar() {
    return (
        <div className={cls.TopBarContainer}>
            <LeftZone/>
            <RightZone/>
        </div>
    );
}

function LeftZone() {
    const crumbs = useBreadcrumbs((s) => s.crumbs);

    if (crumbs.length === 0) return <div className={cls.LeftZoneContainer}/>;

    return (
        <div className={cls.LeftZoneContainer}>
            <BreadcrumbsBar crumbs={crumbs}/>
        </div>
    );
}

function RightZone() {
    const pluginsQuery = ListPluginsQuery();
    const nodesQuery = ListNodesQuery();

    const isLoading = pluginsQuery.isLoading && nodesQuery.isLoading

    const isStateFullMode = IsStatefullModeEnabled()

    return (
        <div className={cls.RightZoneContainer}>
            {!isLoading && (isStateFullMode ? <NodesHealthStatus/> : <SingleNodeStub/>)}
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
