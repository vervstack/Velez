import { queryOptions } from '@tanstack/react-query'
import {
    ListServices,
    FetchService,
    FetchDeployments,
    FetchServiceMetrics,
    FetchServiceResources,
    FetchServiceGraph,
    FetchServiceEnvironments,
    FetchVervonomicon,
} from '@/processes/api/service'

const LIST_REQ = { paging: { limit: '50', offset: '0' } }

export function listServicesQuery() {
    return queryOptions({
        queryKey: ['services'] as const,
        queryFn: () => ListServices(LIST_REQ),
    })
}

export function getServiceQuery(key: string) {
    return queryOptions({
        queryKey: ['service', key] as const,
        queryFn: () => FetchService(key),
    })
}

export function listDeploymentsQuery(serviceId: string) {
    return queryOptions({
        queryKey: ['deployments', serviceId] as const,
        queryFn: () => FetchDeployments(serviceId),
    })
}

export function getServiceMetricsQuery(serviceId: string) {
    return queryOptions({
        queryKey: ['service-metrics', serviceId] as const,
        queryFn: () => FetchServiceMetrics(serviceId),
    })
}

export function getServiceResourcesQuery(serviceId: string) {
    return queryOptions({
        queryKey: ['service-resources', serviceId] as const,
        queryFn: () => FetchServiceResources(serviceId),
    })
}

export function getServiceGraphQuery(serviceId: string) {
    return queryOptions({
        queryKey: ['service-graph', serviceId] as const,
        queryFn: () => FetchServiceGraph(serviceId),
    })
}

export function listServiceEnvsQuery(serviceId: string) {
    return queryOptions({
        queryKey: ['service-envs', serviceId] as const,
        queryFn: () => FetchServiceEnvironments(serviceId),
    })
}

export function getVervonomiconQuery(serviceId: string) {
    return queryOptions({
        queryKey: ['vervonomicon', serviceId] as const,
        queryFn: () => FetchVervonomicon(serviceId),
    })
}
