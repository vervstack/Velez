import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DropPgInstanceMutation} from "@/processes/queries/pg_instances.ts"
import PgInstanceDropDialog from "@/dialogs/PgInstanceDropDialog/PgInstanceDropDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/pg_instances.ts", () => ({DropPgInstanceMutation: vi.fn()}))

type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>
type DropMutation = ReturnType<typeof DropPgInstanceMutation>

function renderDialog(mutateAsync = vi.fn().mockResolvedValue({})) {
    const CloseDialog = vi.fn()
    const bake = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake, catchGrpc} as Partial<Toaster> as Toaster)
    vi.mocked(DropPgInstanceMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<DropMutation> as DropMutation
    )

    render(<PgInstanceDropDialog name="main"/>)
    return {CloseDialog, bake, catchGrpc, mutateAsync}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("PgInstanceDropDialog", () => {
    it("does not drop before the confirmation button is clicked", () => {
        const {mutateAsync} = renderDialog()

        expect(screen.getByText("Drop database")).toBeInTheDocument()
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
