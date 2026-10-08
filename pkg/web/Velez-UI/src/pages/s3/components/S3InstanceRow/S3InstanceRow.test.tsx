import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen, waitFor} from "@testing-library/react"
import {MemoryRouter} from "react-router-dom"

import type {S3Instance} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {DropS3InstanceMutation} from "@/processes/queries/s3.ts"
import S3InstanceRow from "@/pages/s3/components/S3InstanceRow/S3InstanceRow.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/s3.ts", () => ({DropS3InstanceMutation: vi.fn()}))

type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>
type DropMutation = ReturnType<typeof DropS3InstanceMutation>

function renderRow(instance: S3Instance, mutateAsync = vi.fn().mockResolvedValue({})) {
    const OpenDialog = vi.fn()
    const CloseDialog = vi.fn()
    const bake = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(useDialog).mockReturnValue({OpenDialog, CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({bake, catchGrpc} as Partial<Toaster> as Toaster)
    vi.mocked(DropS3InstanceMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<DropMutation> as DropMutation
    )

    render(
        <MemoryRouter>
            <S3InstanceRow instance={instance} isSelected={false} onSelect={vi.fn()}/>
        </MemoryRouter>
    )
    return {OpenDialog, CloseDialog, bake, catchGrpc, mutateAsync}
}

function confirmDrop(OpenDialog: ReturnType<typeof vi.fn>) {
    fireEvent.click(screen.getByRole("button", {name: "Drop"}))
    OpenDialog.mock.calls[0][0].props.onConfirm()
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("S3InstanceRow", () => {
    it("shows the display name and keeps the instance name in the link", () => {
        renderRow({name: "s3_main", displayName: "main"})

        expect(screen.getByRole("link", {name: "main"})).toHaveAttribute("href", expect.stringContaining("s3_main"))
    })

    it("falls back to the instance name when there is no display name", () => {
        renderRow({name: "main", displayName: ""})

        expect(screen.getByRole("link", {name: "main"})).toBeInTheDocument()
    })

    it("asks for confirmation and does not drop before it is confirmed", () => {
        const {OpenDialog, mutateAsync} = renderRow({name: "main"})

        fireEvent.click(screen.getByRole("button", {name: "Drop"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(mutateAsync).not.toHaveBeenCalled()
    })

    it("drops by instance name and closes the dialog without a success toast when confirmed", async () => {
        const {OpenDialog, CloseDialog, bake, mutateAsync} = renderRow({name: "s3_main", displayName: "main"})

        confirmDrop(OpenDialog)

        await waitFor(() => expect(CloseDialog).toHaveBeenCalledTimes(1))
        expect(mutateAsync).toHaveBeenCalledWith("s3_main")
        expect(bake).not.toHaveBeenCalled()
    })

    it("reports the error and keeps the dialog open when the drop fails", async () => {
        const failure = new Error("boom")
        const {OpenDialog, CloseDialog, catchGrpc} = renderRow({name: "main"}, vi.fn().mockRejectedValue(failure))

        confirmDrop(OpenDialog)

        await waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
        expect(CloseDialog).not.toHaveBeenCalled()
    })
})
