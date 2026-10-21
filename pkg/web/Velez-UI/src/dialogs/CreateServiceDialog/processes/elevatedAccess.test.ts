import {describe, expect, it} from "vitest"

import type {DockerContainer} from "@/app/api/velez"
import {
    elevatedAccessItemsOf,
    elevatedAccessStateOf,
    isElevatedAccessPresent,
} from "@/dialogs/CreateServiceDialog/processes/elevatedAccess.ts"

const PRIVILEGED: DockerContainer = {elevatedAccess: {isPrivileged: true}}

describe("elevatedAccess", () => {
    it("is not elevated without any privilege, device or capability", () => {
        expect(isElevatedAccessPresent(undefined)).toBe(false)
        expect(isElevatedAccessPresent({capAdd: [], devices: []})).toBe(false)
    })

    it("is elevated with any single item", () => {
        expect(isElevatedAccessPresent({capAdd: ["NET_ADMIN"]})).toBe(true)
        expect(isElevatedAccessPresent({devices: [{hostPath: "/dev/net/tun"}]})).toBe(true)
        expect(isElevatedAccessPresent({isHostUserns: true})).toBe(true)
    })

    it("blocks a recreation of an elevated container with a non-whitelisted image until it is allowed", () => {
        expect(elevatedAccessStateOf(PRIVILEGED, true, false))
            .toEqual({isElevated: true, isAllowRequired: true, isBlocked: true})
        expect(elevatedAccessStateOf(PRIVILEGED, true, true))
            .toEqual({isElevated: true, isAllowRequired: true, isBlocked: false})
    })

    it("does not require allowing when the image is whitelisted", () => {
        const container: DockerContainer = {...PRIVILEGED, isImageWhitelisted: true}

        expect(elevatedAccessStateOf(container, true, false))
            .toEqual({isElevated: true, isAllowRequired: false, isBlocked: false})
    })

    it("is not elevated when the container is not recreated", () => {
        expect(elevatedAccessStateOf(PRIVILEGED, false, false))
            .toEqual({isElevated: false, isAllowRequired: false, isBlocked: false})
    })

    it("lists privileged, devices, capabilities and host namespaces as labels", () => {
        const items = elevatedAccessItemsOf({
            isPrivileged: true,
            devices: [{hostPath: "/dev/net/tun"}],
            capAdd: ["NET_ADMIN"],
            isHostPid: true,
            isHostIpc: true,
            isHostUserns: true,
        })

        expect(items).toEqual([
            "Privileged", "Device /dev/net/tun", "Cap NET_ADMIN", "Host PID", "Host IPC", "Host user namespace",
        ])
    })
})
