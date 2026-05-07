import cls from '@/components/node/NodeCard.module.css';
import cn from 'classnames';
import StatusDot from '@/components/base/StatusDot';
import Badge from '@/components/base/Badge';
import MiniBar from '@/components/base/MiniBar';
import IconButton from '@/components/base/IconButton';
import {NodeBaseInfo, NodeStatus} from "@/app/api/velez";

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

export default function NodeCard({node, onShell, onDrain}: NodeCardProps) {
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
                {/*TODO add region to Node description*/}
                <div className={cls.region}>{'spb-1'}</div>
            </div>

            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>CPU</span>
                    <span className={cn(cls.metricValue, {
                        [cls.valueRed]: false, // node.cpu > 80,
                        [cls.valueAmber]: true, // node.cpu > 60 && node.cpu <= 80,
                    })}>
                        {/*TODO*/}
                        {50}%
                    </span>
                </div>
                <MiniBar val={50}/>
            </div>

            <div className={cls.metric}>
                <div className={cls.metricHeader}>
                    <span className={cls.metricLabel}>Memory</span>
                    <span className={
                        cn(cls.metricValue, {
                            [cls.valueRed]: true,//node.mem > 80,
                            [cls.valueAmber]: false, //node.mem > 60 && node.mem <= 80,
                        })}>
                        {/*TODO*/}
                        {81}%
                    </span>
                </div>
                <MiniBar val={81}/>
            </div>

            <div className={cls.servicesBlock}>
                {/*TODO Add services count*/}
                <span className={cls.servicesCount}>{5}</span>
                <span className={cls.servicesLabel}>services</span>
            </div>

            <div className={cls.actions}>
                <IconButton label="shell" title="Open terminal" onClick={onShell}/>
                <IconButton label="drain" title="Drain node" danger onClick={onDrain}/>
            </div>
        </div>
    );
}
