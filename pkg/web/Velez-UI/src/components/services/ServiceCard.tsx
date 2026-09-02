import cls from '@/components/services/ServiceCard.module.css';
import serviceCls from '@/components/service/ServiceCard.module.css';
import StatusDot from '@/components/base/StatusDot';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import FreezeChip from '@/components/base/chips/FreezeChip';
import ServiceLabelBadge from '@/components/service/ServiceLabelBadge';
import { ServiceListItem } from '@/processes/mappings/smerds';

interface ServiceCardProps {
    service: ServiceListItem;
    onOpen: (name: string) => void;
    onDeploy: (name: string) => void;
}

export default function ServiceCard({ service, onOpen, onDeploy }: ServiceCardProps) {
    function handleCardClick() {
        onOpen(service.name);
    }

    function handleOpenClick(e: React.MouseEvent) {
        e.stopPropagation();
        onOpen(service.name);
    }

    function handleDeployClick(e: React.MouseEvent) {
        e.stopPropagation();
        onDeploy(service.name);
    }

    return (
        <div
            className={cls.ServiceCardContainer}
            onClick={handleCardClick}
        >
            <div className={serviceCls.nameRow}>
                <StatusDot status={service.status} pulse />
                <span className={serviceCls.name}>{service.name}</span>
            </div>

            <div className={serviceCls.chips}>
                <EnvChip env={service.env} />
                {service.incident && <IncidentChip />}
                {service.releaseFrozen && <FreezeChip />}
            </div>

            <ServiceLabelBadge labels={service.labels} />

            <div className={serviceCls.image}>{service.image}</div>

            <div className={serviceCls.node}>
                <StatusDot status={service.node.status} />
                <span className={serviceCls.nodeId}>{service.node.id}</span>
                <span className={serviceCls.nodeHost}>{service.node.host}</span>
            </div>

            <div className={cls.extra}>
                <div className={cls.divider} />
                <div className={cls.detailsGrid}>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>deployments</span>
                        <span className={cls.detailValue}>{service.deployments}</span>
                    </div>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>last deploy</span>
                        <span className={cls.detailValue}>{service.lastDeployed}</span>
                    </div>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>config</span>
                        <span className={cls.detailValue}>{service.configSource}</span>
                    </div>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>version</span>
                        <span className={cls.detailValue}>{service.version}</span>
                    </div>
                </div>
                <div className={cls.actionButtons}>
                    <button className={cls.actionButton} onClick={handleOpenClick}>
                        ↗ Open
                    </button>
                    <button className={cls.actionButton} onClick={handleDeployClick}>
                        ▶ Deploy
                    </button>
                </div>
            </div>
        </div>
    );
}
