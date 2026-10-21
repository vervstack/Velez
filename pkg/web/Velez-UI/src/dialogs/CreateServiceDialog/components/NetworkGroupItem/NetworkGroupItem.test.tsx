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
    const onRetry = vi.fn()

    render(
        <NetworkGroupItem
            container={CONTAINER}
            roles={{isRoot: false, isService: false, ...roles}}
            description={description}
            preview={{isLoading: false, isError: false, isDisabled: true, onRetry}}
        />
    )
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

    it("shows badges for the current service and root", () => {
        renderItem({roles: {isRoot: true, isService: true}})

        expect(screen.getByText("Service")).toBeInTheDocument()
        expect(screen.getByText("Network root")).toBeInTheDocument()
    })

    it("offers no role buttons", () => {
        renderItem()
        fireEvent.click(headerButton())

        expect(screen.queryByRole("button", {name: "Make service"})).not.toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Make network root"})).not.toBeInTheDocument()
    })
})
