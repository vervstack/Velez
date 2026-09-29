import {describe, it, expect} from "vitest"

import {parseContainerSearch} from "./parseContainerSearch"

describe("parseContainerSearch", () => {
    it("returns substring mode for plain text", () => {
        expect(parseContainerSearch("artel")).toEqual({mode: "substring", value: "artel"})
    })

    it("returns service mode when prefixed with service:", () => {
        expect(parseContainerSearch("service: artel")).toEqual({mode: "service", value: "artel"})
    })

    it("is case-insensitive on the service: prefix", () => {
        expect(parseContainerSearch("SERVICE:artel")).toEqual({mode: "service", value: "artel"})
    })

    it("trims surrounding whitespace", () => {
        expect(parseContainerSearch("  artel  ")).toEqual({mode: "substring", value: "artel"})
    })

    it("returns an empty substring value for an empty query", () => {
        expect(parseContainerSearch("")).toEqual({mode: "substring", value: ""})
    })
})
