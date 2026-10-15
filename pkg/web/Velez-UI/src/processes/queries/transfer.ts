import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {StartTransferRequest} from "@/app/api/velez/transfer_api.pb"
import {hasActiveTransfer} from "@/processes/mappings/transfer.ts"
import {transferService} from "@/processes/api/transfer.ts"

export const TRANSFERS_QUERY_KEY = ["transfers"]

const ACTIVE_POLL_INTERVAL_MS = 2000

export function useListTransfersQuery() {
    return useQuery({
        queryKey: TRANSFERS_QUERY_KEY,
        queryFn: () => transferService.listTransfers(),
        refetchInterval: (query) => hasActiveTransfer(query.state.data?.transfers) ? ACTIVE_POLL_INTERVAL_MS : false,
    })
}

export function CreateTransferTokenMutation() {
    return useMutation({
        mutationFn: () => transferService.createTransferToken(),
    })
}

export function PlanTransferMutation() {
    return useMutation({
        mutationFn: (containerName: string) => transferService.planTransfer(containerName),
    })
}

export function StartTransferMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: StartTransferRequest) => transferService.startTransfer(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: TRANSFERS_QUERY_KEY})
        },
    })
}

export function CancelTransferMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (transferId: string) => transferService.cancelTransfer(transferId),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: TRANSFERS_QUERY_KEY})
        },
    })
}
