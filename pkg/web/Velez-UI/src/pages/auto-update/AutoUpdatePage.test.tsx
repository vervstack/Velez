import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {ListAutoUpdateServicesResponseService} from "@/app/api/velez/auto_update_api.pb"
import {useListAutoUpdateServicesQuery} from "@/processes/queries/autoUpdate.ts"
import AutoUpdatePage from "@/pages/auto-update/AutoUpdatePage.tsx"

vi.mock("@/processes/queries/autoUpdate.ts", () => ({useListAutoUpdateServicesQuery: vi.fn()}))
vi.mock("@/pages/auto-update/components/AutoUpdateRow/AutoUpdateRow.tsx", () => ({
    default: ({service}: { service: ListAutoUpdateServicesResponseService }) => <span>row {service.name}</span>,
}))
vi.mock("@/pages/auto-update/components/AutoUpdateTableSkeleton/AutoUpdateTableSkeleton.tsx", () => ({
    default: () => <span>auto-update skeleton</span>,
}))

type ServicesQuery = ReturnType<typeof useListAutoUpdateServicesQuery>

function renderPage(query: Partial<ServicesQuery>) {
    const refetch = vi.fn()
    vi.mocked(useListAutoUpdateServicesQuery).mockReturnValue(
        {refetch, ...query} as Partial<ServicesQuery> as ServicesQuery
    )

    render(<AutoUpdatePage/>)
    return {refetch}
}

function loaded(
    services: ListAutoUpdateServicesResponseService[],
    isWorkerEnabled = true,
): Partial<ServicesQuery> {
    return {isSuccess: true, data: {services, checkPeriodSeconds: "300", isWorkerEnabled}}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("AutoUpdatePage", () => {
    it("lists services sorted by name with the watched count and the check period", () => {
        renderPage(loaded([
            {name: "zeta", isEnabled: true},
            {name: "alpha", isEnabled: false},
        ]))

        const rows = screen.getAllByText(/^row /).map((el) => el.textContent)
        expect(rows).toEqual(["row alpha", "row zeta"])
        expect(screen.getByText("1 watched")).toBeInTheDocument()
        expect(screen.getByText("Checks every 5 minutes")).toBeInTheDocument()
    })

    it("shows the skeleton instead of the empty state while loading", () => {
        renderPage({isLoading: true})

        expect(screen.getByText("auto-update skeleton")).toBeInTheDocument()
        expect(screen.queryByText("No Velez app services on this node.")).not.toBeInTheDocument()
    })

    it("shows a retry-capable error state that refetches", () => {
        const {refetch} = renderPage({isError: true})

        expect(screen.getByText("Failed to load auto-update services.")).toBeInTheDocument()
        fireEvent.click(screen.getByText("Retry"))

        expect(refetch).toHaveBeenCalledTimes(1)
    })

    it("shows the empty state when there are no services", () => {
        renderPage(loaded([]))

        expect(screen.getByText("No Velez app services on this node.")).toBeInTheDocument()
    })

    it("shows the warning banner when the worker is turned off by config", () => {
        renderPage(loaded([{name: "api"}], false))

        expect(screen.getByRole("alert")).toHaveTextContent("Auto-update worker is turned off by config")
    })

    it("hides the warning banner when the worker is running", () => {
        renderPage(loaded([{name: "api"}], true))

        expect(screen.queryByRole("alert")).not.toBeInTheDocument()
    })
})
