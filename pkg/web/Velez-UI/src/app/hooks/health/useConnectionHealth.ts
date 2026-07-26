import {useQuery} from '@tanstack/react-query'

import {serviceService} from '@/processes/api/service.ts'

export type ConnectionHealthStatus = 'checking' | 'healthy' | 'unhealthy'

export interface ConnectionHealth {
    status: ConnectionHealthStatus;
    lastCheckedAt: Date | null;
    error: string | null;
    recheck: () => void;
}

export function useConnectionHealth(): ConnectionHealth {
    const query = useQuery({
        queryKey: ['connection-health'] as const,
        queryFn: () => serviceService.listServices({}),
        retry: false,
        refetchOnWindowFocus: false,
    })

    function recheck() {
        query.refetch()
    }

    let status: ConnectionHealthStatus = 'checking'
    if (!query.isFetching) {
        status = query.isError ? 'unhealthy' : 'healthy'
    }

    const lastCheckedMs = query.errorUpdatedAt > query.dataUpdatedAt ? query.errorUpdatedAt : query.dataUpdatedAt

    return {
        status,
        lastCheckedAt: lastCheckedMs ? new Date(lastCheckedMs) : null,
        error: query.error ? (query.error as Error).message : null,
        recheck,
    }
}
