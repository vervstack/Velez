import {describe, expect, it} from "vitest"

import {SmerdStatus} from "@/app/api/velez"
import {sidecarToEnvironment, toServiceSidecars} from "@/processes/mappings/serviceSidecars"

describe("toServiceSidecars", () => {
    it("returns nothing when the response carries no sidecars", () => {
        expect(toServiceSidecars(undefined)).toEqual([])
    })

    it("drops sidecars without a container id", () => {
        const result = toServiceSidecars([{containerName: "ghost"}, {containerId: "c1", containerName: "vpn"}])

        expect(result.map(sidecar => sidecar.containerId)).toEqual(["c1"])
    })

    it("falls back to the container id when the name is empty", () => {
        const result = toServiceSidecars([{containerId: "c1", containerName: ""}])

        expect(result[0].name).toBe("c1")
    })

    it("keeps image and status", () => {
        const result = toServiceSidecars([
            {
                containerId: "c1",
                containerName: "vpn",
                imageName: "tailscale/tailscale:latest",
                status: SmerdStatus.running,
            },
        ])

        expect(result[0]).toEqual({
            containerId: "c1",
            name: "vpn",
            imageName: "tailscale/tailscale:latest",
            status: SmerdStatus.running,
        })
    })
})

describe("sidecarToEnvironment", () => {
    it("maps a running sidecar to a running environment carrying its id, name and image", () => {
        const environment = sidecarToEnvironment({
            containerId: "c1",
            name: "garage-webui",
            imageName: "khairul169/garage-webui:1.0",
            status: SmerdStatus.running,
        })

        expect(environment).toMatchObject({
            id: "c1",
            label: "sidecar",
            status: "running",
            version: "khairul169/garage-webui:1.0",
            containerId: "c1",
            containerName: "garage-webui",
            health: "healthy",
        })
    })

    it("maps an exited sidecar to a stopped environment", () => {
        const environment = sidecarToEnvironment({
            containerId: "c2",
            name: "vpn",
            imageName: "tailscale/tailscale",
            status: SmerdStatus.exited,
        })

        expect(environment.status).toBe("stopped")
        expect(environment.health).toBe("unhealthy")
    })
})
