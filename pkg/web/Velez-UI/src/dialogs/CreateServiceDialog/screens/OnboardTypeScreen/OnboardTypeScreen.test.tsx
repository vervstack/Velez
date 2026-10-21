import {describe, expect, it, vi} from "vitest"
import {fireEvent, render, screen} from "@testing-library/react"

import OnboardTypeScreen from "@/dialogs/CreateServiceDialog/screens/OnboardTypeScreen/OnboardTypeScreen.tsx"
import {initialOnboardContext, OnboardContext} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"

const app = {id: "app", name: "app"}
const vpn = {id: "vpn", name: "vpn", networkOwnerContainerId: "app"}

function renderScreen(context: OnboardContext) {
    const updateContext = vi.fn()
    render(<OnboardTypeScreen {...context} updateContext={updateContext}/>)
    return {updateContext}
}

describe("OnboardTypeScreen", () => {
    it("auto-selects the sole enabled type when the roles changed", () => {
        const context = initialOnboardContext(app, [vpn])
        const {updateContext} = renderScreen({...context, roles: {rootId: "vpn"}, screen: "postgres"})

        expect(updateContext).toHaveBeenCalledWith({screen: "generic"})
        expect(screen.getByText(/only the Generic type is available/)).toBeInTheDocument()
    })

    it("updates the context when a type card is clicked", () => {
        const {updateContext} = renderScreen(initialOnboardContext(app, [vpn]))

        fireEvent.click(screen.getByText("PostgreSQL"))

        expect(updateContext).toHaveBeenCalledWith({screen: "postgres"})
    })

    it("shows the hint only while no type is selected", () => {
        const context = initialOnboardContext(app, [vpn])
        const {unmount} = render(<OnboardTypeScreen {...context} updateContext={vi.fn()}/>)
        expect(screen.getByRole("status")).toHaveTextContent("Choose the type of this service to continue.")
        unmount()

        renderScreen({...context, screen: "postgres"})

        expect(screen.queryByRole("status")).not.toBeInTheDocument()
    })
})
