import {describe, expect, it} from "vitest"

import {
    initialRoles,
    isGenericOnly,
    pickRoot,
    pickService,
    roleIdsOf,
    serviceIdOf,
} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

const MEMBERS = ["a", "b", "c"]

describe("registerRoles", () => {
    it("starts with the original root as both root and service", () => {
        const roles = initialRoles("a")

        expect(roles.rootId).toBe("a")
        expect(serviceIdOf(roles)).toBe("a")
        expect(isGenericOnly(roles, "a")).toBe(false)
        expect(roleIdsOf(roles, "a")).toEqual({})
    })

    it("makes a picked member the service and keeps the root", () => {
        const roles = pickService(initialRoles("a"), "b")

        expect(roles.rootId).toBe("a")
        expect(serviceIdOf(roles)).toBe("b")
        expect(isGenericOnly(roles, "a")).toBe(true)
        expect(roleIdsOf(roles, "a")).toEqual({serviceContainerId: "b"})
    })

    it("follows the root again when the root is picked as the service", () => {
        const roles = pickService(pickService(initialRoles("a"), "b"), "a")

        expect(roles.explicitServiceId).toBeUndefined()
        expect(roleIdsOf(roles, "a")).toEqual({})
    })

    it("moves the service along with the root while no service was picked", () => {
        const roles = pickRoot(initialRoles("a"), "b", MEMBERS)

        expect(serviceIdOf(roles)).toBe("b")
        expect(isGenericOnly(roles, "a")).toBe(true)
        expect(roleIdsOf(roles, "a")).toEqual({rootContainerId: "b"})
    })

    it("keeps an explicitly picked service when the root changes", () => {
        const roles = pickRoot(pickService(initialRoles("a"), "c"), "b", MEMBERS)

        expect(roles.rootId).toBe("b")
        expect(serviceIdOf(roles)).toBe("c")
        expect(roleIdsOf(roles, "a")).toEqual({rootContainerId: "b", serviceContainerId: "c"})
    })

    it("follows the new root when the explicit service becomes the root", () => {
        const roles = pickRoot(pickService(initialRoles("a"), "b"), "b", MEMBERS)

        expect(roles.explicitServiceId).toBeUndefined()
        expect(roleIdsOf(roles, "a")).toEqual({rootContainerId: "b"})
    })

    it("drops an explicit service that no longer exists in the group", () => {
        const roles = pickRoot(pickService(initialRoles("a"), "gone"), "b", MEMBERS)

        expect(serviceIdOf(roles)).toBe("b")
    })

    it("is generic-only after switching the root and back-to-original sends nothing", () => {
        const switched = pickRoot(initialRoles("a"), "b", MEMBERS)
        const restored = pickRoot(switched, "a", MEMBERS)

        expect(isGenericOnly(switched, "a")).toBe(true)
        expect(isGenericOnly(restored, "a")).toBe(false)
        expect(roleIdsOf(restored, "a")).toEqual({})
    })

    it("sends both ids when the root moved away and another member stays the service", () => {
        const roles = pickRoot({rootId: "c", explicitServiceId: "a"}, "b", MEMBERS)

        expect(roleIdsOf(roles, "a")).toEqual({rootContainerId: "b", serviceContainerId: "a"})
        expect(isGenericOnly(roles, "a")).toBe(true)
    })
})
