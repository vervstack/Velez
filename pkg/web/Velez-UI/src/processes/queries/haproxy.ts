import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {CreateHaproxyInstanceRequest, SelectHaproxyMasterRequest} from "@/app/api/velez/haproxy_api.pb"
import {haproxyService} from "@/processes/api/haproxy.ts"
import {provisioningRefetchInterval} from "@/processes/mappings/provisioning.ts"

export const HAPROXY_INSTANCES_QUERY_KEY = ["haproxy-instances"]

export function useListHaproxyInstancesQuery() {
    return useQuery({
        queryKey: HAPROXY_INSTANCES_QUERY_KEY,
        queryFn: () => haproxyService.listHaproxyInstances(),
        refetchInterval: (query) => provisioningRefetchInterval(query.state.data?.provisioning),
    })
}

export function useCreateHaproxyInstanceMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateHaproxyInstanceRequest) => haproxyService.createHaproxyInstance(req),
        onSuccess: () => queryClient.invalidateQueries({queryKey: HAPROXY_INSTANCES_QUERY_KEY}),
    })
}

export function useDropHaproxyInstanceMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (name: string) => haproxyService.dropHaproxyInstance(name),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: HAPROXY_INSTANCES_QUERY_KEY})
            queryClient.invalidateQueries({queryKey: ["services"]})
        },
    })
}

export function useSelectHaproxyMasterMutation() {
    return useMutation({
        mutationFn: (req: SelectHaproxyMasterRequest) => haproxyService.selectHaproxyMaster(req),
    })
}
