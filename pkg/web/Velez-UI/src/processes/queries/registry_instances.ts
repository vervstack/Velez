import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import {registryaasService} from "@/processes/api/registryaas"
import {CreateRegistryInstanceRequest} from "@/app/api/velez"

export const REGISTRY_INSTANCES_QUERY_KEY = ["registry-instances"]
const LIST_REQ = {paging: {limit: "50", offset: "0"}}

export function useListRegistryInstancesQuery() {
    return useQuery({
        queryKey: REGISTRY_INSTANCES_QUERY_KEY,
        queryFn: () => registryaasService.listRegistryInstances(LIST_REQ),
    })
}

export function CreateRegistryInstanceMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateRegistryInstanceRequest) => registryaasService.createRegistryInstance(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: REGISTRY_INSTANCES_QUERY_KEY})
        },
    })
}

export function DropRegistryInstanceMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (name: string) => registryaasService.dropRegistryInstance(name),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: REGISTRY_INSTANCES_QUERY_KEY})
        },
    })
}

// GetRegistryInstanceCredentialsQuery is disabled by default — it must only run when the caller
// explicitly triggers refetch() (the "Reveal" / "Copy docker login" actions), never on mount
// alongside the instance list.
export function GetRegistryInstanceCredentialsQuery(name: string) {
    return useQuery({
        queryKey: ["registry-instance-credentials", name],
        queryFn: () => registryaasService.getRegistryInstanceCredentials(name),
        enabled: false,
    })
}
