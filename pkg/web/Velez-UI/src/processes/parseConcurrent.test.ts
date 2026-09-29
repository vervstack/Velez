import {describe, it, expect} from "vitest"
import {parseConcurrent} from "@/processes/parseConcurrent.ts"

describe("parseConcurrent", () => {
    it("does parse a simple whole number", () => {
        expect(parseConcurrent("4")).toBe(4)
    })

    it("does parse a number with surrounding whitespace", () => {
        expect(parseConcurrent(" 4 ")).toBe(4)
    })

    it("does return undefined for zero", () => {
        expect(parseConcurrent("0")).toBeUndefined()
    })

    it("does return undefined for negative numbers", () => {
        expect(parseConcurrent("-1")).toBeUndefined()
    })

    it("does return undefined for decimal numbers", () => {
        expect(parseConcurrent("1.5")).toBeUndefined()
    })

    it("does return undefined for non-numeric input", () => {
        expect(parseConcurrent("abc")).toBeUndefined()
    })

    it("does return undefined for a number followed by text", () => {
        expect(parseConcurrent("4abc")).toBeUndefined()
    })

    it("does return undefined for exponent notation", () => {
        expect(parseConcurrent("1e3")).toBeUndefined()
    })

    it("does return undefined for empty string", () => {
        expect(parseConcurrent("")).toBeUndefined()
    })

    it("does return 1 for the minimum valid value", () => {
        expect(parseConcurrent("1")).toBe(1)
    })

    it("does return large numbers", () => {
        expect(parseConcurrent("100")).toBe(100)
    })
})
