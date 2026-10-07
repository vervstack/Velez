import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import NetworkClusterBanner from "@/components/NetworkClusterBanner/NetworkClusterBanner.tsx"

function renderBanner() {
    const onOpenVcn = vi.fn()
    render(<NetworkClusterBanner onOpenVcn={onOpenVcn}/>)
    return {onOpenVcn}
}

describe("NetworkClusterBanner", () => {
    it("explains that cross-node traffic may not be routed", () => {
        renderBanner()

        expect(screen.getByText(
            "Cluster mode is on and the closed network (VCN) is not connected. " +
            "Traffic between services on different nodes may not be routed. Set up VCN to connect them.",
        )).toBeInTheDocument()
    })

    it("opens VCN when Set up VCN is clicked", () => {
        const {onOpenVcn} = renderBanner()

        fireEvent.click(screen.getByRole("button", {name: "Set up VCN"}))

        expect(onOpenVcn).toHaveBeenCalledTimes(1)
    })
})
