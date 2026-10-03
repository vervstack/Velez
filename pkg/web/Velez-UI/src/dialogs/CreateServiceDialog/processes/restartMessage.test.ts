import {describe, expect, it} from "vitest"

import {restartMessage} from "@/dialogs/CreateServiceDialog/processes/restartMessage.ts"

describe("restartMessage", () => {
    it("names the container and says the old one is kept until onboarding is finished", () => {
        expect(restartMessage("db-1")).toBe(
            "db-1 will be recreated under a new container id. The old one is kept until you finish onboarding."
        )
    })

    it("says no ports are mapped when the container publishes none", () => {
        expect(restartMessage("db-1", false)).toContain("publishes no ports")
        expect(restartMessage("db-1", true)).not.toContain("publishes no ports")
    })

    it("falls back to a generic subject without a name", () => {
        expect(restartMessage(undefined)).toMatch(/^The container will be recreated/)
    })
})
