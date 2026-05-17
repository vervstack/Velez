import { useQuery } from '@tanstack/react-query'
import { FetchSmerds, FetchSmerd, FetchSmerdsByServiceName } from '@/processes/api/velez'

export function useListSmerdsQuery() {
    return useQuery({
        queryKey: ['smerds'] as const,
        queryFn: FetchSmerds,
    })
}

export function useGetSmerdQuery(name: string, enabled = true) {
    return useQuery({
        queryKey: ['smerd', name] as const,
        queryFn: () => FetchSmerd(name),
        enabled,
    })
}

export function useListSmerdsByServiceQuery(serviceName: string, enabled = true) {
    return useQuery({
        queryKey: ['service-smerds', serviceName] as const,
        queryFn: () => FetchSmerdsByServiceName(serviceName),
        enabled,
    })
}
