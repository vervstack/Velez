import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {CreateDindMutation} from "@/processes/queries/dinds.ts"
import Button from "@/components/base/Button.tsx"
import CreateDindDialog from "@/dialogs/CreateDindDialog/CreateDindDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/dinds.ts", () => ({
    CreateDindMutation: vi.fn(),
    DINDS_QUERY_KEY: ["dinds"],
}))
vi.mock("@/dialogs/CreateDindDialog/components/CreateDindDialogForm/CreateDindDialogForm.tsx", () => ({
    default: ({onSubmit, onCancel}: { onSubmit(req: { name: string }): void, onCancel(): void }) => (
        <>
            <Button variant="primary" onClick={() => onSubmit({name: "ci"})}>form submit</Button>
            <Button variant="secondary" onClick={onCancel}>form cancel</Button>
        </>
    ),
}))
vi.mock("@/widgets/TaskProgressScreen/TaskProgressScreen.tsx", () => ({
    default: ({start}: { start(): Promise<unknown> }) => (
        <Button onClick={() => void start()}>start task</Button>
    ),
}))

type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>
type CreateMutation = ReturnType<typeof CreateDindMutation>

function renderDialog() {
    const CloseDialog = vi.fn()
    const mutateAsync = vi.fn().mockResolvedValue({entityId: "ci", action: "create_dind"})
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc: vi.fn()} as Partial<Toaster> as Toaster)
    vi.mocked(CreateDindMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<CreateMutation> as CreateMutation
    )

    render(<CreateDindDialog/>)
    return {CloseDialog, mutateAsync}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("CreateDindDialog", () => {
    it("shows the title", () => {
        renderDialog()

        expect(screen.getByText("Create Docker daemon")).toBeInTheDocument()
    })

    it("swaps to the progress screen and creates the daemon when the form is submitted", () => {
        const {mutateAsync} = renderDialog()

        fireEvent.click(screen.getByText("form submit"))
        fireEvent.click(screen.getByText("start task"))

        expect(mutateAsync).toHaveBeenCalledWith({name: "ci"})
    })

    it("closes the dialog when the form is cancelled", () => {
        const {CloseDialog} = renderDialog()

        fireEvent.click(screen.getByText("form cancel"))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
    })
})
