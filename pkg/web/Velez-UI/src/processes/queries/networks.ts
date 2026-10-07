import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {
    ConnectContainerRequest,
    CreateNetworkRequest,
    DeleteNetworkRequest,
    DisconnectContainerRequest,
} from "@/app/api/velez/network_api.pb"
import {networksService} from "@/processes/api/networks.ts"
import {CONTAINERS_QUERY_KEY} from "@/processes/queries/containers.ts"

export const NETWORKS_QUERY_KEY = ["networks"]
export const NETWORK_STATUS_QUERY_KEY = ["network-status"]

export function useNetworkStatusQuery() {
    return useQuery({
        queryKey: NETWORK_STATUS_QUERY_KEY,
        queryFn: () => networksService.getStatus(),
    })
}

export function useListNetworksQuery(environment: string, isForeignIncluded: boolean) {
    return useQuery({
        queryKey: [...NETWORKS_QUERY_KEY, environment, isForeignIncluded],
        queryFn: () => networksService.listNetworks(environment, isForeignIncluded),
    })
}

export function CreateNetworkMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateNetworkRequest) => networksService.createNetwork(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: NETWORKS_QUERY_KEY})
        },
    })
}

export function DeleteNetworkMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: DeleteNetworkRequest) => networksService.deleteNetwork(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: NETWORKS_QUERY_KEY})
        },
    })
}

export function ConnectContainerMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: ConnectContainerRequest) => networksService.connectContainer(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: NETWORKS_QUERY_KEY})
            queryClient.invalidateQueries({queryKey: ["services"]})
            queryClient.invalidateQueries({queryKey: CONTAINERS_QUERY_KEY})
        },
    })
}

export function DisconnectContainerMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: DisconnectContainerRequest) => networksService.disconnectContainer(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: NETWORKS_QUERY_KEY})
            queryClient.invalidateQueries({queryKey: ["services"]})
            queryClient.invalidateQueries({queryKey: CONTAINERS_QUERY_KEY})
        },
    })
}
