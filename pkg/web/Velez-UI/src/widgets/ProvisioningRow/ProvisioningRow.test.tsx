import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {ProvisioningTaskStatus} from "@/app/api/velez/velez_common.pb"
import type {ProvisioningTask} from "@/app/api/velez/velez_common.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DismissTaskMutation} from "@/processes/queries/provisioning.ts"
import ProvisioningRow from "@/widgets/ProvisioningRow/ProvisioningRow.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/provisioning.ts", () => ({DismissTaskMutation: vi.fn()}))
vi.mock("@/dialogs/ProvisioningProgressDialog/ProvisioningProgressDialog.tsx", () => ({
    default: ({entityId}: { entityId: string }) => <span>progress dialog for {entityId}</span>,
}))

type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>
type DismissMutation = ReturnType<typeof DismissTaskMutation>

const queryKey = ["pg-instances"]

function renderRow(task: Partial<ProvisioningTask>) {
    const OpenDialog = vi.fn()
    const mutateAsync = vi.fn().mockResolvedValue(undefined)
    vi.mocked(useDialog).mockReturnValue({OpenDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)
    vi.mocked(DismissTaskMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<DismissMutation> as DismissMutation
    )

    const fullTask: ProvisioningTask = {
        taskId: "7",
        entityId: "main",
        action: "create_pg_instance",
        status: ProvisioningTaskStatus.RUNNING,
        ...task,
    }
    render(<ProvisioningRow task={fullTask} title="Creating Postgres instance" queryKey={queryKey}/>)
    return {OpenDialog, mutateAsync, fullTask}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ProvisioningRow", () => {
    it("shows the current step and progress when the task is running", () => {
        renderRow({
            jobs: [
                {name: "pull image", status: ProvisioningTaskStatus.DONE},
                {name: "start container", status: ProvisioningTaskStatus.RUNNING},
                {name: "create user", status: ProvisioningTaskStatus.PENDING},
            ],
        })

        expect(screen.getByText("main")).toBeInTheDocument()
        expect(screen.getByText("start container")).toBeInTheDocument()
        expect(screen.getByText("1/3")).toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Dismiss"})).not.toBeInTheDocument()
    })

    it("shows the error and dismisses the task when the task failed and Dismiss is clicked", () => {
        const {mutateAsync, fullTask} = renderRow({status: ProvisioningTaskStatus.FAILED, error: "image pull failed"})

        expect(screen.getByText("image pull failed")).toBeInTheDocument()

        fireEvent.click(screen.getByRole("button", {name: "Dismiss"}))

        expect(mutateAsync).toHaveBeenCalledWith(fullTask)
        expect(DismissTaskMutation).toHaveBeenCalledWith(queryKey)
    })

    it("opens the progress dialog for the task when Details is clicked", () => {
        const {OpenDialog} = renderRow({})

        fireEvent.click(screen.getByRole("button", {name: "Details"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        const opened = OpenDialog.mock.calls[0][0]
        expect(opened.props).toEqual({
            title: "Creating Postgres instance",
            entityId: "main",
            action: "create_pg_instance",
            queryKey,
        })
    })

    it("offers Details on a failed task", () => {
        renderRow({status: ProvisioningTaskStatus.FAILED, error: "boom"})

        expect(screen.getByRole("button", {name: "Details"})).toBeInTheDocument()
    })
})
