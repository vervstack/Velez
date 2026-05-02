import cls from '@/components/node/NodeCard.module.css';
import cn from 'classnames';
import StatusDot from '@/components/base/StatusDot';
import Badge from '@/components/base/Badge';
import MiniBar from '@/components/base/MiniBar';
import IconButton from '@/components/base/IconButton';

export interface NodeCardData {
    id: string;
    host: string;
    region: string;
    status: 'online' | 'degraded' | 'offline';
    cpu: number;
    mem: number;
    services: number;
}

interface NodeCardProps {
    node: NodeCardData;
    onShell?: () => void;
    onDrain?: () => void;
}

export default function NodeCard({ node, onShell, onDrain }: NodeCardProps) {
    return (
        <div className={cn(cls.NodeCardContainer, { [cls.degraded]: node.status === 'degraded' })}>
            <div className={cls.identity}>
                <div className={cls.nameRow}>
                    <StatusDot status={node.status} pulse />
                    <span className={cls.nodeId}>{node.id}</span>
                    {node.status === 'degraded' && (
                        <Badge label="degraded" color="var(--amber)" />
                    )}
                </div>
                <div className={cls.host}>{node.host}</div>
                <div className={cls.region}>{node.region}</div>
            </div>

            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>CPU</span>
                    <span className={cn(cls.metricValue, {
                        [cls.valueRed]: node.cpu > 80,
                        [cls.valueAmber]: node.cpu > 60 && node.cpu <= 80,
                    })}>
                        {node.cpu}%
                    </span>
                </div>
                <MiniBar val={node.cpu} />
            </div>

            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>Memory</span>
                    <span className={cn(cls.metricValue, {
                        [cls.valueRed]: node.mem > 80,
                        [cls.valueAmber]: node.mem > 60 && node.mem <= 80,
                    })}>
                        {node.mem}%
                    </span>
                </div>
                <MiniBar val={node.mem} />
            </div>

            <div className={cls.servicesBlock}>
                <span className={cls.servicesCount}>{node.services}</span>
                <span className={cls.servicesLabel}>services</span>
            </div>

            <div className={cls.actions}>
                <IconButton label="shell" title="Open terminal" onClick={onShell} />
                <IconButton label="drain" title="Drain node" danger onClick={onDrain} />
            </div>
        </div>
    );
}
