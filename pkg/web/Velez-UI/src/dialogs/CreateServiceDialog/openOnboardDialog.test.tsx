import {afterEach, describe, expect, it, vi} from "vitest"

import {OpenWizardDialog} from "@/dialogs/StepsDialog/openWizardDialog.tsx"
import {openOnboardDialog} from "@/dialogs/CreateServiceDialog/openOnboardDialog.tsx"
import {initialOnboardContext} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"

vi.mock("@/dialogs/StepsDialog/openWizardDialog.tsx", () => ({OpenWizardDialog: vi.fn()}))

const app = {id: "app", name: "app"}
const vpn = {id: "vpn", name: "vpn", networkOwnerContainerId: "app"}

function openedProps() {
    return vi.mocked(OpenWizardDialog).mock.calls[0][0]
}

afterEach(() => {
    vi.clearAllMocks()
})

describe("openOnboardDialog", () => {
    it("starts on the type step for a single container", () => {
        openOnboardDialog(app)

        expect(openedProps().initialStepIndex).toBe(1)
    })

    it("starts on the containers step for a network group", () => {
        openOnboardDialog(app, [vpn])

        expect(openedProps().initialStepIndex).toBe(0)
        expect(openedProps().initialContext).toEqual(initialOnboardContext(app, [vpn]))
    })

    it("lists the containers, type and review steps and names the container in the header", () => {
        openOnboardDialog(app)

        expect(openedProps().steps.map((step) => step.label)).toEqual(["Containers", "Type", "Review"])
        expect(openedProps().steps[2].nextLabel).toBe("Onboard")
        expect(openedProps().header).toEqual({icon: "⇪", eyebrow: "Onboard container", eyebrowDetail: "app"})
    })

    it("lets the review step proceed only with a built request", () => {
        openOnboardDialog(app)
        const context = initialOnboardContext(app, [])
        const canProceed = openedProps().steps[2].canProceed!

        expect(canProceed(context)).toBe(false)
        expect(canProceed({...context, form: {request: {}, isRecreating: false, hasPublishedPorts: false}})).toBe(true)
    })

    it("confirms the finish only when the form recreates the container", () => {
        openOnboardDialog(app)
        const context = initialOnboardContext(app, [])
        const shouldConfirm = openedProps().shouldConfirmFinish!

        expect(shouldConfirm(context)).toBe(true)
        expect(shouldConfirm({...context, form: {request: {}, isRecreating: false, hasPublishedPorts: false}}))
            .toBe(false)
    })
})
