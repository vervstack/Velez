import { queryOptions } from '@tanstack/react-query'
import { FetchSmerds, FetchSmerd, FetchSmerdsByServiceName } from '@/processes/api/velez'

export function listSmerdsQuery() {
    return queryOptions({
        queryKey: ['smerds'] as const,
        queryFn: FetchSmerds,
    })
}

export function getSmerdQuery(name: string) {
    return queryOptions({
        queryKey: ['smerd', name] as const,
        queryFn: () => FetchSmerd(name),
    })
}

export function listSmerdsByServiceQuery(serviceName: string) {
    return queryOptions({
        queryKey: ['service-smerds', serviceName] as const,
        queryFn: () => FetchSmerdsByServiceName(serviceName),
    })
}
