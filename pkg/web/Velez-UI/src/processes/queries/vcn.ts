import {useIsMutating, useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {ConnectServiceRequest, CreateVcnNetworkRequest} from "@/app/api/velez/verv_closed_network.pb"
import {vcnService} from "@/processes/api/vcn.ts"

const VCN_NETWORKS_QUERY_KEY = ["vcn", "networks"]
const VCN_NETWORK_DETAILS_QUERY_KEY = ["vcn", "network"]

function vcnNetworkQueryKey(networkId: string) {
    return [...VCN_NETWORK_DETAILS_QUERY_KEY, networkId]
}

export function useListVcnNetworksQuery() {
    return useQuery({
        queryKey: VCN_NETWORKS_QUERY_KEY,
        queryFn: () => vcnService.listNetworks().then((response) => response.networks ?? []),
    })
}

export function useVcnNetworkQuery(networkId: string) {
    return useQuery({
        queryKey: vcnNetworkQueryKey(networkId),
        queryFn: () => vcnService.getNetwork(networkId),
        enabled: Boolean(networkId),
    })
}

export function useCreateVcnNetworkMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateVcnNetworkRequest) => vcnService.createNetwork(req),
        gcTime: 0,
        onSuccess: () => {
            void queryClient.invalidateQueries({queryKey: VCN_NETWORKS_QUERY_KEY})
            void queryClient.invalidateQueries({queryKey: ["services"]})
        },
    })
}

export function useDropVcnNetworkMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (id: string) => vcnService.dropNetwork(id),
        onSuccess: () => {
            void queryClient.invalidateQueries({queryKey: ["services"]})
            return queryClient.invalidateQueries({queryKey: VCN_NETWORKS_QUERY_KEY})
        },
    })
}

export function useIssueNetworkApiKeyMutation(networkId: string) {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: () => vcnService.issueApiKey(networkId),
        gcTime: 0,
        onSuccess: () => {
            void queryClient.invalidateQueries({queryKey: vcnNetworkQueryKey(networkId)})
        },
    })
}

export function useExpireNetworkApiKeyMutation(networkId: string) {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (prefix: string) => vcnService.expireApiKey(networkId, prefix),
        onSuccess: () => queryClient.invalidateQueries({queryKey: vcnNetworkQueryKey(networkId)}),
    })
}

function connectServiceFilter(serviceName: string) {
    return {mutationKey: ["vcn", "connect-service", serviceName], exact: true}
}

export function useConnectServiceMutation(serviceName: string) {
    const queryClient = useQueryClient()
    const filters = connectServiceFilter(serviceName)
    const isPending = useIsMutating(filters) > 0
    const mutation = useMutation({
        mutationKey: filters.mutationKey,
        mutationFn: (req: ConnectServiceRequest) => vcnService.connectService(req),
        gcTime: 1_000,
        onSuccess: (_, req) => {
            if (req.networkId) {
                void queryClient.invalidateQueries({queryKey: vcnNetworkQueryKey(req.networkId)})
            }
        },
    })

    function connectService(req: ConnectServiceRequest) {
        if (queryClient.isMutating(filters)) return

        return mutation.mutateAsync(req).finally(mutation.reset)
    }

    return {connectService, isPending}
}

export function useIsConnectServicePending(serviceName: string) {
    return useIsMutating(connectServiceFilter(serviceName)) > 0
}
