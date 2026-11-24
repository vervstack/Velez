import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import {pgaasService} from "@/processes/api/pgaas"
import {provisioningRefetchInterval} from "@/processes/mappings/provisioning.ts"
import {CreatePgInstanceRequest, EnablePgReplicationRequest, SetupPgReplicationRequest} from "@/app/api/velez"

export const PG_INSTANCES_QUERY_KEY = ["pg-instances"]
const PG_REPLICATION_STATUS_QUERY_KEY = ["pg-replication-status"]
const PG_INSTANCE_CREDENTIALS_QUERY_KEY = ["pg-instance-credentials"]
const PG_REPLICATION_STATUS_REFETCH_MS = 5000
const LIST_REQ = {paging: {limit: "50", offset: "0"}}

export function useListPgInstancesQuery(isUnpaged = false) {
    return useQuery({
        queryKey: isUnpaged ? [...PG_INSTANCES_QUERY_KEY, "all"] : PG_INSTANCES_QUERY_KEY,
        queryFn: () => pgaasService.listPgInstances(isUnpaged ? {} : LIST_REQ),
        refetchInterval: (query) => provisioningRefetchInterval(query.state.data?.provisioning),
    })
}

export function CreatePgInstanceMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreatePgInstanceRequest) => pgaasService.createPgInstance(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: PG_INSTANCES_QUERY_KEY})
        },
    })
}

export function DropPgInstanceMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (name: string) => pgaasService.dropPgInstance(name),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: PG_INSTANCES_QUERY_KEY})
            queryClient.invalidateQueries({queryKey: ["services"]})
        },
    })
}

export function EnablePgReplicationMutation() {
    return useMutation({
        mutationFn: (req: EnablePgReplicationRequest) => pgaasService.enablePgReplication(req),
    })
}

export function SetupPgReplicationMutation() {
    return useMutation({
        mutationFn: (req: SetupPgReplicationRequest) => pgaasService.setupPgReplication(req),
    })
}

export function usePgReplicationStatusQuery(instance: string) {
    return useQuery({
        queryKey: [...PG_REPLICATION_STATUS_QUERY_KEY, instance],
        queryFn: () => pgaasService.getPgReplicationStatus(instance),
        refetchInterval: PG_REPLICATION_STATUS_REFETCH_MS,
    })
}

export function useRefreshPgReplication() {
    const queryClient = useQueryClient()

    return function refreshPgReplication(source: string, target: string) {
        queryClient.invalidateQueries({queryKey: PG_INSTANCES_QUERY_KEY})
        queryClient.invalidateQueries({queryKey: [...PG_INSTANCE_CREDENTIALS_QUERY_KEY, target]})
        queryClient.invalidateQueries({queryKey: [...PG_REPLICATION_STATUS_QUERY_KEY, source]})
        queryClient.invalidateQueries({queryKey: [...PG_REPLICATION_STATUS_QUERY_KEY, target]})
    }
}

// The credentials response carries the real password, so callers must keep it unrendered until
// the user asks to reveal it. The query itself runs on mount: the masked DSNs are always shown.
export function GetPgInstanceCredentialsQuery(name: string) {
    return useQuery({
        queryKey: [...PG_INSTANCE_CREDENTIALS_QUERY_KEY, name],
        queryFn: () => pgaasService.getPgInstanceCredentials(name),
    })
}
