import {useQuery} from '@tanstack/react-query'

import {ContainerFilter} from '@/app/api/velez'
import {FetchContainers, FetchContainer} from '@/processes/api/velez'

export function useListContainersQuery(filters?: ContainerFilter[]) {
    return useQuery({
        queryKey: ['containers', filters] as const,
        queryFn: () => FetchContainers(filters),
    })
}

export function useGetContainerQuery(id: string, enabled = true) {
    return useQuery({
        queryKey: ['container', id] as const,
        queryFn: () => FetchContainer(id),
        enabled,
    })
}
