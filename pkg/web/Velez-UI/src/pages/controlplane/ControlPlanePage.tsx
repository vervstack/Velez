import cls from '@/pages/controlplane/ControlPlanePage.module.css';
import StatCard from '@/components/base/StatCard';
import NodeHealthList from '@/widgets/controlplane/NodeHealthList';
import PluginMatrix from '@/widgets/controlplane/PluginMatrix';
import { type NodeCardData } from '@/components/node/NodeCard';

const MOCK_NODES: NodeCardData[] = [
    { id: 'node01', host: '192.168.1.10', region: 'eu-west', status: 'online',   cpu: 38, mem: 52, services: 5 },
    { id: 'node02', host: '192.168.1.42', region: 'eu-west', status: 'online',   cpu: 71, mem: 68, services: 3 },
    { id: 'node03', host: '10.0.0.15',    region: 'us-east', status: 'degraded', cpu: 89, mem: 81, services: 2 },
];

const MOCK_PLUGINS = [
    { pluginName: 'Matreshka', tag: 'config',  nodeStatuses: { node01: 'enabled' as const, node02: 'enabled' as const, node03: 'enabled'  as const } },
    { pluginName: 'Makosh',    tag: 'gRPC',    nodeStatuses: { node01: 'enabled' as const, node02: 'enabled' as const, node03: 'disabled' as const } },
    { pluginName: 'Svarog',    tag: 'secrets', nodeStatuses: { node01: 'enabled' as const, node02: 'enabled' as const, node03: 'enabled'  as const } },
];

export default function ControlPlanePage() {
    const onlineCount  = MOCK_NODES.filter(n => n.status === 'online').length;
    const degradedCount = MOCK_NODES.filter(n => n.status === 'degraded').length;
    const totalServices = MOCK_NODES.reduce((sum, n) => sum + n.services, 0);

    function handleShell(nodeId: string) {
        console.log('shell', nodeId);
    }

    function handleDrain(nodeId: string) {
        console.log('drain', nodeId);
    }

    return (
        <div className={cls.ControlPlanePageContainer}>
            {/* Summary stats */}
            <div className={cls.statsGrid}>
                <StatCard value={MOCK_NODES.length} label="Total nodes"    color="var(--cyan)"  />
                <StatCard value={onlineCount}        label="Online"         color="var(--green)" />
                <StatCard value={degradedCount}      label="Degraded"       color="var(--amber)" />
                <StatCard value={totalServices}      label="Total services" color="var(--cyan)"  />
            </div>

            {/* Node health */}
            <NodeHealthList
                nodes={MOCK_NODES}
                onShell={handleShell}
                onDrain={handleDrain}
            />

            {/* Plugin matrix */}
            <PluginMatrix
                nodes={MOCK_NODES}
                plugins={MOCK_PLUGINS}
            />
        </div>
    );
}
