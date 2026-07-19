import {useQuery} from '@tanstack/react-query'
import {serviceService} from '@/processes/api/service'
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

const LIST_REQ = {paging: {limit: '50', offset: '0'}}

export function useListServicesQuery() {
    return useQuery({
        queryKey: ['services'] as const,
        queryFn: () => serviceService.listServices(LIST_REQ),
    })
}

export function GetServiceByNameQuery(name: string) {
    return useQuery({
        queryKey: ['service', 'name', name] as const,
        queryFn: () => serviceService.getServiceByName(name),
    })
}

export function ListDeploymentsByServiceNameQuery(serviceName: string) {
    const toaster = useToaster();

    return useQuery({
        queryKey: ['deployments', serviceName] as const,
        queryFn: () => serviceService
            .fetchDeploymentsByServiceName(serviceName)
            .catch(toaster.catchGrpc),
    })
}

export function useGetServiceAboutQuery(name: string) {
    return useQuery({
        queryKey: ['service', 'about', name] as const,
        queryFn: () => serviceService.fetchServiceAbout(name),
    })
}

export function useGetServiceMetricsQuery(serviceName: string) {
    return useQuery({
        queryKey: ['service-metrics', serviceName] as const,
        queryFn: () => serviceService.fetchServiceMetrics(serviceName),
    })
}

export function useGetServiceResourcesQuery(serviceName: string) {
    return useQuery({
        queryKey: ['service-resources', serviceName] as const,
        queryFn: () => serviceService.fetchServiceResources(serviceName),
    })
}

export function useGetServiceGraphQuery(serviceName: string) {
    const toaster = useToaster()
    return useQuery({
        queryKey: ['service-graph', serviceName] as const,
        queryFn: () =>
            serviceService.fetchServiceGraph(serviceName)
                .catch(toaster.catchGrpc),
        enabled: !!serviceName,
    })
}

export function useListServiceEnvsQuery(serviceName: string) {
    return useQuery({
        queryKey: ['service-envs', serviceName] as const,
        queryFn: () => serviceService.fetchServiceEnvironments(serviceName),
    })
}

export function useGetVervonomiconQuery(serviceName: string) {
    return useQuery({
        queryKey: ['vervonomicon', serviceName] as const,
        queryFn: () => serviceService.fetchVervonomicon(serviceName),
    })
}
