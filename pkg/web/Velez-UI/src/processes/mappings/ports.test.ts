import {describe, expect, it} from "vitest"

import {formatPortMapping} from "@/processes/mappings/ports.ts"

describe("formatPortMapping", () => {
    it("shows container and host port when the host port is published", () => {
        expect(formatPortMapping(5432, 18714)).toBe("5432 → 18714")
    })

    it("shows only the container port when the host port is absent", () => {
        expect(formatPortMapping(5432)).toBe("5432")
    })

    it("shows only the container port when the host port is zero", () => {
        expect(formatPortMapping(5432, 0)).toBe("5432")
    })

    it("shows a dash when the container port is missing", () => {
        expect(formatPortMapping(undefined, 18714)).toBe("-")
    })

    it("shows a dash when the container port is zero", () => {
        expect(formatPortMapping(0, 0)).toBe("-")
    })
})
