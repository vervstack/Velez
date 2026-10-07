import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {UpdateSettingsRequest} from "@/app/api/velez/settings_api.pb"
import {settingsService} from "@/processes/api/settings.ts"

export const SETTINGS_QUERY_KEY = ["settings"]
const SYSBOX_STATUS_QUERY_KEY = ["sysbox_status"]
const ADDRESSES_REBUILD_STATUS_QUERY_KEY = ["addresses_rebuild_status"]
const REBUILD_POLL_INTERVAL_MS = 1000

export function useGetSettingsQuery() {
    return useQuery({
        queryKey: SETTINGS_QUERY_KEY,
        queryFn: () => settingsService.getSettings(),
    })
}

export function UpdateSettingsMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: UpdateSettingsRequest) => settingsService.updateSettings(req),
        onSettled: () => {
            queryClient.invalidateQueries({queryKey: SETTINGS_QUERY_KEY})
        },
    })
}

export function useSysboxStatusQuery() {
    return useQuery({
        queryKey: SYSBOX_STATUS_QUERY_KEY,
        queryFn: () => settingsService.getSysboxStatus(),
    })
}

export function RunSysboxSmokeTestMutation() {
    return useMutation({
        mutationFn: () => settingsService.runSysboxSmokeTest(),
    })
}

export function useAddressesRebuildStatusQuery() {
    return useQuery({
        queryKey: ADDRESSES_REBUILD_STATUS_QUERY_KEY,
        queryFn: () => settingsService.getAddressesRebuildStatus(),
        refetchInterval: (query) => query.state.data?.isRunning ? REBUILD_POLL_INTERVAL_MS : false,
    })
}

export function RebuildAddressesMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: () => settingsService.rebuildAddresses(),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ADDRESSES_REBUILD_STATUS_QUERY_KEY})
        },
    })
}
