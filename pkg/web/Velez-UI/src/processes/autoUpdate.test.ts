import {describe, expect, it} from "vitest"

import {AutoUpdateBumpLevel} from "@/app/api/velez/auto_update_api.pb"
import {
    BUMP_LEVEL_OPTIONS,
    bumpLevelName,
    confirmTitle,
    describeCheckPeriod,
    initialBumpLevel,
    restartWarning,
    sortServices,
} from "@/processes/autoUpdate.ts"

describe("initialBumpLevel", () => {
    it("falls back to PATCH when the level is unspecified or missing", () => {
        expect(initialBumpLevel(AutoUpdateBumpLevel.AUTO_UPDATE_BUMP_LEVEL_UNSPECIFIED)).toBe(AutoUpdateBumpLevel.PATCH)
        expect(initialBumpLevel(undefined)).toBe(AutoUpdateBumpLevel.PATCH)
    })

    it("keeps a specified level", () => {
        expect(initialBumpLevel(AutoUpdateBumpLevel.MAJOR)).toBe(AutoUpdateBumpLevel.MAJOR)
    })
})

describe("bumpLevelName", () => {
    it("names every level the user can pick", () => {
        expect(BUMP_LEVEL_OPTIONS.map((o) => bumpLevelName(o.id))).toEqual([
            "Major only",
            "Major + minor",
            "All versions (major, minor, patch)",
        ])
    })

    it("names an unspecified level as the default one", () => {
        expect(bumpLevelName(AutoUpdateBumpLevel.AUTO_UPDATE_BUMP_LEVEL_UNSPECIFIED))
            .toBe("All versions (major, minor, patch)")
    })
})

describe("sortServices", () => {
    it("sorts by name without mutating the input", () => {
        const input = [{name: "zeta"}, {name: "alpha"}]

        expect(sortServices(input).map((s) => s.name)).toEqual(["alpha", "zeta"])
        expect(input[0].name).toBe("zeta")
    })
})

describe("describeCheckPeriod", () => {
    it("describes seconds, minutes and hours with the right plural", () => {
        expect(describeCheckPeriod("30")).toBe("Checks every 30 seconds")
        expect(describeCheckPeriod("1")).toBe("Checks every 1 second")
        expect(describeCheckPeriod("60")).toBe("Checks every 1 minute")
        expect(describeCheckPeriod("300")).toBe("Checks every 5 minutes")
        expect(describeCheckPeriod(7200)).toBe("Checks every 2 hours")
    })

    it("says the period is not configured for a missing or zero value", () => {
        expect(describeCheckPeriod(undefined)).toBe("Check period is not configured")
        expect(describeCheckPeriod("0")).toBe("Check period is not configured")
    })
})

describe("confirmTitle", () => {
    it("asks to enable, disable or change the level depending on what flips", () => {
        expect(confirmTitle(false, true)).toBe("Enable auto-update?")
        expect(confirmTitle(true, false)).toBe("Disable auto-update?")
        expect(confirmTitle(true, true)).toBe("Change bump level?")
    })
})

describe("restartWarning", () => {
    it("says the service restarts once, in both directions", () => {
        expect(restartWarning("api", true)).toContain("api restarts once")
        expect(restartWarning("api", false)).toContain("api restarts once")
    })
})
