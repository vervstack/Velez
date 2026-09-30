import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {SmerdStatus} from "@/app/api/velez"
import ContainerImageTile from "@/components/smerd/ContainerImageTile/ContainerImageTile.tsx"

function renderTile() {
    const onOpen = vi.fn()

    render(
        <ContainerImageTile
            imageName="tailscale/tailscale:latest"
            label="vpn-sidecar"
            status={SmerdStatus.running}
            onOpen={onOpen}
        />
    )
    return {onOpen}
}

describe("ContainerImageTile", () => {
    it("shows the container name and the image monogram", () => {
        renderTile()

        expect(screen.getByRole("button", {name: /vpn-sidecar/})).toBeTruthy()
        expect(screen.getByText("T")).toBeTruthy()
    })

    it("calls onOpen when clicked", () => {
        const {onOpen} = renderTile()

        fireEvent.click(screen.getByRole("button", {name: /vpn-sidecar/}))

        expect(onOpen).toHaveBeenCalledTimes(1)
    })
})
