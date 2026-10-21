import {describe, expect, it} from "vitest"

import type {DockerContainer} from "@/app/api/velez"
import {
    enabledScreensOf,
    initialOnboardContext,
    isRolesChanged,
    isRootChanged,
    isTypeStepReady,
    nameOfContainer,
    networkModeOf,
    OnboardContext,
    selectedScreenOf,
    serviceNameOf,
} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {ADOPTABLE_SCREENS} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

const root: DockerContainer = {id: "root", name: "app", networks: [{networkName: "bridge"}, {networkName: "backend"}]}
const sidecar: DockerContainer = {id: "side", name: "vpn", networkOwnerContainerId: "root"}

function newContext(overrides: Partial<OnboardContext> = {}): OnboardContext {
    return {...initialOnboardContext(root, [sidecar]), ...overrides}
}

describe("initialOnboardContext", () => {
    it("lists the clicked container first and roots the roles on it", () => {
        const context = initialOnboardContext(root, [sidecar])

        expect(context.containers).toEqual([root, sidecar])
        expect(context.roles).toEqual({rootId: "root"})
        expect(context.screen).toBeUndefined()
    })
})

describe("nameOfContainer", () => {
    it("returns the name of a known container and an empty string otherwise", () => {
        const context = newContext()

        expect(nameOfContainer(context, "side")).toBe("vpn")
        expect(nameOfContainer(context, "missing")).toBe("")
    })
})

describe("serviceNameOf", () => {
    it("prefers the draft over the service container name", () => {
        expect(serviceNameOf(newContext({serviceNameDraft: "custom"}))).toBe("custom")
    })

    it("falls back to the name of the service container", () => {
        expect(serviceNameOf(newContext())).toBe("app")
        expect(serviceNameOf(newContext({roles: {rootId: "root", explicitServiceId: "side"}}))).toBe("vpn")
    })
})

describe("role flags", () => {
    const cases = [
        {name: "untouched", roles: {rootId: "root"}, isRootChanged: false, isRolesChanged: false},
        {name: "service moved", roles: {rootId: "root", explicitServiceId: "side"}, isRootChanged: false, isRolesChanged: true},
        {name: "root moved", roles: {rootId: "side"}, isRootChanged: true, isRolesChanged: true},
    ]

    for (const tc of cases) {
        it(`reports ${tc.name}`, () => {
            const context = newContext({roles: tc.roles})

            expect(isRootChanged(context)).toBe(tc.isRootChanged)
            expect(isRolesChanged(context)).toBe(tc.isRolesChanged)
        })
    }
})

describe("screens", () => {
    it("enables every adoptable screen while the roles are untouched", () => {
        expect(enabledScreensOf(newContext())).toEqual(ADOPTABLE_SCREENS)
    })

    it("enables only generic once the roles change", () => {
        expect(enabledScreensOf(newContext({roles: {rootId: "side"}}))).toEqual(["generic"])
    })

    it("keeps the chosen screen while it is enabled", () => {
        expect(selectedScreenOf(newContext({screen: "postgres"}))).toBe("postgres")
    })

    it("selects nothing without a choice among several screens", () => {
        const context = newContext()

        expect(selectedScreenOf(context)).toBeUndefined()
        expect(isTypeStepReady(context)).toBe(false)
    })

    it("selects the sole enabled screen and drops a stale choice", () => {
        const context = newContext({roles: {rootId: "side"}, screen: "postgres"})

        expect(selectedScreenOf(context)).toBe("generic")
        expect(isTypeStepReady(context)).toBe(true)
    })
})

describe("networkModeOf", () => {
    it("joins the network names of the original root", () => {
        expect(networkModeOf(root, newContext())).toBe("bridge, backend")
    })

    it("falls back to default without networks", () => {
        expect(networkModeOf({id: "x", name: "x"}, newContext())).toBe("default")
    })

    it("names the owner container for a network member", () => {
        expect(networkModeOf(sidecar, newContext())).toBe("container:app")
    })

    it("falls back to the original container name when the owner is unknown", () => {
        const orphan: DockerContainer = {id: "o", name: "o", networkOwnerContainerId: "gone"}

        expect(networkModeOf(orphan, newContext())).toBe("container:app")
    })
})
