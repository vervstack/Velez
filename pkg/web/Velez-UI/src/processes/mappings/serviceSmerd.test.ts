import {describe, expect, it} from "vitest"

import type {Smerd} from "@/app/api/velez"
import {findSmerdOfService, isSmerdOfService} from "@/processes/mappings/serviceSmerd"

describe("isSmerdOfService", () => {
    it("matches by the VERV_SERVICE label when the container name differs", () => {
        const smerd: Smerd = {name: "app_web_1", labels: {VERV_SERVICE: "app"}}

        expect(isSmerdOfService(smerd, "app")).toBe(true)
    })

    it("rejects a labelled smerd whose label names another service despite a matching name", () => {
        const smerd: Smerd = {name: "app", labels: {VERV_SERVICE: "other"}}

        expect(isSmerdOfService(smerd, "app")).toBe(false)
    })

    it("falls back to the exact container name when the smerd has no label", () => {
        expect(isSmerdOfService({name: "app"}, "app")).toBe(true)
    })

    it("falls back to the container name prefix when the smerd has no label", () => {
        expect(isSmerdOfService({name: "app-1"}, "app")).toBe(true)
    })

    it("does not match an unlabelled smerd with an unrelated name", () => {
        expect(isSmerdOfService({name: "application"}, "app")).toBe(false)
    })

    it("does not match anything when the service name is empty", () => {
        expect(isSmerdOfService({name: ""}, "")).toBe(false)
    })
})

describe("findSmerdOfService", () => {
    it("skips a smerd labelled for another service and returns the one labelled for this service", () => {
        const foreign: Smerd = {name: "app", labels: {VERV_SERVICE: "other"}}
        const own: Smerd = {name: "app_web", labels: {VERV_SERVICE: "app"}}

        expect(findSmerdOfService([foreign, own], "app")).toBe(own)
    })

    it("returns undefined when nothing matches", () => {
        expect(findSmerdOfService([{name: "x"}], "app")).toBeUndefined()
    })
})
