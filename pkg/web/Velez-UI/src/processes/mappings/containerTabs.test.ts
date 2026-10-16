import {describe, expect, it} from "vitest"

import {parseContainerTab} from "@/processes/mappings/containerTabs.ts"

describe("parseContainerTab", () => {
    it("returns the tab when the value is a known tab id", () => {
        expect(parseContainerTab("logs")).toBe("logs")
        expect(parseContainerTab("inspect")).toBe("inspect")
    })

    it("falls back to details when the value is missing or unknown", () => {
        expect(parseContainerTab(null)).toBe("details")
        expect(parseContainerTab("nope")).toBe("details")
    })
})
