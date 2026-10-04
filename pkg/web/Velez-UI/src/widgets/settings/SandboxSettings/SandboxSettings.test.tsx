import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {UpdateSettingsMutation, useGetSettingsQuery} from "@/processes/queries/settings.ts"
import SandboxSettings from "@/widgets/settings/SandboxSettings/SandboxSettings.tsx"

vi.mock("@/processes/queries/settings.ts", () => ({
    useGetSettingsQuery: vi.fn(),
    UpdateSettingsMutation: vi.fn(),
}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/SysboxSetupChecklist.tsx", () => ({
    default: () => null,
}))

type SettingsQuery = ReturnType<typeof useGetSettingsQuery>
type SettingsMutation = ReturnType<typeof UpdateSettingsMutation>
type Toaster = ReturnType<typeof useToaster>

interface RenderOptions {
    query?: Partial<SettingsQuery>
    mutation?: Partial<SettingsMutation>
}

const SYSBOX_LABEL = "Run containers in Sysbox"
const WHITELIST_LABEL = "Apply Sysbox to whitelisted containers too"

function renderSettings({query, mutation}: RenderOptions = {}) {
    const refetch = vi.fn()
    const mutate = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(useGetSettingsQuery).mockReturnValue({refetch, ...query} as Partial<SettingsQuery> as SettingsQuery)
    vi.mocked(UpdateSettingsMutation).mockReturnValue(
        {mutate, ...mutation} as Partial<SettingsMutation> as SettingsMutation,
    )
    vi.mocked(useToaster).mockReturnValue({catchGrpc} as Partial<Toaster> as Toaster)

    render(<SandboxSettings/>)
    return {refetch, mutate, catchGrpc}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("SandboxSettings", () => {
    it("renders both toggles with the values from the settings query", () => {
        renderSettings({query: {data: {settings: {isSysboxEnabled: true, isSysboxWhitelistIgnored: false}}}})

        expect(screen.getByRole("switch", {name: SYSBOX_LABEL})).toBeChecked()
        expect(screen.getByRole("switch", {name: WHITELIST_LABEL})).not.toBeChecked()
    })

    it("sends only isSysboxEnabled when the Sysbox toggle is flipped", () => {
        const {mutate} = renderSettings({query: {data: {settings: {isSysboxEnabled: false}}}})

        fireEvent.click(screen.getByRole("switch", {name: SYSBOX_LABEL}))

        expect(mutate).toHaveBeenCalledTimes(1)
        expect(mutate.mock.calls[0][0]).toEqual({isSysboxEnabled: true})
    })

    it("sends only isSysboxWhitelistIgnored when the whitelist toggle is flipped", () => {
        const {mutate} = renderSettings({query: {data: {settings: {isSysboxEnabled: true}}}})

        fireEvent.click(screen.getByRole("switch", {name: WHITELIST_LABEL}))

        expect(mutate.mock.calls[0][0]).toEqual({isSysboxWhitelistIgnored: true})
    })

    it("routes a failed update through the toaster error surface", () => {
        const failure = new Error("update failed")
        const {mutate, catchGrpc} = renderSettings({query: {data: {settings: {}}}})
        mutate.mockImplementation((_req, options) => options.onError(failure))

        fireEvent.click(screen.getByRole("switch", {name: SYSBOX_LABEL}))

        expect(catchGrpc).toHaveBeenCalledWith(failure)
    })

    it("disables both toggles while an update is pending", () => {
        renderSettings({query: {data: {settings: {}}}, mutation: {isPending: true}})

        expect(screen.getByRole("switch", {name: SYSBOX_LABEL})).toBeDisabled()
        expect(screen.getByRole("switch", {name: WHITELIST_LABEL})).toBeDisabled()
    })

    it("shows no toggles and no error while settings are loading", () => {
        renderSettings({query: {isLoading: true}})

        expect(screen.queryByRole("switch")).not.toBeInTheDocument()
        expect(screen.queryByText("Retry")).not.toBeInTheDocument()
    })

    it("shows a retry-capable error state that refetches when settings fail to load", () => {
        const {refetch} = renderSettings({query: {isError: true}})

        expect(screen.getByText("Failed to load settings.")).toBeInTheDocument()
        expect(screen.queryByRole("switch")).not.toBeInTheDocument()

        fireEvent.click(screen.getByText("Retry"))

        expect(refetch).toHaveBeenCalledTimes(1)
    })
})
