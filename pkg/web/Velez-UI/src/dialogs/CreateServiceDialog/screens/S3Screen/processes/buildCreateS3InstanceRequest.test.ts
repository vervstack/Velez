import {describe, expect, it} from "vitest"

import {
    buildCreateS3InstanceRequest,
    CreateS3InstanceFormState,
} from "@/dialogs/CreateServiceDialog/screens/S3Screen/processes/buildCreateS3InstanceRequest.ts"

const BASE: CreateS3InstanceFormState = {
    name: "  main  ",
    environment: "",
    box: "small",
    isPortExposed: false,
    port: "",
    region: "",
    isWebUiEnabled: false,
}

describe("buildCreateS3InstanceRequest", () => {
    it("returns null when the name is blank", () => {
        expect(buildCreateS3InstanceRequest({...BASE, name: "   "})).toBeNull()
    })

    it("returns null when the name violates the instance name pattern", () => {
        expect(buildCreateS3InstanceRequest({...BASE, name: "Main Bucket"})).toBeNull()
    })

    it("trims the name, locks replication to 1 and omits the optional fields", () => {
        expect(buildCreateS3InstanceRequest(BASE)).toEqual({
            name: "main",
            environment: undefined,
            box: "small",
            exposeToPort: undefined,
            replicationFactor: 1,
            region: undefined,
            enableWebUi: false,
        })
    })

    it("sends the port only when exposing is on and a port is typed", () => {
        expect(buildCreateS3InstanceRequest({...BASE, isPortExposed: true, port: "3900"})?.exposeToPort).toBe(3900)
        expect(buildCreateS3InstanceRequest({...BASE, isPortExposed: true, port: ""})?.exposeToPort).toBeUndefined()
        expect(buildCreateS3InstanceRequest({...BASE, isPortExposed: false, port: "3900"})?.exposeToPort)
            .toBeUndefined()
    })

    it("forwards the region and the web UI flag", () => {
        const req = buildCreateS3InstanceRequest({...BASE, region: " eu ", isWebUiEnabled: true})

        expect(req?.region).toBe("eu")
        expect(req?.enableWebUi).toBe(true)
    })
})
