import {describe, expect, it} from "vitest"

import {ServicePattern} from "@/app/api/velez"
import {formatContainerTimestamp, shortContainerId, suggestedPatternHint} from "@/processes/mappings/containers.ts"

describe("shortContainerId", () => {
    it("cuts the id to the 12-character docker short form when the id is longer", () => {
        expect(shortContainerId("0123456789abcdef0123456789abcdef")).toBe("0123456789ab")
    })

    it("returns an empty string when the id is empty", () => {
        expect(shortContainerId("")).toBe("")
    })
})

describe("suggestedPatternHint", () => {
    it("words a hint for every recognised pattern", () => {
        expect(suggestedPatternHint(ServicePattern.SERVICE_PATTERN_POSTGRES)).toBe("Looks like PostgreSQL")
        expect(suggestedPatternHint(ServicePattern.SERVICE_PATTERN_GITLAB_RUNNER)).toBe("Looks like a GitLab runner")
    })

    it("has no hint for an unspecified or missing pattern", () => {
        expect(suggestedPatternHint(ServicePattern.SERVICE_PATTERN_UNSPECIFIED)).toBeUndefined()
        expect(suggestedPatternHint(undefined)).toBeUndefined()
    })
})

describe("formatContainerTimestamp", () => {
    it("renders the RFC3339 string grpc-gateway sends", () => {
        const expected = new Date("2026-10-16T14:22:01Z").toLocaleString()
        expect(formatContainerTimestamp("2026-10-16T14:22:01.5Z")).toBe(expected)
    })

    it("renders the seconds and nanos object", () => {
        expect(formatContainerTimestamp({seconds: "1700000000"})).toBe(new Date(1700000000000).toLocaleString())
    })

    it("returns a dash when the timestamp is missing", () => {
        expect(formatContainerTimestamp(undefined)).toBe("—")
    })
})
