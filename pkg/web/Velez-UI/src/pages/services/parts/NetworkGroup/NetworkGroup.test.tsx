import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import type {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import NetworkGroup from "@/pages/services/parts/NetworkGroup/NetworkGroup.tsx"

vi.mock("@/app/hooks/dialog/Dialog.tsx", () => ({useDialog: vi.fn()}))
vi.mock("@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx", () => ({default: () => null}))

function renderGroup(root: DockerContainer, members: DockerContainer[]) {
    const OpenDialog = vi.fn()
    vi.mocked(useDialog).mockReturnValue(
        {OpenDialog} as Partial<ReturnType<typeof useDialog>> as ReturnType<typeof useDialog>
    )

    render(<NetworkGroup root={root} members={members} onOpen={vi.fn()} onFilterByService={vi.fn()}/>)
    return {OpenDialog}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("NetworkGroup", () => {
    it("shows the shared network caption", () => {
        renderGroup({id: "root", name: "vpn"}, [{id: "s1", name: "app"}])

        expect(screen.getByText("Shared network")).toBeInTheDocument()
    })

    it("shows the root network names in the caption", () => {
        renderGroup(
            {id: "root", name: "vpn", networks: [{networkName: "bridge"}, {networkName: "backend"}]},
            [{id: "s1", name: "app"}],
        )

        expect(screen.getByText("Shared network: bridge, backend")).toBeInTheDocument()
    })

    it("offers a single Register when a member is not registered", () => {
        renderGroup({id: "root", name: "vpn", isRegistered: true}, [
            {id: "s1", name: "app", isRegistered: false},
            {id: "s2", name: "worker", isRegistered: true},
        ])

        expect(screen.getAllByRole("button", {name: "Onboard"})).toHaveLength(1)
    })

    it("offers no Register when every container is registered", () => {
        renderGroup({id: "root", name: "vpn", isRegistered: true}, [{id: "s1", name: "app", isRegistered: true}])

        expect(screen.queryByRole("button", {name: "Onboard"})).not.toBeInTheDocument()
    })

    it("opens the dialog for the root with the sidecars when Register is clicked", () => {
        const root = {id: "root", name: "vpn"}
        const members = [{id: "s1", name: "app"}]
        const {OpenDialog} = renderGroup(root, members)

        fireEvent.click(screen.getByRole("button", {name: "Onboard"}))

        expect(OpenDialog).toHaveBeenCalledTimes(1)
        expect(OpenDialog.mock.calls[0][0].props.container).toBe(root)
        expect(OpenDialog.mock.calls[0][0].props.networkMembers).toBe(members)
    })
})
