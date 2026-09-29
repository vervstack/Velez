import {describe, expect, it} from "vitest"

import {isPgLoginMissing} from "@/dialogs/CreateServiceDialog/processes/pgLogin.ts"

describe("isPgLoginMissing", () => {
    it("is false when both the user and the password are in the env", () => {
        expect(isPgLoginMissing({POSTGRES_USER: "admin", POSTGRES_PASSWORD: "secret"})).toBe(false)
    })

    it("is true when either is absent or empty", () => {
        expect(isPgLoginMissing({POSTGRES_USER: "admin"})).toBe(true)
        expect(isPgLoginMissing({POSTGRES_PASSWORD: "secret"})).toBe(true)
        expect(isPgLoginMissing({POSTGRES_USER: "", POSTGRES_PASSWORD: "secret"})).toBe(true)
    })

    it("is true when the container has no env at all", () => {
        expect(isPgLoginMissing(undefined)).toBe(true)
    })
})
