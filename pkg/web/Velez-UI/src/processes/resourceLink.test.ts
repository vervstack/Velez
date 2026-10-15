import {describe, it, expect} from "vitest"

import {resourcePath} from "@/processes/resourceLink"

describe("resourcePath", () => {
    it("returns the service path when a linked service is set even if a container id exists", () => {
        expect(resourcePath({linkedServiceName: "dind-main", containerId: "abc"})).toBe("/service/dind-main")
    })

    it("returns the container path when only a container id is set", () => {
        expect(resourcePath({containerId: "abc"})).toBe("/container/abc")
    })

    it("returns undefined when neither a service nor a container is linked", () => {
        expect(resourcePath({})).toBeUndefined()
    })
})
