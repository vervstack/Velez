import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import type {AddressesRebuildStatus} from "@/model/settings/AddressesRebuildStatus.ts"
import {RebuildAddressesMutation, useAddressesRebuildStatusQuery} from "@/processes/queries/settings.ts"
import AddressesSettings from "@/widgets/settings/AddressesSettings/AddressesSettings.tsx"

vi.mock("@/processes/queries/settings.ts", () => ({
    RebuildAddressesMutation: vi.fn(),
    useAddressesRebuildStatusQuery: vi.fn(),
}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))

type Mutation = ReturnType<typeof RebuildAddressesMutation>
type StatusQuery = ReturnType<typeof useAddressesRebuildStatusQuery>
type Toaster = ReturnType<typeof useToaster>

const IDLE: AddressesRebuildStatus = {isRunning: false, totalSteps: 0, doneSteps: 0, lastError: ""}
const RUNNING: AddressesRebuildStatus = {isRunning: true, totalSteps: 10, doneSteps: 3, lastError: ""}

interface RenderOptions {
    status?: AddressesRebuildStatus
    query?: Partial<StatusQuery>
    mutation?: Partial<Mutation>
}

const bake = vi.fn()
const catchGrpc = vi.fn()
const mutate = vi.fn()
const refetch = vi.fn()

function mockHooks({status = IDLE, query, mutation}: RenderOptions) {
    vi.mocked(useAddressesRebuildStatusQuery).mockReturnValue(
        {data: status, refetch, ...query} as Partial<StatusQuery> as StatusQuery,
    )
    vi.mocked(RebuildAddressesMutation).mockReturnValue({mutate, ...mutation} as Partial<Mutation> as Mutation)
    vi.mocked(useToaster).mockReturnValue({bake, catchGrpc} as Partial<Toaster> as Toaster)
}

function renderSettings(options: RenderOptions = {}) {
    mockHooks(options)
    const view = render(<AddressesSettings/>)

    function rerenderWith(next: RenderOptions) {
        mockHooks(next)
        view.rerender(<AddressesSettings/>)
    }

    return {rerenderWith}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("AddressesSettings", () => {
    it("shows the title and the description", () => {
        renderSettings()

        expect(screen.getByText("Addresses")).toBeInTheDocument()
        expect(screen.getByText(/Re-crawl Docker and the VCN/)).toBeInTheDocument()
    })

    it("shows only the enabled button when no rebuild is running", () => {
        renderSettings()

        expect(screen.getByText("Rebuild addresses").closest("button")).toBeEnabled()
        expect(screen.queryByRole("progressbar")).not.toBeInTheDocument()
    })

    it("starts the rebuild when the button is clicked", () => {
        renderSettings()

        fireEvent.click(screen.getByText("Rebuild addresses"))

        expect(mutate).toHaveBeenCalledTimes(1)
    })

    it("routes a failed rebuild start through the toaster error surface", () => {
        const failure = new Error("already exists")
        mutate.mockImplementation((_vars, options) => options.onError(failure))
        renderSettings()

        fireEvent.click(screen.getByText("Rebuild addresses"))

        expect(catchGrpc).toHaveBeenCalledWith(failure)
    })

    it("disables the button while the start request is pending", () => {
        renderSettings({mutation: {isPending: true}})

        expect(screen.getByText("Rebuild addresses").closest("button")).toBeDisabled()
    })

    it("disables the button and shows the progress when a rebuild is running", () => {
        renderSettings({status: RUNNING})

        expect(screen.getByText("Rebuild addresses").closest("button")).toBeDisabled()
        expect(screen.getByText("Rebuilding addresses… 3/10")).toBeInTheDocument()
        expect(screen.getByRole("progressbar")).toHaveAttribute("value", "3")
    })

    it("shows the last error when the previous rebuild failed", () => {
        renderSettings({status: {...IDLE, lastError: "docker unreachable"}})

        expect(screen.getByText("Last rebuild failed: docker unreachable")).toBeInTheDocument()
    })

    it("shows a success toast once when a running rebuild finishes without an error", () => {
        const {rerenderWith} = renderSettings({status: RUNNING})

        rerenderWith({status: IDLE})
        rerenderWith({status: IDLE})

        expect(bake).toHaveBeenCalledTimes(1)
        expect(bake).toHaveBeenCalledWith(expect.objectContaining({title: "Addresses rebuilt", level: "Info"}))
    })

    it("shows no success toast when a running rebuild finishes with an error", () => {
        const {rerenderWith} = renderSettings({status: RUNNING})

        rerenderWith({status: {...IDLE, lastError: "boom"}})

        expect(bake).not.toHaveBeenCalled()
        expect(screen.getByText("Last rebuild failed: boom")).toBeInTheDocument()
    })

    it("shows no success toast when the page loads with an idle status", () => {
        renderSettings()

        expect(bake).not.toHaveBeenCalled()
    })

    it("shows a retry-capable error state when the status query fails", () => {
        renderSettings({query: {isError: true, data: undefined}})

        fireEvent.click(screen.getByText("Retry"))

        expect(refetch).toHaveBeenCalledTimes(1)
    })
})
