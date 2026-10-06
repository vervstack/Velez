import {describe, it, expect} from "vitest"

import {resourceWebUiUrl} from "./resourceWebUi"

describe("resourceWebUiUrl", () => {
    const mockLocation = {
        protocol: "http:",
        hostname: "localhost",
    }

    it("returns undefined when webUiPort is undefined", () => {
        const resource = {webUiPort: undefined, webUiHost: undefined}
        const url = resourceWebUiUrl(resource, mockLocation)
        expect(url).toBeUndefined()
    })

    it("returns undefined when webUiPort is 0", () => {
        const resource = {webUiPort: 0, webUiHost: undefined}
        const url = resourceWebUiUrl(resource, mockLocation)
        expect(url).toBeUndefined()
    })

    it("uses window location hostname when webUiHost is not set", () => {
        const resource = {webUiPort: 3909, webUiHost: undefined}
        const url = resourceWebUiUrl(resource, mockLocation)
        expect(url).toBe("http://localhost:3909")
    })

    it("uses webUiHost when it is set", () => {
        const resource = {webUiPort: 3909, webUiHost: "example.com"}
        const url = resourceWebUiUrl(resource, mockLocation)
        expect(url).toBe("http://example.com:3909")
    })

    it("uses the protocol from location", () => {
        const httpsLocation = {protocol: "https:", hostname: "localhost"}
        const resource = {webUiPort: 3909, webUiHost: undefined}
        const url = resourceWebUiUrl(resource, httpsLocation)
        expect(url).toBe("https://localhost:3909")
    })
})
