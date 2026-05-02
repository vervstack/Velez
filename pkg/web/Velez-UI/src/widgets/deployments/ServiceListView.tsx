import cls from '@/widgets/deployments/ServiceListView.module.css';
import ServiceListRow from '@/components/service/ServiceListRow';
import { type ServiceCardData } from '@/components/service/ServiceCard';

interface ServiceListViewProps {
    services: ServiceCardData[];
    statusFilter?: Set<string>;
    onServiceAction: (serviceId: string, action: string) => void;
}

const COLUMN_GROUPS = [
    { id: 'running',  label: 'Running',  color: 'var(--green)'   },
    { id: 'degraded', label: 'Degraded', color: 'var(--amber)'   },
    { id: 'stopped',  label: 'Stopped',  color: 'var(--fg-faint)'},
];

function buildActions(svc: ServiceCardData, onAction: (id: string, action: string) => void) {
    const actions: Array<{ label: string; icon: string; danger?: boolean; onClick: () => void }> = [];
    if (svc.status === 'running') {
        actions.push({ label: 'Restart', icon: '↺', onClick: () => onAction(svc.name, 'restart') });
        actions.push({ label: 'Stop',    icon: '◼', danger: true, onClick: () => onAction(svc.name, 'stop') });
    }
    if (svc.status === 'stopped') {
        actions.push({ label: 'Start',   icon: '▶', onClick: () => onAction(svc.name, 'start') });
    }
    if (svc.status === 'degraded') {
        actions.push({ label: 'Restart', icon: '↺', onClick: () => onAction(svc.name, 'restart') });
    }
    actions.push({ label: 'View logs', icon: '≈', onClick: () => onAction(svc.name, 'logs') });
    actions.push({ label: 'Shell',     icon: '$', onClick: () => onAction(svc.name, 'shell') });
    return actions;
}

export default function ServiceListView({ services, statusFilter, onServiceAction }: ServiceListViewProps) {
    const hasAny = services.length > 0;

    return (
        <div className={cls.ServiceListViewContainer}>
            {/* Sticky header */}
            <div className={cls.tableHeader}>
                {['', 'Service', 'Node', 'CPU', 'Memory', 'Uptime', ''].map(function renderHeader(h, i) {
                    return <span key={i} className={cls.headerCell}>{h}</span>;
                })}
            </div>

            {!hasAny && (
                <div className={cls.empty}>no services match current filters</div>
            )}

            {COLUMN_GROUPS.map(function renderGroup(group) {
                if (statusFilter && statusFilter.size > 0 && !statusFilter.has(group.id)) return null;
                const groupServices = services.filter(s => s.status === group.id);
                if (groupServices.length === 0) return null;

                return (
                    <div key={group.id}>
                        <div className={cls.groupLabel}>
                            <span className={cls.groupDot} style={{ background: group.color }} />
                            <span className={cls.groupName}>{group.label}</span>
                            <span className={cls.groupCount}>{groupServices.length}</span>
                        </div>
                        {groupServices.map(function renderRow(svc) {
                            return (
                                <ServiceListRow
                                    key={svc.name}
                                    service={svc}
                                    menuActions={buildActions(svc, onServiceAction)}
                                />
                            );
                        })}
                    </div>
                );
            })}
        </div>
    );
}
