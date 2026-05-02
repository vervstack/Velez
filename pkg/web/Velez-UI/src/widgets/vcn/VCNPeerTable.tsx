import cls from '@/widgets/vcn/VCNPeerTable.module.css';
import VCNPeerRow, { type VCNPeerData } from '@/components/vcn/VCNPeerRow';
import SectionLabel from '@/components/base/SectionLabel';

interface VCNPeerTableProps {
    peers: VCNPeerData[];
}

const TABLE_HEADERS = ['Name', 'Type', 'Latency', 'Status', 'RX', 'TX'];

export default function VCNPeerTable({ peers }: VCNPeerTableProps) {
    return (
        <div className={cls.VCNPeerTableContainer}>
            <div className={cls.sectionHeader}>
                <SectionLabel>Peers</SectionLabel>
            </div>
            <div className={cls.table}>
                <div className={cls.tableHeader}>
                    {TABLE_HEADERS.map(function renderHeader(h) {
                        return <span key={h} className={cls.headerCell}>{h}</span>;
                    })}
                </div>
                {peers.map(function renderRow(peer, i) {
                    return (
                        <VCNPeerRow
                            key={peer.id}
                            peer={peer}
                            last={i === peers.length - 1}
                        />
                    );
                })}
            </div>
        </div>
    );
}
