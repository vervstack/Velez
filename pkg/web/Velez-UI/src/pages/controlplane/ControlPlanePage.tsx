import {useQuery, UseQueryResult} from '@tanstack/react-query';

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {ListNodesResponse, NodeStatus} from "@/app/api/velez";

import StatCard, {Level} from '@/components/base/StatCard';
import NodeHealthList from '@/widgets/controlplane/NodeHealthList';
import PluginMatrix from '@/widgets/controlplane/PluginMatrix';
import SkeletonNodeCard from '@/components/node/SkeletonNodeCard';

import {useToaster} from '@/app/hooks/toaster/Toaster';
import {ListNodes, ListPlugins} from '@/processes/api/control_plane';
import {CacheKey} from "@/app/query/Cache.ts";
import useSettings from '@/app/settings/state';
import {Service} from '@/model/services/Services';

export default function ControlPlanePage() {
    const toaster = useToaster();
    const {initReq} = useSettings();

    const nodesQuery = useQuery({
        queryKey: [CacheKey.Nodes],
        queryFn: () =>
            ListNodes()
                .catch(toaster.catchGrpc),
    });

    const pluginsQuery = useQuery({
        queryKey: [CacheKey.Plugins],
        queryFn: () =>
            ListPlugins(initReq())
                .catch(toaster.catchGrpc)
    });

    return (
        <div className={cls.ControlPlanePageContainer}>
            <StatsGrid
                nodesQuery={nodesQuery}/>

            <NodeList
                nodesQuery={nodesQuery}
            />

            <PluginsList
                pluginsQuery={pluginsQuery}
                nodesQuery={nodesQuery}
            />

        </div>
    );
}

interface StatsGridProps {
    nodesQuery: UseQueryResult<void | ListNodesResponse, Error>;
}

function StatsGrid({nodesQuery}: StatsGridProps) {
    const nodes = nodesQuery.data?.nodes || [];

    const onlineCount = nodes
        .filter(n => n.status === NodeStatus.NodeStatus_Online).length;
    const degradedCount = nodes
        .filter(n => n.status === NodeStatus.NodeStatus_Degraded).length;

    // TODO add listing services
    const offlineCount = 0;
    return (
        <div className={cls.StatsGridContainer}>
            <StatCard value={nodes.length}
                      label="Total nodes"
                      level={Level.INFO}
            />
            <StatCard value={onlineCount}
                      label="Online"
                      level={Level.Good}
            />
            <StatCard value={degradedCount}
                      label="Degraded"
                      level={Level.WARN}
            />
            <StatCard value={offlineCount}
                      label="Offline"
                      level={offlineCount == 0 ? Level.INFO : Level.ERROR}
            />
        </div>
    )
}

interface NodeListProps {
    nodesQuery: UseQueryResult<void | ListNodesResponse, Error>;
}

function NodeList({nodesQuery}: NodeListProps) {
    function handleShell() {
        alert('shell is not available yet');
    }

    function handleDrain() {
        alert('drain is not available yet');
    }

    if (nodesQuery.isLoading) {
        return (
            <div className={cls.skeletonList}>
                <SkeletonNodeCard/>
                <SkeletonNodeCard/>
                <SkeletonNodeCard/>
            </div>
        );
    }

    return (
        <NodeHealthList
            nodes={nodesQuery.data?.nodes || []}
            onShell={handleShell}
            onDrain={handleDrain}
        />
    );
}

interface PluginsListProps {
    pluginsQuery: UseQueryResult<void | Service[], Error>;
    nodesQuery: UseQueryResult<void | ListNodesResponse, Error>;
}

function PluginsList({pluginsQuery, nodesQuery}: PluginsListProps) {
    function handleEnable(pluginName: string, nodeId: string) {
        alert(`not implemented: enable ${pluginName} on ${nodeId}`);
    }

    function handleDisable(pluginName: string, nodeId: string) {
        alert(`not implemented: disable ${pluginName} on ${nodeId}`);
    }


    return (
        <PluginMatrix
            nodes={nodesQuery.data?.nodes || []}
            plugins={pluginsQuery.data || []}
            onEnable={handleEnable}
            onDisable={handleDisable}/>
    )
}

