import {useIsMutating, useMutation, useQueryClient} from "@tanstack/react-query"

import {ConnectServiceRequest} from "@/app/api/velez"
import {vcnService} from "@/processes/api/vcn.ts"

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
