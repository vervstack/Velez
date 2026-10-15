import {describe, expect, it} from "vitest"

import {ServicePattern} from "@/app/api/velez"
import {shortContainerId, suggestedPatternHint} from "@/processes/mappings/containers.ts"

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
