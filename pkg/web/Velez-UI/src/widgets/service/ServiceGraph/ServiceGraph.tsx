import type { ServiceGraphNode } from '@/model/service_page/ServicePageModel'
import { useGetServiceGraphQuery } from '@/processes/queries/services'

import cls from './ServiceGraph.module.css'

interface ServiceGraphProps {
    serviceName: string
}

const SVG_W = 980
const SVG_H = 700
const CENTER_X = SVG_W / 2
const CENTER_Y = SVG_H / 2
const INCOMING_X = 130
const OUTGOING_X = SVG_W - 130
const CENTER_R = 42
const CENTER_RING_R = 62
const NODE_R = 44
const RECT_W = 116
const RECT_H = 40
const LINE_HEIGHT = 11

function labelLines(name: string): string[] {
    const MAX = 15
    const lines: string[] = []
    for (let i = 0; i < name.length; i += MAX) {
        lines.push(name.slice(i, i + MAX))
    }
    return lines.length > 0 ? lines : [name]
}

function spreadY(index: number, total: number): number {
    if (total === 1) return CENTER_Y
    const step = NODE_R * 2 + 62
    const span = (total - 1) * step
    return CENTER_Y - span / 2 + index * step
}

function nodeEdgeX(node: ServiceGraphNode, side: 'right' | 'left'): number {
    if (node.kind === 'resource') return side === 'right' ? RECT_W / 2 : -RECT_W / 2
    return side === 'right' ? NODE_R : -NODE_R
}

function renderNode(node: ServiceGraphNode, x: number, y: number, direction: 'incoming' | 'outgoing') {
    const isResource = node.kind === 'resource'
    const shapeClass = isResource
        ? (direction === 'incoming' ? cls.incomingResourceNode : cls.outgoingResourceNode)
        : (direction === 'incoming' ? cls.incomingServiceNode : cls.outgoingServiceNode)

    const shape = isResource ? (
        <rect
            x={x - RECT_W / 2}
            y={y - RECT_H / 2}
            width={RECT_W}
            height={RECT_H}
            rx={7}
            className={shapeClass}
        />
    ) : (
        <circle cx={x} cy={y} r={NODE_R} className={shapeClass} />
    )

    const lines = labelLines(node.id)
    const labelStartY = isResource ? y + RECT_H / 2 + 12 : y + NODE_R + 12

    return (
        <g key={node.id}>
            {shape}
            {lines.map(function renderLabelLine(part: string, i: number) {
                return <text key={i} x={x} y={labelStartY + i * LINE_HEIGHT} className={cls.nodeLabel}>{part}</text>
            })}
            <text x={x} y={labelStartY + lines.length * LINE_HEIGHT + 2} className={cls.nodeMeta}>{node.proto} · {node.rate}</text>
        </g>
    )
}

export default function ServiceGraph({serviceName}: ServiceGraphProps) {
    const {data, isLoading, isError} = useGetServiceGraphQuery(serviceName)

    const incoming: ServiceGraphNode[] = data?.incoming ?? []
    const outgoing: ServiceGraphNode[] = data?.outgoing ?? []

    if (isLoading) {
        return (
            <div className={cls.ServiceGraphContainer}>
                <p className={cls.Empty}>Loading graph…</p>
            </div>
        )
    }

    if (isError) {
        return (
            <div className={cls.ServiceGraphContainer}>
                <p className={cls.Empty}>Failed to load graph</p>
            </div>
        )
    }

    if (incoming.length === 0 && outgoing.length === 0) {
        return (
            <div className={cls.ServiceGraphContainer}>
                <p className={cls.Empty}>No graph data available</p>
            </div>
        )
    }

    return (
        <div className={cls.ServiceGraphContainer}>
            <div className={cls.SectionHeader}>
                <h2 className={cls.SectionTitle}>Service Graph</h2>
                <span className={cls.SectionSubtitle}>incoming on left · outgoing on right · live edge animation reflects request rate</span>
            </div>
            <div className={cls.SvgWrapper}>
                <div className={cls.LaneLabelsRow}>
                    <span className={cls.LaneLabel}>← incoming · {incoming.length} callers</span>
                    <span className={cls.LaneLabel}>outgoing · {outgoing.length} deps →</span>
                </div>

                <svg viewBox={`0 0 ${SVG_W} ${SVG_H}`} width="100%" className={cls.Graph}>
                    <defs>
                        <radialGradient id={`sg-glow-${serviceName}`} cx="50%" cy="50%" r="50%">
                            <stop offset="0%" stopColor="#ed2f32" stopOpacity="0.18" />
                            <stop offset="100%" stopColor="#ed2f32" stopOpacity="0" />
                        </radialGradient>
                        <marker id={`sg-arr-in-${serviceName}`} viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6"
                                markerHeight="6" orient="auto-start-reverse">
                            <path d="M0,1 L9,5 L0,9 z" fill="#0ab7ee" opacity="0.7" />
                        </marker>
                        <marker id={`sg-arr-out-${serviceName}`} viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6"
                                markerHeight="6" orient="auto-start-reverse">
                            <path d="M0,1 L9,5 L0,9 z" fill="#f5a623" opacity="0.75" />
                        </marker>
                    </defs>

                    <ellipse cx={CENTER_X} cy={CENTER_Y} rx="180" ry="140" fill={`url(#sg-glow-${serviceName})`}/>

                    <line x1={INCOMING_X + 90} y1={36} x2={INCOMING_X + 90} y2={SVG_H - 20} className={cls.separator} />
                    <line x1={OUTGOING_X - 90} y1={36} x2={OUTGOING_X - 90} y2={SVG_H - 20} className={cls.separator} />

                    {incoming.map(function renderIncomingEdge(node: ServiceGraphNode, i: number) {
                        const y = spreadY(i, incoming.length)
                        const startX = INCOMING_X + nodeEdgeX(node, 'right')
                        const c1x = startX + 160
                        const c2x = CENTER_X - 140
                        return (
                            <path
                                key={`in-edge-${node.id}`}
                                d={`M ${startX} ${y} C ${c1x} ${y}, ${c2x} ${CENTER_Y}, ${CENTER_X - CENTER_R} ${CENTER_Y}`}
                                className={cls.incomingEdge}
                                markerEnd={`url(#sg-arr-in-${serviceName})`}
                            />
                        )
                    })}

                    {outgoing.map(function renderOutgoingEdge(node: ServiceGraphNode, i: number) {
                        const y = spreadY(i, outgoing.length)
                        const endX = OUTGOING_X + nodeEdgeX(node, 'left')
                        const c1x = CENTER_X + 140
                        const c2x = endX - 160
                        return (
                            <path
                                key={`out-edge-${node.id}`}
                                d={`M ${CENTER_X + CENTER_R} ${CENTER_Y} C ${c1x} ${CENTER_Y}, ${c2x} ${y}, ${endX} ${y}`}
                                className={cls.outgoingEdge}
                                markerEnd={`url(#sg-arr-out-${serviceName})`}
                            />
                        )
                    })}

                    <circle cx={CENTER_X} cy={CENTER_Y} r={CENTER_RING_R} className={cls.centerRing} />
                    <circle cx={CENTER_X} cy={CENTER_Y} r={CENTER_R} className={cls.centerNode} />
                    <text x={CENTER_X} y={CENTER_Y} className={cls.centerMeta}>·service·</text>
                    {labelLines(serviceName).map(function renderCenterPart(part: string, i: number) {
                        return <text key={i} x={CENTER_X} y={CENTER_Y + CENTER_RING_R + 14 + i * LINE_HEIGHT} className={cls.centerLabel}>{part}</text>
                    })}

                    {incoming.map(function renderIncomingNode(node: ServiceGraphNode, i: number) {
                        return renderNode(node, INCOMING_X, spreadY(i, incoming.length), 'incoming')
                    })}

                    {outgoing.map(function renderOutgoingNode(node: ServiceGraphNode, i: number) {
                        return renderNode(node, OUTGOING_X, spreadY(i, outgoing.length), 'outgoing')
                    })}
                </svg>

                <div className={cls.Legend}>
                <div className={cls.LegendItem}>
                    <span className={`${cls.LegendLine} ${cls.legendIncoming}`} />
                    <span>incoming (callers)</span>
                </div>
                <div className={cls.LegendItem}>
                    <span className={`${cls.LegendLine} ${cls.legendOutgoing}`} />
                    <span>outgoing (deps)</span>
                </div>
                <div className={cls.LegendItem}>
                    <svg width="14" height="9" className={cls.LegendShape}>
                        <rect x="0" y="0" width="14" height="9" rx="2" />
                    </svg>
                    <span>resource</span>
                </div>
                <div className={cls.LegendItem}>
                    <svg width="10" height="10" className={cls.LegendShape}>
                        <circle cx="5" cy="5" r="4.5" />
                    </svg>
                    <span>service</span>
                </div>
                </div>
            </div>
        </div>
    )
}
