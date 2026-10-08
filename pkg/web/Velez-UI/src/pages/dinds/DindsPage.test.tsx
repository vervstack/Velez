import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DindInfo} from "@/app/api/velez/dind_api.pb"
import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DropDindMutation, useListDindsQuery} from "@/processes/queries/dinds.ts"
import DindsPage from "@/pages/dinds/DindsPage.tsx"
import CreateDindDialog from "@/dialogs/CreateDindDialog/CreateDindDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/dinds.ts", () => ({
    useListDindsQuery: vi.fn(),
    DropDindMutation: vi.fn(),
    DINDS_QUERY_KEY: ["dinds"],
}))
vi.mock("@/widgets/ProvisioningRow/ProvisioningRow.tsx", () => ({
    default: ({task}: { task: ProvisioningTask }) => <span>provisioning {task.entityId}</span>,
}))
vi.mock("@/dialogs/CreateDindDialog/CreateDindDialog.tsx", () => ({default: () => null}))
vi.mock("@/pages/dinds/components/DindsTableSkeleton/DindsTableSkeleton.tsx", () => ({
    default: () => <span>dinds skeleton</span>,
}))

type DindsQuery = ReturnType<typeof useListDindsQuery>
type DropMutation = ReturnType<typeof DropDindMutation>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function renderPage(query: Partial<DindsQuery>) {
    const refetch = vi.fn()
    const OpenDialog = vi.fn()
    vi.mocked(useListDindsQuery).mockReturnValue({refetch, ...query} as Partial<DindsQuery> as DindsQuery)
    vi.mocked(DropDindMutation).mockReturnValue({mutateAsync: vi.fn()} as Partial<DropMutation> as DropMutation)
    vi.mocked(useDialog).mockReturnValue({OpenDialog, CloseDialog: vi.fn()} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)

    render(<DindsPage/>)
    return {refetch, OpenDialog}
}

function loaded(dinds: DindInfo[], provisioning: ProvisioningTask[] = []): Partial<DindsQuery> {
    return {data: {dinds, provisioning}}
}

function runningCreateTask(entityId: string): ProvisioningTask {
    return {taskId: "1", entityId, action: "create_dind", status: ProvisioningTaskStatus.RUNNING}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("DindsPage", () => {
    it("lists daemons sorted by name with address and Sysbox yes/no", () => {
        renderPage(loaded([
            {name: "zeta", address: "tcp://zeta:2375", isSysboxEnabled: false},
            {name: "alpha", address: "tcp://alpha:2375", isSysboxEnabled: true},
        ]))

        const names = screen.getAllByText(/^(alpha|zeta)$/).map((el) => el.textContent)
        expect(names).toEqual(["alpha", "zeta"])
        expect(screen.getByText("tcp://alpha:2375")).toBeInTheDocument()
        expect(screen.getByText("Yes")).toBeInTheDocument()
        expect(screen.getByText("No")).toBeInTheDocument()
        expect(screen.getByText("2 daemons")).toBeInTheDocument()
    })

    it("shows the count skeleton instead of a zero count while loading", () => {
        renderPage({isLoading: true})

        expect(screen.queryByText("0 daemons")).not.toBeInTheDocument()
        expect(screen.getByText("daemons")).toHaveAttribute("aria-busy", "true")
    })

    it("shows the empty state when there are no daemons", () => {
        renderPage(loaded([]))

        expect(screen.getByText("No Docker daemons on this node.")).toBeInTheDocument()
    })

    it("renders a provisioning row and no empty state when only a create task exists", () => {
        renderPage(loaded([], [runningCreateTask("ci")]))

        expect(screen.getByText("provisioning ci")).toBeInTheDocument()
        expect(screen.queryByText("No Docker daemons on this node.")).not.toBeInTheDocument()
    })

    it("hides the real row of a daemon that is still provisioning but keeps counting it", () => {
        renderPage(loaded([{name: "ci", address: "tcp://ci:2375"}], [runningCreateTask("ci")]))

        expect(screen.queryByText("tcp://ci:2375")).not.toBeInTheDocument()
        expect(screen.getByText("1 daemons")).toBeInTheDocument()
    })

    it("shows the skeleton while loading", () => {
        renderPage({isLoading: true})

        expect(screen.getByText("dinds skeleton")).toBeInTheDocument()
        expect(screen.queryByText("No Docker daemons on this node.")).not.toBeInTheDocument()
    })

    it("shows a retry-capable error state that refetches", () => {
        const {refetch} = renderPage({isError: true})

        expect(screen.getByText("Failed to load Docker daemons.")).toBeInTheDocument()
        fireEvent.click(screen.getByText("Retry"))

        expect(refetch).toHaveBeenCalledTimes(1)
    })

    it("opens CreateDindDialog when Create Docker daemon is clicked", () => {
        const {OpenDialog} = renderPage(loaded([]))

        fireEvent.click(screen.getByText("Create Docker daemon"))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(OpenDialog.mock.calls[0][0].type).toBe(CreateDindDialog)
    })
})
