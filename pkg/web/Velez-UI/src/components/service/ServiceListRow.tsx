import { useState } from 'react';
import cn from 'classnames';
import cls from '@/components/service/ServiceListRow.module.css';
import StatusDot from '@/components/base/StatusDot';
import MiniBar from '@/components/base/MiniBar';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import FreezeChip from '@/components/base/chips/FreezeChip';
import ThreeDotMenu from '@/components/complex/ThreeDotMenu/ThreeDotMenu';
import { type ServiceCardData } from '@/components/service/ServiceCard';

interface ServiceListRowProps {
    service: ServiceCardData;
    menuActions: Array<{ label: string; icon: string; danger?: boolean; onClick?: () => void }>;
}

export default function ServiceListRow({ service, menuActions }: ServiceListRowProps) {
    const [hovered, setHovered] = useState(false);

    function handleMouseEnter() { setHovered(true); }
    function handleMouseLeave() { setHovered(false); }

    function formatMem(mb: number) {
        return mb < 1000 ? mb + 'M' : (mb / 1000).toFixed(1) + 'G';
    }

    return (
        <div
            className={cn(cls.ServiceListRowContainer, { [cls.hovered]: hovered })}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
        >
            <StatusDot status={service.status} pulse />

            <div className={cls.nameCell}>
                <div className={cls.nameRow}>
                    <span className={cls.name}>{service.name}</span>
                    {service.restarts > 0 && (
                        <span className={cls.restarts}>↺{service.restarts}</span>
                    )}
                </div>
                <div className={cls.chips}>
                    <EnvChip env={service.env} />
                    {service.incident && <IncidentChip />}
                    {service.releaseFrozen && <FreezeChip />}
                </div>
            </div>

            <div className={cls.nodeCell}>
                <StatusDot status={service.node.status} />
                <span className={cls.nodeId}>{service.node.id}</span>
            </div>

            <div className={cls.barCell}>
                {service.status !== 'stopped' ? (
                    <>
                        <div className={cls.barHeader}>
                            <span className={cls.barLabel}>CPU</span>
                            <span className={cn(cls.barValue, {
                                [cls.valueRed]: service.cpu > 80,
                                [cls.valueAmber]: service.cpu > 60 && service.cpu <= 80,
                            })}>
                                {service.cpu}%
                            </span>
                        </div>
                        <MiniBar val={service.cpu} />
                    </>
                ) : <span className={cls.stopped}>—</span>}
            </div>

            <div className={cls.barCell}>
                {service.status !== 'stopped' ? (
                    <>
                        <div className={cls.barHeader}>
                            <span className={cls.barLabel}>MEM</span>
                            <span className={cls.barValue}>{formatMem(service.mem)}</span>
                        </div>
                        <MiniBar val={service.mem} max={2048} />
                    </>
                ) : <span className={cls.stopped}>—</span>}
            </div>

            <div className={cls.uptimeCell}>
                {service.status !== 'stopped'
                    ? service.uptime
                    : <span className={cls.stopped}>stopped</span>
                }
            </div>

            <ThreeDotMenu actions={menuActions} />
        </div>
    );
}
