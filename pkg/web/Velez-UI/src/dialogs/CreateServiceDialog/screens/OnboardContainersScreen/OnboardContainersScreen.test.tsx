import {beforeEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import OnboardContainersScreen from "@/dialogs/CreateServiceDialog/screens/OnboardContainersScreen/OnboardContainersScreen.tsx"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import {initialOnboardContext, OnboardContext} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {pickRoot} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

vi.mock("@/processes/queries/control_plane.ts", () => ({IsStatefullModeEnabled: vi.fn()}))

const app = {id: "app", name: "app", imageName: "nginx:1"}
const vpn = {id: "vpn", name: "vpn", imageName: "tailscale:1", networkOwnerContainerId: "app"}

function renderScreen(context: OnboardContext) {
    const updateContext = vi.fn()
    render(<OnboardContainersScreen {...context} updateContext={updateContext}/>)
    return {updateContext}
}

describe("OnboardContainersScreen", () => {
    beforeEach(() => {
        vi.mocked(IsStatefullModeEnabled).mockReturnValue(false)
    })

    it("renders a row for every container", () => {
        renderScreen(initialOnboardContext(app, [vpn]))

        expect(screen.getByText("tailscale:1")).toBeInTheDocument()
        expect(screen.getByText("nginx:1")).toBeInTheDocument()
        expect(screen.getAllByRole("button", {name: "Network root"})).toHaveLength(2)
    })

    it("updates the roles with the pickRoot result when a root is picked", () => {
        const context = initialOnboardContext(app, [vpn])
        const {updateContext} = renderScreen(context)

        fireEvent.click(screen.getAllByRole("button", {name: "Network root"})[1])

        expect(updateContext).toHaveBeenCalledWith({roles: pickRoot(context.roles, "vpn", ["app", "vpn"])})
    })

    it("shows the root change warning only after the root changed", () => {
        const context = initialOnboardContext(app, [vpn])
        const {unmount} = render(<OnboardContainersScreen {...context} updateContext={vi.fn()}/>)
        expect(screen.queryByRole("status")).not.toBeInTheDocument()
        unmount()

        renderScreen({...context, roles: {rootId: "vpn"}})

        expect(screen.getByRole("status")).toHaveTextContent("vpn becomes the network root")
    })

    it("disables the role buttons for a single container", () => {
        renderScreen(initialOnboardContext(app, []))

        expect(screen.getByRole("button", {name: "Network root"})).toBeDisabled()
        expect(screen.getByRole("button", {name: "Service"})).toBeDisabled()
        expect(screen.getByText("Only one container: nothing to choose.")).toBeInTheDocument()
    })
})
