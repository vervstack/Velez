import {describe, expect, it} from "vitest"

import {buildCreateS3BucketRequest} from "@/dialogs/CreateS3BucketDialog/processes/buildCreateS3BucketRequest.ts"

describe("buildCreateS3BucketRequest", () => {
    it("returns null when the bucket name is blank", () => {
        expect(buildCreateS3BucketRequest("s3", "  ", "")).toBeNull()
    })

    it("trims the name and omits an empty owner service", () => {
        expect(buildCreateS3BucketRequest("s3", " assets ", "")).toEqual({
            instanceName: "s3",
            bucketName: "assets",
            ownerService: undefined,
        })
    })

    it("forwards the owner service when one is picked", () => {
        expect(buildCreateS3BucketRequest("s3", "assets", "web")?.ownerService).toBe("web")
    })
})
