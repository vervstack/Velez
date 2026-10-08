import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import ProvisioningProgressDialog
    from "@/dialogs/ProvisioningProgressDialog/ProvisioningProgressDialog.tsx"

const started = vi.fn()

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/widgets/TaskProgressScreen/TaskProgressScreen.tsx", () => ({
    default: ({start}: { start(): Promise<unknown> }) => (
        <Button onClick={() => void start().then(started)}>start task</Button>
    ),
}))

type Dialog = ReturnType<typeof useDialog>

function renderDialog() {
    const CloseDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)

    render(
        <ProvisioningProgressDialog
            title="Creating S3 instance"
            entityId="main"
            action="create_s3_instance"
            queryKey={["s3-instances"]}
        />
    )
    return {CloseDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ProvisioningProgressDialog", () => {
    it("shows the title in the dialog header", () => {
        renderDialog()

        expect(screen.getByRole("heading", {name: "Creating S3 instance"})).toBeInTheDocument()
    })

    it("re-attaches to the existing task by entity id and action when the task starts", async () => {
        renderDialog()

        fireEvent.click(screen.getByRole("button", {name: "start task"}))

        await vi.waitFor(() => {
            expect(started).toHaveBeenCalledWith({entityId: "main", action: "create_s3_instance"})
        })
    })

    it("closes the dialog when the close button is clicked", () => {
        const {CloseDialog} = renderDialog()

        fireEvent.click(screen.getByRole("button", {name: "✕"}))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
    })
})
