import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DropDindMutation} from "@/processes/queries/dinds.ts"
import DindRow from "@/pages/dinds/components/DindRow/DindRow.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/dinds.ts", () => ({DropDindMutation: vi.fn()}))

type DropMutation = ReturnType<typeof DropDindMutation>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function renderRow(mutateAsync = vi.fn().mockResolvedValue(undefined)) {
    const OpenDialog = vi.fn()
    const CloseDialog = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(DropDindMutation).mockReturnValue({mutateAsync} as Partial<DropMutation> as DropMutation)
    vi.mocked(useDialog).mockReturnValue({OpenDialog, CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake: vi.fn(), catchGrpc} as Partial<Toaster> as Toaster)

    render(<DindRow dind={{name: "ci", address: "tcp://ci:2375", isSysboxEnabled: true}}/>)
    return {mutateAsync, OpenDialog, CloseDialog, catchGrpc}
}

function openConfirmDialog(OpenDialog: ReturnType<typeof vi.fn>) {
    fireEvent.click(screen.getByText("Drop"))
    render(OpenDialog.mock.calls[0][0])
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("DindRow", () => {
    it("shows the name, address and Sysbox state", () => {
        renderRow()

        expect(screen.getByText("ci")).toBeInTheDocument()
        expect(screen.getByText("tcp://ci:2375")).toBeInTheDocument()
        expect(screen.getByText("Yes")).toBeInTheDocument()
    })

    it("asks for confirmation instead of dropping immediately when Drop is clicked", () => {
        const {mutateAsync, OpenDialog} = renderRow()

        openConfirmDialog(OpenDialog)

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(screen.getByText("Drop Docker daemon?")).toBeInTheDocument()
        expect(mutateAsync).not.toHaveBeenCalled()
    })

    it("drops the daemon by name and closes the dialog once confirmed", async () => {
        const {mutateAsync, OpenDialog, CloseDialog} = renderRow()
        openConfirmDialog(OpenDialog)

        fireEvent.click(screen.getAllByText("Drop").at(-1) as HTMLElement)

        await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith("ci"))
        await waitFor(() => expect(CloseDialog).toHaveBeenCalled())
    })

    it("routes a failed drop through the toaster error surface", async () => {
        const failure = new Error("drop failed")
        const {OpenDialog, catchGrpc} = renderRow(vi.fn().mockRejectedValue(failure))
        openConfirmDialog(OpenDialog)

        fireEvent.click(screen.getAllByText("Drop").at(-1) as HTMLElement)

        await waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
    })
})
