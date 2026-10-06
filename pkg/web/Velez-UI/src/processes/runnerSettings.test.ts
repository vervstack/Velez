import {describe, expect, it} from "vitest"

import {RunnerPullPolicy} from "@/app/api/velez"
import {
    arePolicyListsEqual,
    availablePullPolicies,
    canAddFallback,
    nextFallbackPolicy,
    parseNonNegativeInt,
    toggleAllowedPolicy,
} from "@/processes/runnerSettings.ts"

const ALWAYS = RunnerPullPolicy.RUNNER_PULL_POLICY_ALWAYS
const IF_NOT_PRESENT = RunnerPullPolicy.RUNNER_PULL_POLICY_IF_NOT_PRESENT
const NEVER = RunnerPullPolicy.RUNNER_PULL_POLICY_NEVER

describe("parseNonNegativeInt", () => {
    it("does parse a whole number", () => {
        expect(parseNonNegativeInt("30")).toBe(30)
    })

    it("does parse zero", () => {
        expect(parseNonNegativeInt("0")).toBe(0)
    })

    it("does treat an empty string as zero", () => {
        expect(parseNonNegativeInt("")).toBe(0)
        expect(parseNonNegativeInt("   ")).toBe(0)
    })

    it("does trim surrounding whitespace", () => {
        expect(parseNonNegativeInt(" 7 ")).toBe(7)
    })

    it("does return undefined for negative, decimal and exponent input", () => {
        expect(parseNonNegativeInt("-1")).toBeUndefined()
        expect(parseNonNegativeInt("1.5")).toBeUndefined()
        expect(parseNonNegativeInt("1e3")).toBeUndefined()
        expect(parseNonNegativeInt("abc")).toBeUndefined()
    })

    it("does return undefined for an unsafe integer", () => {
        expect(parseNonNegativeInt("99999999999999999999")).toBeUndefined()
    })
})

describe("availablePullPolicies", () => {
    it("does exclude policies chosen in other rows", () => {
        const ids = availablePullPolicies([ALWAYS, NEVER], 0).map((o) => o.id)

        expect(ids).toEqual([ALWAYS, IF_NOT_PRESENT])
    })

    it("does keep the row's own policy", () => {
        const ids = availablePullPolicies([IF_NOT_PRESENT], 0).map((o) => o.id)

        expect(ids).toEqual([ALWAYS, IF_NOT_PRESENT, NEVER])
    })
})

describe("nextFallbackPolicy", () => {
    it("does return the first unused policy", () => {
        expect(nextFallbackPolicy([ALWAYS])).toBe(IF_NOT_PRESENT)
    })

    it("does return undefined when every policy is used", () => {
        expect(nextFallbackPolicy([ALWAYS, IF_NOT_PRESENT, NEVER])).toBeUndefined()
    })
})

describe("canAddFallback", () => {
    it("does allow adding to a short chain", () => {
        expect(canAddFallback([])).toBe(true)
        expect(canAddFallback([ALWAYS, NEVER])).toBe(true)
    })

    it("does forbid adding once three are chosen", () => {
        expect(canAddFallback([ALWAYS, IF_NOT_PRESENT, NEVER])).toBe(false)
    })
})

describe("toggleAllowedPolicy", () => {
    it("does add a policy in canonical order", () => {
        expect(toggleAllowedPolicy([NEVER], ALWAYS, true)).toEqual([ALWAYS, NEVER])
    })

    it("does remove an unchecked policy", () => {
        expect(toggleAllowedPolicy([ALWAYS, NEVER], ALWAYS, false)).toEqual([NEVER])
    })

    it("does not duplicate an already checked policy", () => {
        expect(toggleAllowedPolicy([ALWAYS], ALWAYS, true)).toEqual([ALWAYS])
    })
})

describe("arePolicyListsEqual", () => {
    it("does treat identical ordered lists as equal", () => {
        expect(arePolicyListsEqual([ALWAYS, NEVER], [ALWAYS, NEVER])).toBe(true)
    })

    it("does treat a different order as unequal", () => {
        expect(arePolicyListsEqual([ALWAYS, NEVER], [NEVER, ALWAYS])).toBe(false)
    })

    it("does treat different lengths as unequal", () => {
        expect(arePolicyListsEqual([ALWAYS], [])).toBe(false)
    })
})
