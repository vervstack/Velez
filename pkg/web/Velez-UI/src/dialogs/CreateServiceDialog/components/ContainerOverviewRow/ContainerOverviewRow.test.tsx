import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import ContainerOverviewRow from "@/dialogs/CreateServiceDialog/components/ContainerOverviewRow/ContainerOverviewRow.tsx"

function renderRow(overrides: {isReadOnly?: boolean; isRoot?: boolean; isService?: boolean} = {}) {
    const onPickRoot = vi.fn()
    const onPickService = vi.fn()
    render(
        <ContainerOverviewRow
            container={{id: "c1", name: "app", imageName: "nginx:1"}}
            networkMode="container:vpn"
            isRoot={overrides.isRoot ?? false}
            isService={overrides.isService ?? false}
            isReadOnly={overrides.isReadOnly ?? false}
            onPickRoot={onPickRoot}
            onPickService={onPickService}
        />,
    )
    return {onPickRoot, onPickService}
}

describe("ContainerOverviewRow", () => {
    it("shows the name, image and network mode", () => {
        renderRow()

        expect(screen.getByText("app")).toBeInTheDocument()
        expect(screen.getByText("nginx:1")).toBeInTheDocument()
        expect(screen.getByText("network: container:vpn")).toBeInTheDocument()
    })

    it("calls the pick callbacks with the container id when a role button is clicked", () => {
        const {onPickRoot, onPickService} = renderRow()

        fireEvent.click(screen.getByRole("button", {name: "Make root container"}))
        fireEvent.click(screen.getByRole("button", {name: "Make service container"}))

        expect(onPickRoot).toHaveBeenCalledWith("c1")
        expect(onPickService).toHaveBeenCalledWith("c1")
    })

    it("disables both role buttons when read only", () => {
        renderRow({isReadOnly: true})

        expect(screen.getByRole("button", {name: "Make root container"})).toBeDisabled()
        expect(screen.getByRole("button", {name: "Make service container"})).toBeDisabled()
    })

    it("shows a chip instead of the button for a role the container holds", () => {
        renderRow({isRoot: true, isService: true})

        expect(screen.getByText("Network root")).toBeInTheDocument()
        expect(screen.getByText("Service")).toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Make root container"})).not.toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Make service container"})).not.toBeInTheDocument()
    })
})
