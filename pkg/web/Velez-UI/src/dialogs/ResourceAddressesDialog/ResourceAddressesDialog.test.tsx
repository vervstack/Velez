import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import type {ResourceAddress} from "@/model/service_page/ServicePageModel"
import ResourceAddressesDialog from "@/dialogs/ResourceAddressesDialog/ResourceAddressesDialog.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/dialogs/ResourceAddressesDialog/components/AddressRow/AddressRow.tsx", () => ({
    default: ({address}: {address: ResourceAddress}) => <span>{`row ${address.host}:${address.port}`}</span>,
}))

type Dialog = ReturnType<typeof useDialog>

function renderDialog(addresses: ResourceAddress[]) {
    const CloseDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue({CloseDialog} as Partial<Dialog> as Dialog)

    render(<ResourceAddressesDialog resourceName="garage" addresses={addresses}/>)
    return {CloseDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ResourceAddressesDialog", () => {
    it("shows the resource name in the title", () => {
        renderDialog([])

        expect(screen.getByText("garage — addresses")).toBeInTheDocument()
    })

    it("renders one row per address", () => {
        renderDialog([
            {host: "10.0.0.5", port: 3909, scope: "docker"},
            {host: "100.64.0.2", port: 3909, scope: "vcn"},
        ])

        expect(screen.getByText("row 10.0.0.5:3909")).toBeInTheDocument()
        expect(screen.getByText("row 100.64.0.2:3909")).toBeInTheDocument()
    })

    it("closes the dialog when the close button is clicked", () => {
        const {CloseDialog} = renderDialog([])

        fireEvent.click(screen.getByText("✕"))

        expect(CloseDialog).toHaveBeenCalledTimes(1)
    })
})
