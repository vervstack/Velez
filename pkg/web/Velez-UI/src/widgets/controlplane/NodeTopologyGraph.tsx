import cls from '@/widgets/controlplane/NodeTopologyGraph.module.css';

import SectionLabel from '@/components/base/SectionLabel';
import DatabasePixelIcon from '@/assets/icons/services/database-pixel.svg';

import {NodeBaseInfo, NodeStatus, VervPluginType} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins';

interface NodeTopologyGraphProps {
    nodes: NodeBaseInfo[];
    plugins: VervPlugin[];
}

const STATUS_COLOR: Record<NodeStatus, string> = {
    [NodeStatus.NodeStatus_Online]: 'var(--green)',
    [NodeStatus.NodeStatus_Degraded]: 'var(--amber)',
    [NodeStatus.NodeStatus_Offline]: 'var(--red)',
    [NodeStatus.NodeStatus_Unknown]: 'var(--fg-dim)',
};

const VIEW_WIDTH = 520;
const VIEW_HEIGHT = 320;
const CENTER_X = VIEW_WIDTH / 2;
const CENTER_Y = VIEW_HEIGHT / 2;
const RADIUS = 120;
const CENTER_NODE_RADIUS = 26;
const OUTER_NODE_RADIUS = 20;

function statusColor(status?: NodeStatus): string {
    return STATUS_COLOR[status ?? NodeStatus.NodeStatus_Unknown] ?? 'var(--fg-dim)';
}

function findMasterNodeId(plugins: VervPlugin[]): string | undefined {
    return plugins.find(p => p.type === VervPluginType.statefull_pg)?.masterNodeId;
}

interface Positioned {
    node: NodeBaseInfo;
    x: number;
    y: number;
}

function layoutOuterNodes(outerNodes: NodeBaseInfo[]): Positioned[] {
    const count = outerNodes.length;
    return outerNodes.map(function toPositioned(node, i) {
        const angle = (2 * Math.PI * i) / count;
        const x = CENTER_X + RADIUS * Math.cos(angle);
        const y = CENTER_Y + RADIUS * Math.sin(angle);
        return {node, x, y};
    });
}

export default function NodeTopologyGraph({nodes, plugins}: NodeTopologyGraphProps) {
    const masterNodeId = findMasterNodeId(plugins);
    const centerNode = nodes.find(n => n.id === masterNodeId) ?? nodes[0];
    const outerNodes = nodes.filter(n => n.id !== centerNode?.id);
    const positioned = layoutOuterNodes(outerNodes);

    return (
        <div className={cls.NodeTopologyGraphContainer}>
            <div className={cls.Header}>
                <SectionLabel>Node Topology</SectionLabel>
            </div>

            <svg
                className={cls.svg}
                viewBox={`0 0 ${VIEW_WIDTH} ${VIEW_HEIGHT}`}
                preserveAspectRatio="xMidYMid meet"
            >
                <defs>
                    <pattern id="cp-topology-grid" x="0" y="0" width="30" height="30" patternUnits="userSpaceOnUse">
                        <path d="M 30 0 L 0 0 0 30" fill="none" stroke="var(--border)" strokeWidth="0.5"/>
                    </pattern>
                </defs>

                <rect width={VIEW_WIDTH} height={VIEW_HEIGHT} fill="url(#cp-topology-grid)"/>

                {centerNode && positioned.map(function renderEdge(p) {
                    return (
                        <line
                            key={`edge-${p.node.id}`}
                            x1={CENTER_X} y1={CENTER_Y}
                            x2={p.x} y2={p.y}
                            stroke="var(--cyan)"
                            strokeWidth={1.5}
                            strokeOpacity={0.4}
                        />
                    );
                })}

                {positioned.map(function renderNode(p) {
                    const color = statusColor(p.node.status);
                    return (
                        <g key={p.node.id} className={cls.node} data-testid="topology-node">
                            <circle cx={p.x} cy={p.y} r={OUTER_NODE_RADIUS} fill="var(--bg2)" stroke={color} strokeWidth="1.5" strokeOpacity="0.7"/>
                            <circle cx={p.x} cy={p.y} r={5} fill={color} fillOpacity="0.85"/>
                            <text x={p.x} y={p.y + OUTER_NODE_RADIUS + 14} textAnchor="middle" fontFamily="var(--font-mono)" fontSize="10" fill="var(--fg-dim)">
                                {p.node.name ?? p.node.id}
                            </text>
                        </g>
                    );
                })}

                {centerNode && (
                    <g className={cls.node} data-testid="topology-node" data-center="true">
                        <circle cx={CENTER_X} cy={CENTER_Y} r={CENTER_NODE_RADIUS} fill="var(--bg2)" stroke={statusColor(centerNode.status)} strokeWidth="2" strokeOpacity="0.8"/>
                        <circle cx={CENTER_X} cy={CENTER_Y} r={6} fill={statusColor(centerNode.status)} fillOpacity="0.9"/>
                        <image
                            href={DatabasePixelIcon}
                            x={CENTER_X + 10}
                            y={CENTER_Y - CENTER_NODE_RADIUS - 8}
                            width={18}
                            height={18}
                        />
                        <text x={CENTER_X} y={CENTER_Y + CENTER_NODE_RADIUS + 16} textAnchor="middle" fontFamily="var(--font-mono)" fontSize="11" fill="var(--fg)">
                            {centerNode.name ?? centerNode.id}
                        </text>
                    </g>
                )}

                <g transform={`translate(${VIEW_WIDTH - 130}, ${VIEW_HEIGHT - 24})`}>
                    <image href={DatabasePixelIcon} x="0" y="-9" width="12" height="12"/>
                    <text x="16" y="0" fontFamily="var(--font-mono)" fontSize="9" fill="var(--fg-dim)">postgres master</text>
                </g>
            </svg>
        </div>
    );
}
