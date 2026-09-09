import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import {pgaasService} from "@/processes/api/pgaas"
import {CreatePgInstanceRequest} from "@/app/api/velez"

const PG_INSTANCES_QUERY_KEY = ["pg-instances"]
const LIST_REQ = {paging: {limit: "50", offset: "0"}}

export function useListPgInstancesQuery() {
    return useQuery({
        queryKey: PG_INSTANCES_QUERY_KEY,
        queryFn: () => pgaasService.listPgInstances(LIST_REQ),
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
        },
    })
}

// GetPgInstanceCredentialsQuery is disabled by default — it must only run when the caller
// explicitly triggers refetch() (the "Reveal" / "Copy DSN" actions), never on mount alongside
// the instance list.
export function GetPgInstanceCredentialsQuery(name: string) {
    return useQuery({
        queryKey: ["pg-instance-credentials", name],
        queryFn: () => pgaasService.getPgInstanceCredentials(name),
        enabled: false,
    })
}
