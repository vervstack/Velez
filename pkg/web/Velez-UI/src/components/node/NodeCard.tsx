import cls from '@/components/node/NodeCard.module.css';
import cn from 'classnames';
import StatusDot from '@/components/base/StatusDot';
import Badge from '@/components/base/Badge';
import MiniBar from '@/components/base/MiniBar';
import IconButton from '@/components/base/IconButton';
import {NodeBaseInfo, NodeStatus} from "@/app/api/velez";

interface NodeCardProps {
    node: NodeBaseInfo;

    onShell?: () => void;
    onDrain?: () => void;
}

function mapNodeStatus(status?: NodeStatus): 'online' | 'offline' | 'degraded' | 'stopped' {
    switch (status) {
        case NodeStatus.NodeStatus_Online: return 'online';
        case NodeStatus.NodeStatus_Degraded: return 'degraded';
        case NodeStatus.NodeStatus_Offline: return 'offline';
        default: return 'stopped';
    }
}

function isMetricRed(percent: number): boolean {
    return percent > 80;
}

function isMetricAmber(percent: number): boolean {
    return percent > 60 && percent <= 80;
}

export default function NodeCard({node, onShell, onDrain}: NodeCardProps) {
    const cpuPercent = node.cpuPercent ?? 0;
    const memPercent = node.memPercent ?? 0;
    const servicesCount = Number(node.servicesCount ?? 0);

    return (
        <div className={
            cn(cls.NodeCardContainer,
                {
                    [cls.degraded]: node.status === NodeStatus.NodeStatus_Degraded,
                }
            )}>
            <div className={cls.identity}>
                <div className={cls.nameRow}>
                    <StatusDot status={mapNodeStatus(node.status)} pulse/>
                    <span className={cls.nodeId}>{node.id}</span>

                    {node.status === NodeStatus.NodeStatus_Degraded && (
                        <Badge label="degraded" color="var(--amber)"/>
                    )}
                </div>
                <div className={cls.host}>{node.addr}</div>
                <div className={cls.region}>{node.region}</div>
            </div>

            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>CPU</span>
                    <span className={cn(cls.metricValue, {
                        [cls.valueRed]: isMetricRed(cpuPercent),
                        [cls.valueAmber]: isMetricAmber(cpuPercent),
                    })}>
                        {Math.round(cpuPercent)}%
                    </span>
                </div>
                <MiniBar val={cpuPercent}/>
            </div>

            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>Memory</span>
                    <span className={
                        cn(cls.metricValue, {
                            [cls.valueRed]: isMetricRed(memPercent),
                            [cls.valueAmber]: isMetricAmber(memPercent),
                        })}>
                        {Math.round(memPercent)}%
                    </span>
                </div>
                <MiniBar val={memPercent}/>
            </div>

            <div className={cls.servicesBlock}>
                <span className={cls.servicesCount}>{servicesCount}</span>
                <span className={cls.servicesLabel}>services</span>
            </div>

            <div className={cls.actions}>
                <IconButton label="shell" title="Not implemented yet" onClick={onShell} disabled/>
                <IconButton label="drain" title="Not implemented yet" danger onClick={onDrain} disabled/>
            </div>
        </div>
    );
}
