import {describe, expect, it} from "vitest"

import {expandAvailablePorts, isPortAvailable} from "@/processes/availablePorts.ts"

describe("availablePorts", () => {
    it("expands single ports and ranges", () => {
        const availablePorts = expandAvailablePorts(["18501:18503", "5432"])

        expect(availablePorts).toEqual(new Set([5432, 18501, 18502, 18503]))
    })

    it("ignores malformed ranges", () => {
        const availablePorts = expandAvailablePorts(["0", "abc", "18503:18501", "5432:5433:5434"])

        expect(availablePorts).toEqual(new Set())
    })

    it("checks whether a port belongs to the pool", () => {
        const availablePorts = expandAvailablePorts(["18501:18503"])

        expect(isPortAvailable(availablePorts, 18502)).toBe(true)
        expect(isPortAvailable(availablePorts, 5432)).toBe(false)
    })
})
