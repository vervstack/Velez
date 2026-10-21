import {describe, expect, it, vi} from "vitest"
import {render, screen} from "@testing-library/react"

import type {RegisterContainerRequest} from "@/app/api/velez"
import OnboardProgress from "@/dialogs/CreateServiceDialog/components/OnboardProgress/OnboardProgress.tsx"
import {initialOnboardContext, OnboardContext} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"

vi.mock("@/dialogs/CreateServiceDialog/components/RegisterProgress/RegisterProgress.tsx", () => ({
    default: ({request, containerName}: {request: RegisterContainerRequest, containerName?: string}) => (
        <span>{`progress ${request.serviceName} for ${containerName}`}</span>
    ),
}))

function renderProgress(context: OnboardContext) {
    render(<OnboardProgress context={context} onClose={vi.fn()}/>)
}

describe("OnboardProgress", () => {
    it("shows the register progress of the built request", () => {
        const context = initialOnboardContext({id: "app", name: "app"}, [])
        const request: RegisterContainerRequest = {serviceName: "web"}

        renderProgress({...context, form: {request, isRecreating: true, hasPublishedPorts: false}})

        expect(screen.getByText("progress web for app")).toBeInTheDocument()
    })

    it("renders nothing without a request", () => {
        const context = initialOnboardContext({id: "app", name: "app"}, [])

        renderProgress({...context, form: {request: null, isRecreating: true, hasPublishedPorts: false}})

        expect(screen.queryByText(/progress/)).not.toBeInTheDocument()
    })
})
