import {describe, expect, it} from "vitest"

import {restartMessage} from "@/dialogs/CreateServiceDialog/processes/restartMessage.ts"

describe("restartMessage", () => {
    it("names the container and warns it is unavailable while recreated", () => {
        expect(restartMessage("db-1", false)).toBe(
            "db-1 will be recreated under a new container id. It is unavailable while the new one starts."
        )
    })

    it("mentions the stop-first downtime when keeping the exact ports", () => {
        expect(restartMessage("db-1", true)).toContain("stopped first")
    })

    it("falls back to a generic subject without a name", () => {
        expect(restartMessage(undefined, false)).toMatch(/^The container will be recreated/)
    })
})
