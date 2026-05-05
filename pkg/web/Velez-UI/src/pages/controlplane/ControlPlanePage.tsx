import {useQuery} from '@tanstack/react-query';

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {NodeStatus, VervServiceState} from "@/app/api/velez";

import StatCard, {Level} from '@/components/base/StatCard';
import NodeHealthList from '@/widgets/controlplane/NodeHealthList';
import PluginMatrix from '@/widgets/controlplane/PluginMatrix';
import SkeletonNodeCard from '@/components/node/SkeletonNodeCard';

import {useToaster} from '@/app/hooks/toaster/Toaster';
import {ListNodes, ListVervServices} from '@/processes/api/control_plane';
import {CacheKey} from "@/app/query/Cache.ts";
import useSettings from '@/app/settings/state';
import {Service} from '@/model/services/Services';

export default function ControlPlanePage() {
    const toaster = useToaster();
    const { initReq } = useSettings();

    const nodesQuery = useQuery({
        queryKey: [CacheKey.Nodes],
        queryFn: () =>
            ListNodes()
                .catch(toaster.catchGrpc),
    });

    const pluginsQuery = useQuery({
        queryKey: ['vervServices'],
        queryFn: () =>
            ListVervServices(initReq())
                .catch((e) => { toaster.catchGrpc(e); return []; }),
    });

    const nodes = nodesQuery.data?.nodes || [];
    const vervServices = pluginsQuery.data || [];

    const onlineCount = nodes
        .filter(n => n.status === NodeStatus.NodeStatus_Online).length;
    const degradedCount = nodes
        .filter(n => n.status === NodeStatus.NodeStatus_Degraded).length;

    // TODO add listing services
    const offlineCount = 0;

    function handleShell() {
        alert('shell is not available yet');
    }

    function handleDrain() {
        alert('drain is not available yet');
    }

    function renderNodeList() {
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
                nodes={nodes}
                onShell={handleShell}
                onDrain={handleDrain}
            />
        );
    }

    return (
        <div className={cls.ControlPlanePageContainer}>
            <div className={cls.StatsGrid}>
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

            {renderNodeList()}

            {mapPluginsForDisplay(vervServices, nodes)}
        </div>
    );
}

function mapPluginsForDisplay(vervServices: Service[], nodes: any[]) {
    const plugins = vervServices.map(function mapService(service) {
        return {
            pluginName: service.type || 'unknown',
            tag: 'service',
            nodeStatuses: nodes.reduce((acc, node) => {
                acc[node.id || ''] = mapVervServiceStateToPluginStatus(service.state) as 'enabled' | 'disabled';
                return acc;
            }, {} as Record<string, 'enabled' | 'disabled'>),
            serviceKey: service.type || 'unknown',
        };
    });

    function handleEnable(pluginName: string, nodeId: string) {
        alert(`not implemented: enable ${pluginName} on ${nodeId}`);
    }

    function handleDisable(pluginName: string, nodeId: string) {
        alert(`not implemented: disable ${pluginName} on ${nodeId}`);
    }

    return <PluginMatrix nodes={nodes} plugins={plugins} onEnable={handleEnable} onDisable={handleDisable} />;
}

function mapVervServiceStateToPluginStatus(state?: VervServiceState): 'enabled' | 'disabled' | 'error' {
    switch (state) {
        case VervServiceState.running:
            return 'enabled';
        case VervServiceState.disabled:
            return 'disabled';
        case VervServiceState.warning:
        case VervServiceState.dead:
            return 'error';
        default:
            return 'disabled';
    }
}

