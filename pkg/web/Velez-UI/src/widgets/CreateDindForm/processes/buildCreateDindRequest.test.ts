import {describe, expect, it} from "vitest"

import {buildCreateDindRequest} from "@/widgets/CreateDindForm/processes/buildCreateDindRequest.ts"

describe("buildCreateDindRequest", () => {
    it("returns null when the name is blank", () => {
        expect(buildCreateDindRequest({name: "  ", environment: "", isSysboxEnabled: true})).toBeNull()
    })

    it("returns null when the name is not a valid instance name", () => {
        expect(buildCreateDindRequest({name: "Bad Name", environment: "", isSysboxEnabled: true})).toBeNull()
    })

    it("trims the name and omits an empty environment", () => {
        expect(buildCreateDindRequest({name: " ci ", environment: " ", isSysboxEnabled: true})).toEqual({
            name: "ci",
            environment: undefined,
            isSysboxEnabled: true,
        })
    })

    it("passes the environment and the sysbox flag through", () => {
        expect(buildCreateDindRequest({name: "ci", environment: "PROD", isSysboxEnabled: false})).toEqual({
            name: "ci",
            environment: "PROD",
            isSysboxEnabled: false,
        })
    })
})
