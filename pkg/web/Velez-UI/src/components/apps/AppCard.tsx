import cls from '@/components/apps/AppCard.module.css';
import serviceCls from '@/components/service/ServiceCard.module.css';
import StatusDot from '@/components/base/StatusDot';
import EnvChip from '@/components/base/chips/EnvChip';
import IncidentChip from '@/components/base/chips/IncidentChip';
import FreezeChip from '@/components/base/chips/FreezeChip';
import { AppData } from '@/processes/mappings/smerds';

interface AppCardProps {
    app: AppData;
    onOpen: (name: string) => void;
    onDeploy: (name: string) => void;
}

export default function AppCard({ app, onOpen, onDeploy }: AppCardProps) {
    function handleCardClick() {
        onOpen(app.name);
    }

    function handleOpenClick(e: React.MouseEvent) {
        e.stopPropagation();
        onOpen(app.name);
    }

    function handleDeployClick(e: React.MouseEvent) {
        e.stopPropagation();
        onDeploy(app.name);
    }

    return (
        <div
            className={cls.AppCardContainer}
            onClick={handleCardClick}
        >
            <div className={serviceCls.nameRow}>
                <StatusDot status={app.status} pulse />
                <span className={serviceCls.name}>{app.name}</span>
            </div>

            <div className={serviceCls.chips}>
                <EnvChip env={app.env} />
                {app.incident && <IncidentChip />}
                {app.releaseFrozen && <FreezeChip />}
            </div>

            <div className={serviceCls.image}>{app.image}</div>

            <div className={serviceCls.node}>
                <StatusDot status={app.node.status} />
                <span className={serviceCls.nodeId}>{app.node.id}</span>
                <span className={serviceCls.nodeHost}>{app.node.host}</span>
            </div>

            <div className={cls.extra}>
                <div className={cls.divider} />
                <div className={cls.detailsGrid}>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>deployments</span>
                        <span className={cls.detailValue}>{app.deployments}</span>
                    </div>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>last deploy</span>
                        <span className={cls.detailValue}>{app.lastDeployed}</span>
                    </div>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>config</span>
                        <span className={cls.detailValue}>{app.configSource}</span>
                    </div>
                    <div className={cls.detailRow}>
                        <span className={cls.detailLabel}>version</span>
                        <span className={cls.detailValue}>{app.version}</span>
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
