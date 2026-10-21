import {afterEach, describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import OnboardConfirm from "@/dialogs/CreateServiceDialog/components/OnboardConfirm/OnboardConfirm.tsx"
import {initialOnboardContext, OnboardContext} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"

vi.mock("@/processes/queries/control_plane.ts", () => ({IsStatefullModeEnabled: vi.fn()}))

const app = {id: "app", name: "app"}
const vpn = {id: "vpn", name: "vpn", networkOwnerContainerId: "app"}

function renderConfirm(context: OnboardContext) {
    vi.mocked(IsStatefullModeEnabled).mockReturnValue(false)
    const onConfirm = vi.fn()
    const onCancel = vi.fn()
    render(<OnboardConfirm context={context} onConfirm={onConfirm} onCancel={onCancel}/>)
    return {onConfirm, onCancel}
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("OnboardConfirm", () => {
    it("asks to restart the container when the network root is unchanged", () => {
        const {onConfirm} = renderConfirm(initialOnboardContext(app, [vpn]))

        expect(screen.getByText("Restart container?")).toBeInTheDocument()
        fireEvent.click(screen.getByRole("button", {name: "Onboard and restart"}))

        expect(onConfirm).toHaveBeenCalledTimes(1)
    })

    it("asks to switch the network root when another container becomes the root", () => {
        const context = {...initialOnboardContext(app, [vpn]), roles: {rootId: "vpn"}}
        const {onCancel} = renderConfirm(context)

        expect(screen.getByText("Make network root?")).toBeInTheDocument()
        fireEvent.click(screen.getByRole("button", {name: "Cancel"}))

        expect(onCancel).toHaveBeenCalledTimes(1)
    })
})
