import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DropRunnerMutation} from "@/processes/queries/runners.ts"
import RunnerDropDialog from "@/dialogs/RunnerDropDialog/RunnerDropDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/runners.ts", () => ({DropRunnerMutation: vi.fn()}))

type DropMutation = ReturnType<typeof DropRunnerMutation>
type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>

function renderDialog(mutateAsync: ReturnType<typeof vi.fn>) {
    const CloseDialog = vi.fn()
    const bake = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(DropRunnerMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<DropMutation> as DropMutation,
    )
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake, catchGrpc} as Partial<Toaster> as Toaster)

    render(<RunnerDropDialog name="github_runner_build"/>)
    return {CloseDialog, bake, catchGrpc}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RunnerDropDialog", () => {
    it("asks for confirmation instead of dropping immediately", () => {
        const mutateAsync = vi.fn().mockResolvedValue(undefined)
        renderDialog(mutateAsync)

        expect(screen.getByText("Drop runner")).toBeInTheDocument()
        expect(mutateAsync).not.toHaveBeenCalled()
    })

    it("drops the runner by name and closes without a success toast when Drop is clicked", async () => {
        const mutateAsync = vi.fn().mockResolvedValue(undefined)
        const {CloseDialog, bake} = renderDialog(mutateAsync)

        fireEvent.click(screen.getByRole("button", {name: "Drop"}))

        expect(mutateAsync).toHaveBeenCalledWith("github_runner_build")
        await waitFor(() => expect(CloseDialog).toHaveBeenCalledTimes(1))
        expect(bake).not.toHaveBeenCalled()
    })

    it("keeps the dialog open and reports the error when the drop fails", async () => {
        const failure = new Error("drop failed")
        const {CloseDialog, catchGrpc} = renderDialog(vi.fn().mockRejectedValue(failure))

        fireEvent.click(screen.getByRole("button", {name: "Drop"}))

        await waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
        expect(CloseDialog).not.toHaveBeenCalled()
    })
})
