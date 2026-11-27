import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {
    CreateHaproxyInstanceRequest,
    SelectHaproxyMasterRequest,
    ValidateHaproxyTargetRequest,
} from "@/app/api/velez/haproxy_api.pb"
import {haproxyService} from "@/processes/api/haproxy.ts"
import {provisioningRefetchInterval} from "@/processes/mappings/provisioning.ts"

export const HAPROXY_INSTANCES_QUERY_KEY = ["haproxy-instances"]

export function useValidateHaproxyTargetQuery(req: ValidateHaproxyTargetRequest | undefined) {
    return useQuery({
        queryKey: ["haproxy-target", req],
        queryFn: req ? () => haproxyService.validateHaproxyTarget(req) : undefined,
        enabled: Boolean(req),
        retry: false,
    })
}

export function useDockerHaproxyTargetQuery(master: string, environment: string) {
    return useValidateHaproxyTargetQuery(master ? {
        masterPgInstance: master,
        environment: environment || undefined,
        connectionTarget: {dockerNetwork: ""},
    } : undefined)
}

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
