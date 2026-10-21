import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DockerContainer} from "@/app/api/velez"
import NetworkGroupItem, {ItemRoles}
    from "@/dialogs/CreateServiceDialog/components/NetworkGroupItem/NetworkGroupItem.tsx"

const CONTAINER: DockerContainer = {id: "c1", name: "app", imageName: "nginx:1"}

interface RenderOptions {
    roles?: Partial<ItemRoles>
    description?: string
}

function renderItem({roles = {}, description = "Resource of web"}: RenderOptions = {}) {
    const onMakeService = vi.fn()
    const onMakeRoot = vi.fn()
    const onRetry = vi.fn()

    render(
        <NetworkGroupItem
            container={CONTAINER}
            roles={{isRoot: false, isService: false, canSwitch: true, ...roles}}
            description={description}
            preview={{isLoading: false, isError: false, isDisabled: true, onRetry}}
            onMakeService={onMakeService}
            onMakeRoot={onMakeRoot}
        />
    )
    return {onMakeService, onMakeRoot}
}

function headerButton(): HTMLElement {
    return screen.getAllByRole("button")[0]
}

describe("NetworkGroupItem", () => {
    it("shows the name and image of a closed item", () => {
        renderItem()

        expect(screen.getByText("app")).toBeInTheDocument()
        expect(screen.getByText("nginx:1")).toBeInTheDocument()
        expect(headerButton()).toHaveAttribute("aria-expanded", "false")
    })

    it("opens and closes when the header is clicked", () => {
        renderItem()

        fireEvent.click(headerButton())
        expect(headerButton()).toHaveAttribute("aria-expanded", "true")

        fireEvent.click(headerButton())
        expect(headerButton()).toHaveAttribute("aria-expanded", "false")
    })

    it("explains the role of the container when it is open", () => {
        renderItem({description: "Resource of web"})

        fireEvent.click(headerButton())

        expect(screen.getByText("Resource of web")).toBeInTheDocument()
    })

    it("makes the container the service in one click", () => {
        const {onMakeService, onMakeRoot} = renderItem()
        fireEvent.click(headerButton())

        fireEvent.click(screen.getByRole("button", {name: "Make service"}))

        expect(onMakeService).toHaveBeenCalledWith("c1")
        expect(onMakeRoot).not.toHaveBeenCalled()
    })

    it("asks to make the container the network root without applying it itself", () => {
        const {onMakeService, onMakeRoot} = renderItem()
        fireEvent.click(headerButton())

        fireEvent.click(screen.getByRole("button", {name: "Make network root"}))

        expect(onMakeRoot).toHaveBeenCalledWith("c1")
        expect(onMakeService).not.toHaveBeenCalled()
    })

    it("shows badges instead of buttons for the current service and root", () => {
        renderItem({roles: {isRoot: true, isService: true}})
        fireEvent.click(headerButton())

        expect(screen.getByText("Service")).toBeInTheDocument()
        expect(screen.getByText("Network root")).toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Make service"})).not.toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Make network root"})).not.toBeInTheDocument()
    })

    it("offers no role buttons when there is nothing to switch", () => {
        renderItem({roles: {canSwitch: false, isRoot: true, isService: true}})
        fireEvent.click(headerButton())

        expect(screen.queryByRole("button", {name: "Make service"})).not.toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Make network root"})).not.toBeInTheDocument()
    })
})
