import {useMutation, useQuery} from "@tanstack/react-query"

import type {SetServiceAutoUpdateRequest} from "@/app/api/velez/auto_update_api.pb"
import {autoUpdateService} from "@/processes/api/autoUpdate.ts"

export const AUTO_UPDATE_QUERY_KEY = ["auto-update"]

export function useListAutoUpdateServicesQuery() {
    return useQuery({
        queryKey: AUTO_UPDATE_QUERY_KEY,
        queryFn: () => autoUpdateService.listServices(),
    })
}

export function SetServiceAutoUpdateMutation() {
    return useMutation({
        mutationFn: (req: SetServiceAutoUpdateRequest) => autoUpdateService.setService(req),
    })
}
