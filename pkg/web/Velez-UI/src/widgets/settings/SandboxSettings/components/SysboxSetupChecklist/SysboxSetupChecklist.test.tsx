import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {GetSysboxStatusResponse} from "@/app/api/velez/settings_api.pb"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {
    RunSysboxSmokeTestMutation,
    useGetSettingsQuery,
    useSysboxStatusQuery,
} from "@/processes/queries/settings.ts"
import SysboxSetupChecklist from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/SysboxSetupChecklist.tsx"

vi.mock("@/processes/queries/settings.ts", () => ({
    useSysboxStatusQuery: vi.fn(),
    useGetSettingsQuery: vi.fn(),
    RunSysboxSmokeTestMutation: vi.fn(),
}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))

const healthyStatus: GetSysboxStatusResponse = {
    osType: "linux",
    kernelVersion: "5.15.0",
    isRuntimeRegistered: true,
    containersTotal: 3,
    containersOnSysbox: 2,
}

type StatusQuery = ReturnType<typeof useSysboxStatusQuery>
type SettingsQuery = ReturnType<typeof useGetSettingsQuery>
type SmokeTest = ReturnType<typeof RunSysboxSmokeTestMutation>
type Toaster = ReturnType<typeof useToaster>

interface RenderOptions {
    status?: Partial<StatusQuery>
    smokeTest?: Partial<SmokeTest>
}

function renderChecklist({status, smokeTest}: RenderOptions) {
    const refetch = vi.fn()
    const mutate = vi.fn()
    vi.mocked(useSysboxStatusQuery).mockReturnValue({refetch, ...status} as Partial<StatusQuery> as StatusQuery)
    vi.mocked(useGetSettingsQuery).mockReturnValue(
        {data: {settings: {isSysboxEnabled: true}}} as Partial<SettingsQuery> as SettingsQuery,
    )
    vi.mocked(RunSysboxSmokeTestMutation).mockReturnValue({mutate, ...smokeTest} as Partial<SmokeTest> as SmokeTest)
    vi.mocked(useToaster).mockReturnValue({catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)

    render(<SysboxSetupChecklist/>)
    return {refetch, mutate}
}

describe("SysboxSetupChecklist", () => {
    it("shows the passed-check count in the heading once the status loaded", () => {
        renderChecklist({status: {data: healthyStatus}})

        expect(screen.getByText("6 of 8 checks passed")).toBeInTheDocument()
    })

    it("marks a failing step as failed and shows its fix copy", () => {
        renderChecklist({status: {data: {...healthyStatus, isRuntimeRegistered: false}}})
        fireEvent.click(screen.getByText("Sysbox setup checklist (host requirements)"))

        expect(screen.getByText("docker info | grep -i runtimes")).toBeVisible()
        expect(screen.getAllByRole("img", {name: "Failed"}).length).toBeGreaterThan(0)
    })

    it("shows how many containers run under Sysbox when the setting is on", () => {
        renderChecklist({status: {data: healthyStatus}})

        expect(screen.getByText("2 of 3 containers run under Sysbox — recreate the rest to move them"))
            .toBeInTheDocument()
    })

    it("shows the smoke test failure text after a failed run", () => {
        renderChecklist({
            status: {data: healthyStatus},
            smokeTest: {data: {isPassed: false, failure: "runtime exited 125"}},
        })

        expect(screen.getByText("runtime exited 125")).toBeInTheDocument()
    })

    it("shows a retry-capable error state when the status failed to load", () => {
        const {refetch} = renderChecklist({status: {isError: true}})

        fireEvent.click(screen.getByText("Retry"))

        expect(screen.getByText("Failed to load Sysbox status.")).toBeInTheDocument()
        expect(refetch).toHaveBeenCalled()
    })

    it("runs the smoke test when its button is clicked", () => {
        const {mutate} = renderChecklist({status: {data: healthyStatus}})

        fireEvent.click(screen.getByText("Run smoke test"))

        expect(mutate).toHaveBeenCalled()
    })

    it("disables the smoke test button and shows Running while it is pending", () => {
        renderChecklist({status: {data: healthyStatus}, smokeTest: {isPending: true}})

        expect(screen.getByRole("button", {name: "Running…"})).toBeDisabled()
    })

    it("re-fetches the status when Re-check is clicked", () => {
        const {refetch} = renderChecklist({status: {data: healthyStatus}})

        fireEvent.click(screen.getByText("Re-check"))

        expect(refetch).toHaveBeenCalled()
    })
})
