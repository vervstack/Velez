import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import CreateDindDialog from "@/dialogs/CreateDindDialog/CreateDindDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/widgets/CreateDindForm/CreateDindForm.tsx", () => ({
    default: ({onCreated, onCancel}: { onCreated(name: string): void, onCancel(): void }) => (
        <>
            <Button variant="primary" onClick={() => onCreated("ci")}>form created</Button>
            <Button variant="secondary" onClick={onCancel}>form cancel</Button>
        </>
    ),
}))

type Dialog = ReturnType<typeof useDialog>

function renderDialog() {
    const CloseDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)

    render(<CreateDindDialog/>)
    return {CloseDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("CreateDindDialog", () => {
    it("shows the title", () => {
        renderDialog()

        expect(screen.getByText("Create Docker daemon")).toBeInTheDocument()
    })

    it("closes the dialog when the form reports the daemon created", () => {
        const {CloseDialog} = renderDialog()

        fireEvent.click(screen.getByText("form created"))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
    })

    it("closes the dialog when the form is cancelled", () => {
        const {CloseDialog} = renderDialog()

        fireEvent.click(screen.getByText("form cancel"))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
    })
})
