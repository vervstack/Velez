import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {CreateDindRequest} from "@/app/api/velez/dind_api.pb"
import {dindsService} from "@/processes/api/dinds.ts"
import {provisioningRefetchInterval} from "@/processes/mappings/provisioning.ts"

export const DINDS_QUERY_KEY = ["dinds"]

export function useListDindsQuery() {
    return useQuery({
        queryKey: DINDS_QUERY_KEY,
        queryFn: () => dindsService.listDinds(),
        refetchInterval: (query) => provisioningRefetchInterval(query.state.data?.provisioning),
    })
}

export function CreateDindMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateDindRequest) => dindsService.createDind(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: DINDS_QUERY_KEY})
            queryClient.invalidateQueries({queryKey: ["services"]})
        },
    })
}

export function DropDindMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (name: string) => dindsService.dropDind(name),
        onSuccess: (_data, name) => {
            queryClient.invalidateQueries({queryKey: DINDS_QUERY_KEY})
            queryClient.invalidateQueries({queryKey: ["services"]})
            queryClient.invalidateQueries({queryKey: ["service", "name", name]})
        },
    })
}
