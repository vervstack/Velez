import {describe, it, expect} from "vitest"

import type {ResourceAddress} from "@/model/service_page/ServicePageModel"

import {resourceGoToUrl} from "./resourceGoTo"

describe("resourceGoToUrl", () => {
    const mockLocation = {
        protocol: "http:",
        hostname: "localhost",
    }

    it("uses the only address when there is exactly one", () => {
        const addresses: ResourceAddress[] = [{host: "10.0.0.5", port: 8080, scope: "vcn"}]
        const url = resourceGoToUrl(addresses, mockLocation)
        expect(url).toBe("http://10.0.0.5:8080")
    })

    it("uses the page hostname when the only address has an empty host", () => {
        const addresses: ResourceAddress[] = [{host: "", port: 8080, scope: "docker"}]
        const url = resourceGoToUrl(addresses, mockLocation)
        expect(url).toBe("http://localhost:8080")
    })

    it("uses the address whose host matches the page hostname when there are several", () => {
        const addresses: ResourceAddress[] = [
            {host: "10.0.0.5", port: 8080, scope: "vcn"},
            {host: "localhost", port: 9090, scope: "docker"},
        ]
        const url = resourceGoToUrl(addresses, mockLocation)
        expect(url).toBe("http://localhost:9090")
    })

    it("treats an empty host as the page hostname when several addresses are present", () => {
        const addresses: ResourceAddress[] = [
            {host: "10.0.0.5", port: 8080, scope: "vcn"},
            {host: "", port: 9090, scope: "docker"},
        ]
        const url = resourceGoToUrl(addresses, mockLocation)
        expect(url).toBe("http://localhost:9090")
    })

    it("returns undefined when several addresses match no page hostname", () => {
        const addresses: ResourceAddress[] = [
            {host: "10.0.0.5", port: 8080, scope: "vcn"},
            {host: "10.0.0.6", port: 9090, scope: "docker"},
        ]
        const url = resourceGoToUrl(addresses, mockLocation)
        expect(url).toBeUndefined()
    })

    it("returns undefined when there are no addresses", () => {
        const url = resourceGoToUrl([], mockLocation)
        expect(url).toBeUndefined()
    })
})
