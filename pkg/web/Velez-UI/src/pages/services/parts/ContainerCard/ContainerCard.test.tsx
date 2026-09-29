import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {ServicePattern} from "@/app/api/velez"
import type {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import ContainerCard from "@/pages/services/parts/ContainerCard/ContainerCard.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx", () => ({default: () => null}))

function renderCard(container: DockerContainer) {
    const OpenDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue(
        {OpenDialog} as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>
    )
    const onOpen = vi.fn()

    render(<ContainerCard container={container} onOpen={onOpen} onFilterByService={vi.fn()}/>)
    return {OpenDialog, onOpen}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("ContainerCard", () => {
    it("offers Register for a container that is not linked to a service", () => {
        renderCard({id: "c1", name: "db"})

        expect(screen.getByRole("button", {name: "Register"})).toBeInTheDocument()
    })

    it("offers no Register for a container linked to a service", () => {
        renderCard({id: "c1", name: "db", linkedServiceName: "db-service"})

        expect(screen.queryByRole("button", {name: "Register"})).not.toBeInTheDocument()
    })

    it("opens the adopt dialog for the container without navigating to it", () => {
        const container = {id: "c1", name: "db"}
        const {OpenDialog, onOpen} = renderCard(container)

        fireEvent.click(screen.getByRole("button", {name: "Register"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(OpenDialog.mock.calls[0][0].props.container).toBe(container)
        expect(onOpen).not.toHaveBeenCalled()
    })

    it("hints at the suggested pattern of an unlinked container", () => {
        renderCard({id: "c1", name: "db", suggestedPattern: ServicePattern.SERVICE_PATTERN_POSTGRES})

        expect(screen.getByText("Looks like PostgreSQL")).toBeInTheDocument()
    })
})
