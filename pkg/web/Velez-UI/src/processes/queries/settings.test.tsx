import {afterEach, describe, expect, it, vi} from "vitest"
import {renderHook, waitFor} from "@testing-library/react"
import {QueryClient, QueryClientProvider} from "@tanstack/react-query"
import type {ReactNode} from "react"

import type {GetSettingsResponse, UpdateSettingsRequest} from "@/app/api/velez/settings_api.pb"
import {settingsService} from "@/processes/api/settings.ts"
import {UpdateSettingsMutation, useGetSettingsQuery} from "@/processes/queries/settings.ts"

vi.mock("@/processes/api/settings.ts", () => ({
    settingsService: {getSettings: vi.fn(), updateSettings: vi.fn()},
}))

function renderSettingsHooks() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})

    function Wrapper({children}: { children: ReactNode }) {
        return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    }

    return renderHook(
        () => ({query: useGetSettingsQuery(), mutation: UpdateSettingsMutation()}),
        {wrapper: Wrapper},
    )
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("UpdateSettingsMutation", () => {
    it("refetches the settings after a failed update so the toggle reverts", async () => {
        const stored: GetSettingsResponse = {settings: {isSysboxEnabled: false}}
        vi.mocked(settingsService.getSettings).mockImplementation(() => Promise.resolve(stored))
        vi.mocked(settingsService.updateSettings).mockRejectedValue(new Error("update failed"))
        const {result} = renderSettingsHooks()
        await waitFor(() => expect(result.current.query.data).toEqual(stored))

        const req: UpdateSettingsRequest = {isSysboxEnabled: true}
        result.current.mutation.mutate(req)

        await waitFor(() => expect(settingsService.getSettings).toHaveBeenCalledTimes(2))
        expect(result.current.query.data?.settings?.isSysboxEnabled).toBe(false)
    })

    it("refetches the settings after a successful update", async () => {
        let stored: GetSettingsResponse = {settings: {isSysboxEnabled: false}}
        vi.mocked(settingsService.getSettings).mockImplementation(() => Promise.resolve(stored))
        vi.mocked(settingsService.updateSettings).mockImplementation(function writeSettings(req) {
            stored = {settings: {...stored.settings, ...req}}
            return Promise.resolve({})
        })
        const {result} = renderSettingsHooks()
        await waitFor(() => expect(result.current.query.data).toEqual(stored))

        result.current.mutation.mutate({isSysboxEnabled: true})

        await waitFor(() => expect(result.current.query.data?.settings?.isSysboxEnabled).toBe(true))
    })
})
