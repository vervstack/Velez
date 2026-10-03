import {useMutation, useQuery} from '@tanstack/react-query'

import {ContainerFilter, FinishOnboardingRequest, RegisterContainerRequest} from '@/app/api/velez'
import {FetchContainers, FetchContainer, FetchImageVersions, velezService} from '@/processes/api/velez'

export const CONTAINERS_QUERY_KEY = ['containers'] as const
export const CONTAINER_QUERY_KEY = ['container'] as const
export const IMAGE_VERSIONS_QUERY_KEY = ['image-versions'] as const

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

export function useImageVersionsQuery(containerId: string, enabled = true) {
    return useQuery({
        queryKey: [...IMAGE_VERSIONS_QUERY_KEY, containerId] as const,
        queryFn: () => FetchImageVersions(containerId),
        enabled,
    })
}

export function FinishOnboardingMutation() {
    return useMutation({
        mutationFn: (req: FinishOnboardingRequest) => velezService.finishOnboarding(req),
    })
}
