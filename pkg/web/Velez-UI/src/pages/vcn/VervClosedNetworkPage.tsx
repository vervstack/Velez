import cls from '@/pages/vcn/VervClosedNetwork.module.css';
import StatCard from '@/components/base/StatCard';
import NetworkTopologyMap from '@/widgets/vcn/NetworkTopologyMap';
import VCNPeerTable from '@/widgets/vcn/VCNPeerTable';
import CodeBlock from '@/components/complex/CodeBlock/CodeBlock';
import { type VCNPeerData } from '@/components/vcn/VCNPeerRow';

const MOCK_NODES = [
    { id: 'node01', host: '192.168.1.10', status: 'online'   as const },
    { id: 'node02', host: '192.168.1.42', status: 'online'   as const },
    { id: 'node03', host: '10.0.0.15',    status: 'degraded' as const },
];

const MOCK_EDGES = [
    { from: 'node01', to: 'node02', status: 'active'   as const },
    { from: 'node01', to: 'node03', status: 'active'   as const },
    { from: 'node02', to: 'node03', status: 'degraded' as const },
];

const MOCK_PEERS: VCNPeerData[] = [
    { id: 'p1', name: 'node01 ↔ node02', type: 'mesh',    latency: '1.2ms', status: 'active',   rx: '2.4 GB', tx: '1.8 GB' },
    { id: 'p2', name: 'node01 ↔ node03', type: 'mesh',    latency: '18ms',  status: 'active',   rx: '840 MB', tx: '620 MB' },
    { id: 'p3', name: 'node02 ↔ node03', type: 'mesh',    latency: '19ms',  status: 'degraded', rx: '120 MB', tx: '88 MB'  },
    { id: 'p4', name: 'edge-01',          type: 'gateway', latency: '—',     status: 'active',   rx: '5.2 GB', tx: '3.1 GB' },
    { id: 'p5', name: 'vpn-client-1',     type: 'client',  latency: '42ms',  status: 'active',   rx: '220 MB', tx: '90 MB'  },
];

const VCN_CONFIG = `# /etc/velez/vcn.yaml
network:
  mode: mesh
  subnet: 10.100.0.0/24
  peers:
    - host: 192.168.1.42 # node02
    - host: 10.0.0.15   # node03
  encryption: wireguard`;

const activeCount  = MOCK_PEERS.filter(p => p.status === 'active').length;
const totalRx      = '8.8 GB';
const totalTx      = '5.7 GB';

export default function VervClosedNetworkPage() {
    return (
        <div className={cls.VervClosedNetworkPageContainer}>
            {/* Stats */}
            <div className={cls.statsGrid}>
                <StatCard value={MOCK_PEERS.length} label="Tunnel peers"   color="var(--cyan)"  />
                <StatCard value={activeCount}        label="Active tunnels" color="var(--green)" />
                <StatCard value={totalRx}            label="Total RX"       color="var(--cyan)"  />
                <StatCard value={totalTx}            label="Total TX"       color="var(--cyan)"  />
            </div>

            <NetworkTopologyMap nodes={MOCK_NODES} edges={MOCK_EDGES} />

            <VCNPeerTable peers={MOCK_PEERS} />

            <CodeBlock title="vcn config — node01" code={VCN_CONFIG} />
        </div>
    );
}
