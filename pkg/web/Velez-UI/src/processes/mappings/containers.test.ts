import {describe, expect, it} from "vitest"

import {ServicePattern} from "@/app/api/velez"
import {suggestedPatternHint} from "@/processes/mappings/containers.ts"

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
