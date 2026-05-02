import cls from '@/components/vcn/VCNPeerRow.module.css';
import cn from 'classnames';
import Badge from '@/components/base/Badge';

export interface VCNPeerData {
    id: string;
    name: string;
    type: 'mesh' | 'gateway' | 'client';
    latency: string;
    status: 'active' | 'degraded';
    rx: string;
    tx: string;
}

interface VCNPeerRowProps {
    peer: VCNPeerData;
    last?: boolean;
}

export default function VCNPeerRow({ peer, last }: VCNPeerRowProps) {
    const statusColor = peer.status === 'active' ? 'var(--green)' : 'var(--amber)';

    return (
        <div className={cn(cls.VCNPeerRowContainer, { [cls.last]: last })}>
            <span className={cls.name}>{peer.name}</span>
            <span className={cls.type}>{peer.type}</span>
            <span className={cls.latency}>{peer.latency}</span>
            <div>
                <Badge label={peer.status} color={statusColor} />
            </div>
            <span className={cls.traffic}>{peer.rx}</span>
            <span className={cls.traffic}>{peer.tx}</span>
        </div>
    );
}
