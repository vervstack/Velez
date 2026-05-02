import { useState, useMemo } from 'react';
import cls from '@/pages/deployments/DeploymentsPage.module.css';
import DeploymentFilters from '@/widgets/deployments/DeploymentFilters';
import KanbanBoard from '@/widgets/deployments/KanbanBoard';
import ServiceListView from '@/widgets/deployments/ServiceListView';
import { type ServiceCardData } from '@/components/service/ServiceCard';

type ViewMode = 'kanban' | 'list';

const MOCK_SERVICES: ServiceCardData[] = [
    { name: 'matreshka-be', image: 'godverv/matreshka:latest', status: 'running',  cpu: 12, mem: 180,  uptime: '14d 3h',  restarts: 0,  env: 'prod',  incident: false, releaseFrozen: false, node: { id: 'node01', host: '192.168.1.10', status: 'online'   } },
    { name: 'api-gateway',  image: 'internal/gateway:v2.1',    status: 'running',  cpu: 33, mem: 512,  uptime: '6d 4h',   restarts: 0,  env: 'prod',  incident: true,  releaseFrozen: false, node: { id: 'node02', host: '192.168.1.42', status: 'online'   } },
    { name: 'postgres-main',image: 'postgres:16-alpine',        status: 'degraded', cpu: 89, mem: 1800, uptime: '3d 7h',   restarts: 12, env: 'prod',  incident: true,  releaseFrozen: false, node: { id: 'node03', host: '10.0.0.15',    status: 'degraded' } },
    { name: 'redis-cache',  image: 'redis:7-alpine',            status: 'degraded', cpu: 44, mem: 320,  uptime: '1h 20m',  restarts: 5,  env: 'prod',  incident: false, releaseFrozen: false, node: { id: 'node02', host: '192.168.1.42', status: 'online'   } },
    { name: 'prometheus',   image: 'prom/prometheus:latest',    status: 'stopped',  cpu: 0,  mem: 0,    uptime: 'stopped', restarts: 0,  env: 'stage', incident: false, releaseFrozen: false, node: { id: 'node01', host: '192.168.1.10', status: 'online'   } },
];

export default function DeploymentsPage() {
    const [search,        setSearch]        = useState('');
    const [statusFilters, setStatusFilters] = useState<Set<string>>(new Set());
    const [envFilters,    setEnvFilters]    = useState<Set<string>>(new Set());
    const [viewMode,      setViewMode]      = useState<ViewMode>('kanban');

    function handleToggleStatus(id: string) {
        setStatusFilters(prev => {
            const next = new Set(prev);
            next.has(id) ? next.delete(id) : next.add(id);
            return next;
        });
    }

    function handleToggleEnv(id: string) {
        setEnvFilters(prev => {
            const next = new Set(prev);
            next.has(id) ? next.delete(id) : next.add(id);
            return next;
        });
    }

    function handleClearAll() {
        setSearch('');
        setStatusFilters(new Set());
        setEnvFilters(new Set());
    }

    function handleServiceAction(serviceId: string, action: string) {
        console.log('action', action, serviceId);
    }

    const filtered = useMemo(function computeFiltered() {
        let result = MOCK_SERVICES;
        if (search.trim()) {
            const q = search.trim().toLowerCase();
            result = result.filter(s =>
                s.name.toLowerCase().includes(q) ||
                s.image.toLowerCase().includes(q)
            );
        }
        if (statusFilters.size > 0) {
            result = result.filter(s => statusFilters.has(s.status));
        }
        if (envFilters.size > 0) {
            result = result.filter(s => envFilters.has(s.env));
        }
        return result;
    }, [search, statusFilters, envFilters]);

    return (
        <div className={cls.DeploymentsPageContainer}>
            <DeploymentFilters
                search={search}
                onSearchChange={setSearch}
                statusFilters={statusFilters}
                onToggleStatus={handleToggleStatus}
                envFilters={envFilters}
                onToggleEnv={handleToggleEnv}
                onClearAll={handleClearAll}
                viewMode={viewMode}
                onViewModeChange={setViewMode}
                totalCount={filtered.length}
            />

            {viewMode === 'kanban' && (
                <KanbanBoard
                    services={filtered}
                    statusFilter={statusFilters}
                    onServiceAction={handleServiceAction}
                />
            )}

            {viewMode === 'list' && (
                <ServiceListView
                    services={filtered}
                    statusFilter={statusFilters}
                    onServiceAction={handleServiceAction}
                />
            )}
        </div>
    );
}
