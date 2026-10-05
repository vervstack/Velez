import {describe, expect, it} from "vitest"

import {
    buildCreateS3KeyRequest,
    newKeyPermissionRow,
} from "@/dialogs/CreateS3KeyDialog/processes/buildCreateS3KeyRequest.ts"

describe("buildCreateS3KeyRequest", () => {
    it("returns null when the key name is blank", () => {
        expect(buildCreateS3KeyRequest("s3", " ", [])).toBeNull()
    })

    it("drops rows without a bucket and maps the flags of the rest", () => {
        const picked = {...newKeyPermissionRow(1), bucketName: "assets", isWrite: true}
        const empty = newKeyPermissionRow(2)

        expect(buildCreateS3KeyRequest("s3", " ci ", [picked, empty])).toEqual({
            instanceName: "s3",
            keyName: "ci",
            access: [{bucketName: "assets", isRead: true, isWrite: true, isOwner: false}],
        })
    })

    it("allows a key without any bucket permission", () => {
        expect(buildCreateS3KeyRequest("s3", "ci", [])?.access).toEqual([])
    })
})
