import {useQuery} from '@tanstack/react-query';

import cls from '@/pages/controlplane/ControlPlanePage.module.css';

import {NodeStatus} from "@/app/api/velez";

import StatCard, {Level} from '@/components/base/StatCard';
import NodeHealthList from '@/widgets/controlplane/NodeHealthList';
import PluginMatrix from '@/widgets/controlplane/PluginMatrix';
import SkeletonNodeCard from '@/components/node/SkeletonNodeCard';

import {useToaster} from '@/app/hooks/toaster/Toaster';
import {ListNodes} from '@/processes/api/control_plane';
import {CacheKey} from "@/app/query/Cache.ts";

const MOCK_PLUGINS = [
    {
        pluginName: 'Matreshka',
        tag: 'config',
        nodeStatuses: {node01: 'enabled' as const, node02: 'enabled' as const, node03: 'enabled' as const}
    },
    {
        pluginName: 'Makosh',
        tag: 'gRPC',
        nodeStatuses: {node01: 'enabled' as const, node02: 'enabled' as const, node03: 'disabled' as const}
    },
    {
        pluginName: 'Svarog',
        tag: 'secrets',
        nodeStatuses: {node01: 'enabled' as const, node02: 'enabled' as const, node03: 'enabled' as const}
    },
];

export default function ControlPlanePage() {
    const toaster = useToaster();

    const nodesQuery = useQuery({
        queryKey: [CacheKey.Nodes],
        queryFn: () =>
            ListNodes()
                .catch(toaster.catchGrpc),
    });

    const nodes = nodesQuery.data?.nodes || [];

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

            <PluginMatrix
                nodes={nodes}
                plugins={MOCK_PLUGINS}
            />
        </div>
    );
}
