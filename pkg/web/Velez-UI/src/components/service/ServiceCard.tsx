import { useState } from 'react';
import cn from 'classnames';
import cls from '@/components/service/ServiceCard.module.css';
import StatusDot from '@/components/base/StatusDot';
import MiniBar from '@/components/base/MiniBar';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import FreezeChip from '@/components/base/chips/FreezeChip';
import ThreeDotMenu from '@/components/complex/ThreeDotMenu/ThreeDotMenu';

export interface ServiceCardData {
    name: string;
    image: string;
    status: 'running' | 'degraded' | 'stopped';
    cpu: number;
    mem: number;
    uptime: string;
    restarts: number;
    env: string;
    incident: boolean;
    releaseFrozen: boolean;
    node: { id: string; host: string; status: 'online' | 'degraded' | 'offline' };
}

interface ServiceCardProps {
    service: ServiceCardData;
    menuActions: Array<{ label: string; icon: string; danger?: boolean; onClick?: () => void }>;
}

export default function ServiceCard({ service, menuActions }: ServiceCardProps) {
    const [hovered, setHovered] = useState(false);

    function handleMouseEnter() { setHovered(true); }
    function handleMouseLeave() { setHovered(false); }

    function formatMem(mb: number) {
        return mb < 1000 ? mb + 'M' : (mb / 1000).toFixed(1) + 'G';
    }

    return (
        <div
            className={cn(cls.ServiceCardContainer, { [cls.hovered]: hovered })}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
        >
            <div className={cls.header}>
                <div className={cls.nameRow}>
                    <StatusDot status={service.status} pulse />
                    <span className={cls.name}>{service.name}</span>
                    {service.restarts > 0 && (
                        <span className={cls.restartsBadge} title={`${service.restarts} restarts`}>
                            ↺{service.restarts}
                        </span>
                    )}
                </div>
                <ThreeDotMenu actions={menuActions} />
            </div>

            <div className={cls.chips}>
                <EnvChip env={service.env} />
                {service.incident && <IncidentChip />}
                {service.releaseFrozen && <FreezeChip />}
            </div>

            <div className={cls.image}>{service.image}</div>

            <div className={cls.node}>
                <StatusDot status={service.node.status} />
                <span className={cls.nodeId}>{service.node.id}</span>
                <span className={cls.nodeHost}>{service.node.host}</span>
            </div>

            {service.status !== 'stopped' ? (
                <div className={cls.stats}>
                    <div className={cls.statRow}>
                        <div className={cls.statItem}>
                            <div className={cls.statHeader}>
                                <span className={cls.statLabel}>CPU</span>
                                <span className={cn(cls.statValue, {
                                    [cls.valueRed]: service.cpu > 80,
                                    [cls.valueAmber]: service.cpu > 60 && service.cpu <= 80,
                                })}>
                                    {service.cpu}%
                                </span>
                            </div>
                            <MiniBar val={service.cpu} />
                        </div>
                        <div className={cls.statItem}>
                            <div className={cls.statHeader}>
                                <span className={cls.statLabel}>MEM</span>
                                <span className={cls.statValue}>{formatMem(service.mem)}</span>
                            </div>
                            <MiniBar val={service.mem} max={2048} />
                        </div>
                    </div>
                    <div className={cls.uptime}>
                        <span className={cls.statLabel}>uptime</span>
                        <span className={cls.statValue}>{service.uptime}</span>
                    </div>
                </div>
            ) : (
                <div className={cls.stopped}>container stopped</div>
            )}
        </div>
    );
}
