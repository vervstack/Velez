import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {UpdateSettingsRequest} from "@/app/api/velez/settings_api.pb"
import {settingsService} from "@/processes/api/settings.ts"

export const SETTINGS_QUERY_KEY = ["settings"]
const SYSBOX_STATUS_QUERY_KEY = ["sysbox_status"]

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
