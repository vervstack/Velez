import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import cls from '@/pages/deployments/DeploymentsPage.module.css';
import DeploymentFilters from '@/widgets/deployments/DeploymentFilters';
import KanbanBoard from '@/widgets/deployments/KanbanBoard';
import ServiceListView from '@/widgets/deployments/ServiceListView';
import { type ServiceCardData } from '@/components/service/ServiceCard';
import { FetchSmerds } from '@/processes/api/velez';
import { mapSmerdToServiceCard } from '@/processes/mappings/smerds';
import { useToaster } from '@/app/hooks/toaster/Toaster';

type ViewMode = 'kanban' | 'list';

export default function DeploymentsPage() {
    const toaster = useToaster();
    const [search,        setSearch]        = useState('');
    const [statusFilters, setStatusFilters] = useState<Set<string>>(new Set());
    const [envFilters,    setEnvFilters]    = useState<Set<string>>(new Set());
    const [viewMode,      setViewMode]      = useState<ViewMode>('kanban');

    const smerdsQuery = useQuery({
        queryKey: ['smerds'],
        queryFn: () => FetchSmerds().catch((e) => { toaster.catchGrpc(e); return { smerds: [] }; }),
    });

    const services: ServiceCardData[] = (smerdsQuery.data?.smerds ?? []).map(mapSmerdToServiceCard);

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
        let result = services;
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
    }, [services, search, statusFilters, envFilters]);

    if (smerdsQuery.isLoading) {
        return (
            <div className={cls.DeploymentsPageContainer}>
                <div className={cls.loadingSpinner}>Loading services...</div>
            </div>
        );
    }

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
