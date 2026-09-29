import {useMutation, useQuery} from '@tanstack/react-query'

import {ContainerFilter, RegisterContainerRequest} from '@/app/api/velez'
import {FetchContainers, FetchContainer, velezService} from '@/processes/api/velez'

export const CONTAINERS_QUERY_KEY = ['containers'] as const
export const CONTAINER_QUERY_KEY = ['container'] as const

export function useListContainersQuery(filters?: ContainerFilter[]) {
    return useQuery({
        queryKey: [...CONTAINERS_QUERY_KEY, filters] as const,
        queryFn: () => FetchContainers(filters),
    })
}

export function useGetContainerQuery(id: string, enabled = true) {
    return useQuery({
        queryKey: [...CONTAINER_QUERY_KEY, id] as const,
        queryFn: () => FetchContainer(id),
        enabled,
    })
}

export function RegisterContainerMutation() {
    return useMutation({
        mutationFn: (req: RegisterContainerRequest) => velezService.registerContainer(req),
    })
}
