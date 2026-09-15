import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"
import {MemoryRouter} from "react-router-dom"

import RegistryInstanceRow from "@/pages/container-registry/components/RegistryInstanceRow/RegistryInstanceRow.tsx"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import type {RegistryInstance} from "@/app/api/velez"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))

vi.mock("@/pages/container-registry/components/RegistryInstanceCredentials/RegistryInstanceCredentials.tsx", () => ({
    default: () => null,
}))

function renderRow(instance: Partial<RegistryInstance>) {
    const OpenDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue(
        {OpenDialog} as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>
    )

    const fullInstance: RegistryInstance = {
        name: "my-registry", environment: "prod", port: 5000, username: "verv",
        status: "running", uiPort: 0, ...instance,
    }

    render(
        <MemoryRouter>
            <RegistryInstanceRow instance={fullInstance}/>
        </MemoryRouter>
    )
    return {OpenDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("RegistryInstanceRow", () => {
    it("hides the image browser button when ui_port is 0", () => {
        renderRow({uiPort: 0})

        expect(screen.queryByText(/Open image browser/)).not.toBeInTheDocument()
    })

    it("shows and opens the image browser link when ui_port is set and the instance is running", () => {
        const openSpy = vi.spyOn(window, "open").mockImplementation(() => null)
        renderRow({uiPort: 8080, status: "running"})

        fireEvent.click(screen.getByText(/Open image browser/))

        expect(openSpy).toHaveBeenCalledWith(expect.stringContaining(":8080"), "_blank")
    })

    it("opens the drop dialog with the instance name when Drop is clicked", () => {
        const {OpenDialog} = renderRow({name: "my-registry"})

        fireEvent.click(screen.getByText("Drop"))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        const openedElement = OpenDialog.mock.calls[0][0]
        expect(openedElement.props.name).toBe("my-registry")
    })
})
