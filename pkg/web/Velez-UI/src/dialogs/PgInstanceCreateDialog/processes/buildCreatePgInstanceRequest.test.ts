import {describe, expect, it} from "vitest"

import {buildCreatePgInstanceRequest} from "@/dialogs/PgInstanceCreateDialog/processes/buildCreatePgInstanceRequest.ts"

describe("buildCreatePgInstanceRequest", () => {
    it("returns null when the name is blank", () => {
        const req = buildCreatePgInstanceRequest({
            name: "  ",
            environment: "",
            box: "small",
            exposePort: false,
            port: "",
            ownerService: "",
        })

        expect(req).toBeNull()
    })

    it("omits optional fields when they are not set", () => {
        const req = buildCreatePgInstanceRequest({
            name: "my-db",
            environment: "",
            box: "small",
            exposePort: false,
            port: "",
            ownerService: "",
        })

        expect(req).toEqual({
            name: "my-db",
            environment: undefined,
            box: "small",
            exposeToPort: undefined,
            ownerService: undefined,
        })
    })

    it("includes exposeToPort only when exposePort is checked and a port is set", () => {
        const req = buildCreatePgInstanceRequest({
            name: "my-db",
            environment: "prod",
            box: "large",
            exposePort: true,
            port: "5432",
            ownerService: "my-service",
        })

        expect(req).toEqual({
            name: "my-db",
            environment: "prod",
            box: "large",
            exposeToPort: 5432,
            ownerService: "my-service",
        })
    })

    it("omits exposeToPort when exposePort is checked but the port is blank", () => {
        const req = buildCreatePgInstanceRequest({
            name: "my-db",
            environment: "",
            box: "small",
            exposePort: true,
            port: "  ",
            ownerService: "",
        })

        expect(req?.exposeToPort).toBeUndefined()
    })
})
