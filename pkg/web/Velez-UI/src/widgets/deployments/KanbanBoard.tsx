import cls from '@/widgets/deployments/KanbanBoard.module.css';
import ServiceCard, { type ServiceCardData } from '@/components/service/ServiceCard';

interface KanbanBoardProps {
    services: ServiceCardData[];
    statusFilter?: Set<string>;
    onServiceAction: (serviceId: string, action: string) => void;
}

const COLUMNS = [
    { id: 'running',  label: 'Running',  color: 'var(--green)'   },
    { id: 'degraded', label: 'Degraded', color: 'var(--amber)'   },
    { id: 'stopped',  label: 'Stopped',  color: 'var(--fg-faint)'},
];

function buildActions(svc: ServiceCardData, onAction: (id: string, action: string) => void) {
    const actions: Array<{ label: string; icon: string; danger?: boolean; onClick: () => void }> = [];
    if (svc.status === 'running') {
        actions.push({ label: 'Restart',   icon: '↺', onClick: () => onAction(svc.name, 'restart') });
        actions.push({ label: 'Stop',      icon: '◼', danger: true, onClick: () => onAction(svc.name, 'stop') });
    }
    if (svc.status === 'stopped') {
        actions.push({ label: 'Start',     icon: '▶', onClick: () => onAction(svc.name, 'start') });
    }
    if (svc.status === 'degraded') {
        actions.push({ label: 'Restart',   icon: '↺', onClick: () => onAction(svc.name, 'restart') });
    }
    actions.push({ label: 'View logs', icon: '≈', onClick: () => onAction(svc.name, 'logs') });
    actions.push({ label: 'Shell',     icon: '$', onClick: () => onAction(svc.name, 'shell') });
    return actions;
}

export default function KanbanBoard({ services, statusFilter, onServiceAction }: KanbanBoardProps) {
    return (
        <div className={cls.KanbanBoardContainer}>
            {COLUMNS.map(function renderColumn(col) {
                if (statusFilter && statusFilter.size > 0 && !statusFilter.has(col.id)) return null;
                const colServices = services.filter(s => s.status === col.id);

                return (
                    <div key={col.id} className={cls.column}>
                        <div className={cls.columnHeader}>
                            <span className={cls.columnDot} style={{ background: col.color }} />
                            <span className={cls.columnLabel}>{col.label}</span>
                            <span className={cls.columnCount}>{colServices.length}</span>
                        </div>

                        <div className={cls.columnBody}>
                            {colServices.length === 0 && (
                                <div className={cls.empty}>
                                    <div className={cls.emptyDash}>—</div>
                                    <div className={cls.emptyText}>no services</div>
                                </div>
                            )}
                            {colServices.map(function renderCard(svc) {
                                return (
                                    <ServiceCard
                                        key={svc.name}
                                        service={svc}
                                        menuActions={buildActions(svc, onServiceAction)}
                                    />
                                );
                            })}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
