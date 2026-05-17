import { useQuery } from '@tanstack/react-query'
import { serviceService } from '@/processes/api/service'

const LIST_REQ = { paging: { limit: '50', offset: '0' } }

export function useListServicesQuery() {
    return useQuery({
        queryKey: ['services'] as const,
        queryFn: () => serviceService.listServices(LIST_REQ),
    })
}

export function useGetServiceQuery(key: string, enabled = true) {
    return useQuery({
        queryKey: ['service', key] as const,
        queryFn: () => serviceService.fetchService(key),
        enabled,
    })
}

export function useListDeploymentsQuery(serviceId: string, enabled = true) {
    return useQuery({
        queryKey: ['deployments', serviceId] as const,
        queryFn: () => serviceService.fetchDeployments(serviceId),
        enabled,
    })
}

export function useGetServiceMetricsQuery(serviceId: string) {
    return useQuery({
        queryKey: ['service-metrics', serviceId] as const,
        queryFn: () => serviceService.fetchServiceMetrics(serviceId),
    })
}

export function useGetServiceResourcesQuery(serviceId: string) {
    return useQuery({
        queryKey: ['service-resources', serviceId] as const,
        queryFn: () => serviceService.fetchServiceResources(serviceId),
    })
}

export function useGetServiceGraphQuery(serviceId: string) {
    return useQuery({
        queryKey: ['service-graph', serviceId] as const,
        queryFn: () => serviceService.fetchServiceGraph(serviceId),
    })
}

export function useListServiceEnvsQuery(serviceId: string) {
    return useQuery({
        queryKey: ['service-envs', serviceId] as const,
        queryFn: () => serviceService.fetchServiceEnvironments(serviceId),
    })
}

export function useGetVervonomiconQuery(serviceId: string) {
    return useQuery({
        queryKey: ['vervonomicon', serviceId] as const,
        queryFn: () => serviceService.fetchVervonomicon(serviceId),
    })
}
