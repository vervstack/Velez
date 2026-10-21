import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"
import type {ReactElement} from "react"

import type {WhitelistedImage} from "@/app/api/velez/settings_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {RemoveWhitelistedImageMutation} from "@/processes/queries/settings.ts"
import WhitelistedImageRow from "@/widgets/settings/SandboxSettings/components/WhitelistedImages/components/WhitelistedImageRow/WhitelistedImageRow.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/app/hooks/toaster/Toaster.ts", () => ({useToaster: vi.fn()}))
vi.mock("@/processes/queries/settings.ts", () => ({RemoveWhitelistedImageMutation: vi.fn()}))

type Dialog = ReturnType<typeof useDialog>
type Toaster = ReturnType<typeof useToaster>
type RemoveMutation = ReturnType<typeof RemoveWhitelistedImageMutation>

function renderRow(image: WhitelistedImage, mutateAsync = vi.fn().mockResolvedValue(undefined)) {
    const OpenDialog = vi.fn()
    const CloseDialog = vi.fn()
    const catchGrpc = vi.fn()
    vi.mocked(useDialog).mockReturnValue({OpenDialog, CloseDialog} as Partial<Dialog> as Dialog)
    vi.mocked(useToaster).mockReturnValue({catchGrpc} as Partial<Toaster> as Toaster)
    vi.mocked(RemoveWhitelistedImageMutation).mockReturnValue(
        {mutateAsync, isPending: false} as Partial<RemoveMutation> as RemoveMutation,
    )

    render(<WhitelistedImageRow image={image}/>)
    return {OpenDialog, CloseDialog, catchGrpc, mutateAsync}
}

function openedDialog(OpenDialog: ReturnType<typeof vi.fn>): ReactElement<{
    message: string
    onConfirm(): void
}> {
    return OpenDialog.mock.calls[0][0]
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("WhitelistedImageRow", () => {
    it("shows a built-in chip and no remove button for a built-in image", () => {
        renderRow({repository: "tailscale/tailscale", isBuiltin: true})

        expect(screen.getByText("tailscale/tailscale")).toBeInTheDocument()
        expect(screen.getByText("built-in")).toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Remove"})).not.toBeInTheDocument()
    })

    it("asks for confirmation instead of removing immediately when Remove is clicked", () => {
        const {OpenDialog, mutateAsync} = renderRow({id: "1", repository: "my/image"})

        fireEvent.click(screen.getByRole("button", {name: "Remove"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(openedDialog(OpenDialog).props.message).toBe(
            "Remove my/image from the whitelist? " +
            "Containers from this image can no longer be created with elevated access.",
        )
        expect(mutateAsync).not.toHaveBeenCalled()
    })

    it("removes the image by id and closes the dialog when the removal is confirmed", async () => {
        const {OpenDialog, CloseDialog, mutateAsync} = renderRow({id: "1", repository: "my/image"})
        fireEvent.click(screen.getByRole("button", {name: "Remove"}))

        openedDialog(OpenDialog).props.onConfirm()

        expect(mutateAsync).toHaveBeenCalledWith({id: "1"})
        await vi.waitFor(() => expect(CloseDialog).toHaveBeenCalledTimes(1))
    })

    it("routes a failed removal through the toaster error surface", async () => {
        const failure = new Error("remove failed")
        const {OpenDialog, catchGrpc} = renderRow(
            {id: "1", repository: "my/image"},
            vi.fn().mockRejectedValue(failure),
        )
        fireEvent.click(screen.getByRole("button", {name: "Remove"}))

        openedDialog(OpenDialog).props.onConfirm()

        await vi.waitFor(() => expect(catchGrpc).toHaveBeenCalledWith(failure))
    })
})
