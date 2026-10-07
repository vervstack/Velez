import {afterEach, describe, expect, it, vi} from "vitest"

import {pingAddress, resourceAddressUrl} from "@/processes/resourceAddresses.ts"

const location = {protocol: "http:", hostname: "velez.local"}

afterEach(() => {
    vi.unstubAllGlobals()
})

describe("resourceAddressUrl", () => {
    it("builds the url from the address host and port", () => {
        expect(resourceAddressUrl({host: "10.0.0.5", port: 3909}, location)).toBe("http://10.0.0.5:3909")
    })

    it("falls back to the location hostname when the host is empty", () => {
        expect(resourceAddressUrl({host: "", port: 3909}, location)).toBe("http://velez.local:3909")
    })

    it("keeps the location protocol", () => {
        expect(resourceAddressUrl({host: "h", port: 1}, {protocol: "https:", hostname: "x"})).toBe("https://h:1")
    })
})

describe("pingAddress", () => {
    it("resolves true when the request resolves", async () => {
        const fetchMock = vi.fn().mockResolvedValue({})
        vi.stubGlobal("fetch", fetchMock)

        await expect(pingAddress("http://h:1")).resolves.toBe(true)
        expect(fetchMock).toHaveBeenCalledWith("http://h:1", expect.objectContaining({mode: "no-cors"}))
    })

    it("resolves false when the request rejects", async () => {
        vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")))

        await expect(pingAddress("http://h:1")).resolves.toBe(false)
    })
})
