import {describe, expect, it} from "vitest"

import {SmerdStatus} from "@/app/api/velez"
import {imageMonogram, toServiceSidecars} from "@/processes/mappings/serviceSidecars"

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

describe("imageMonogram", () => {
    it("uses the first letter of the repository name without registry, owner or tag", () => {
        expect(imageMonogram("registry.io/redsockruf/zpotify:1.2")).toBe("Z")
    })

    it("strips a digest", () => {
        expect(imageMonogram("nginx@sha256:abc")).toBe("N")
    })

    it("returns a placeholder for an empty image name", () => {
        expect(imageMonogram("")).toBe("?")
        expect(imageMonogram(undefined)).toBe("?")
    })
})
