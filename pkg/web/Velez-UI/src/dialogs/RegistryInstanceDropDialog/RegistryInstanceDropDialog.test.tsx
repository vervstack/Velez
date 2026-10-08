import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DropRegistryInstanceMutation} from "@/processes/queries/registry_instances.ts"
import RegistryInstanceDropDialog from "@/dialogs/RegistryInstanceDropDialog/RegistryInstanceDropDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/registry_instances.ts", () => ({DropRegistryInstanceMutation: vi.fn()}))

type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>
type DropMutation = ReturnType<typeof DropRegistryInstanceMutation>

function renderDialog(mutateAsync = vi.fn().mockResolvedValue({})) {
    const CloseDialog = vi.fn()
    const bake = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake, catchGrpc} as Partial<Toaster> as Toaster)
    vi.mocked(DropRegistryInstanceMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<DropMutation> as DropMutation
    )

    render(<RegistryInstanceDropDialog name="main"/>)
    return {CloseDialog, bake, catchGrpc, mutateAsync}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RegistryInstanceDropDialog", () => {
    it("does not drop before the confirmation button is clicked", () => {
        const {mutateAsync} = renderDialog()

        expect(screen.getByText("Drop registry")).toBeInTheDocument()
        expect(mutateAsync).not.toHaveBeenCalled()
    })

    it("drops the instance and closes the dialog without a success toast when Drop is clicked", async () => {
        const {CloseDialog, bake, mutateAsync} = renderDialog()

        fireEvent.click(screen.getByRole("button", {name: "Drop"}))

        await waitFor(() => expect(CloseDialog).toHaveBeenCalledTimes(1))
        expect(mutateAsync).toHaveBeenCalledWith("main")
        expect(bake).not.toHaveBeenCalled()
    })

    it("reports the error and keeps the dialog open when the drop fails", async () => {
        const failure = new Error("boom")
        const {CloseDialog, catchGrpc} = renderDialog(vi.fn().mockRejectedValue(failure))

        fireEvent.click(screen.getByRole("button", {name: "Drop"}))

        await waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
        expect(CloseDialog).not.toHaveBeenCalled()
    })
})
