import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {Network, NetworkCapability} from "@/app/api/velez/network_api.pb"
import NetworkRow from "@/pages/networks/components/NetworkRow/NetworkRow.tsx"

const MANAGED_CAPABILITIES = [
    "NETWORK_CAPABILITY_DELETE" as NetworkCapability,
    "NETWORK_CAPABILITY_ATTACH" as NetworkCapability,
]

function renderRow(network: Network) {
    const onAttach = vi.fn()
    const onDelete = vi.fn()
    const onDetach = vi.fn()

    render(<NetworkRow network={network} onAttach={onAttach} onDelete={onDelete} onDetach={onDetach}/>)
    return {onAttach, onDelete, onDetach}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("NetworkRow", () => {
    it("shows Attach and an enabled Delete for an empty managed network", () => {
        renderRow({name: "app-net", isManaged: true, capabilities: MANAGED_CAPABILITIES})

        expect(screen.getByText("Velez")).toBeInTheDocument()
        expect(screen.getByRole("button", {name: "Attach container"})).toBeEnabled()
        expect(screen.getByRole("button", {name: "Delete"})).toBeEnabled()
    })

    it("disables Delete with a hint while containers are attached", () => {
        renderRow({
            name: "app-net",
            isManaged: true,
            capabilities: MANAGED_CAPABILITIES,
            members: [{containerName: "api"}],
        })

        const deleteButton = screen.getByRole("button", {name: "Delete"})
        expect(deleteButton).toBeDisabled()
        expect(deleteButton).toHaveAttribute("data-tooltip-content", "Detach all containers first")
    })

    it("hides every action for a system network without capabilities", () => {
        renderRow({name: "bridge", members: [{containerName: "api"}]})

        expect(screen.getByText("Docker (foreign)")).toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Attach container"})).not.toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Delete"})).not.toBeInTheDocument()
        expect(screen.queryByRole("button", {name: "Detach"})).not.toBeInTheDocument()
    })

    it("shows flag chips for internal, isolation and subnet", () => {
        renderRow({name: "n", isInternal: true, isIccEnabled: false, subnet: "10.0.0.0/24"})

        expect(screen.getByText("internal")).toBeInTheDocument()
        expect(screen.getByText("container isolation")).toBeInTheDocument()
        expect(screen.getByText("10.0.0.0/24")).toBeInTheDocument()
    })

    it("calls the callbacks with the network and container", () => {
        const network: Network = {
            name: "app-net",
            isManaged: true,
            capabilities: MANAGED_CAPABILITIES,
            members: [{containerName: "api", ipAddress: "10.0.0.2", aliases: ["backend"]}],
        }
        const {onAttach, onDelete, onDetach} = renderRow(network)

        fireEvent.click(screen.getByRole("button", {name: "Attach container"}))
        fireEvent.click(screen.getByRole("button", {name: "Detach"}))
        fireEvent.click(screen.getByRole("button", {name: "Delete"}))

        expect(onAttach).toHaveBeenCalledWith(network)
        expect(onDetach).toHaveBeenCalledWith(network, "api")
        expect(onDelete).not.toHaveBeenCalled()
        expect(screen.getByText("backend")).toBeInTheDocument()
        expect(screen.getByText("10.0.0.2")).toBeInTheDocument()
    })

    it("calls onDelete for an empty managed network", () => {
        const network: Network = {name: "n", isManaged: true, capabilities: MANAGED_CAPABILITIES}
        const {onDelete} = renderRow(network)

        fireEvent.click(screen.getByRole("button", {name: "Delete"}))

        expect(onDelete).toHaveBeenCalledWith(network)
    })
})
